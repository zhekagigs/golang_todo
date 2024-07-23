package main

import (
	"strings"
	"testing"
	"time"
)


func TestNewTask(t *testing.T) {
	originalTimeNow := timeNow

	defer func() {
		timeNow = originalTimeNow
	}()

	mockTime := time.Date(2023, 7, 23, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return mockTime
	}

	id := 1
	taskDescription := "Brew IPA"
	category := Brewing
	plannedAt := time.Date(2023, 7, 24, 10, 0, 0, 0, time.UTC)

	task := NewTask(id, taskDescription, category, plannedAt)

	if task.id != id {
		t.Errorf("Expected id %d, got %d", id, task.id)
	}
	if task.task != taskDescription {
		t.Errorf("Expected task %s, got %s", taskDescription, task.task)
	}
	if task.category != category {
		t.Errorf("Expected category %d, got %d", category, task.category)
	}
	if !task.createdAt.Equal(mockTime) {
		t.Errorf("Expected createdAt %v, got %v", mockTime, task.createdAt)
	}
	if !task.plannedAt.Equal(plannedAt) {
		t.Errorf("Expected plannedAt %v, got %v", plannedAt, task.plannedAt)
	}
	if task.done {
		t.Error("Expected done to be false")
	}
}

func TestString(t *testing.T) {
	originalTime := timeNow
	defer func() {
		timeNow = originalTime
	}()
	timeNow = func() time.Time { return time.Date(2024, 07, 22, 16, 43, 00, 00, time.Local) }
	
	task := NewTask(1, "Brew Beer", 0, TimeExample)
	got := strings.Split(task.String(), ",")
	want := strings.Split("id:1,[Brewing] Brew Beer, created: Monday, July 22, 2024 at 16:43, planned: Sunday, July 14, 2024 at 12:45", ",")

	for i := 0; i < len(want); i++ {
		if got[i] != want[i] {
			t.Errorf("got %v want %v", got[i], want[i])
		}
	}
}
