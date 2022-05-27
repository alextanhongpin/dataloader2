package dataloader2

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

const (
	batchDuration = 16 * time.Millisecond
)

type BatchFunc[K comparable, T any] func(keys []K) (map[K]T, error)

type Dataloader[K comparable, T any] struct {
	batchCap      int
	batchDuration time.Duration
	batchFunc     BatchFunc[K, T]
	cache         map[K]*Thunk[K, T]
	ch            chan *Thunk[K, T]
	done          chan struct{}
	init          sync.Once
	mu            sync.RWMutex
	wg            sync.WaitGroup
}

func New[K comparable, T any](batchFunc BatchFunc[K, T], options ...Option[K, T]) (*Dataloader[K, T], func()) {
	dl := &Dataloader[K, T]{
		batchCap:      0,
		batchDuration: batchDuration,
		batchFunc:     batchFunc,
		cache:         make(map[K]*Thunk[K, T], 16),
		ch:            make(chan *Thunk[K, T]),
		done:          make(chan struct{}),
	}

	for _, opt := range options {
		opt(dl)
	}

	var once sync.Once
	return dl, func() {
		once.Do(func() {
			close(dl.done)
			dl.wg.Wait()
		})
	}
}

// Prime sets the cache data if it does not exists. Does not overwrite existing
// cache data.
func (l *Dataloader[K, T]) Prime(key K, res T) bool {
	l.mu.RLock()
	t, ok := l.cache[key]
	l.mu.RUnlock()

	if ok && t.done() {
		return false
	}

	if ok {
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

func (l *Dataloader[K, T]) LoadMany(keys []K) map[K]*Result[T] {
	if len(keys) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(keys))

	result := make([]*Result[T], len(keys))
	for i, key := range keys {
		go func(i int, key K) {
			defer wg.Done()

			result[i] = NewResult[T](l.Load(key))
		}(i, key)
	}

	wg.Wait()

	resultByKey := make(map[K]*Result[T])
	for i, res := range result {
		resultByKey[keys[i]] = res
	}

	return resultByKey
}

func (l *Dataloader[K, T]) Load(key K) (T, error) {
	l.init.Do(func() {
		l.wg.Add(1)
		go l.loop()
	})

	l.mu.RLock()
	t, ok := l.cache[key]
	l.mu.RUnlock()

	if ok {
		return t.Wait()
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

	keys := make([]K, 0, 16)

	ticker := time.NewTicker(l.batchDuration)
	defer ticker.Stop()

	for {
		select {
		case <-l.done:
			return
		case <-ticker.C:
			l.flush(keys)
			keys = nil
		case thunk := <-l.ch:
			ticker.Reset(l.batchDuration)

			keys = append(keys, thunk.key)
			if l.batchCap == 0 || len(keys) < l.batchCap {
				continue
			}

			l.flush(keys)
			keys = nil

		}
	}
}

func (l *Dataloader[K, T]) flush(keys []K) {
	if len(keys) == 0 {
		return
	}

	res, err := l.batchFunc(keys)
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
