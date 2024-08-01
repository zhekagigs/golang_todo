package view

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/zhekagigs/golang_todo/internal"
	"github.com/zhekagigs/golang_todo/logger"
)

//go:embed templates/*.html
var templateFiles embed.FS

// type TemplateData struct {
// 	Tasks []in.Task
// }

type Renderer interface {
	RenderTaskList(http.ResponseWriter, []internal.Task) error
	RenderCreateForm(http.ResponseWriter) error
	RenderTaskUpdate(http.ResponseWriter, *internal.Task) error
}

type TaskRenderer struct {
	templates *template.Template
}

type TaskListData struct {
	Tasks []internal.Task
}

func renderErrCheck(err error) error {
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}
	return nil
}

func NewRenderer() (*TaskRenderer, error) {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 02, 2006 15:04")
		},
		"toLowerCase": strings.ToLower,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &TaskRenderer{templates: tmpl}, nil
}

func (r *TaskRenderer) RenderTaskList(w http.ResponseWriter, tasks []internal.Task) error {
	logger.Info.Printf("Rendering Task List")
	data := TaskListData{
		Tasks: tasks,
	}

	err := r.templates.ExecuteTemplate(w, "index.html", data)
	return renderErrCheck(err)
}

func (r *TaskRenderer) RenderCreateForm(w http.ResponseWriter) error {
	err := r.templates.ExecuteTemplate(w, "create.html", nil)
	return renderErrCheck(err)
}

func (r *TaskRenderer) RenderTaskUpdate(w http.ResponseWriter, task *internal.Task) error {
	logger.Info.Println("Rendering update task form")
	data := struct {
		Task *internal.Task
	}{
		Task: task,
	}
	err := r.templates.ExecuteTemplate(w, "update.html", data)
	return renderErrCheck(err)
}
