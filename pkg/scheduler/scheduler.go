package scheduler

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Scheduler implements a cron-like scheduler for automated tasks.
// Jobs can be scheduled with:
//   - Duration: "30m", "2h", "1d"
//   - Interval phrase: "every 10m", "every 2h", "every monday 9am"
//   - ISO timestamp: "2026-06-15T09:00:00Z" (one-shot)
type Scheduler struct {
	mu       sync.Mutex
	jobs     map[string]*Job
	stopCh   chan struct{}
	tick     time.Duration
	running  bool
	history  []RunRecord
}

// Job represents a scheduled automation.
type Job struct {
	Name     string
	Schedule string       // user-provided schedule string
	Handler  func()
	LastRun  time.Time
	NextRun  time.Time
	Interval time.Duration // computed interval for duration/interval schedules
	OneShot  time.Time     // zero if not one-shot
	Paused   bool
}

// RunRecord tracks a single job execution.
type RunRecord struct {
	JobName   string
	StartedAt time.Time
	Duration  time.Duration
	Success   bool
	Error     string
}

// New creates a new scheduler with a default tick interval (5 seconds).
func New() *Scheduler {
	return &Scheduler{
		jobs:   make(map[string]*Job),
		stopCh: make(chan struct{}),
		tick:   5 * time.Second,
	}
}

// NewWithTick creates a new scheduler with a custom tick interval.
func NewWithTick(tick time.Duration) *Scheduler {
	s := New()
	s.tick = tick
	return s
}

// AddJob registers a new scheduled job. The Schedule string is parsed to
// determine the next run time.
func (s *Scheduler) AddJob(j *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[j.Name]; exists {
		return fmt.Errorf("job %q already exists", j.Name)
	}

	// Parse the schedule
	next, interval, oneShot, err := parseSchedule(j.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule %q: %w", j.Schedule, err)
	}
	j.NextRun = next
	j.Interval = interval
	j.OneShot = oneShot
	j.Paused = false

	s.jobs[j.Name] = j
	return nil
}

// RemoveJob removes a scheduled job by name.
func (s *Scheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.jobs[name]; !exists {
		return fmt.Errorf("job %q not found", name)
	}
	delete(s.jobs, name)
	return nil
}

// PauseJob pauses a job without removing it.
func (s *Scheduler) PauseJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %q not found", name)
	}
	j.Paused = true
	return nil
}

// ResumeJob resumes a paused job.
func (s *Scheduler) ResumeJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %q not found", name)
	}
	j.Paused = false
	// Recompute next run
	next, _, _, err := parseSchedule(j.Schedule)
	if err == nil {
		j.NextRun = next
	}
	return nil
}

// ListJobs returns all registered jobs.
func (s *Scheduler) ListJobs() []*Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	return jobs
}

// RunHistory returns recent run history.
func (s *Scheduler) RunHistory() []RunRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]RunRecord, len(s.history))
	copy(cp, s.history)
	return cp
}

// Start begins the scheduler tick loop in a background goroutine.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	go s.runLoop(ctx)
}

// Stop halts the scheduler tick loop.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
}

// runLoop is the main tick loop, running in a goroutine.
func (s *Scheduler) runLoop(ctx context.Context) {
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tickJobs()
		}
	}
}

// tickJobs checks all jobs and runs any that are due.
func (s *Scheduler) tickJobs() {
	s.mu.Lock()
	due := make([]*Job, 0)
	now := time.Now().UTC()

	for _, j := range s.jobs {
		if j.Paused {
			continue
		}
		if now.After(j.NextRun) || now.Equal(j.NextRun) {
			due = append(due, j)
		}
	}
	s.mu.Unlock()

	// Execute due jobs outside the lock
	for _, j := range due {
		s.executeJob(j)
	}
}

// executeJob runs a job's handler in a goroutine.
func (s *Scheduler) executeJob(j *Job) {
	start := time.Now().UTC()

	// Run handler in goroutine with recovery
	go func() {
		var success bool
		var errStr string

		func() {
			defer func() {
				if r := recover(); r != nil {
					errStr = fmt.Sprintf("panic: %v", r)
				}
			}()
			j.Handler()
			success = true
		}()

		// Record execution
		duration := time.Since(start)

		s.mu.Lock()
		j.LastRun = start
		s.history = append(s.history, RunRecord{
			JobName:   j.Name,
			StartedAt: start,
			Duration:  duration,
			Success:   success,
			Error:     errStr,
		})

		// Compute next run time
		if j.OneShot != (time.Time{}) {
			// One-shot job: remove after execution
			log.Printf("[hermes:scheduler] One-shot job %q completed, removing", j.Name)
			delete(s.jobs, j.Name)
		} else if j.Interval > 0 {
			j.NextRun = start.Add(j.Interval)
		}
		s.mu.Unlock()

		result := "OK"
		if !success {
			result = fmt.Sprintf("ERROR: %s", errStr)
		}
		log.Printf("[hermes:scheduler] Job %q completed (%s) in %v", j.Name, result, duration.Round(time.Millisecond))
	}()
}

// parseSchedule parses a schedule string and returns:
// - nextRun: the first scheduled time (in the future)
// - interval: repeating interval (0 for one-shot)
// - oneShot: the exact time for one-shot schedules (zero if repeating)
func parseSchedule(schedule string) (nextRun time.Time, interval time.Duration, oneShot time.Time, err error) {
	schedule = strings.TrimSpace(schedule)
	now := time.Now().UTC()

	// Try ISO timestamp (one-shot)
	if t, err := time.Parse(time.RFC3339, schedule); err == nil {
		if t.Before(now) {
			return now, 0, t, nil // Already passed, run immediately then done
		}
		return t, 0, t, nil
	}

	// Try "every N..." or "every <day> ..."
	if strings.HasPrefix(strings.ToLower(schedule), "every ") {
		return parseEverySchedule(strings.TrimPrefix(strings.ToLower(schedule), "every "), now)
	}

	// Try duration string: "30m", "2h", "1d", "7d"
	d, err := parseDuration(schedule)
	if err == nil && d > 0 {
		nextRun = now.Add(d)
		return nextRun, d, time.Time{}, nil
	}

	return time.Time{}, 0, time.Time{}, fmt.Errorf("unrecognized schedule format: %q", schedule)
}

// parseEverySchedule parses "every <interval>" or "every <day> <time>" formats.
func parseEverySchedule(every string, now time.Time) (nextRun time.Time, interval time.Duration, oneShot time.Time, err error) {
	every = strings.TrimSpace(every)

	// Try "every <duration>" like "every 10m", "every 2h"
	if d, err := parseDuration(every); err == nil && d > 0 {
		nextRun = now.Add(d)
		return nextRun, d, time.Time{}, nil
	}

	// Try "every <weekday> <time>" like "every monday 9am", "every sunday 12:00"
	parts := strings.Fields(every)
	if len(parts) >= 2 {
		dayName := parts[0]
		timeStr := parts[1]

		weekday, ok := parseWeekday(dayName)
		if ok {
			hour, min, err := parseTime(timeStr)
			if err == nil {
				// Compute next occurrence of this weekday at this time
				return computeNextWeekdayTime(now, weekday, hour, min), 7 * 24 * time.Hour, time.Time{}, nil
			}

			// Try hour-only like "9am", "2pm"
			if h, err := parseHourOnly(timeStr); err == nil {
				return computeNextWeekdayTime(now, weekday, h, 0), 7 * 24 * time.Hour, time.Time{}, nil
			}
		}
	}

	return time.Time{}, 0, time.Time{}, fmt.Errorf("could not parse 'every' schedule: %q", every)
}

// parseDuration parses duration strings like "30m", "2h", "1d", "7d".
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Try standard Go duration first
	if d, err := time.ParseDuration(s); err == nil {
		if d <= 0 {
			return 0, fmt.Errorf("non-positive duration: %q", s)
		}
		return d, nil
	}

	// Handle "Nd" format
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		n, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid day duration: %q", s)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}

	// Handle "Nw" format
	if strings.HasSuffix(s, "w") {
		numStr := strings.TrimSuffix(s, "w")
		n, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid week duration: %q", s)
		}
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	}

	return 0, fmt.Errorf("unrecognized duration: %q", s)
}

// parseWeekday converts a day name to time.Weekday.
func parseWeekday(name string) (time.Weekday, bool) {
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
		"sun":       time.Sunday,
		"mon":       time.Monday,
		"tue":       time.Tuesday,
		"wed":       time.Wednesday,
		"thu":       time.Thursday,
		"fri":       time.Friday,
		"sat":       time.Saturday,
	}
	d, ok := days[strings.ToLower(name)]
	return d, ok
}

// parseTime parses "HH:MM" format.
func parseTime(s string) (hour, min int, err error) {
	if !strings.Contains(s, ":") {
		return 0, 0, fmt.Errorf("no colon in time: %q", s)
	}
	parts := strings.SplitN(s, ":", 2)
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %q", parts[0])
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %q", parts[1])
	}
	return h, m, nil
}

// parseHourOnly parses "9am", "2pm", "12PM" etc.
func parseHourOnly(s string) (int, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	isPM := strings.HasSuffix(s, "pm")
	isAM := strings.HasSuffix(s, "am")

	if !isPM && !isAM {
		return 0, fmt.Errorf("not am/pm format: %q", s)
	}

	numStr := strings.TrimSuffix(strings.TrimSuffix(s, "pm"), "am")
	h, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid hour number: %q", numStr)
	}

	if h < 1 || h > 12 {
		return 0, fmt.Errorf("hour out of range (1-12): %d", h)
	}

	if isPM && h != 12 {
		h += 12
	}
	if isAM && h == 12 {
		h = 0
	}

	if h < 0 || h > 23 {
		return 0, fmt.Errorf("hour out of range: %d", h)
	}
	return h, nil
}

// computeNextWeekdayTime finds the next occurrence of a given weekday+time.
func computeNextWeekdayTime(now time.Time, weekday time.Weekday, hour, min int) time.Time {
	daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
	target := time.Date(now.Year(), now.Month(), now.Day()+daysUntil, hour, min, 0, 0, time.UTC)

	if target.Before(now) || target.Equal(now) {
		// This week's occurrence already passed, go to next week
		target = target.AddDate(0, 0, 7)
	}

	return target
}

// NextRunInterval helper: returns a human-readable description of the next run.
func NextRunInterval(j *Job) string {
	if j.Paused {
		return "paused"
	}
	if j.OneShot != (time.Time{}) {
		return fmt.Sprintf("one-shot at %s", j.OneShot.Format("2006-01-02 15:04 UTC"))
	}
	if j.Interval > 0 {
		return fmt.Sprintf("every %s, next at %s",
			humanDuration(j.Interval),
			j.NextRun.Format("2006-01-02 15:04 UTC"))
	}
	return "unknown"
}

// humanDuration converts a duration to a human-readable string.
func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(math.Mod(d.Seconds(), 60))
		if s > 0 {
			return fmt.Sprintf("%dm %ds", m, s)
		}
		return fmt.Sprintf("%dm", m)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(math.Mod(d.Minutes(), 60))
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	}
	days := int(d.Hours() / 24)
	hours := int(math.Mod(d.Hours(), 24))
	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}
