package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mboeck/obmanager/obsidian"
	"github.com/mjbmb/TaskAlarm/internal/model"
	"github.com/mjbmb/TaskAlarm/internal/storage"
)

func showAddTaskDialog(win fyne.Window, store *storage.Storage, onAdd func()) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("z.B. E-Mails beantworten")

	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder("z.B. 30  oder  1h30m  oder  45m")

	items := []*widget.FormItem{
		widget.NewFormItem("Aufgabe", nameEntry),
		widget.NewFormItem("Dauer", durationEntry),
	}

	dlg := dialog.NewForm("Neue Aufgabe", "Hinzufügen", "Abbrechen", items, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			dialog.ShowError(simpleErr("Bitte einen Namen eingeben."), win)
			return
		}
		dur, err := model.ParseDuration(durationEntry.Text)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		store.AddTask(model.NewTask(name, dur))
		onAdd()
	}, win)
	dlg.Resize(fyne.NewSize(440, 210))
	dlg.Show()
	win.Canvas().Focus(nameEntry)
}

func showSettingsDialog(win fyne.Window, store *storage.Storage, onSave func()) {
	bufEntry := widget.NewEntry()
	bufEntry.SetText(strconv.Itoa(store.Settings.BufferMinutes))

	endTimeEntry := widget.NewEntry()
	endTimeEntry.SetText(fmt.Sprintf("%02d:%02d", store.Settings.EndHour, store.Settings.EndMinute))
	endTimeEntry.SetPlaceHolder("HH:MM  z.B. 20:00")

	items := []*widget.FormItem{
		widget.NewFormItem("Pufferzeit (Minuten)", bufEntry),
		widget.NewFormItem("Tagesende (Uhrzeit)", endTimeEntry),
	}

	dialog.ShowForm("Einstellungen", "Speichern", "Abbrechen", items, func(ok bool) {
		if !ok {
			return
		}
		buf, err := strconv.Atoi(strings.TrimSpace(bufEntry.Text))
		if err != nil || buf < 0 {
			dialog.ShowError(simpleErr("Pufferzeit: bitte eine gültige Zahl eingeben."), win)
			return
		}
		h, m, err := parseTime(strings.TrimSpace(endTimeEntry.Text))
		if err != nil {
			dialog.ShowError(simpleErr("Tagesende: bitte im Format HH:MM eingeben (z.B. 20:00)."), win)
			return
		}
		store.Settings.BufferMinutes = buf
		store.Settings.EndHour = h
		store.Settings.EndMinute = m
		_ = store.Save()
		onSave()
	}, win)
}

func parseTime(s string) (int, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, simpleErr("ungültiges Format")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, 0, simpleErr("ungültige Stunde")
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, 0, simpleErr("ungültige Minute")
	}
	return h, m, nil
}

func showWelcomeDialog(win fyne.Window, onContinue func()) {
	content := container.NewVBox(
		widget.NewLabel("Was steht heute bei dir an?"),
		widget.NewLabel("Füge deine Aufgaben mit der \"+\" Schaltfläche hinzu.\nDie App berechnet ob du rechtzeitig fertig wirst."),
	)
	dlg := dialog.NewCustom("Guten Morgen! 👋", "Los geht's", content, win)
	dlg.SetOnClosed(onContinue)
	dlg.Show()
}

// showCarryOverDialog asks whether to carry unfinished tasks from yesterday.
func showCarryOverDialog(win fyne.Window, tasks []*model.Task, onCarryOver func([]*model.Task)) {
	checks := make([]*widget.Check, len(tasks))
	rows := make([]fyne.CanvasObject, len(tasks))
	for i, t := range tasks {
		check := widget.NewCheck(fmt.Sprintf("%s  (%s)", t.Name, t.FormatDuration()), nil)
		check.SetChecked(true)
		checks[i] = check
		rows[i] = check
	}

	content := container.NewVBox(append(
		[]fyne.CanvasObject{widget.NewLabel("Folgende Aufgaben von gestern sind noch offen:")},
		rows...,
	)...)

	dlg := dialog.NewCustomConfirm(
		"Aufgaben übernehmen?",
		"Übernehmen", "Nein danke",
		content,
		func(ok bool) {
			if !ok {
				return
			}
			var selected []*model.Task
			for i, ch := range checks {
				if ch.Checked {
					selected = append(selected, tasks[i])
				}
			}
			if len(selected) > 0 {
				onCarryOver(selected)
			}
		},
		win,
	)
	dlg.Resize(fyne.NewSize(460, 300))
	dlg.Show()
}

func showObsidianSetupDialog(win fyne.Window, store *storage.Storage, onDone func()) {
	vaults, _ := obsidian.ListVaults()
	sort.Slice(vaults, func(i, j int) bool {
		return vaults[i].Path < vaults[j].Path
	})

	const newVaultOption = "Neue Vault erstellen…"
	options := make([]string, 0, len(vaults)+1)
	pathByOption := make(map[string]string, len(vaults))
	for _, v := range vaults {
		label := filepath.Base(v.Path) + "  (" + v.Path + ")"
		options = append(options, label)
		pathByOption[label] = v.Path
	}
	options = append(options, newVaultOption)

	radio := widget.NewRadioGroup(options, nil)
	if len(vaults) > 0 {
		radio.SetSelected(options[0])
	} else {
		radio.SetSelected(newVaultOption)
	}

	intro := widget.NewLabel("Obsidian ist installiert! TaskAlarm kann deine Aufgaben\nals Markdown-Notizen in deiner Vault speichern.")

	content := container.NewVBox(intro, widget.NewSeparator(), radio)

	handled := false

	var dlg *dialog.CustomDialog

	skipBtn := widget.NewButton("Ohne Obsidian fortfahren", func() {
		handled = true
		store.Settings.UseObsidian = false
		store.Settings.ObsidianSetupDone = true
		_ = store.Save()
		dlg.Hide()
		onDone()
	})
	skipBtn.Importance = widget.LowImportance

	saveBtn := widget.NewButton("Speichern", func() {
		sel := radio.Selected
		if sel == "" {
			return
		}
		handled = true
		dlg.Hide()
		if sel == newVaultOption {
			showNewVaultDialog(win, store, onDone)
			return
		}
		store.Settings.UseObsidian = true
		store.Settings.ObsidianVaultPath = pathByOption[sel]
		store.Settings.ObsidianSetupDone = true
		_ = store.Save()
		onDone()
	})
	saveBtn.Importance = widget.HighImportance

	dlg = dialog.NewCustom(
		"Obsidian-Integration",
		"Schließen",
		container.NewVBox(content, widget.NewSeparator(), container.NewCenter(container.NewHBox(skipBtn, saveBtn))),
		win,
	)
	dlg.SetOnClosed(func() {
		if !handled {
			store.Settings.ObsidianSetupDone = true
			_ = store.Save()
			onDone()
		}
	})
	dlg.Resize(fyne.NewSize(520, 360))
	dlg.Show()
}

func showNewVaultDialog(win fyne.Window, store *storage.Storage, onDone func()) {
	home, _ := os.UserHomeDir()
	pathEntry := widget.NewEntry()
	pathEntry.SetText(filepath.Join(home, "Dokumente", "TaskAlarm"))

	items := []*widget.FormItem{
		widget.NewFormItem("Vault-Pfad", pathEntry),
	}
	dialog.ShowForm("Neue Vault erstellen", "Erstellen", "Abbrechen", items, func(ok bool) {
		if !ok {
			showObsidianSetupDialog(win, store, onDone)
			return
		}
		vaultPath := strings.TrimSpace(pathEntry.Text)
		if vaultPath == "" {
			dialog.ShowError(simpleErr("Bitte einen Pfad eingeben."), win)
			showObsidianSetupDialog(win, store, onDone)
			return
		}
		if err := obsidian.RegisterVault(vaultPath); err != nil {
			dialog.ShowError(fmt.Errorf("Vault konnte nicht erstellt werden: %w", err), win)
			showObsidianSetupDialog(win, store, onDone)
			return
		}
		store.Settings.UseObsidian = true
		store.Settings.ObsidianVaultPath = vaultPath
		store.Settings.ObsidianSetupDone = true
		_ = store.Save()
		onDone()
	}, win)
}

type simpleError struct{ msg string }

func (e *simpleError) Error() string { return e.msg }

func simpleErr(s string) error { return &simpleError{s} }
