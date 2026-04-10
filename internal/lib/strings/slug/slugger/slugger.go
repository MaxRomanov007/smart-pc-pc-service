package slugger

import (
	"context"
	"fmt"

	"github.com/gosimple/slug"
)

type Slugger[T any] struct {
	retries uint
	fn      func(string) (T, error, bool)
}

func New[T any](retries uint, slugFunc func(string) (T, error, bool)) *Slugger[T] {
	return &Slugger[T]{retries: retries, fn: slugFunc}
}

func (sl *Slugger[T]) DoCtx(ctx context.Context, s string) (T, error) {
	var result T
	var err error
	for i := uint(0); i < sl.retries+1; i++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		candidate := s
		if i != 0 {
			candidate = fmt.Sprintf("%s-%d", s, i)
		}

		var continueNeeded bool
		result, err, continueNeeded = sl.fn(slug.Make(candidate))
		if continueNeeded {
			continue
		}
		break
	}
	return result, err
}
