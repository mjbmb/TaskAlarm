package sound

import (
	"os"
	"os/exec"
	"sync"
)

var mu sync.Mutex

// alarmStop cancels a running alarm loop when closed.
var alarmStop chan struct{}
var alarmCmd *exec.Cmd

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

func playOgaOnce(files []string) {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		cmd := exec.Command("paplay", f)
		mu.Lock()
		alarmCmd = cmd
		mu.Unlock()
		if cmd.Run() == nil {
			return
		}
		// If paplay was killed by StopAlarm, don't fall through to aplay.
		// aplay can't decode .oga and would output the raw bytes as noise.
		mu.Lock()
		stopped := alarmStop == nil
		mu.Unlock()
		if stopped {
			return
		}
		cmd2 := exec.Command("aplay", f)
		mu.Lock()
		alarmCmd = cmd2
		mu.Unlock()
		if cmd2.Run() == nil {
			return
		}
	}
	print("\007")
}

// Reminder plays a gentle notification sound once.
func Reminder() {
	go func() {
		for _, f := range reminderSounds {
			if _, err := os.Stat(f); err != nil {
				continue
			}
			if exec.Command("paplay", f).Run() == nil {
				return
			}
		}
		print("\007")
	}()
}

// StartAlarm begins a continuous alarm loop until StopAlarm is called.
func StartAlarm() {
	mu.Lock()
	if alarmStop != nil {
		// already running
		mu.Unlock()
		return
	}
	stop := make(chan struct{})
	alarmStop = stop
	mu.Unlock()

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				playOgaOnce(alarmSounds)
				// brief pause between repetitions so select can fire
				select {
				case <-stop:
					return
				default:
				}
			}
		}
	}()
}

// StopAlarm stops the current alarm loop immediately.
func StopAlarm() {
	mu.Lock()
	defer mu.Unlock()
	if alarmStop != nil {
		close(alarmStop)
		alarmStop = nil
	}
	if alarmCmd != nil {
		_ = alarmCmd.Process.Kill()
		alarmCmd = nil
	}
}

// IsAlarmRunning reports whether the alarm is currently active.
func IsAlarmRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return alarmStop != nil
}

// Notify sends a desktop notification.
func Notify(title, body string) {
	exec.Command("notify-send", "--urgency=normal", "--expire-time=8000", title, body).Run()
}

// NotifyUrgent sends an urgent desktop notification.
func NotifyUrgent(title, body string) {
	exec.Command("notify-send", "--urgency=critical", "--expire-time=0", title, body).Run()
}
