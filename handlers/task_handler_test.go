package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/zhekagigs/golang_todo/internal"
)

// Mock TaskHolder
type mockTaskHolder struct {
	tasks []internal.Task
}

func (m *mockTaskHolder) Read() []internal.Task {
	return m.tasks
}

func (m *mockTaskHolder) CreateTask(task *internal.TaskOptional) *internal.Task {
	newTask := &internal.Task{
		Id:        len(m.tasks) + 1,
		Msg:       *task.Msg,
		Category:  *task.Category,
		PlannedAt: *task.PlannedAt,
	}
	m.tasks = append(m.tasks, *newTask)
	return newTask
}

func (m *mockTaskHolder) FindTaskById(id int) (*internal.Task, error) {
	for _, task := range m.tasks {
		if task.Id == id {
			return &task, nil
		}
	}
	return nil, internal.ErrNotFound
}

func (m *mockTaskHolder) PartialUpdateTask(id int, update *internal.TaskOptional) error {
	for i, task := range m.tasks {
		if task.Id == id {
			if update.Msg != nil {
				m.tasks[i].Msg = *update.Msg
			}
			if update.Category != nil {
				m.tasks[i].Category = *update.Category
			}
			if update.PlannedAt != nil {
				m.tasks[i].PlannedAt = *update.PlannedAt
			}
			if update.Done != nil {
				m.tasks[i].Done = *update.Done
			}
			return nil
		}
	}
	return internal.ErrNotFound
}

func (m *mockTaskHolder) DeleteTask(id int) error {
	for i, task := range m.tasks {
		if task.Id == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return internal.ErrNotFound
}

// Mock Renderer
type mockRenderer struct {
	renderTaskListCalled   bool
	renderCreateFormCalled bool
	renderTaskUpdateCalled bool
}

func (m *mockRenderer) RenderTaskList(w http.ResponseWriter, tasks []internal.Task) error {
	m.renderTaskListCalled = true
	return nil
}

func (m *mockRenderer) RenderCreateForm(w http.ResponseWriter) error {
	m.renderCreateFormCalled = true
	return nil
}

func (m *mockRenderer) RenderTaskUpdate(w http.ResponseWriter, task *internal.Task) error {
	m.renderTaskUpdateCalled = true
	return nil
}

func TestHandleTaskListRead(t *testing.T) {
	mockService := &mockTaskHolder{
		tasks: []internal.Task{
			{Id: 1, Msg: "Task 1"},
			{Id: 2, Msg: "Task 2"},
		},
	}
	mockRenderer := &mockRenderer{}
	handler := NewTaskHandler(mockService, mockRenderer)

	req, err := http.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleTaskListRead(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !mockRenderer.renderTaskListCalled {
		t.Errorf("RenderTaskList was not called")
	}
}

func TestHandleTaskCreate(t *testing.T) {
	mockService := &mockTaskHolder{}
	mockRenderer := &mockRenderer{}
	handler := NewTaskHandler(mockService, mockRenderer)

	t.Run("GET request", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/tasks/create", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.HandleTaskCreate(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		if !mockRenderer.renderCreateFormCalled {
			t.Errorf("RenderCreateForm was not called")
		}
	})

	t.Run("POST request", func(t *testing.T) {
		form := url.Values{}
		form.Add("msg", "New Task")
		form.Add("category", "1")
		form.Add("plannedAt", time.Now().Format("2006-01-02T15:04"))

		req, err := http.NewRequest("POST", "/tasks/create", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.HandleTaskCreate(rr, req)

		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
		}

		if len(mockService.tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(mockService.tasks))
		}
	})
}

func TestHandleTaskUpdate(t *testing.T) {
	mockService := &mockTaskHolder{
		tasks: []internal.Task{
			{Id: 1, Msg: "Task 1"},
		},
	}
	mockRenderer := &mockRenderer{}
	handler := NewTaskHandler(mockService, mockRenderer)

	t.Run("GET request", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/tasks/update?id=1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.HandleTaskUpdate(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		if !mockRenderer.renderTaskUpdateCalled {
			t.Errorf("RenderTaskUpdate was not called")
		}
	})

	t.Run("POST request", func(t *testing.T) {
		form := url.Values{}
		form.Add("msg", "Updated Task")
		form.Add("category", "2")
		form.Add("plannedAt", time.Now().Format("2006-01-02T15:04"))

		req, err := http.NewRequest("POST", "/tasks/update?id=1", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.HandleTaskUpdate(rr, req)

		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
		}

		updatedTask, _ := mockService.FindTaskById(1)
		if updatedTask.Msg != "Updated Task" {
			t.Errorf("Expected task message to be 'Updated Task', got '%s'", updatedTask.Msg)
		}
	})
}

func TestHandleTaskDelete(t *testing.T) {
	mockService := &mockTaskHolder{
		tasks: []internal.Task{
			{Id: 1, Msg: "Task 1"},
		},
	}
	mockRenderer := &mockRenderer{}
	handler := NewTaskHandler(mockService, mockRenderer)

	req, err := http.NewRequest("DELETE", "/tasks/delete?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleTaskDelete(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if len(mockService.tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(mockService.tasks))
	}
}

func TestExtractFormValues(t *testing.T) {
	form := url.Values{}
	form.Add("msg", "Test Task")
	form.Add("category", "1")
	form.Add("plannedAt", "2023-05-01T10:00")
	form.Add("done", "true")

	req, err := http.NewRequest("POST", "/tasks", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	taskOptional, err := ExtractFormValues(req)
	if err != nil {
		t.Fatalf("ExtractFormValues failed: %v", err)
	}

	if *taskOptional.Msg != "Test Task" {
		t.Errorf("Expected msg 'Test Task', got '%s'", *taskOptional.Msg)
	}

	if *taskOptional.Category != internal.TaskCategory(1) {
		t.Errorf("Expected category 1, got %d", *taskOptional.Category)
	}

	expectedTime, _ := time.Parse("2006-01-02T15:04", "2023-05-01T10:00")
	if !taskOptional.PlannedAt.Equal(expectedTime) {
		t.Errorf("Expected plannedAt %v, got %v", expectedTime, *taskOptional.PlannedAt)
	}

	if !*taskOptional.Done {
		t.Errorf("Expected done to be true")
	}
}
