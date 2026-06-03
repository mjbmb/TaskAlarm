package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mjbmb/TaskAlarm/internal/model"
)

const rowHeightEstimate = 90.0 // pixels, used to calculate drag distance → slots

// dragHandle is a grip icon that implements fyne.Draggable for list reordering.
type dragHandle struct {
	widget.BaseWidget
	accumulated float32
	onDrop      func(deltaSlots int)
}

func newDragHandle(onDrop func(deltaSlots int)) *dragHandle {
	h := &dragHandle{onDrop: onDrop}
	h.ExtendBaseWidget(h)
	return h
}

func (h *dragHandle) CreateRenderer() fyne.WidgetRenderer {
	grip := canvas.NewText("⠿", colorGray)
	grip.TextSize = 22
	return widget.NewSimpleRenderer(container.NewCenter(grip))
}

func (h *dragHandle) MinSize() fyne.Size { return fyne.NewSize(28, 44) }

func (h *dragHandle) Dragged(e *fyne.DragEvent) {
	h.accumulated += e.Dragged.DY
}

func (h *dragHandle) DragEnd() {
	slots := int(h.accumulated / rowHeightEstimate)
	h.accumulated = 0
	if slots != 0 {
		h.onDrop(slots)
	}
}

// buildTaskRow builds a single task card with all controls.
func buildTaskRow(
	task *model.Task,
	isFirst, isLast bool,
	onMoveUp, onMoveDown func(),
	onMoveBySlots func(int),
	onStart, onComplete, onRemove func(),
) fyne.CanvasObject {

	// ── Status colour ────────────────────────────────────────────────────────
	var stripColor color.Color
	var statusText string
	var statusColor color.Color
	switch task.Status {
	case model.StatusPending:
		stripColor = colorGray
		statusText = "ausstehend"
		statusColor = colorGray
	case model.StatusActive:
		stripColor = colorBlue
		statusText = "● aktiv"
		statusColor = colorBlue
	case model.StatusDone:
		stripColor = colorGreen
		statusText = "✓ erledigt"
		statusColor = colorGreen
	}

	strip := canvas.NewRectangle(stripColor)
	strip.SetMinSize(fyne.NewSize(5, 0))

	// ── Text ─────────────────────────────────────────────────────────────────
	nameLabel := widget.NewLabelWithStyle(task.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	durText := canvas.NewText(task.FormatDuration(), theme.Color(theme.ColorNameForeground))
	durText.TextSize = 13

	statusBadge := canvas.NewText(statusText, statusColor)
	statusBadge.TextSize = 12
	statusBadge.TextStyle = fyne.TextStyle{Italic: true}

	info := container.NewVBox(
		nameLabel,
		container.NewHBox(durText, widget.NewSeparator(), statusBadge),
	)

	// ── Drag handle ──────────────────────────────────────────────────────────
	handle := newDragHandle(onMoveBySlots)

	// ── Arrow buttons ────────────────────────────────────────────────────────
	upBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), onMoveUp)
	downBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), onMoveDown)
	upBtn.Importance = widget.LowImportance
	downBtn.Importance = widget.LowImportance
	if isFirst {
		upBtn.Disable()
	}
	if isLast {
		downBtn.Disable()
	}
	orderCol := container.NewVBox(handle, upBtn, downBtn)

	// ── Action button ─────────────────────────────────────────────────────────
	var actionBtn *widget.Button
	switch task.Status {
	case model.StatusPending:
		actionBtn = widget.NewButton("▶  Starten", onStart)
		actionBtn.Importance = widget.HighImportance
	case model.StatusActive:
		actionBtn = widget.NewButton("✓  Erledigt", onComplete)
		actionBtn.Importance = widget.SuccessImportance
	case model.StatusDone:
		actionBtn = widget.NewButton("✓  Fertig", func() {})
		actionBtn.Disable()
	}

	removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), onRemove)
	removeBtn.Importance = widget.DangerImportance

	actions := container.NewHBox(actionBtn, removeBtn)

	// ── Background ────────────────────────────────────────────────────────────
	var bgColor color.NRGBA
	switch task.Status {
	case model.StatusDone:
		bgColor = color.NRGBA{R: 0x1E, G: 0x22, B: 0x1E, A: 0xFF}
	case model.StatusActive:
		bgColor = color.NRGBA{R: 0x1E, G: 0x24, B: 0x30, A: 0xFF}
	default:
		bgColor = color.NRGBA{R: 0x2A, G: 0x2A, B: 0x2E, A: 0xFF}
	}
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 8

	row := container.NewHBox(
		strip,
		container.NewPadded(info),
		widget.NewSeparator(),
		orderCol,
		widget.NewSeparator(),
		actions,
	)

	return container.NewStack(bg, container.NewPadded(row))
}
