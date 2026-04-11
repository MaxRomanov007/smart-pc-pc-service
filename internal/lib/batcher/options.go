package batcher

import "time"

type Options[T any] struct {
	Flush        func([]T) error
	MaxSize      int
	Timeout      time.Duration
	OnFlushError func(error)
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
