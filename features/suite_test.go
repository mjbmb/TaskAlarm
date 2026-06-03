package features_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/mjbmb/TaskAlarm/internal/model"
	"github.com/mjbmb/TaskAlarm/internal/reminder"
	"github.com/mjbmb/TaskAlarm/internal/storage"
)

// ── Test context ─────────────────────────────────────────────────────────────

type testContext struct {
	store       *storage.Storage
	configDir   string
	now         time.Time
	checker     *reminder.Checker
	mockNotify  *mockNotifier
}

type mockNotifier struct {
	reminderFired bool
	alarmFired    bool
	reminderTask  string
}

func (m *mockNotifier) Remind(taskName string) {
	m.reminderFired = true
	m.reminderTask = taskName
}

func (m *mockNotifier) Alarm(projected, deadline string) {
	m.alarmFired = true
}

func newTestContext() *testContext {
	dir, _ := os.MkdirTemp("", "taskalarm-test-*")
	tc := &testContext{configDir: dir}
	tc.resetAll()
	return tc
}

func (tc *testContext) resetAll() {
	// Wipe all daily files so scenarios don't bleed into each other.
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_ = os.Remove(filepath.Join(tc.configDir, today+".json"))
	_ = os.Remove(filepath.Join(tc.configDir, yesterday+".json"))
	_ = os.Remove(filepath.Join(tc.configDir, "settings.json"))

	tc.now = time.Now()
	tc.store = storage.NewWithDir(tc.configDir)
	tc.mockNotify = &mockNotifier{}
	tc.checker = reminder.NewChecker(tc.store, tc.mockNotify)
}

func (tc *testContext) cleanup() {
	os.RemoveAll(tc.configDir)
}

func (tc *testContext) taskByName(name string) *model.Task {
	for _, t := range tc.store.GetTasks() {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// ── Step definitions ──────────────────────────────────────────────────────────

func (tc *testContext) ichHabeKeineAufgaben() error {
	tc.resetAll()
	return nil
}

func (tc *testContext) esGibtKeineGestrigen() error {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_ = os.Remove(filepath.Join(tc.configDir, yesterday+".json"))
	return nil
}

func (tc *testContext) dasTagesende(uhrzeit string) error {
	h, m, err := parseHHMM(uhrzeit)
	if err != nil {
		return err
	}
	tc.store.Settings.EndHour = h
	tc.store.Settings.EndMinute = m
	return nil
}

func (tc *testContext) diePufferzeit(minuten int) error {
	tc.store.Settings.BufferMinutes = minuten
	return nil
}

func (tc *testContext) dieAktuelleZeit(uhrzeit string) error {
	h, m, err := parseHHMM(uhrzeit)
	if err != nil {
		return err
	}
	today := time.Now()
	tc.now = time.Date(today.Year(), today.Month(), today.Day(), h, m, 0, 0, today.Location())
	return nil
}

func (tc *testContext) ichFuegeAufgabeHinzu(name, dauer string) error {
	dur, err := model.ParseDuration(dauer)
	if err != nil {
		return err
	}
	tc.store.AddTask(model.NewTask(name, dur))
	return nil
}

func (tc *testContext) ichEntferneAufgabe(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	tc.store.RemoveTask(t.ID)
	return nil
}

func (tc *testContext) ichVerschiebeNachOben(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	tc.store.MoveUp(t.ID)
	return nil
}

func (tc *testContext) ichVerschiebeNachUnten(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	tc.store.MoveDown(t.ID)
	return nil
}

func (tc *testContext) ichStarteAufgabe(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	tc.store.StartTask(t.ID)
	return nil
}

func (tc *testContext) ichStarteAufgabeUm(name, uhrzeit string) error {
	h, m, err := parseHHMM(uhrzeit)
	if err != nil {
		return err
	}
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	today := time.Now()
	startTime := time.Date(today.Year(), today.Month(), today.Day(), h, m, 0, 0, today.Location())
	tc.store.StartTaskAt(t.ID, startTime)
	return nil
}

func (tc *testContext) ichSchliesseAufgabeAb(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	tc.store.CompleteTask(t.ID)
	return nil
}

func (tc *testContext) minutenVergehen(min int) error {
	tc.now = tc.now.Add(time.Duration(min) * time.Minute)
	return nil
}

// ── Assertions: tasks ─────────────────────────────────────────────────────────

func (tc *testContext) enthaeltDieListe(n int) error {
	got := len(tc.store.GetTasks())
	if got != n {
		return fmt.Errorf("erwartet %d Aufgaben, got %d", n, got)
	}
	return nil
}

func (tc *testContext) dieErsteAufgabeHeisst(name string) error {
	tasks := tc.store.GetTasks()
	if len(tasks) == 0 {
		return fmt.Errorf("keine Aufgaben vorhanden")
	}
	if tasks[0].Name != name {
		return fmt.Errorf("erwartet %q, got %q", name, tasks[0].Name)
	}
	return nil
}

func (tc *testContext) dieZweiteAufgabeHeisst(name string) error {
	tasks := tc.store.GetTasks()
	if len(tasks) < 2 {
		return fmt.Errorf("weniger als 2 Aufgaben vorhanden")
	}
	if tasks[1].Name != name {
		return fmt.Errorf("erwartet %q, got %q", name, tasks[1].Name)
	}
	return nil
}

func (tc *testContext) dieErsteAufgabeHatDauer(min int) error {
	tasks := tc.store.GetTasks()
	if len(tasks) == 0 {
		return fmt.Errorf("keine Aufgaben vorhanden")
	}
	got := int(tasks[0].Duration.Minutes())
	if got != min {
		return fmt.Errorf("erwartet %d Minuten, got %d", min, got)
	}
	return nil
}

func (tc *testContext) dieErsteAufgabeIstAusstehend() error {
	tasks := tc.store.GetTasks()
	if len(tasks) == 0 {
		return fmt.Errorf("keine Aufgaben vorhanden")
	}
	if tasks[0].Status != model.StatusPending {
		return fmt.Errorf("erwartet StatusPending, got %v", tasks[0].Status)
	}
	return nil
}

func (tc *testContext) dieErsteAufgabeHatKeineStartzeit() error {
	tasks := tc.store.GetTasks()
	if len(tasks) == 0 {
		return fmt.Errorf("keine Aufgaben vorhanden")
	}
	if tasks[0].StartedAt != nil {
		return fmt.Errorf("Startzeit sollte nil sein, ist aber %v", tasks[0].StartedAt)
	}
	return nil
}

func (tc *testContext) istDieErsteAufgabe(name string) error {
	return tc.dieErsteAufgabeHeisst(name)
}

func (tc *testContext) istDieZweiteAufgabe(name string) error {
	return tc.dieZweiteAufgabeHeisst(name)
}

func (tc *testContext) istDieAufgabeAktiv(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	if t.Status != model.StatusActive {
		return fmt.Errorf("erwartet StatusActive, got %v", t.Status)
	}
	return nil
}

func (tc *testContext) istDieAufgabeErledigt(name string) error {
	t := tc.taskByName(name)
	if t == nil {
		return fmt.Errorf("Aufgabe %q nicht gefunden", name)
	}
	if t.Status != model.StatusDone {
		return fmt.Errorf("erwartet StatusDone, got %v", t.Status)
	}
	return nil
}

// ── Assertions: timeline ──────────────────────────────────────────────────────

func (tc *testContext) liegtDasVoraussichtlicheEnde(uhrzeit string) error {
	h, m, err := parseHHMM(uhrzeit)
	if err != nil {
		return err
	}
	projected := tc.store.ProjectedEndTime(tc.now)
	if projected.Hour() != h || projected.Minute() != m {
		return fmt.Errorf("erwartet Ende um %02d:%02d, got %02d:%02d",
			h, m, projected.Hour(), projected.Minute())
	}
	return nil
}

func (tc *testContext) istDerZeitplanImGruenenBereich() error {
	if !tc.store.IsOnTime(tc.now) {
		projected := tc.store.ProjectedEndTime(tc.now)
		return fmt.Errorf("Zeitplan überschritten: Ende um %s", projected.Format("15:04"))
	}
	return nil
}

func (tc *testContext) istDerZeitplanUeberschritten() error {
	if tc.store.IsOnTime(tc.now) {
		return fmt.Errorf("Zeitplan sollte überschritten sein, ist aber im grünen Bereich")
	}
	return nil
}

// ── Assertions: settings ──────────────────────────────────────────────────────

func (tc *testContext) betragtDiePufferzeit(min int) error {
	got := tc.store.Settings.BufferMinutes
	if got != min {
		return fmt.Errorf("erwartet %d Minuten Puffer, got %d", min, got)
	}
	return nil
}

func (tc *testContext) istDasTagesende(uhrzeit string) error {
	h, m, err := parseHHMM(uhrzeit)
	if err != nil {
		return err
	}
	if tc.store.Settings.EndHour != h || tc.store.Settings.EndMinute != m {
		return fmt.Errorf("erwartet Tagesende %02d:%02d, got %02d:%02d",
			h, m, tc.store.Settings.EndHour, tc.store.Settings.EndMinute)
	}
	return nil
}

func (tc *testContext) ichSetzeDiePufferzeit(min int) error {
	tc.store.Settings.BufferMinutes = min
	return nil
}

func (tc *testContext) ichSetzeDasTagesende(uhrzeit string) error {
	return tc.dasTagesende(uhrzeit)
}

// ── Carryover steps ───────────────────────────────────────────────────────────

func (tc *testContext) gesternGabEsAufgabe(status, name, dauer string) error {
	dur, err := model.ParseDuration(dauer)
	if err != nil {
		return err
	}
	t := model.NewTask(name, dur)
	switch status {
	case "eine erledigte":
		t.Status = model.StatusDone
	case "eine aktive":
		t.Status = model.StatusActive
		now := time.Now()
		t.StartedAt = &now
	case "eine ausstehende":
		t.Status = model.StatusPending
	default:
		return fmt.Errorf("unbekannter Status: %q", status)
	}
	return tc.writeYesterdayTask(t)
}

func (tc *testContext) writeYesterdayTask(t *model.Task) error {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	file := filepath.Join(tc.configDir, yesterday+".json")

	// Load existing
	var existing storage.DayDataExport
	if data, err := os.ReadFile(file); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	existing.Tasks = append(existing.Tasks, t)
	existing.Date = yesterday

	data, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

func (tc *testContext) gibtEsKeineUebertragbaren() error {
	got := tc.store.YesterdayUnfinished()
	if len(got) != 0 {
		return fmt.Errorf("erwartet 0 übertragbare Aufgaben, got %d", len(got))
	}
	return nil
}

func (tc *testContext) gibtEsNUebertragbare(n int) error {
	got := tc.store.YesterdayUnfinished()
	if len(got) != n {
		return fmt.Errorf("erwartet %d übertragbare Aufgaben, got %d", n, len(got))
	}
	return nil
}

func (tc *testContext) dieUebertragbareHeisst(name string) error {
	tasks := tc.store.YesterdayUnfinished()
	for _, t := range tasks {
		if t.Name == name {
			return nil
		}
	}
	return fmt.Errorf("Aufgabe %q nicht in übertragbaren Aufgaben", name)
}

func (tc *testContext) ichUebernehmeAlleGestrigen() error {
	tasks := tc.store.YesterdayUnfinished()
	tc.store.CarryOverTasks(tasks)
	return nil
}

// ── Reminder steps ────────────────────────────────────────────────────────────

func (tc *testContext) esWurdeNochKeineErinnerungAusgeloest() error {
	tc.mockNotify = &mockNotifier{}
	tc.checker = reminder.NewChecker(tc.store, tc.mockNotify)
	return nil
}

func (tc *testContext) derErinnerungsCheckLief(description string) error {
	var offset time.Duration
	switch description {
	case "vor 5 Minuten":
		offset = -5 * time.Minute
	case "vor 10 Minuten":
		offset = -10 * time.Minute
	default:
		return fmt.Errorf("unbekannte Zeit: %q", description)
	}
	past := tc.now.Add(offset)
	tc.checker.SetLastReminder(past)
	return nil
}

func (tc *testContext) derAlarmLief(description string) error {
	var offset time.Duration
	switch description {
	case "vor 5 Minuten":
		offset = -5 * time.Minute
	case "vor 10 Minuten":
		offset = -10 * time.Minute
	default:
		return fmt.Errorf("unbekannte Zeit: %q", description)
	}
	past := tc.now.Add(offset)
	tc.checker.SetLastAlarm(past)
	return nil
}

func (tc *testContext) derErinnerungsCheckLaeuft() error {
	tc.checker.Check(tc.now)
	return nil
}

func (tc *testContext) wirdEineErinnerungAusgeloest() error {
	if !tc.mockNotify.reminderFired {
		return fmt.Errorf("erwartet eine Erinnerung, aber keine wurde ausgelöst")
	}
	return nil
}

func (tc *testContext) wirdKeineErinnerungAusgeloest() error {
	if tc.mockNotify.reminderFired {
		return fmt.Errorf("erwartet keine Erinnerung, aber eine wurde ausgelöst (Task: %q)", tc.mockNotify.reminderTask)
	}
	return nil
}

func (tc *testContext) wirdEinAlarmAusgeloest() error {
	if !tc.mockNotify.alarmFired {
		return fmt.Errorf("erwartet einen Alarm, aber keiner wurde ausgelöst")
	}
	return nil
}

func (tc *testContext) wirdKeinAlarmAusgeloest() error {
	if tc.mockNotify.alarmFired {
		return fmt.Errorf("erwartet keinen Alarm, aber einer wurde ausgelöst")
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseHHMM(s string) (int, int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, 0, fmt.Errorf("ungültige Uhrzeit %q (erwartet HH:MM)", s)
	}
	return h, m, nil
}

// ── Suite setup ───────────────────────────────────────────────────────────────

func initSuite(tc *testContext) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		// Background
		sc.Step(`^ich habe keine Aufgaben$`, tc.ichHabeKeineAufgaben)
		sc.Step(`^es gibt keine gestrigen Aufgaben$`, tc.esGibtKeineGestrigen)
		sc.Step(`^das Tagesende ist um (\d+:\d+) Uhr$`, tc.dasTagesende)
		sc.Step(`^die Pufferzeit beträgt (\d+) Minuten$`, tc.diePufferzeit)
		sc.Step(`^die aktuelle Zeit ist (\d+:\d+) Uhr$`, tc.dieAktuelleZeit)
		sc.Step(`^es wurde noch keine Erinnerung ausgelöst$`, tc.esWurdeNochKeineErinnerungAusgeloest)

		// Actions
		sc.Step(`^ich füge eine Aufgabe "([^"]+)" mit Dauer "([^"]+)" hinzu$`, tc.ichFuegeAufgabeHinzu)
		sc.Step(`^ich entferne die Aufgabe "([^"]+)"$`, tc.ichEntferneAufgabe)
		sc.Step(`^ich verschiebe "([^"]+)" nach oben$`, tc.ichVerschiebeNachOben)
		sc.Step(`^ich verschiebe "([^"]+)" nach unten$`, tc.ichVerschiebeNachUnten)
		sc.Step(`^ich starte die Aufgabe "([^"]+)"$`, tc.ichStarteAufgabe)
		sc.Step(`^ich starte die Aufgabe "([^"]+)" um (\d+:\d+) Uhr$`, tc.ichStarteAufgabeUm)
		sc.Step(`^ich schließe die Aufgabe "([^"]+)" ab$`, tc.ichSchliesseAufgabeAb)
		sc.Step(`^(\d+) Minuten vergehen$`, tc.minutenVergehen)
		sc.Step(`^ich setze die Pufferzeit auf (\d+) Minuten$`, tc.ichSetzeDiePufferzeit)
		sc.Step(`^ich setze das Tagesende auf (\d+:\d+) Uhr$`, tc.ichSetzeDasTagesende)

		// Carryover actions
		sc.Step(`^gestern gab es (eine erledigte|eine aktive|eine ausstehende) Aufgabe "([^"]+)" mit Dauer "([^"]+)"$`, tc.gesternGabEsAufgabe)
		sc.Step(`^ich übernehme alle gestrigen Aufgaben$`, tc.ichUebernehmeAlleGestrigen)

		// Reminder actions
		sc.Step(`^der Erinnerungs-Check lief (vor \d+ Minuten)$`, tc.derErinnerungsCheckLief)
		sc.Step(`^der Alarm lief (vor \d+ Minuten)$`, tc.derAlarmLief)
		sc.Step(`^der Erinnerungs-Check läuft$`, tc.derErinnerungsCheckLaeuft)

		// Assertions: tasks
		sc.Step(`^enthält die Liste (\d+) Aufgaben?$`, tc.enthaeltDieListe)
		sc.Step(`^die erste Aufgabe heißt "([^"]+)"$`, tc.dieErsteAufgabeHeisst)
		sc.Step(`^die zweite Aufgabe heißt "([^"]+)"$`, tc.dieZweiteAufgabeHeisst)
		sc.Step(`^die erste Aufgabe hat eine Dauer von (\d+) Minuten$`, tc.dieErsteAufgabeHatDauer)
		sc.Step(`^die erste Aufgabe ist ausstehend$`, tc.dieErsteAufgabeIstAusstehend)
		sc.Step(`^die erste Aufgabe hat keine Startzeit$`, tc.dieErsteAufgabeHatKeineStartzeit)
		sc.Step(`^ist die erste Aufgabe "([^"]+)"$`, tc.istDieErsteAufgabe)
		sc.Step(`^ist die zweite Aufgabe "([^"]+)"$`, tc.istDieZweiteAufgabe)
		sc.Step(`^ist die Aufgabe "([^"]+)" aktiv$`, tc.istDieAufgabeAktiv)
		sc.Step(`^ist die Aufgabe "([^"]+)" erledigt$`, tc.istDieAufgabeErledigt)

		// Assertions: timeline
		sc.Step(`^liegt das voraussichtliche Ende um (\d+:\d+) Uhr$`, tc.liegtDasVoraussichtlicheEnde)
		sc.Step(`^ist der Zeitplan im grünen Bereich$`, tc.istDerZeitplanImGruenenBereich)
		sc.Step(`^ist der Zeitplan überschritten$`, tc.istDerZeitplanUeberschritten)

		// Assertions: settings
		sc.Step(`^beträgt die Pufferzeit (\d+) Minuten$`, tc.betragtDiePufferzeit)
		sc.Step(`^ist das Tagesende um (\d+:\d+) Uhr$`, tc.istDasTagesende)

		// Assertions: carryover
		sc.Step(`^gibt es keine übertragbaren Aufgaben vom Vortag$`, tc.gibtEsKeineUebertragbaren)
		sc.Step(`^gibt es (\d+) übertragbare Aufgaben? vom Vortag$`, tc.gibtEsNUebertragbare)
		sc.Step(`^die übertragbare Aufgabe heißt "([^"]+)"$`, tc.dieUebertragbareHeisst)

		// Assertions: reminders
		sc.Step(`^wird eine Erinnerung ausgelöst$`, tc.wirdEineErinnerungAusgeloest)
		sc.Step(`^wird keine Erinnerung ausgelöst$`, tc.wirdKeineErinnerungAusgeloest)
		sc.Step(`^keine Erinnerung wird ausgelöst$`, tc.wirdKeineErinnerungAusgeloest)
		sc.Step(`^wird ein Alarm ausgelöst$`, tc.wirdEinAlarmAusgeloest)
		sc.Step(`^wird keine Alarm ausgelöst$`, tc.wirdKeinAlarmAusgeloest)
		sc.Step(`^kein Alarm wird ausgelöst$`, tc.wirdKeinAlarmAusgeloest)

	}
}

func TestFeatures(t *testing.T) {
	tc := newTestContext()
	defer tc.cleanup()

	suite := godog.TestSuite{
		Name:                "taskalarm",
		ScenarioInitializer: initSuite(tc),
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog: nicht alle Szenarien erfolgreich")
	}
}
