package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const TASK_TIME_FORMAT = "2006-01-02 15:04"

var jsonMarshal = json.MarshalIndent //for monkey patching mock

// error definitions
var (
	ErrNotFound = errors.New("not found")
)

type InvalidCategoryError struct {
	Category TaskCategory
}

func (e *InvalidCategoryError) Error() string {
	return fmt.Sprintf("invalid task category: %v", e.Category)
}

type EmptyTaskValueError struct{}

func (e *EmptyTaskValueError) Error() string {
	return "task value cannot be empty"
}

type PastPlannedTimeError struct {
	PlannedTime time.Time
}

func (e *PastPlannedTimeError) Error() string {
	return fmt.Sprintf("planned time %v is in the past", e.PlannedTime)
}

type TaskCategory int

const (
	Brewing TaskCategory = iota
	Marketing
	Logistics
	Quality
)

func (tc TaskCategory) String() string {
	return [...]string{"Brewing", "Marketing", "Logistics", "Quality"}[tc]
}

type Task struct {
	Id        int
	Msg       string
	Category  TaskCategory
	Done      bool
	CreatedAt time.Time
	PlannedAt time.Time
}

func NewTask(id int, task string, category TaskCategory, plannedAt time.Time) Task {
	return Task{
		Id:        id,
		Msg:       task,
		Category:  category,
		Done:      false,
		CreatedAt: timeNow().Round(0),
		PlannedAt: plannedAt.Round(0),
	}
}

func (t *Task) String() string {
	return fmt.Sprintf("id:%d,[%s] %s, created: %s, planned: %s, finished: %v",
		t.Id,
		t.Category.String(),
		t.Msg,
		formatDatetime(t.CreatedAt),
		formatDatetime(t.PlannedAt),
		t.Done)
}

func PrintTasks(out io.Writer, tasks ...Task) {
	for _, task := range tasks {
		fmt.Fprintln(out, task.String())
	}
}

func WriteToJson(filePath string, tasks ...Task) error {
	if !strings.HasSuffix(filePath, ".json") {
		return fmt.Errorf("invalid file extension: %s. Expected a .json file", filePath)
	}

	data, err := jsonMarshal(tasks, "", "  ")
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
		return nil, fmt.Errorf("invalid file extension: %s. Expected a .json file", filePath)
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
