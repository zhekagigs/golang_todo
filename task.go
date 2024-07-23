package main

import (
	"fmt"
	"io"
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
	id        int
	task      string
	category  TaskCategory
	done      bool
	createdAt time.Time
	plannedAt time.Time
}

func NewTask(id int, task string, category TaskCategory, plannedAt time.Time) Task {
	return Task{
		id:        id,
		task:      task,
		category:  category,
		done:      false,
		createdAt: timeNow(),
		plannedAt: plannedAt,
	}
}

func (t *Task) String() string {
	categoryName := [...]string{"Brewing", "Marketing", "Logistics", "Quality"}[t.category]

	return fmt.Sprintf("id:%d,[%s] %s, created: %s, planned: %s",
		t.id,
		categoryName,
		t.task,
		formatDatetime(t.createdAt),
		formatDatetime(t.plannedAt))
}

func PrintTasks(out io.Writer, tasks ...Task) {
	for _, task := range tasks {
		fmt.Fprintln(out, task.String())
	}
}
