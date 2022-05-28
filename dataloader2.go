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
	Size          = 16
	batchDuration = Size * time.Millisecond
)

type BatchFunc[K comparable, T any] func(ctx context.Context, keys []K) (map[K]T, error)

type Dataloader[K comparable, T any] struct {
	batchCap      int
	batchDuration time.Duration
	batchFunc     BatchFunc[K, T]
	cache         map[K]*Thunk[K, T]
	ch            chan *Thunk[K, T]
	ctx           context.Context
	done          chan struct{}
	init          sync.Once
	mu            sync.RWMutex
	wg            sync.WaitGroup
	worker        int
}

func New[K comparable, T any](ctx context.Context, batchFunc BatchFunc[K, T], options ...Option[K, T]) (*Dataloader[K, T], func()) {
	dl := &Dataloader[K, T]{
		batchCap:      0,
		batchDuration: batchDuration,
		batchFunc:     batchFunc,
		cache:         make(map[K]*Thunk[K, T], Size),
		ch:            make(chan *Thunk[K, T], Size),
		done:          make(chan struct{}),
		worker:        1,
		ctx:           ctx,
	}

	for _, opt := range options {
		opt(dl)
	}

	var once sync.Once
	return dl, func() {
		dl.init.Do(func() {
			// Waste the init so that it doesn't setup the worker.
			// Useful when calling flush before the `Load` method.
		})
		once.Do(func() {
			close(dl.done)
			dl.wg.Wait()
		})
	}
}

func (l *Dataloader[K, T]) isDone() bool {
	select {
	case <-l.done:
		return true
	default:
		return false
	}
}

// Prime sets the cache data if it does not exists, or overwrites the data if it already exists.
func (l *Dataloader[K, T]) Prime(key K, res T) bool {
	if l.isDone() {
		return false
	}

	l.mu.RLock()
	t, ok := l.cache[key]
	l.mu.RUnlock()

	if ok && t.ok() {
		return false
	}

	if ok && t.pending() {
		t.resolve(res)

		return true
	}

	l.mu.Lock()
	thunk := NewThunk[K, T](key)
	thunk.resolve(res)
	l.cache[key] = thunk
	l.mu.Unlock()

	return true
}

func (l *Dataloader[K, T]) LoadMany(keys []K) (map[K]*Result[T], error) {
	if l.isDone() {
		return nil, ErrAborted
	}

	if len(keys) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	wg.Add(len(keys))

	result := make([]*Result[T], len(keys))
	for i, key := range keys {
		go func(i int, key K) {
			defer wg.Done()

			result[i] = NewResult(l.Load(key))
		}(i, key)
	}

	wg.Wait()

	resultByKey := make(map[K]*Result[T])
	for i, res := range result {
		resultByKey[keys[i]] = res
	}

	return resultByKey, nil
}

func (l *Dataloader[K, T]) Load(key K) (T, error) {
	l.init.Do(func() {
		if l.isDone() {
			return
		}

		l.wg.Add(l.worker)
		for i := 0; i < l.worker; i++ {
			go l.loop()
		}
	})

	l.mu.RLock()
	t, ok := l.cache[key]
	l.mu.RUnlock()

	if ok {
		return t.Wait()
	}

	if l.isDone() {
		var t T
		return t, ErrAborted
	}

	l.mu.Lock()
	thunk := NewThunk[K, T](key)
	l.cache[key] = thunk
	l.mu.Unlock()

	l.ch <- thunk

	return thunk.Wait()
}

func (l *Dataloader[K, T]) loop() {
	defer l.wg.Done()

	keys := make([]K, 0, l.batchCap)

	ticker := time.NewTicker(l.batchDuration)
	defer ticker.Stop()

	for {
		select {
		case <-l.done:
			l.mu.Lock()
			defer l.mu.Unlock()

			for key := range l.cache {
				l.cache[key].reject(ErrAborted)
			}

			return
		case <-ticker.C:
			go l.flush(keys)
			keys = nil
		case thunk := <-l.ch:
			ticker.Reset(l.batchDuration)

			keys = append(keys, thunk.key)
			if l.batchCap == 0 || len(keys) < l.batchCap {
				continue
			}

			go l.flush(keys)
			keys = nil
		}
	}
}

func (l *Dataloader[K, T]) flush(keys []K) {
	if len(keys) == 0 {
		return
	}

	res, err := l.batchFunc(l.ctx, keys)
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		if err != nil {
			l.cache[key].reject(err)

			continue
		}

		val, ok := res[key]
		if !ok {
			l.cache[key].reject(fmt.Errorf("%w: %v", ErrKeyNotFound, key))
		} else {
			l.cache[key].resolve(val)
		}
	}
}
