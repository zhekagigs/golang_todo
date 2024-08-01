package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	TimeFormat = "2006-01-02T15:04:05Z07:00"
)

var ErrTimeNilPointer = errors.New("time: nil pointer")

type TaskOptional struct {
	Done      *bool         `json:"done"`
	Msg       *string       `json:"msg"`
	Category  *TaskCategory `json:"category"`
	PlannedAt *CustomTime   `json:"plannedAt"`
}

func AdapterTaskOptional(task Task) TaskOptional {
	custmTime, err := NewCustomTime(&task.PlannedAt)
	if err != nil && err != ErrTimeNilPointer {
		panic(err)
	}
	return TaskOptional{
		Done:      &task.Done,
		Msg:       &task.Msg,
		Category:  &task.Category,
		PlannedAt: custmTime,
	}
}

type CustomTime struct {
	Time time.Time
}

func NewCustomTime(timePtr *time.Time) (*CustomTime, error) {
	if timePtr == nil {
		return nil, ErrTimeNilPointer
	} else {
		return &CustomTime{Time: *timePtr}, nil
	}
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(TimeFormat, s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(TimeFormat))), nil
}

func (ct *CustomTime) AsTime() time.Time {
	return ct.Time
}

// mainly to mock test
type TaskService interface {
	Read() []Task
	CreateTask(*TaskOptional) *Task
	FindTaskById(int) (*Task, error)
	PartialUpdateTask(int, *TaskOptional) error
	DeleteTask(int) error
}

// implements TaskService interface
type TaskHolder struct {
	latestId int
	Tasks    []Task
	DiskPath string
}

func NewTaskHolder(diskPath string) *TaskHolder {
	return &TaskHolder{DiskPath: diskPath}
}

func (t *TaskHolder) Read() []Task {
	return append([]Task(nil), t.Tasks...)
}

// returns latestId and len of tasks
func (t *TaskHolder) Count() (int, int) {
	return t.latestId, len(t.Tasks)
}

func (t *TaskHolder) Add(task Task) {
	t.latestId++
	// if task.Id < t.latestId {
	// 	task.Id = t.latestId
	// }
	t.Tasks = append(t.Tasks, task)

}

func (t *TaskHolder) CreateTask(update *TaskOptional) *Task {
	t.latestId++

	var msg string
	if update.Msg != nil {
		msg = *update.Msg
	}

	var category TaskCategory
	if update.Category != nil {
		category = *update.Category
	}

	var plannedAt time.Time
	if update.PlannedAt != nil {
		plannedAt = update.PlannedAt.Time
	}

	task := NewTask(t.latestId, msg, category, plannedAt)
	t.Tasks = append(t.Tasks, task)
	return &task
}

func (t *TaskHolder) FindTaskById(taskId int) (*Task, error) {
	for i := range t.Tasks {
		if t.Tasks[i].Id == taskId {
			return &t.Tasks[i], nil
		}
	}
	return nil, ErrNotFound
}

func (t *TaskHolder) PartialUpdateTask(taskId int, update *TaskOptional) error {
	task, err := t.FindTaskById(taskId)
	if err != nil {
		return err
	}

	if update.Done != nil {
		task.Done = *update.Done
	}

	if update.Msg != nil {
		if len(*update.Msg) == 0 {
			return &EmptyTaskValueError{}
		}
		task.Msg = *update.Msg
	}

	if update.Category != nil {
		if !isValidTaskCategory(*update.Category) {
			return &InvalidCategoryError{Category: *update.Category}
		}
		task.Category = *update.Category
	}

	if update.PlannedAt != nil {
		if update.PlannedAt.Time.Before(time.Now()) {
			return &PastPlannedTimeError{PlannedTime: *&update.PlannedAt.Time}
		}
		task.PlannedAt = *&update.PlannedAt.Time
	}

	return nil
}

func (t *TaskHolder) DeleteTask(taskId int) error {
	index := -1
	topId := -1
	for i, task := range t.Tasks {
		if task.Id == taskId {
			index = i

		}
		if task.Id > topId {
			topId = task.Id
		}
	}

	if index == -1 {
		return fmt.Errorf("task with ID %d not found", taskId)
	}

	// deletedIndex = t.Tasks[index]
	t.Tasks = append(t.Tasks[:index], t.Tasks[index+1:]...)

	return nil
}

func isValidTaskCategory(category TaskCategory) bool {
	return Brewing <= category && category <= Quality
}
