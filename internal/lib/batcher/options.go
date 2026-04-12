package batcher

import "time"

type FlushFunc[T any] func([]T) error
type OnFlushErrorFunc func(error)

type Options[T any] struct {
	Flush        FlushFunc[T]
	MaxSize      int
	Timeout      time.Duration
	OnFlushError OnFlushErrorFunc
}

func (o *Options[T]) fill() {
	if o.Flush == nil {
		o.Flush = func([]T) error { return nil }
	}
	if o.MaxSize <= 0 {
		o.MaxSize = 10
	}
	if o.Timeout == 0 {
		o.Timeout = 5 * time.Second
	}
	if o.OnFlushError == nil {
		o.OnFlushError = func(error) {}
	}
}
