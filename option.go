package dataloader2

import (
	"time"
)

type Option[K comparable, T any] func(*Dataloader[K, T])

func WithBatchWorker[K comparable, T any](n int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchWorker = make(chan struct{}, n)
	}
}

func WithBatchCap[K comparable, T any](cap int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchCap = cap
	}
}

func WithChannelCap[K comparable, T any](cap int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.ch = make(chan K, cap)
		d.cache = make(map[K]*Result[T], cap)
	}
}

func WithCacheSize[K comparable, T any](size int) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.cache = make(map[K]*Result[T], size)
	}
}

func WithBatchDuration[K comparable, T any](duration time.Duration) Option[K, T] {
	return func(d *Dataloader[K, T]) {
		d.batchDuration = duration
	}
}
