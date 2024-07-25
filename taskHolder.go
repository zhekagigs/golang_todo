package main

import (
	"fmt"
	"time"
)

type TaskUpdate struct {
	Done      *bool
	Msg       *string
	Category  *TaskCategory
	PlannedAt *time.Time
}

type TaskHolder struct {
	latestId int
	tasks    []Task
}

func NewTaskHolder() *TaskHolder {
	return &TaskHolder{}
}

func (t *TaskHolder) Read() []Task {
	return t.tasks
}

func (t *TaskHolder) Add(task Task) {
	t.tasks = append(t.tasks, task)
}

func (t *TaskHolder) CreateTask(taskValue string, category TaskCategory, plannedAt time.Time) Task {
	t.latestId++
	task := NewTask(t.latestId, taskValue, category, plannedAt)
	t.tasks = append(t.tasks, task)
	return task
}

func (t *TaskHolder) FindTaskById(taskId int) (*Task, error) {
	for _, v := range t.tasks {
		if v.Id == taskId {
			return &v, nil
		}
	}
	return nil, ErrNotFound
}

func (t *TaskHolder) PartialUpdateTask(taskId int, update TaskUpdate) error {
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
	for i, task := range t.tasks {
		if task.Id == taskId {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("task with ID %d not found", taskId)
	}

	t.tasks = append(t.tasks[:index], t.tasks[index+1:]...)

	return nil
}

func isValidTaskCategory(category TaskCategory) bool {
	return Brewing <= category && category <= Quality
}
