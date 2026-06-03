package model

import (
	"fmt"
	"math/rand"
	"time"
)

type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusActive
	StatusDone
)

type Task struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Duration  time.Duration `json:"duration"`
	Status    TaskStatus    `json:"status"`
	StartedAt *time.Time    `json:"started_at,omitempty"`
	DoneAt    *time.Time    `json:"done_at,omitempty"`
	Order     int           `json:"order"`
}

func NewTask(name string, duration time.Duration) *Task {
	return &Task{
		ID:       fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(9999)),
		Name:     name,
		Duration: duration,
		Status:   StatusPending,
	}
}

func (t *Task) RemainingDuration(now time.Time) time.Duration {
	if t.Status == StatusDone {
		return 0
	}
	if t.Status == StatusActive && t.StartedAt != nil {
		remaining := t.Duration - now.Sub(*t.StartedAt)
		if remaining < 0 {
			return 0
		}
		return remaining
	}
	return t.Duration
}

func (t *Task) FormatDuration() string {
	d := t.Duration
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 && m > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", m)
}

func ParseDuration(s string) (time.Duration, error) {
	// Try standard Go format first (e.g. "1h30m", "90m")
	if d, err := time.ParseDuration(s); err == nil {
		if d > 0 {
			return d, nil
		}
	}
	// Try plain number as minutes
	var minutes float64
	if _, err := fmt.Sscanf(s, "%f", &minutes); err == nil && minutes > 0 {
		return time.Duration(minutes*60) * time.Second, nil
	}
	return 0, fmt.Errorf("ungültiges Format: %q (Beispiele: 90, 1h30m, 45m)", s)
}
