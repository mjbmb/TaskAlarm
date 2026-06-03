package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/mjbmb/TaskAlarm/internal/storage"
	"github.com/mjbmb/TaskAlarm/internal/ui"
)

func main() {
	a := app.NewWithID("info.boeck.taskalarm")

	store := storage.New()
	w := ui.NewMainWindow(a, store)
	w.ShowAndRun()
}
