package ui

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mjbmb/TaskAlarm/internal/model"
	"github.com/mjbmb/TaskAlarm/internal/reminder"
	"github.com/mjbmb/TaskAlarm/internal/sound"
	"github.com/mjbmb/TaskAlarm/internal/storage"
)

// soundNotifier routes reminder/alarm events to actual sound + desktop notifications.
// win must be set before use.
type soundNotifier struct {
	win fyne.Window
}

func (s *soundNotifier) Remind(taskName string) {
	sound.Reminder()
	sound.Notify("📋 TaskAlarm – Nächste Aufgabe",
		fmt.Sprintf(`Vergiss nicht: "%s" steht noch aus.`, taskName))
}

func (s *soundNotifier) Alarm(projected, deadline string) {
	if sound.IsAlarmRunning() {
		return // dialog already open
	}
	sound.StartAlarm()
	sound.NotifyUrgent("⚠️ TaskAlarm – Zeitplan überschritten!",
		fmt.Sprintf("Ende: ca. %s Uhr  (Ziel: %s Uhr)", projected, deadline))
	s.showAlarmDialog(projected, deadline)
}

func (s *soundNotifier) showAlarmDialog(projected, deadline string) {
	icon := canvas.NewText("⚠️", colorRed)
	icon.TextSize = 48
	icon.Alignment = fyne.TextAlignCenter

	title := canvas.NewText("Zeitplan überschritten!", colorRed)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	info := widget.NewLabelWithStyle(
		fmt.Sprintf("Voraussichtliches Ende: %s Uhr\nZiel: %s Uhr\n\nBitte Task beginnen oder abschließen!", projected, deadline),
		fyne.TextAlignCenter, fyne.TextStyle{},
	)

	var dlg *dialog.CustomDialog

	stopBtn := widget.NewButton("🔕  Alarm stoppen", func() {
		sound.StopAlarm()
		if dlg != nil {
			dlg.Hide()
		}
	})
	stopBtn.Importance = widget.DangerImportance

	content := container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(title),
		info,
		container.NewCenter(stopBtn),
	)

	dlg = dialog.NewCustom("TaskAlarm – Achtung!", "Schließen", content, s.win)
	dlg.SetOnClosed(func() {
		sound.StopAlarm()
	})
	dlg.Show()
}

type mainWindow struct {
	store   *storage.Storage
	window  fyne.Window
	checker *reminder.Checker

	statusBg    *canvas.Rectangle
	statusIcon  *canvas.Text
	statusLine1 *canvas.Text
	statusLine2 *canvas.Text
	endTimeText *widget.Label
	bufferLabel *widget.Label

	taskList    *fyne.Container
	scroll      *container.Scroll
	sayingIndex int
}

func NewMainWindow(a fyne.App, store *storage.Storage) fyne.Window {
	mw := &mainWindow{
		store:       store,
		sayingIndex: rand.Intn(len(positiveSayings)),
	}

	mw.window = a.NewWindow("TaskAlarm")
	mw.window.Resize(fyne.NewSize(700, 780))
	mw.window.SetMaster()
	mw.window.SetOnClosed(func() {
		sound.StopAlarm()
		_ = store.Save()
	})
	notifier := &soundNotifier{win: mw.window}
	mw.checker = reminder.NewChecker(store, notifier)

	mw.buildUI()
	mw.handleStartup()
	mw.startTickers()

	return mw.window
}

func (mw *mainWindow) buildUI() {
	// ── Status header ─────────────────────────────────────────────────────────
	mw.statusBg = canvas.NewRectangle(colorGray)
	mw.statusBg.SetMinSize(fyne.NewSize(0, 115))
	mw.statusBg.CornerRadius = 12

	mw.statusIcon = canvas.NewText("", colorWhite)
	mw.statusIcon.TextSize = 40
	mw.statusIcon.TextStyle = fyne.TextStyle{Bold: true}
	mw.statusIcon.Alignment = fyne.TextAlignCenter

	mw.statusLine1 = canvas.NewText("Wird berechnet…", colorWhite)
	mw.statusLine1.TextSize = 17
	mw.statusLine1.TextStyle = fyne.TextStyle{Bold: true}
	mw.statusLine1.Alignment = fyne.TextAlignCenter

	mw.statusLine2 = canvas.NewText("", colorWhite)
	mw.statusLine2.TextSize = 13
	mw.statusLine2.Alignment = fyne.TextAlignCenter

	statusContent := container.NewVBox(mw.statusIcon, mw.statusLine1, mw.statusLine2)
	header := container.NewStack(
		mw.statusBg,
		container.NewCenter(container.NewPadded(statusContent)),
	)

	// ── Info bar ──────────────────────────────────────────────────────────────
	mw.endTimeText = widget.NewLabel("")
	mw.bufferLabel = widget.NewLabel(fmt.Sprintf("Puffer: %d min", mw.store.Settings.BufferMinutes))

	today := time.Now().Format("Monday, 02. January 2006")
	dateLabel := widget.NewLabelWithStyle(today, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	settingsBtn := widget.NewButtonWithIcon("Einstellungen", theme.SettingsIcon(), func() {
		showSettingsDialog(mw.window, mw.store, func() {
			mw.bufferLabel.SetText(fmt.Sprintf("Puffer: %d min", mw.store.Settings.BufferMinutes))
			mw.refreshStatus()
		})
	})
	settingsBtn.Importance = widget.LowImportance

	infoBar := container.NewBorder(nil, nil,
		dateLabel,
		container.NewHBox(mw.endTimeText, mw.bufferLabel, settingsBtn),
	)

	// ── Task list ─────────────────────────────────────────────────────────────
	mw.taskList = container.NewVBox()
	mw.scroll = container.NewScroll(mw.taskList)

	// ── Toolbar ───────────────────────────────────────────────────────────────
	addBtn := widget.NewButtonWithIcon("  Aufgabe hinzufügen", theme.ContentAddIcon(), func() {
		showAddTaskDialog(mw.window, mw.store, mw.rebuildTaskList)
	})
	addBtn.Importance = widget.HighImportance

	// ── Root ─────────────────────────────────────────────────────────────────
	root := container.NewBorder(
		container.NewVBox(container.NewPadded(header), infoBar, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewPadded(container.NewCenter(addBtn))),
		nil, nil,
		mw.scroll,
	)

	mw.window.SetContent(root)
	mw.rebuildTaskList()
}

func (mw *mainWindow) handleStartup() {
	tasks := mw.store.GetTasks()

	// Check for unfinished tasks from yesterday
	yesterday := mw.store.YesterdayUnfinished()
	if len(yesterday) > 0 {
		showCarryOverDialog(mw.window, yesterday, func(selected []*model.Task) {
			mw.store.CarryOverTasks(selected)
			mw.rebuildTaskList()
		})
		return
	}

	// First time today: show welcome
	if len(tasks) == 0 {
		showWelcomeDialog(mw.window, func() {
			showAddTaskDialog(mw.window, mw.store, mw.rebuildTaskList)
		})
	}
}

func (mw *mainWindow) rebuildTaskList() {
	tasks := mw.store.GetTasks()

	if len(tasks) == 0 {
		empty := container.NewCenter(container.NewVBox(
			widget.NewLabelWithStyle("Noch keine Aufgaben für heute.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabel(`Klicke auf "Aufgabe hinzufügen" um zu starten.`),
		))
		mw.taskList.Objects = []fyne.CanvasObject{empty}
		mw.taskList.Refresh()
		mw.refreshStatus()
		return
	}

	objects := make([]fyne.CanvasObject, 0, len(tasks)*2)
	for i, task := range tasks {
		t := task
		idx := i
		n := len(tasks)

		row := buildTaskRow(
			t,
			idx == 0, idx == n-1,
			func() { mw.store.MoveUp(t.ID); mw.rebuildTaskList() },
			func() { mw.store.MoveDown(t.ID); mw.rebuildTaskList() },
			func(slots int) { mw.moveBySlots(t.ID, slots) },
			func() { mw.store.StartTask(t.ID); mw.rebuildTaskList() },
			func() { mw.store.CompleteTask(t.ID); mw.rebuildTaskList() },
			func() { mw.confirmRemove(t) },
		)

		objects = append(objects, container.NewPadded(row))
		if idx < n-1 {
			objects = append(objects, widget.NewSeparator())
		}
	}

	mw.taskList.Objects = objects
	mw.taskList.Refresh()
	mw.refreshStatus()
}

func (mw *mainWindow) moveBySlots(id string, slots int) {
	if slots > 0 {
		for i := 0; i < slots; i++ {
			mw.store.MoveDown(id)
		}
	} else {
		for i := 0; i > slots; i-- {
			mw.store.MoveUp(id)
		}
	}
	mw.rebuildTaskList()
}

func (mw *mainWindow) confirmRemove(t *model.Task) {
	dialog.ShowConfirm(
		"Aufgabe entfernen",
		fmt.Sprintf("Möchtest du \"%s\" wirklich entfernen?", t.Name),
		func(ok bool) {
			if ok {
				mw.store.RemoveTask(t.ID)
				mw.rebuildTaskList()
			}
		},
		mw.window,
	)
}

func (mw *mainWindow) refreshStatus() {
	now := time.Now()
	tasks := mw.store.GetTasks()

	deadline := mw.store.Deadline(now)
	deadlineStr := fmt.Sprintf("%02d:%02d", deadline.Hour(), deadline.Minute())

	allDone := len(tasks) > 0
	for _, t := range tasks {
		if t.Status != model.StatusDone {
			allDone = false
			break
		}
	}

	switch {
	case len(tasks) == 0:
		mw.statusBg.FillColor = colorGray
		mw.statusIcon.Text = "📋"
		mw.statusLine1.Text = "Noch keine Aufgaben für heute."
		mw.statusLine2.Text = `Füge Aufgaben hinzu um zu beginnen.`
		mw.endTimeText.SetText("")

	case allDone:
		mw.statusBg.FillColor = colorGreen
		mw.statusIcon.Text = "🎉"
		mw.statusLine1.Text = "Alle Aufgaben erledigt!"
		mw.statusLine2.Text = "Großartig gemacht. Genieße den Rest des Tages!"
		mw.endTimeText.SetText("Fertig!")

	case mw.store.IsOnTime(now):
		projected := mw.store.ProjectedEndTime(now)
		mw.statusBg.FillColor = colorGreen
		mw.statusIcon.Text = "✓"
		mw.statusLine1.Text = fmt.Sprintf("Du wirst bis %s Uhr fertig!  (ca. %s Uhr)", deadlineStr, projected.Format("15:04"))
		mw.statusLine2.Text = positiveSayings[mw.sayingIndex%len(positiveSayings)]
		mw.endTimeText.SetText(fmt.Sprintf("Ende ca. %s Uhr", projected.Format("15:04")))

	default:
		projected := mw.store.ProjectedEndTime(now)
		mw.statusBg.FillColor = colorRed
		mw.statusIcon.Text = "✗"
		mw.statusLine1.Text = fmt.Sprintf("Das geht sich nicht mehr aus!  (ca. %s Uhr)", projected.Format("15:04"))
		mw.statusLine2.Text = "Bitte Aufgaben kürzen, entfernen oder Tagesende anpassen."
		mw.endTimeText.SetText(fmt.Sprintf("Ende ca. %s Uhr", projected.Format("15:04")))
	}

	canvas.Refresh(mw.statusBg)
	canvas.Refresh(mw.statusIcon)
	canvas.Refresh(mw.statusLine1)
	canvas.Refresh(mw.statusLine2)
}

func (mw *mainWindow) startTickers() {
	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for range t.C {
			mw.refreshStatus()
		}
	}()

	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for range t.C {
			mw.checkReminders()
		}
	}()
}

func (mw *mainWindow) checkReminders() {
	mw.checker.Check(time.Now())
}
