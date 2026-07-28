package scheduler

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"30m", 30 * time.Minute},
		{"2h", 2 * time.Hour},
		{"1d", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2w", 14 * 24 * time.Hour},
		{"5s", 5 * time.Second},
		{"90m", 90 * time.Minute},
	}
	for _, tt := range tests {
		got, err := parseDuration(tt.input)
		if err != nil {
			t.Errorf("parseDuration(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseDurationInvalid(t *testing.T) {
	invalid := []string{"", "abc", "xyz", "1.5d", "-1h"}
	for _, s := range invalid {
		_, err := parseDuration(s)
		if err == nil {
			t.Errorf("parseDuration(%q) expected error", s)
		}
	}
}

func TestParseHourOnly(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"9am", 9},
		{"12pm", 12},
		{"12am", 0},
		{"1pm", 13},
		{"11PM", 23},
		{"6AM", 6},
	}
	for _, tt := range tests {
		got, err := parseHourOnly(tt.input)
		if err != nil {
			t.Errorf("parseHourOnly(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseHourOnly(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseHourOnlyInvalid(t *testing.T) {
	invalid := []string{"", "abc", "25pm", "13am", "-1pm"}
	for _, s := range invalid {
		_, err := parseHourOnly(s)
		if err == nil {
			t.Errorf("parseHourOnly(%q) expected error", s)
		}
	}
}

func TestParseWeekday(t *testing.T) {
	d, ok := parseWeekday("monday")
	if !ok || d != time.Monday {
		t.Errorf("expected Monday, got %v", d)
	}
	d, ok = parseWeekday("sun")
	if !ok || d != time.Sunday {
		t.Errorf("expected Sunday, got %v", d)
	}
	d, ok = parseWeekday("INVALID")
	if ok {
		t.Errorf("expected false for invalid weekday")
	}
}

func TestParseScheduleDuration(t *testing.T) {
	now := time.Now().UTC()

	// Duration schedule "30m"
	next, interval, oneShot, err := parseSchedule("30m")
	if err != nil {
		t.Fatalf("parseSchedule('30m') error: %v", err)
	}
	if next.Before(now.Add(25*time.Minute)) || next.After(now.Add(35*time.Minute)) {
		t.Errorf("next run out of range: %v (now: %v)", next, now)
	}
	if interval != 30*time.Minute {
		t.Errorf("interval = %v, want 30m", interval)
	}
	if oneShot != (time.Time{}) {
		t.Error("expected oneShot zero for repeating schedule")
	}
}

func TestParseScheduleEveryDuration(t *testing.T) {
	now := time.Now().UTC()

	// "every 2h"
	next, interval, oneShot, err := parseSchedule("every 2h")
	if err != nil {
		t.Fatalf("parseSchedule('every 2h') error: %v", err)
	}
	if next.Before(now.Add(1*time.Hour + 50*time.Minute)) || next.After(now.Add(2*time.Hour+10*time.Minute)) {
		t.Errorf("next run out of range: %v (now: %v)", next, now)
	}
	if interval != 2*time.Hour {
		t.Errorf("interval = %v, want 2h", interval)
	}
	if oneShot != (time.Time{}) {
		t.Error("expected oneShot zero for repeating schedule")
	}
}

func TestParseScheduleOneShot(t *testing.T) {
	// ISO timestamp (past — should return immediately)
	next, interval, oneShot, err := parseSchedule("2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parseSchedule(ISO) error: %v", err)
	}
	// Past time — next should be near now
	if next.IsZero() {
		t.Error("next should not be zero")
	}
	if interval != 0 {
		t.Errorf("expected zero interval for one-shot, got %v", interval)
	}
	if oneShot.IsZero() {
		t.Error("oneShot should be set")
	}

	// ISO timestamp (future)
	future := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	next, interval, oneShot, err = parseSchedule(future)
	if err != nil {
		t.Fatalf("parseSchedule(future ISO) error: %v", err)
	}
	if next.Before(time.Now().UTC()) {
		t.Error("next should be in the future")
	}
	if interval != 0 {
		t.Errorf("expected zero interval for one-shot, got %v", interval)
	}
	if oneShot.IsZero() {
		t.Error("oneShot should be set")
	}
}

func TestParseScheduleInvalid(t *testing.T) {
	invalid := []string{"", "banana", "next tuesday", "every"}
	for _, s := range invalid {
		_, _, _, err := parseSchedule(s)
		if err == nil {
			t.Errorf("parseSchedule(%q) expected error", s)
		}
	}
}

func TestComputeNextWeekdayTime(t *testing.T) {
	// On Monday 00:00, compute next Monday 9am — same day, future time
	monday := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	next := computeNextWeekdayTime(monday, time.Monday, 9, 0)
	if next.Weekday() != time.Monday || next.Hour() != 9 {
		t.Errorf("expected Monday 9am, got %v", next)
	}
	// Should be same day (9am hasn't passed yet), not next week
	if next.Day() != monday.Day() {
		t.Errorf("expected same day (time hasn't passed), got day %d vs %d", next.Day(), monday.Day())
	}

	// On Monday 10am, compute next Monday 9am (time already passed — should be next week)
	monday10am := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	next = computeNextWeekdayTime(monday10am, time.Monday, 9, 0)
	if next.Weekday() != time.Monday || next.Hour() != 9 {
		t.Errorf("expected Monday 9am, got %v", next)
	}
	if next.Day()-monday10am.Day() != 7 {
		t.Errorf("expected 7 days later, got %d days", next.Day()-monday10am.Day())
	}
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{90 * time.Second, "1m 30s"},
		{2 * time.Hour, "2h"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{24 * time.Hour, "1d"},
		{3*24*time.Hour + 12*time.Hour, "3d 12h"},
		{7 * 24 * time.Hour, "7d"},
	}
	for _, tt := range tests {
		got := humanDuration(tt.d)
		if got != tt.want {
			t.Errorf("humanDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestAddJob(t *testing.T) {
	s := New()
	err := s.AddJob(&Job{
		Name:     "test-job",
		Schedule: "1h",
		Handler:  func() {},
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}
	jobs := s.ListJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Name != "test-job" {
		t.Errorf("expected job name 'test-job', got %q", jobs[0].Name)
	}
	if jobs[0].NextRun.IsZero() {
		t.Error("NextRun should be set")
	}
}

func TestAddJobDuplicate(t *testing.T) {
	s := New()
	s.AddJob(&Job{Name: "dup", Schedule: "1h", Handler: func() {}})
	err := s.AddJob(&Job{Name: "dup", Schedule: "1h", Handler: func() {}})
	if err == nil {
		t.Fatal("expected error for duplicate job")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestAddJobInvalidSchedule(t *testing.T) {
	s := New()
	err := s.AddJob(&Job{Name: "bad", Schedule: "xyz", Handler: func() {}})
	if err == nil {
		t.Fatal("expected error for invalid schedule")
	}
}

func TestRemoveJob(t *testing.T) {
	s := New()
	s.AddJob(&Job{Name: "rm-me", Schedule: "1h", Handler: func() {}})
	if len(s.ListJobs()) != 1 {
		t.Fatal("expected 1 job")
	}
	err := s.RemoveJob("rm-me")
	if err != nil {
		t.Fatalf("RemoveJob error: %v", err)
	}
	if len(s.ListJobs()) != 0 {
		t.Fatal("expected 0 jobs after removal")
	}
}

func TestRemoveJobNonExistent(t *testing.T) {
	s := New()
	err := s.RemoveJob("not-found")
	if err == nil {
		t.Fatal("expected error for non-existent job")
	}
}

func TestPauseResumeJob(t *testing.T) {
	s := New()
	s.AddJob(&Job{Name: "toggle", Schedule: "1h", Handler: func() {}})

	// Pause
	err := s.PauseJob("toggle")
	if err != nil {
		t.Fatalf("PauseJob error: %v", err)
	}
	jobs := s.ListJobs()
	if len(jobs) != 1 || !jobs[0].Paused {
		t.Error("expected job to be paused")
	}

	// Resume
	err = s.ResumeJob("toggle")
	if err != nil {
		t.Fatalf("ResumeJob error: %v", err)
	}
	jobs = s.ListJobs()
	if len(jobs) != 1 || jobs[0].Paused {
		t.Error("expected job to be resumed")
	}
}

func TestPauseNonExistent(t *testing.T) {
	s := New()
	err := s.PauseJob("no-such-job")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResumeNonExistent(t *testing.T) {
	s := New()
	err := s.ResumeJob("no-such-job")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJobExecution(t *testing.T) {
	s := NewWithTick(100 * time.Millisecond)
	var callCount atomic.Int32
	done := make(chan struct{})

	err := s.AddJob(&Job{
		Name:     "counter",
		Schedule: "500ms",
		Handler: func() {
			if callCount.Add(1) >= 2 {
				close(done)
			}
		},
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	select {
	case <-done:
		// Success — job ran at least twice
	case <-time.After(3 * time.Second):
		t.Fatalf("job did not run enough times (count=%d)", callCount.Load())
	}

	s.Stop()
}

func TestOneShotJobExecution(t *testing.T) {
	s := NewWithTick(50 * time.Millisecond)
	var called atomic.Bool

	err := s.AddJob(&Job{
		Name:     "one-shot",
		Schedule: time.Now().UTC().Add(100*time.Millisecond).Format(time.RFC3339),
		Handler: func() {
			called.Store(true)
		},
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	time.Sleep(1 * time.Second)

	if !called.Load() {
		t.Fatal("one-shot job was not called")
	}

	// Should be removed after execution
	jobs := s.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("one-shot job should be removed after execution, found %d jobs", len(jobs))
	}

	s.Stop()
}

func TestJobPanicRecovery(t *testing.T) {
	s := NewWithTick(100 * time.Millisecond)

	err := s.AddJob(&Job{
		Name:     "panic-job",
		Schedule: "200ms",
		Handler: func() {
			panic("intentional panic")
		},
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	// Wait for the job to trigger and "panic"
	time.Sleep(1 * time.Second)

	// Scheduler should still be running (panic recovered)
	// Check run history
	history := s.RunHistory()
	found := false
	for _, r := range history {
		if r.JobName == "panic-job" {
			found = true
			if r.Success {
				t.Error("expected panic job to be marked as failed")
			}
			if r.Error == "" {
				t.Error("expected error message for panic")
			}
			break
		}
	}
	if !found {
		t.Error("panic-job not found in run history")
	}

	s.Stop()
}

func TestStartStop(t *testing.T) {
	s := New()
	ctx, cancel := context.WithCancel(context.Background())

	s.Start(ctx)
	// Starting twice should be safe
	s.Start(ctx)

	s.Stop()
	// Stopping twice should be safe
	s.Stop()
	cancel()
}

func TestFormatJobInfo(t *testing.T) {
	j := &Job{
		Name:     "test",
		Schedule: "1h",
		NextRun:  time.Now().UTC().Add(time.Hour),
		Interval: time.Hour,
	}
	info := NextRunInterval(j)
	if !strings.Contains(info, "every") {
		t.Errorf("expected 'every' in info, got: %s", info)
	}
	if !strings.Contains(info, "1h") {
		t.Errorf("expected '1h' in info, got: %s", info)
	}

	// Paused
	j.Paused = true
	info = NextRunInterval(j)
	if !strings.Contains(info, "paused") {
		t.Errorf("expected 'paused' in info, got: %s", info)
	}

	// One-shot
	j.Paused = false
	j.OneShot = time.Now().UTC().Add(2 * time.Hour)
	j.Interval = 0
	info = NextRunInterval(j)
	if !strings.Contains(info, "one-shot") {
		t.Errorf("expected 'one-shot' in info, got: %s", info)
	}
}

func TestRunHistory(t *testing.T) {
	s := NewWithTick(100 * time.Millisecond)

	s.AddJob(&Job{
		Name:     "hist",
		Schedule: "200ms",
		Handler:  func() {},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	time.Sleep(1 * time.Second)
	s.Stop()

	history := s.RunHistory()
	if len(history) == 0 {
		t.Fatal("expected at least 1 run in history")
	}
	found := false
	for _, r := range history {
		if r.JobName == "hist" {
			found = true
			if !r.Success {
				t.Error("expected success")
			}
			if r.StartedAt.IsZero() {
				t.Error("expected StartedAt to be set")
			}
			break
		}
	}
	if !found {
		t.Error("job not found in history")
	}
}
