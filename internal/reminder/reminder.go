package reminder

import (
	"time"

	"github.com/mjbmb/TaskAlarm/internal/storage"
)

// Notifier abstracts sound and desktop notifications — injectable for tests.
type Notifier interface {
	Remind(taskName string)
	Alarm(projectedEnd, deadline string)
}

type Checker struct {
	store        *storage.Storage
	notifier     Notifier
	lastReminder time.Time
	lastAlarm    time.Time
}

func NewChecker(store *storage.Storage, notifier Notifier) *Checker {
	return &Checker{store: store, notifier: notifier}
}

// SetLastReminder overrides the last reminder time — used in tests.
func (c *Checker) SetLastReminder(t time.Time) { c.lastReminder = t }

// SetLastAlarm overrides the last alarm time — used in tests.
func (c *Checker) SetLastAlarm(t time.Time) { c.lastAlarm = t }

// Check evaluates whether a reminder or alarm should fire right now.
// Returns "reminder", "alarm", or "" so callers (and tests) can assert.
func (c *Checker) Check(now time.Time) string {
	if now.Hour() < 7 || now.Hour() >= 21 {
		return ""
	}
	tasks := c.store.GetTasks()
	if len(tasks) == 0 {
		return ""
	}

	sinceAlarm := now.Sub(c.lastAlarm)
	sinceReminder := now.Sub(c.lastReminder)

	if !c.store.IsOnTime(now) && sinceAlarm >= 10*time.Minute {
		projected := c.store.ProjectedEndTime(now)
		deadline := c.store.Deadline(now)
		c.lastAlarm = now
		if c.notifier != nil {
			c.notifier.Alarm(projected.Format("15:04"), deadline.Format("15:04"))
		}
		return "alarm"
	}

	next := c.store.NextPendingTask()
	if next != nil && sinceReminder >= 10*time.Minute {
		c.lastReminder = now
		if c.notifier != nil {
			c.notifier.Remind(next.Name)
		}
		return "reminder"
	}

	return ""
}
