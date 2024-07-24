package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewTaskMock(t *testing.T) {
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

	if task.Id != id {
		t.Errorf("Expected id %d, got %d", id, task.Id)
	}
	if task.Task != taskDescription {
		t.Errorf("Expected task %s, got %s", taskDescription, task.Task)
	}
	if task.Category != category {
		t.Errorf("Expected category %d, got %d", category, task.Category)
	}
	if !task.CreatedAt.Equal(mockTime) {
		t.Errorf("Expected createdAt %v, got %v", mockTime, task.CreatedAt)
	}
	if !task.PlannedAt.Equal(plannedAt) {
		t.Errorf("Expected plannedAt %v, got %v", plannedAt, task.PlannedAt)
	}
	if task.Done {
		t.Error("Expected done to be false")
	}
}

func TestNewTaskReal(t *testing.T) {
	id := 1
	taskDescription := "Brew IPA"
	category := Brewing
	plannedAt := time.Date(2023, 7, 24, 10, 0, 0, 0, time.UTC)
	timeBefore := time.Now()
	task := NewTask(id, taskDescription, category, plannedAt)
	timeAfter := time.Now()

	if task.Id != id {
		t.Errorf("Expected id %d, got %d", id, task.Id)
	}
	if task.Task != taskDescription {
		t.Errorf("Expected task %s, got %s", taskDescription, task.Task)
	}
	if task.Category != category {
		t.Errorf("Expected category %d, got %d", category, task.Category)
	}
	if task.CreatedAt.Before(timeBefore) || task.CreatedAt.After(timeAfter) {
		t.Errorf("Task creation time %v is not between %v and %v", task.CreatedAt, timeBefore, timeAfter)
	}
	if !task.PlannedAt.Equal(plannedAt) {
		t.Errorf("Expected plannedAt %v, got %v", plannedAt, task.PlannedAt)
	}
	if task.Done {
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

func TestPrintTasks(t *testing.T) {
	taskA := NewTask(1, "Brew", 0, TimeExample)
	taskB := NewTask(2, "Advertise", 1, TimeExample)

	buffer := &bytes.Buffer{}
	PrintTasks(buffer, taskA, taskB)

	got := buffer.String()
	want := taskA.String() + "\n" + taskB.String() + "\n"
	if got != want {
		t.Errorf("got \n%v\n want \n%v\n", got, want)
	}
}

func TestReadAndWriteJson(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		taskA := NewTask(1, "Brew", 0, TimeExample)
		taskB := NewTask(2, "Advertise", 1, TimeExample)

		SaveToJson("resources/test_tasks.json", taskA, taskB)

		got, err := ReadFromJson("resources/test_tasks.json")
		if err != nil {
			t.Errorf("ReadFromJson failed!")
		}
		want := []Task{taskA, taskB}

		if got[0] != want[0] {
			t.Errorf("got \n%v\n want \n%v\n", got[0], want[0])
		}
		if got[1] != want[1] {
			t.Errorf("got \n%v\n want \n%v\n", got[1], want[1])
		}
	})

	t.Run("Wrong file name read with correct suffix", func(t *testing.T) {
		_, err := ReadFromJson("wrongFileName.json")
		if err == nil {
			t.Errorf("There must be an error")
			return
		}

		// Check if the error is a *fs.PathError
		var pathError *os.PathError
		if !errors.As(err, &pathError) {
			t.Errorf("Expected *fs.PathError, got %T", err)
			return
		}

		// Check if the underlying error is os.ErrNotExist
		if !errors.Is(pathError, os.ErrNotExist) {
			t.Errorf("Expected underlying error to be os.ErrNotExist, got %v", pathError.Err)
		}
	})


	t.Run("Wrong file name with incorrect suffix", func(t *testing.T) {
		_, err := ReadFromJson("wrongFileName.garbage")
		if err == nil {
			t.Errorf("There must be an error")
			return
		}

		// Check if the error is a *fs.PathError
		var pathError *os.PathError
		if !errors.As(err, &pathError) {
			t.Errorf("Expected *fs.PathError, got %T", err)
			return
		}

		// Check if the underlying error is os.ErrNotExist
		if !errors.Is(pathError, os.ErrNotExist) {
			t.Errorf("Expected underlying error to be os.ErrNotExist, got %v", pathError.Err)
		}
	})
}
