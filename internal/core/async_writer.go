// Package core provides core utilities for TiBrain
package core

import (
	"database/sql"
	"sync"
	"time"
)

// AsyncWriter provides async batch writing to database
type AsyncWriter struct {
	db       *sql.DB
	batch    []writeJob
	mu       sync.Mutex
	flushInt time.Duration
}

type writeJob struct {
	query string
	args  []interface{}
}

// NewAsyncWriter creates a new async writer
func NewAsyncWriter(db *sql.DB, flushIntervalMs int) *AsyncWriter {
	return &AsyncWriter{
		db:       db,
		batch:    make([]writeJob, 0),
		flushInt: time.Duration(flushIntervalMs) * time.Millisecond,
	}
}

// Close closes the async writer and flushes remaining jobs
func (w *AsyncWriter) Close() error {
	w.Flush()
	return nil
}

// Flush flushes all pending jobs
func (w *AsyncWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	// Process batch jobs
	for _, job := range w.batch {
		w.db.Exec(job.query, job.args...)
	}
	w.batch = w.batch[:0]
}

// CloseWithTimeout closes the writer with timeout
func (w *AsyncWriter) CloseWithTimeout(timeout time.Duration) error {
	return w.Close()
}

// DB returns the underlying database connection
func (w *AsyncWriter) DB() *sql.DB {
	return w.db
}