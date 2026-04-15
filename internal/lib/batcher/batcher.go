package batcher

import (
	"context"
	"fmt"
	"time"
)

type Batcher[T any] struct {
	input chan T
	done  chan struct{}
	opts  *Options[T]
}

func New[T any](ctx context.Context, opts *Options[T]) *Batcher[T] {
	opts.fill()
	b := &Batcher[T]{
		input: make(chan T, opts.MaxSize*2),
		done:  make(chan struct{}),
		opts:  opts,
	}

	go b.run(ctx)
	return b
}

func (b *Batcher[T]) Send(ctx context.Context, item T) error {
	const op = "lib.batcher.Send"

	select {
	case b.input <- item:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s: context done: %w", op, ctx.Err())
	}
}

func (b *Batcher[T]) Done() <-chan struct{} {
	return b.done
}

func (b *Batcher[T]) run(ctx context.Context) {
	defer close(b.done)

	buf := make([]T, 0, b.opts.MaxSize)
	timer := time.NewTimer(b.opts.Timeout)
	defer timer.Stop()

	for {
		select {
		case item := <-b.input:
			buf = append(buf, item)
			if len(buf) >= b.opts.MaxSize {
				buf = b.flush(buf)
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(b.opts.Timeout)
			}

		case <-timer.C:
			if len(buf) > 0 {
				buf = b.flush(buf)
			}
			timer.Reset(b.opts.Timeout)

		case <-ctx.Done():
			if len(buf) > 0 {
				b.flush(buf)
			}
			return
		}
	}
}

func (b *Batcher[T]) flush(buf []T) []T {
	snapshot := make([]T, len(buf))
	copy(snapshot, buf)
	if err := b.opts.Flush(snapshot); err != nil {
		b.opts.OnFlushError(err)
	}
	return buf[:0]
}
