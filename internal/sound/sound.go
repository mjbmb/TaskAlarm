package sound

import (
	"os"
	"os/exec"
	"sync"
)

var mu sync.Mutex

var reminderSounds = []string{
	"/usr/share/sounds/freedesktop/stereo/message-new-instant.oga",
	"/usr/share/sounds/freedesktop/stereo/message.oga",
	"/usr/share/sounds/freedesktop/stereo/bell.oga",
	"/usr/share/sounds/freedesktop/stereo/complete.oga",
}

var alarmSounds = []string{
	"/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga",
	"/usr/share/sounds/freedesktop/stereo/dialog-warning.oga",
	"/usr/share/sounds/freedesktop/stereo/dialog-error.oga",
}

func playOga(files []string) {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		if exec.Command("paplay", f).Run() == nil {
			return
		}
		if exec.Command("aplay", f).Run() == nil {
			return
		}
	}
	// terminal bell fallback
	print("\007")
}

// Reminder plays a gentle notification sound (next task waiting).
func Reminder() {
	mu.Lock()
	defer mu.Unlock()
	go playOga(reminderSounds)
}

// Alarm plays an urgent alarm (tasks running late, 20:00 missed).
func Alarm() {
	mu.Lock()
	defer mu.Unlock()
	go func() {
		for i := 0; i < 3; i++ {
			playOga(alarmSounds)
		}
	}()
}

// Notify sends a desktop notification with optional sound.
func Notify(title, body string) {
	exec.Command("notify-send", "--urgency=normal", "--expire-time=8000", title, body).Run()
}

// NotifyUrgent sends an urgent desktop notification.
func NotifyUrgent(title, body string) {
	exec.Command("notify-send", "--urgency=critical", "--expire-time=0", title, body).Run()
}
