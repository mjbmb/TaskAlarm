package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mjbmb/TaskAlarm/internal/model"
	"gopkg.in/yaml.v3"
)

type taskFrontmatter struct {
	ID              string `yaml:"id"`
	Name            string `yaml:"name"`
	Status          int    `yaml:"status"`
	DurationMinutes int64  `yaml:"duration_minutes"`
	Order           int    `yaml:"order"`
	StartedAt       string `yaml:"started_at,omitempty"`
	DoneAt          string `yaml:"done_at,omitempty"`
}

func saveMarkdownDay(vaultPath, date string, tasks []*model.Task) error {
	dir := filepath.Join(vaultPath, "TaskAlarm", date)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	for _, t := range tasks {
		if err := writeTaskFile(dir, t); err != nil {
			return err
		}
	}
	return nil
}

func loadMarkdownDay(vaultPath, date string) ([]*model.Task, error) {
	dir := filepath.Join(vaultPath, "TaskAlarm", date)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var tasks []*model.Task
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		t, err := readTaskFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Order < tasks[j].Order
	})
	return tasks, nil
}

func writeTaskFile(dir string, t *model.Task) error {
	fm := taskFrontmatter{
		ID:              t.ID,
		Name:            t.Name,
		Status:          int(t.Status),
		DurationMinutes: int64(t.Duration.Minutes()),
		Order:           t.Order,
	}
	if t.StartedAt != nil {
		fm.StartedAt = t.StartedAt.Format(time.RFC3339)
	}
	if t.DoneAt != nil {
		fm.DoneAt = t.DoneAt.Format(time.RFC3339)
	}
	fmData, err := yaml.Marshal(fm)
	if err != nil {
		return err
	}
	content := fmt.Sprintf("---\n%s---\n\n# %s\n", fmData, t.Name)
	filename := fmt.Sprintf("%02d-%s.md", t.Order, slugify(t.Name))
	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644)
}

func readTaskFile(path string) (*model.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return nil, fmt.Errorf("missing frontmatter")
	}
	rest := s[4:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, fmt.Errorf("unclosed frontmatter")
	}
	var fm taskFrontmatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return nil, err
	}
	t := &model.Task{
		ID:       fm.ID,
		Name:     fm.Name,
		Status:   model.TaskStatus(fm.Status),
		Duration: time.Duration(fm.DurationMinutes) * time.Minute,
		Order:    fm.Order,
	}
	if fm.StartedAt != "" {
		if ts, err := time.Parse(time.RFC3339, fm.StartedAt); err == nil {
			t.StartedAt = &ts
		}
	}
	if fm.DoneAt != "" {
		if ts, err := time.Parse(time.RFC3339, fm.DoneAt); err == nil {
			t.DoneAt = &ts
		}
	}
	return t, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	result := strings.TrimRight(b.String(), "-")
	if len(result) > 40 {
		result = result[:40]
	}
	if result == "" {
		result = "task"
	}
	return result
}
