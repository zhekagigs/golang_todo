package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhekagigs/golang_todo/users"
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
	CreatedBy users.User
}

func NewTask(id int, task string, category TaskCategory, plannedAt time.Time, user *users.User) Task {
	if user == nil {
		user = &users.User{UserName: "Team"}
	}
	return Task{
		Id:        id,
		Msg:       task,
		Category:  category,
		Done:      false,
		CreatedAt: timeNow().Round(0),
		PlannedAt: plannedAt.Round(0),
		CreatedBy: *user,
	}
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func (t *Task) String() string {
	return fmt.Sprintf("id:"+colorPurple+"%d,"+colorReset+colorBlue+"[%s] "+colorReset+colorCyan+"%s,"+colorReset+"\ncreated: %s,\nplanned: %s, \nfinished: %v",
		t.Id,
		t.Category.String(),
		t.Msg,
		formatDatetime(t.CreatedAt),
		formatDatetime(t.PlannedAt),
		t.Done)
}

func PrintTasks(out io.Writer, tasks ...Task) {
	for _, task := range tasks {
		fmt.Fprintln(out, task.String()+"\n")
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

func ReadTasksFromJSON(filePath string) ([]Task, error) {
	if err := validateJSONFile(filePath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	return UnmarshalTasks(data)
}

func UnmarshalTasks(data []byte) ([]Task, error) {
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return tasks, nil
}

func UnmarshalTask(data []byte) (*Task, error) {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &task, nil
}

func validateJSONFile(filePath string) error {
	ext := filepath.Ext(filePath)
	if ext != ".json" {
		return fmt.Errorf("invalid file extension: %s. Expected a .json file", filePath)
	}
	return nil
}
