package view

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zhekagigs/golang_todo/internal"
)

func TestNewRenderer(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}
	if renderer == nil {
		t.Fatal("NewRenderer() returned nil renderer")
	}
	if renderer.templates == nil {
		t.Fatal("NewRenderer() returned renderer with nil templates")
	}
}

func TestTaskRenderer_RenderTaskList(t *testing.T) {
	renderer, _ := NewRenderer()
	tasks := []internal.Task{
		{Id: 1, Msg: "Task 1", Category: internal.Brewing, PlannedAt: time.Now()},
		{Id: 2, Msg: "Task 2", Category: internal.Marketing, PlannedAt: time.Now().Add(24 * time.Hour)},
	}

	w := httptest.NewRecorder()
	err := renderer.RenderTaskList(w, tasks)
	if err != nil {
		t.Fatalf("RenderTaskList() error = %v", err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("RenderTaskList() status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	for _, task := range tasks {
		if !strings.Contains(body, task.Msg) {
			t.Errorf("RenderTaskList() body doesn't contain task message: %v", task.Msg)
		}
	}
}

func TestTaskRenderer_RenderCreateForm(t *testing.T) {
	renderer, _ := NewRenderer()

	w := httptest.NewRecorder()
	err := renderer.RenderCreateForm(w)
	if err != nil {
		t.Fatalf("RenderCreateForm() error = %v", err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("RenderCreateForm() status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<form") || !strings.Contains(body, "create") {
		t.Errorf("RenderCreateForm() body doesn't contain expected form elements")
	}
}

func TestTaskRenderer_RenderTaskUpdate(t *testing.T) {
	renderer, _ := NewRenderer()
	task := &internal.Task{
		Id:        1,
		Msg:       "Update this task",
		Category:  internal.Brewing,
		PlannedAt: time.Now(),
	}

	w := httptest.NewRecorder()
	err := renderer.RenderTaskUpdate(w, task)
	if err != nil {
		t.Fatalf("RenderTaskUpdate() error = %v", err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("RenderTaskUpdate() status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, task.Msg) || !strings.Contains(body, "update") {
		t.Errorf("RenderTaskUpdate() body doesn't contain expected task details or update form")
	}
}

func TestRenderErrCheck(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"No error", nil, false},
		{"With error", fmt.Errorf("template error: %v", template.ErrNoSuchTemplate), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := renderErrCheck(tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderErrCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "error rendering template") {
				t.Errorf("renderErrCheck() error message doesn't contain expected prefix")
			}
		})
	}
}
