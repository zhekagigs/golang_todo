package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestNewTaskMock(t *testing.T) {

	timeNow = ProvideMocktimeNow(t)
	mockTime := timeNow()

	id := 1
	taskDescription := "Brew IPA"
	category := Brewing
	plannedAt := time.Date(2023, 7, 24, 10, 0, 0, 0, time.UTC)

	task := NewTask(id, taskDescription, category, plannedAt)

	if task.Id != id {
		t.Errorf("Expected id %d, got %d", id, task.Id)
	}
	if task.Msg != taskDescription {
		t.Errorf("Expected task %s, got %s", taskDescription, task.Msg)
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
	if task.Msg != taskDescription {
		t.Errorf("Expected task %s, got %s", taskDescription, task.Msg)
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
	timeNow = ProvideMocktimeNow(t)
	task := NewTask(1, "Brew Beer", 0, TimeExample)
	got := strings.Split(task.String(), ",")

	want := strings.Split("id:1,[Brewing] Brew Beer, created: Sunday, July 23, 2023 at 12:00, planned: Sunday, July 14, 2024 at 12:45", ",")

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
	t.Run("Happy Path read and write", func(t *testing.T) {
		taskA := NewTask(1, "Brew", 0, TimeExample)
		taskB := NewTask(2, "Advertise", 1, TimeExample)

		err := WriteToJson("../resources/test_tasks.json", taskA, taskB)

		if err != nil {
			t.Errorf("Unexpected err %v", err)
		}

		got, err := ReadFromJson("../resources/test_tasks.json")
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

	t.Run("Wrong file extension to write", func(t *testing.T) {
		taskA := NewTask(1, "Brew", 0, TimeExample)
		taskB := NewTask(2, "Advertise", 1, TimeExample)

		err := WriteToJson("resources/test_tasks.txt", taskA, taskB)

		if err == nil {
			t.Errorf("ReadFromJson didn't fail as expected with given wrong file extension!")
		}

		want := "invalid file extension: resources/test_tasks.txt. Expected a .json file"

		if !strings.Contains(err.Error(), want) {
			t.Errorf("got %v want %v", err.Error(), want)
		}
	})

	t.Run("Error - Marshal failure to write", func(t *testing.T) {
		taskA := NewTask(1, "Brew", 0, TimeExample)
		oldMarshal := jsonMarshal
		jsonMarshal = func(v any, prefix, indent string) ([]byte, error) {
			return nil, &json.UnsupportedTypeError{Type: nil}
		}
		defer func() {
			jsonMarshal = oldMarshal
		}()

		err := WriteToJson("test.json", taskA)
		if err == nil {
			t.Error("Expected an error for marshal failure, got nil")
		}

		var unsupportedTypeError *json.UnsupportedTypeError
		if !errors.As(err, &unsupportedTypeError) {
			t.Errorf("Expected error chain to contain json.UnsupportedTypeError, got %T", err)
		}
	})

	t.Run("Wrong file path  with correct suffix to read", func(t *testing.T) {
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

	t.Run("Wrong file extension to read", func(t *testing.T) {
		_, err := ReadFromJson("wrongFileName.txt")
		if err == nil {
			t.Errorf("Expected an error, but got nil")
			return
		}

		expectedErrMsg := "invalid file extension: wrongFileName.txt. Expected a .json file"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
		}
	})
}
