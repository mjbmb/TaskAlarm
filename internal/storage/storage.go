package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/mjbmb/TaskAlarm/internal/model"
)

type Settings struct {
	BufferMinutes     int    `json:"buffer_minutes"`
	EndHour           int    `json:"end_hour"`
	EndMinute         int    `json:"end_minute"`
	UseObsidian       bool   `json:"use_obsidian"`
	ObsidianVaultPath string `json:"obsidian_vault_path"`
	ObsidianSetupDone bool   `json:"obsidian_setup_done"`
}

type DayData = DayDataExport

// DayDataExport is exported so tests can write fixture files directly.
type DayDataExport struct {
	Tasks []*model.Task `json:"tasks"`
	Date  string        `json:"date"`
}

type Storage struct {
	mu        sync.RWMutex
	configDir string
	Tasks     []*model.Task
	Settings  Settings
}

func New() *Storage {
	home, _ := os.UserHomeDir()
	return NewWithDir(filepath.Join(home, ".config", "taskalarm"))
}

// NewWithDir creates a Storage using a custom directory — used in tests.
func NewWithDir(configDir string) *Storage {
	_ = os.MkdirAll(configDir, 0755)
	s := &Storage{
		configDir: configDir,
		Settings:  Settings{BufferMinutes: 10, EndHour: 20, EndMinute: 0},
	}
	s.load()
	return s
}

func (s *Storage) todayFile() string {
	return filepath.Join(s.configDir, time.Now().Format("2006-01-02")+".json")
}

func (s *Storage) settingsFile() string {
	return filepath.Join(s.configDir, "settings.json")
}

func (s *Storage) load() {
	if data, err := os.ReadFile(s.settingsFile()); err == nil {
		_ = json.Unmarshal(data, &s.Settings)
	}
	if s.Settings.BufferMinutes == 0 {
		s.Settings.BufferMinutes = 10
	}
	if s.Settings.EndHour == 0 && s.Settings.EndMinute == 0 {
		s.Settings.EndHour = 20
	}

	today := time.Now().Format("2006-01-02")
	if s.Settings.UseObsidian && s.Settings.ObsidianVaultPath != "" {
		if tasks, err := loadMarkdownDay(s.Settings.ObsidianVaultPath, today); err == nil && len(tasks) > 0 {
			s.Tasks = tasks
			return
		}
	}

	if data, err := os.ReadFile(s.todayFile()); err == nil {
		var dayData DayData
		if err := json.Unmarshal(data, &dayData); err == nil {
			s.Tasks = dayData.Tasks
		}
	}
	if s.Tasks == nil {
		s.Tasks = []*model.Task{}
	}

	sort.Slice(s.Tasks, func(i, j int) bool {
		return s.Tasks[i].Order < s.Tasks[j].Order
	})
}

func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if data, err := json.Marshal(s.Settings); err == nil {
		_ = os.WriteFile(s.settingsFile(), data, 0644)
	}

	dayData := DayData{
		Tasks: s.Tasks,
		Date:  time.Now().Format("2006-01-02"),
	}
	data, err := json.Marshal(dayData)
	if err != nil {
		return err
	}
	return os.WriteFile(s.todayFile(), data, 0644)
}

func (s *Storage) AddTask(task *model.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.Order = len(s.Tasks)
	s.Tasks = append(s.Tasks, task)
	s.saveUnlocked()
}

func (s *Storage) RemoveTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.Tasks {
		if t.ID == id {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
			break
		}
	}
	s.reorderUnlocked()
	s.saveUnlocked()
}

func (s *Storage) MoveUp(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.Tasks {
		if t.ID == id && i > 0 {
			s.Tasks[i], s.Tasks[i-1] = s.Tasks[i-1], s.Tasks[i]
			break
		}
	}
	s.reorderUnlocked()
	s.saveUnlocked()
}

func (s *Storage) MoveDown(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.Tasks {
		if t.ID == id && i < len(s.Tasks)-1 {
			s.Tasks[i], s.Tasks[i+1] = s.Tasks[i+1], s.Tasks[i]
			break
		}
	}
	s.reorderUnlocked()
	s.saveUnlocked()
}

func (s *Storage) StartTask(id string) {
	s.StartTaskAt(id, time.Now())
}

// StartTaskAt starts a task at a specific time — used in tests.
func (s *Storage) StartTaskAt(id string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.Tasks {
		if t.ID == id {
			t.Status = model.StatusActive
			t.StartedAt = &at
			break
		}
	}
	s.saveUnlocked()
}

func (s *Storage) CompleteTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for _, t := range s.Tasks {
		if t.ID == id {
			t.Status = model.StatusDone
			t.DoneAt = &now
			break
		}
	}
	s.saveUnlocked()
}

func (s *Storage) reorderUnlocked() {
	for i, t := range s.Tasks {
		t.Order = i
	}
}

func (s *Storage) saveUnlocked() {
	today := time.Now().Format("2006-01-02")
	dayData := DayData{Tasks: s.Tasks, Date: today}
	data, _ := json.Marshal(dayData)
	_ = os.WriteFile(s.todayFile(), data, 0644)
	if data, err := json.Marshal(s.Settings); err == nil {
		_ = os.WriteFile(s.settingsFile(), data, 0644)
	}
	if s.Settings.UseObsidian && s.Settings.ObsidianVaultPath != "" {
		_ = saveMarkdownDay(s.Settings.ObsidianVaultPath, today, s.Tasks)
	}
}

// ProjectedEndTime calculates when all remaining tasks will be done.
func (s *Storage) ProjectedEndTime(now time.Time) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projectedEndTimeUnlocked(now)
}

func (s *Storage) projectedEndTimeUnlocked(now time.Time) time.Time {
	t := now
	buffer := time.Duration(s.Settings.BufferMinutes) * time.Minute
	first := true
	for _, task := range s.Tasks {
		if task.Status == model.StatusDone {
			continue
		}
		if !first {
			t = t.Add(buffer)
		}
		t = t.Add(task.RemainingDuration(now))
		first = false
	}
	return t
}

func (s *Storage) Deadline(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(),
		s.Settings.EndHour, s.Settings.EndMinute, 0, 0, now.Location())
}

func (s *Storage) IsOnTime(now time.Time) bool {
	return !s.ProjectedEndTime(now).After(s.Deadline(now))
}

// YesterdayUnfinished returns incomplete tasks from yesterday's file.
func (s *Storage) YesterdayUnfinished() []*model.Task {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	if s.Settings.UseObsidian && s.Settings.ObsidianVaultPath != "" {
		if tasks, err := loadMarkdownDay(s.Settings.ObsidianVaultPath, yesterday); err == nil && len(tasks) > 0 {
			var unfinished []*model.Task
			for _, t := range tasks {
				if t.Status != model.StatusDone {
					unfinished = append(unfinished, t)
				}
			}
			return unfinished
		}
	}

	file := filepath.Join(s.configDir, yesterday+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var dayData DayData
	if err := json.Unmarshal(data, &dayData); err != nil {
		return nil
	}
	var unfinished []*model.Task
	for _, t := range dayData.Tasks {
		if t.Status != model.StatusDone {
			unfinished = append(unfinished, t)
		}
	}
	return unfinished
}

// CarryOverTasks adds the given tasks to today, resetting their status to pending.
func (s *Storage) CarryOverTasks(tasks []*model.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range tasks {
		t.Status = model.StatusPending
		t.StartedAt = nil
		t.DoneAt = nil
		t.Order = len(s.Tasks)
		s.Tasks = append(s.Tasks, t)
	}
	s.reorderUnlocked()
	s.saveUnlocked()
}

func (s *Storage) HasPendingTasks() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.Tasks {
		if t.Status == model.StatusPending {
			return true
		}
	}
	return false
}

func (s *Storage) NextPendingTask() *model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.Tasks {
		if t.Status == model.StatusPending {
			return t
		}
	}
	return nil
}

func (s *Storage) GetTasks() []*model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Task, len(s.Tasks))
	copy(result, s.Tasks)
	return result
}
