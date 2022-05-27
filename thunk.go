package dataloader2

import (
	"sync"
	"sync/atomic"
)

type status int32

const (
	pending status = iota
	fulfilled
	rejected
)

func (s status) Int32() int32 {
	return int32(s)
}

type Thunk[K comparable, T any] struct {
	_      struct{}
	wg     sync.WaitGroup
	once   sync.Once
	key    K
	status int32

	res T
	err error
}

func NewThunk[K comparable, T any](key K) *Thunk[K, T] {
	t := &Thunk[K, T]{
		key: key,
	}
	t.wg.Add(1)
	return t
}

func (t *Thunk[K, T]) pending() bool {
	if t == nil {
		return false
	}

	return status(atomic.LoadInt32(&t.status)) == pending
}

func (t *Thunk[K, T]) ok() bool {
	if t == nil {
		return false
	}

	return status(atomic.LoadInt32(&t.status)) == fulfilled
}

func (t *Thunk[K, T]) fail() bool {
	if t == nil {
		return false
	}

	return status(atomic.LoadInt32(&t.status)) == rejected
}

func (t *Thunk[K, T]) resolve(res T) {
	t.once.Do(func() {
		t.res = res
		atomic.StoreInt32(&t.status, 1)
		t.wg.Done()
	})
}

func (t *Thunk[K, T]) reject(err error) {
	t.once.Do(func() {
		t.err = err
		atomic.StoreInt32(&t.status, 2)
		t.wg.Done()
	})
}

func (t *Thunk[K, T]) Wait() (T, error) {
	t.wg.Wait()
	return t.res, t.err
}
