package async

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"runtime/debug"
	"sync"
)

// streamResult holds the resolved value or error from a streamed task.
// It is internal; consumers should iterate directly over the returned seq.
type streamResult[T any] struct {
	res T
	err error
}

type streamConfig struct {
	buffer int
}

type StreamOption func(*streamConfig)

func WithBuffer(buffer int) StreamOption {
	return func(c *streamConfig) {
		c.buffer = buffer
	}
}

func applyConfig(opts ...StreamOption) *streamConfig {
	cfg := &streamConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Sender allows producers to lazily submit tasks concurrently up to a specified limit.
type Sender[R any] struct {
	ch chan streamResult[R]
	wg sync.WaitGroup
}

// Stream creates a new bounded multiple-producer single-consumer task channel.
// It returns a Sender (tx) and a pull-based iterator (rx).
// The producer MUST call tx.Close() when they are done.
func Stream[R any](
	opts ...StreamOption,
) (*Sender[R], iter.Seq2[R, error]) {
	cfg := applyConfig(opts...)

	bufSize := cfg.buffer
	if bufSize <= 0 {
		bufSize = 100
	}

	ch := make(chan streamResult[R], bufSize)

	tx := &Sender[R]{
		ch: ch,
		wg: sync.WaitGroup{},
	}

	rx := func(yield func(R, error) bool) {
		for r := range ch {
			if !yield(r.res, r.err) {
				return
			}
		}
	}

	return tx, rx
}

func (s *Sender[R]) Add(delta int) {
	s.wg.Add(delta)
}

func (s *Sender[R]) Done() {
	s.wg.Done()
}

// Send instantly queues a static result, bypassing scope scheduling.
func (s *Sender[R]) Send(ctx context.Context, res R, err error) {
	if IsCanceled(err) {
		return
	}

	select {
	case s.ch <- streamResult[R]{res: res, err: err}:
	case <-ctx.Done():
	}
}

func (s *Sender[R]) CloseWhenDone() {
	go func() {
		s.wg.Wait()
		close(s.ch)
	}()
}

// StreamBatch streams a static batch through a fixed set of workers owned by the group.
// It returns immediately so callers can consume results while the batch is still running.
func StreamBatch[I any, R any](
	g *Group,
	batch []I,
	eval func(context.Context, I) (R, error),
) iter.Seq2[R, error] {
	tx, rx := Stream[R](WithBuffer(len(batch)))

	workers := g.Limit()
	if workers <= 0 || workers > len(batch) {
		workers = len(batch)
	}
	if workers == 0 {
		tx.CloseWhenDone()
		return rx
	}

	jobs := make(chan I, workers)
	tx.Add(workers)

	for range workers {
		g.Go(func(ctx context.Context) error {
			defer tx.Done()

			for it := range jobs {
				if ctx.Err() != nil {
					return nil
				}

				res, err := evalRecover(ctx, it, eval)
				tx.Send(ctx, res, err)
			}

			return nil
		})
	}

	go func() {
		defer close(jobs)

		for _, it := range batch {
			select {
			case jobs <- it:
			case <-g.Context().Done():
				return
			}
		}
	}()

	tx.CloseWhenDone()
	return rx
}

func evalRecover[I any, R any](
	ctx context.Context,
	it I,
	eval func(context.Context, I) (R, error),
) (res R, err error) {
	defer func() {
		if r := recoverPanic(); r != nil {
			err = r
		}
	}()

	return eval(ctx, it)
}

func IsCanceled(err error) bool {
	return errors.Is(err, context.Canceled)
}

func recoverPanic() error {
	r := recover()
	if r == nil {
		return nil
	}

	switch x := r.(type) {
	case error:
		return fmt.Errorf("panic: %w\n%s", x, debug.Stack())
	default:
		return fmt.Errorf("panic: %v\n%s", x, debug.Stack())
	}
}
