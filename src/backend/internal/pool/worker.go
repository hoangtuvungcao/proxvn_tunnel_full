package pool

import (
	"sync"
)

// WorkerPool manages a pool of goroutines for concurrent task execution
type WorkerPool struct {
	maxWorkers int
	tasks      chan func()
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	pool := &WorkerPool{
		maxWorkers: maxWorkers,
		tasks:      make(chan func(), maxWorkers*2), // Buffer = 2x workers
		stopChan:   make(chan struct{}),
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// Submit adds a task to the pool
func (p *WorkerPool) Submit(task func()) {
	select {
	case p.tasks <- task:
		// Task submitted
	case <-p.stopChan:
		// Pool stopped, don't submit
	}
}

// worker processes tasks from the queue
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			task()
		case <-p.stopChan:
			return
		}
	}
}

// Stop gracefully stops the worker pool
func (p *WorkerPool) Stop() {
	close(p.stopChan)
	close(p.tasks)
	p.wg.Wait()
}

// Stats returns pool statistics
func (p *WorkerPool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"max_workers":  p.maxWorkers,
		"queued_tasks": len(p.tasks),
		"buffer_size":  cap(p.tasks),
	}
}
