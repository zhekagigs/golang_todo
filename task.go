package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

func SaveToJson(fileName string, tasks ...Task) {
	data, err := json.Marshal(tasks)
	check(err)

	err = os.WriteFile(fileName, data, 0644)
	check(err)
}

func ReadFromJson(fileName string) []Task {
	bytes, err := os.ReadFile(fileName)
	check(err)
	str := bytes
	var res []Task
	json.Unmarshal([]byte(str), &res)
	return res
}
