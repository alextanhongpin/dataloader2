package dataloader2

import (
	"sync"
	"sync/atomic"
)

type Thunk[K comparable, T any] struct {
	_    struct{}
	wg   sync.WaitGroup
	once sync.Once
	key  K
	n    int32

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

func (t *Thunk[K, T]) done() bool {
	if t == nil {
		return false
	}

	return atomic.LoadInt32(&t.n) == 1
}

func (t *Thunk[K, T]) resolve(res T) {
	t.once.Do(func() {
		t.res = res
		atomic.StoreInt32(&t.n, 1)
		t.wg.Done()
	})
}

func (t *Thunk[K, T]) reject(err error) {
	t.once.Do(func() {
		t.err = err
		atomic.StoreInt32(&t.n, 1)
		t.wg.Done()
	})
}

func (t *Thunk[K, T]) Wait() (T, error) {
	t.wg.Wait()
	return t.res, t.err
}
