package jobpool

import (
	"sync"

	"github.com/pkg/errors"
)

var ErrJobPool = errors.New("job pool error")

// Consumer transforms IN to OUT.
type Consumer[IN any, OUT any] func(IN) OUT

// JobPool runs a sequence of tasks concurrently.
type JobPool[IN any, OUT any] struct {
	consumer      Consumer[IN, OUT]
	consumerCount int
	jobCount      int
	wg            sync.WaitGroup
	inbox         chan job[IN]
	outbox        chan result[OUT]
	producedCount int
}

type job[IN any] struct {
	index int
	task  IN
}

type result[OUT any] struct {
	index  int
	result OUT
}

// NewJobPool creates a new JobPool.
//
// If jobCount is zero, the number of jobs is unbounded.
// The number of consumers is the minimum of maxWorkers and jobCount.
func NewJobPool[IN any, OUT any](jobCount, maxConsumers int, consumer Consumer[IN, OUT]) *JobPool[IN, OUT] {
	return &JobPool[IN, OUT]{
		consumer:      consumer,
		consumerCount: min(maxConsumers, jobCount),
		jobCount:      jobCount,
		inbox:         make(chan job[IN], jobCount),
		outbox:        make(chan result[OUT], jobCount),
	}
}

// Produce adds a job the to pool.
//
// Returns ErrJobPool if the pool was created with a non-zero jobCount
// and all jobs have already been produced.
//
// Note: Produce is not thread-safe.
func (jp *JobPool[IN, OUT]) Produce(in IN) error {
	if jp.jobCount > 0 && jp.producedCount >= jp.jobCount {
		return errors.Wrap(ErrJobPool, "job count exceeded")
	}

	jp.inbox <- job[IN]{jp.producedCount, in}

	jp.producedCount++

	return nil
}

// Start consuming jobs.
func (jp *JobPool[IN, OUT]) Start() {
	for range jp.consumerCount {
		jp.wg.Go(func() {
			for job := range jp.inbox {
				out := jp.consumer(job.task)
				jp.outbox <- result[OUT]{job.index, out}
			}
		})
	}
}

// Wait for all jobs to complete and return their results.
//
// Results are returned in the order that jobs were produced.
func (jp *JobPool[IN, OUT]) Wait() []OUT {
	close(jp.inbox)
	jp.wg.Wait()
	close(jp.outbox)

	results := make([]OUT, len(jp.outbox))
	for result := range jp.outbox {
		results[result.index] = result.result
	}

	return results
}
