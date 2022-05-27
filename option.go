package dataloader2

import "time"

type Option[K comparable, T any] func(*Dataloader[K, T])

func WithWorker[K comparable, T any](n int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.worker = n
	}
}

func WithBatchCap[K comparable, T any](cap int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchCap = cap
	}
}

func WithChannelCap[K comparable, T any](cap int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.ch = make(chan *Thunk[K, T], cap)
		d.cache = make(map[K]*Thunk[K, T], cap)
	}
}

func WithCacheSize[K comparable, T any](size int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.cache = make(map[K]*Thunk[K, T], size)
	}
}

func WithBatchDuration[K comparable, T any](duration time.Duration) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchDuration = duration
	}
}
