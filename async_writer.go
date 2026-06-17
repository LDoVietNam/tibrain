package main

import (
	"database/sql"
	"log"
	"sync"
)

// AsyncWriteJob represents a single database write task.
type AsyncWriteJob struct {
	Query string
	Args  []interface{}
}

// AsyncWriter is the "Mailbox/Pending Message" system for database writes.
// It uses a buffered channel to accept write jobs without blocking the caller,
// and a background worker goroutine to execute them sequentially.
type AsyncWriter struct {
	db       *sql.DB
	jobQueue chan AsyncWriteJob
	wg       sync.WaitGroup
	jobsWg   sync.WaitGroup
	quit     chan struct{}
}

// NewAsyncWriter initializes the asynchronous writer queue and starts the background worker.
// bufferSize determines how many queries can be pending before Enqueue blocks.
func NewAsyncWriter(db *sql.DB, bufferSize int) *AsyncWriter {
	aw := &AsyncWriter{
		db:       db,
		jobQueue: make(chan AsyncWriteJob, bufferSize),
		quit:     make(chan struct{}),
	}
	aw.startWorker()
	return aw
}

// startWorker launches the background goroutine to process the write queue.
func (aw *AsyncWriter) startWorker() {
	aw.wg.Add(1)
	go func() {
		defer aw.wg.Done()
		for {
			select {
			case job := <-aw.jobQueue:
				_, err := aw.db.Exec(job.Query, job.Args...)
				if err != nil {
					// We log the error but continue processing the next jobs
					log.Printf("[AsyncWriter] Failed to execute background write: %v | Query: %s", err, job.Query)
				}
				aw.jobsWg.Done()
			case <-aw.quit:
				return
			}
		}
	}()
}

// flush waits until all queued and in-flight jobs finish.
func (aw *AsyncWriter) flush() {
	aw.jobsWg.Wait()
}

// Enqueue adds a query and its arguments to the write mailbox.
// It returns immediately as long as the buffer is not full.
func (aw *AsyncWriter) Enqueue(query string, args ...interface{}) {
	// If the buffer is full, this will temporarily block, acting as backpressure.
	aw.jobsWg.Add(1)
	aw.jobQueue <- AsyncWriteJob{
		Query: query,
		Args:  args,
	}
}

// Close gracefully shuts down the worker, ensuring all pending jobs are flushed to disk.
func (aw *AsyncWriter) Close() {
	aw.flush()
	close(aw.quit)
	aw.wg.Wait()
}
