package dataloader2

import (
	"errors"
	"sync"
)

var ErrResultNotSet = errors.New("result not set")

type Result[T any] struct {
	res   T
	err   error
	dirty bool
	once  sync.Once
	wg    sync.WaitGroup
}

func NewResult[T any]() *Result[T] {
	r := &Result[T]{}
	r.wg.Add(1)

	return r
}

func (r *Result[T]) IsZero() bool {
	return r == nil || !r.dirty
}

func (r *Result[T]) Result() (t T) {
	r.wg.Wait()

	if r.IsZero() {
		return
	}

	return r.res
}

func (r *Result[T]) Error() error {
	r.wg.Wait()

	if r.IsZero() {
		return ErrResultNotSet
	}

	return r.err
}

func (r *Result[T]) Ok() bool {
	return r.Error() == nil
}

func (r *Result[T]) Unwrap() (T, error) {
	return r.Result(), r.Error()
}

func (r *Result[T]) resolve(t T) *Result[T] {
	r.once.Do(func() {
		r.dirty = true
		r.res = t
		r.wg.Done()
	})

	return r
}

func (r *Result[T]) reject(err error) *Result[T] {
	r.once.Do(func() {
		r.dirty = true
		r.err = err
		r.wg.Done()
	})

	return r
}

// UnwrapAll attempts to unwrap all results, and returns the first error.
func UnwrapAll[T any](res []Result[T]) ([]T, error) {
	result := make([]T, len(res))
	for i, r := range res {
		t, err := r.Unwrap()
		if err != nil {
			return nil, err
		}
		result[i] = t
	}

	return result, nil
}
