package dataloader2

import "time"

type Option[K comparable, T any] func(*Dataloader[K, T])

func WithBatchCap[K comparable, T any](cap int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchCap = cap
	}
}

func WithBatchDuration[K comparable, T any](duration time.Duration) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchDuration = duration
	}
}
