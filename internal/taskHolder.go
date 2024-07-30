package internal

import (
	"fmt"
	"time"
)

// Pointers for optional fields
type TaskOptional struct {
	Done      *bool
	Msg       *string
	Category  *TaskCategory
	PlannedAt *time.Time
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
	return t.Tasks
}

func (t *TaskHolder) Add(task Task) {
	t.latestId++
	if task.Id < t.latestId {
		task.Id = t.latestId
	}
	t.Tasks = append(t.Tasks, task)

}

func (t *TaskHolder) CreateTask(update *TaskOptional) *Task {
	t.latestId++
	task := NewTask(t.latestId, *update.Msg, *update.Category, *update.PlannedAt)
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
		if update.PlannedAt.Before(time.Now()) {
			return &PastPlannedTimeError{PlannedTime: *update.PlannedAt}
		}
		task.PlannedAt = *update.PlannedAt
	}

	return nil
}

func (t *TaskHolder) DeleteTask(taskId int) error {
	index := -1
	for i, task := range t.Tasks {
		if task.Id == taskId {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("task with ID %d not found", taskId)
	}

	t.Tasks = append(t.Tasks[:index], t.Tasks[index+1:]...)

	return nil
}

func isValidTaskCategory(category TaskCategory) bool {
	return Brewing <= category && category <= Quality
}
