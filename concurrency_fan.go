// Package common provides common concurrency utilities.
package common

import (
	"context"
	"sync"
	"time"
)

// FanOutResult is the result of a fan-out operation.
type FanOutResult[T any] struct {
	WorkerID int
	Input    any
	Value    T
	Error    error
}

// FanOutInWithTimeout runs fn over each item in inputs using up to 'workers' goroutines.
// Each invocation is canceled after timeoutPerTask. Returns all FanOutResult[T].
func FanOutInWithTimeout[In any, Out any](
	ctx context.Context,
	inputs []In,
	workers int,
	timeoutPerTask time.Duration,
	fn func(context.Context, In) (Out, error),
) []FanOutResult[Out] {
	ctx, cancelAll := context.WithCancel(ctx)
	defer cancelAll()

	inCh := make(chan In)
	outCh := make(chan FanOutResult[Out])

	var wg sync.WaitGroup

	// worker pool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for item := range inCh {
				taskCtx, cancel := context.WithTimeout(ctx, timeoutPerTask)
				val, err := fn(taskCtx, item)
				cancel()
				outCh <- FanOutResult[Out]{
					WorkerID: workerID,
					Input:    item,
					Value:    val,
					Error:    err,
				}
			}
		}(i)
	}

	// feed inputs
	go func() {
		defer close(inCh)
		for _, it := range inputs {
			select {
			case <-ctx.Done():
				return
			case inCh <- it:
			}
		}
	}()

	// close outCh once all workers done
	go func() {
		wg.Wait()
		close(outCh)
	}()

	var results []FanOutResult[Out]
	for r := range outCh {
		results = append(results, r)
	}
	return results
}
