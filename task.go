package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type TaskCategory int

const (
	Brewing TaskCategory = iota
	Marketing
	Logistics
	Quality
)

type Task struct {
	Id        int
	Task      string
	Category  TaskCategory
	Done      bool
	CreatedAt time.Time
	PlannedAt time.Time
}

func NewTask(id int, task string, category TaskCategory, plannedAt time.Time) Task {
	return Task{
		Id:        id,
		Task:      task,
		Category:  category,
		Done:      false,
		CreatedAt: timeNow().Round(0),
		PlannedAt: plannedAt.Round(0),
	}
}

func (t *Task) String() string {
	categoryName := [...]string{"Brewing", "Marketing", "Logistics", "Quality"}[t.Category]

	return fmt.Sprintf("id:%d,[%s] %s, created: %s, planned: %s",
		t.Id,
		categoryName,
		t.Task,
		formatDatetime(t.CreatedAt),
		formatDatetime(t.PlannedAt))
}

func PrintTasks(out io.Writer, tasks ...Task) {
	for _, task := range tasks {
		fmt.Fprintln(out, task.String())
	}
}

func SaveToJson(filePath string, tasks ...Task) error {
	if !strings.HasSuffix(filePath, ".json") {
		return fmt.Errorf("invalid file extension: %s. Expected a .json file", filePath)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to a file '%s': %w", filePath, err)
	}
	return nil
}

func ReadFromJson(filePath string) ([]Task, error) {
	if !strings.HasSuffix(filePath, ".json") {
		return nil, fmt.Errorf("invalid file extension: %s. Expected a .json file ", filePath)
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read a file '%s': %w", filePath, err)
	}

	var tasks []Task
	if err := json.Unmarshal(bytes, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from file '%s': %w", filePath, err)
	}

	return tasks, nil
}
