// Package common provides shared utilities for the notion.codes organization.
//
// This package contains utilities for managing concurrent operations,
// including job queues and worker pools.
package common

import (
	"context"
	"sync"
	"time"
)

// Result holds the outcome of processing an item.
type Result[T any] struct {
	WorkerID int
	Value    T
	Error    error
}

// ConcurrentJob manages a pool of goroutines to process jobs of type T.
type ConcurrentJob[In any, Out any] struct {
	ctx      context.Context
	cancel   context.CancelFunc
	workers  int
	ch       chan In
	results  chan Result[Out]
	handler  func(context.Context, In) Out
	wg       sync.WaitGroup
	stopOnce sync.Once     // ensure Stop is idempotent
	stopping chan struct{} // signal to abort enqueueing
}

// NewConcurrentJob initializes a job processor with given worker count and handler.
func NewConcurrentJob[In any, Out any](workers int,
	handler func(context.Context, In) Out,
	timeout time.Duration,
) *ConcurrentJob[In, Out] {
	var ctx context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background()) // unlimited
	}

	return &ConcurrentJob[In, Out]{
		ctx:      ctx,
		cancel:   cancel,
		workers:  workers,
		ch:       make(chan In),
		results:  make(chan Result[Out]),
		handler:  handler,
		stopping: make(chan struct{}),
	}
}

// Start launches all workers to begin processing jobs.
func (j *ConcurrentJob[In, Out]) Start() {
	for i := 0; i < j.workers; i++ {
		j.wg.Add(1)
		go func() {
			defer j.wg.Done()
			for {
				select {
				case <-j.ctx.Done():
					return
				case item, ok := <-j.ch:
					if !ok {
						return
					}
					j.results <- j.wrapHandler(j.ctx, item, i)
				}
			}
		}()
	}
}

func (j *ConcurrentJob[In, Out]) Stop() {
	j.stopOnce.Do(func() {
		close(j.stopping) // signal enqueueers to halt
		j.cancel()        // cancel context to break worker loops
		close(j.ch)       // close input channel to terminate workers
		j.wg.Wait()       // wait for all workers to finish
		close(j.results)  // close results channel
	})
}

func (j *ConcurrentJob[In, Out]) wrapHandler(ctx context.Context, item In, workerID int) Result[Out] {
	r := j.handler(ctx, item)

	result := &Result[Out]{
		WorkerID: workerID,
		Value:    r,
	}

	return *result
}

// Enqueue adds an item to the processing queue.
func (j *ConcurrentJob[In, Out]) Enqueue(item In) {
	select {
	case <-j.stopping:
		// early stop signal received, don't enqueue
	default:
		j.ch <- item
	}
}

// Results returns a channel to consume job outcomes.
func (j *ConcurrentJob[In, Out]) Results() <-chan Result[Out] {
	return j.results
}
