package dataloader2

import "errors"

var ErrResultNotSet = errors.New("result not set")

type Result[T any] struct {
	res   T
	err   error
	dirty bool
}

func NewResult[T any](res T, err error) *Result[T] {
	return &Result[T]{
		res:   res,
		err:   err,
		dirty: true,
	}
}

func Resolve[T any](res T) *Result[T] {
	return &Result[T]{
		res:   res,
		dirty: true,
	}
}

func Reject[T any](err error) *Result[T] {
	return &Result[T]{
		err:   err,
		dirty: true,
	}
}

func (r *Result[T]) IsZero() bool {
	return r == nil || !r.dirty
}

func (r *Result[T]) Result() (t T) {
	if r.IsZero() {
		return
	}
	return r.res
}

func (r *Result[T]) Error() error {
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
