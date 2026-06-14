package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// Scheduler implements a cron-like scheduler for automated tasks.
type Scheduler struct {
	mu     sync.Mutex
	jobs   map[string]*Job
	stopCh chan struct{}
}

// Job represents a scheduled automation.
type Job struct {
	Name     string
	CronExpr string
	Handler  func()
	LastRun  time.Time
}

// New creates a new scheduler.
func New() *Scheduler {
	return &Scheduler{
		jobs:   make(map[string]*Job),
		stopCh: make(chan struct{}),
	}
}

// AddJob registers a new scheduled job.
func (s *Scheduler) AddJob(j *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.Name] = j
	fmt.Printf("[hermes:scheduler] Job added: %s (%s)\n", j.Name, j.CronExpr)
}

// Start begins executing scheduled jobs.
func (s *Scheduler) Start() {
	fmt.Println("[hermes:scheduler] Started (stub mode)")
}

// Stop halts the scheduler.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	fmt.Println("[hermes:scheduler] Stopped")
}
