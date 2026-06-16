package async

import "context"

// Task represents an asynchronous operation that will eventually produce a result.
type Task[T any] struct {
	wait func() (T, error)
}

// Wait blocks until the asynchronous operation completes or the scope's context is canceled.
func (f *Task[T]) Wait() (T, error) {
	return f.wait()
}

// Go safely spawns a background task within the provided Scope.
// It returns a Task that can be used to await the strongly-typed result.
func Go[T any](g *Group, f func(context.Context) (T, error)) *Task[T] {
	var res T
	var err error
	done := make(chan struct{})

	g.Go(func(ctx context.Context) error {
		defer func() {
			if r := recoverPanic(); r != nil {
				err = r
			}

			close(done)
		}()

		res, err = f(ctx)
		return nil
	})

	return &Task[T]{
		wait: func() (T, error) {
			select {
			case <-done:
				return res, err
			case <-g.Context().Done():
				var zero T
				return zero, g.Context().Err()
			}
		},
	}
}
