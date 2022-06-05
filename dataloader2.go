package dataloader2

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("dataloader2: key not found")
	ErrAborted     = errors.New("dataloader2: aborted")
)

const (
	defaultBatchDuration = 16 * time.Millisecond
)

type BatchFunc[K comparable, T any] func(ctx context.Context, keys []K) (map[K]T, error)

type Dataloader[K comparable, T any] struct {
	batchCap      int
	batchDuration time.Duration
	batchWorker   chan struct{}
	batchFunc     BatchFunc[K, T]
	cache         map[K]*Result[T]
	ch            chan K
	ctx           context.Context
	done          chan struct{}
	init          sync.Once
	mu            sync.RWMutex
	wg            sync.WaitGroup
}

func New[K comparable, T any](ctx context.Context, batchFunc BatchFunc[K, T], options ...Option[K, T]) (*Dataloader[K, T], func()) {
	dl := &Dataloader[K, T]{
		batchCap:      0,
		batchDuration: defaultBatchDuration,
		batchWorker:   make(chan struct{}, 1),
		batchFunc:     batchFunc,
		cache:         make(map[K]*Result[T]),
		ch:            make(chan K),
		ctx:           ctx,
		done:          make(chan struct{}),
	}

	for _, opt := range options {
		opt(dl)
	}

	var once sync.Once
	return dl, func() {
		dl.init.Do(func() {
			// Waste the init so that it doesn't setup the batchWorker.
			// Useful when calling flush before the `Load` method.
		})
		once.Do(func() {
			close(dl.done)
			dl.wg.Wait()
		})
	}
}

func (l *Dataloader[K, T]) Load(key K) (T, error) {
	return l.load(key).Unwrap()
}

func (l *Dataloader[K, T]) LoadMany(keys []K) (map[K]T, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	resolve := make(map[K]*Result[T])
	for _, key := range keys {
		l.mu.RLock()
		t, found := l.cache[key]
		l.mu.RUnlock()

		if found {
			resolve[key] = t
		} else {
			resolve[key] = l.load(key)
		}
	}

	result := make(map[K]T, len(resolve))
	for key, t := range resolve {
		res, err := t.Unwrap()
		if err != nil {
			return nil, err
		}

		result[key] = res
	}

	return result, nil
}

// Prime sets the cache data if it does not exists, or overwrites the data if it already exists.
func (l *Dataloader[K, T]) Prime(key K, res T) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	t, found := l.cache[key]
	if found && t.IsZero() {
		t.resolve(res)

		return false
	}

	l.cache[key] = NewResult[T]().resolve(res)

	return true
}

func (l *Dataloader[K, T]) load(key K) *Result[T] {
	l.init.Do(func() {
		select {
		case <-l.done:
			return
		default:
			l.loopAsync()
		}
	})

	l.mu.Lock()
	t, found := l.cache[key]
	if found {
		l.mu.Unlock()

		return t
	}

	t = NewResult[T]()
	l.cache[key] = t
	l.mu.Unlock()

	select {
	case <-l.done:
		l.mu.Lock()
		t = l.cache[key].reject(ErrAborted)
		l.mu.Unlock()

		return t
	case l.ch <- key:
		return t
	}
}

func (l *Dataloader[K, T]) loopAsync() {
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()
		l.loop()
	}()
}

func (l *Dataloader[K, T]) loop() {
	keys := make([]K, 0, l.batchCap)

	ticker := time.NewTicker(l.batchDuration)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(l.ctx)
	defer cancel()

	for {
		select {
		case <-l.done:
			l.mu.Lock()

			for _, key := range keys {
				l.cache[key].reject(ErrAborted)
			}

			for key := range l.cache {
				l.cache[key].reject(ErrAborted)
			}

			l.mu.Unlock()

			return
		case <-ticker.C:
			l.flushAsync(ctx, keys)
			keys = nil
		case key := <-l.ch:
			ticker.Reset(l.batchDuration)

			keys = append(keys, key)
			if l.batchCap == 0 || len(keys) < l.batchCap {
				continue
			}

			l.flushAsync(ctx, keys)
			keys = nil
		}
	}
}

func (l *Dataloader[K, T]) flushAsync(ctx context.Context, keys []K) {
	if len(keys) == 0 {
		return
	}

	l.batchWorker <- struct{}{}
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()
		defer func() {
			<-l.batchWorker
		}()

		l.flush(ctx, keys)
	}()
}

func (l *Dataloader[K, T]) flush(ctx context.Context, keys []K) {
	if len(keys) == 0 {
		return
	}

	res, err := l.batchFunc(ctx, keys)

	l.mu.Lock()
	for _, key := range keys {
		if err != nil {
			l.cache[key].reject(err)

			continue
		}

		r, ok := res[key]
		if ok {
			l.cache[key].resolve(r)
		} else {
			l.cache[key].reject(fmt.Errorf("%w: %v", ErrKeyNotFound, key))
		}
	}
	l.mu.Unlock()
}
