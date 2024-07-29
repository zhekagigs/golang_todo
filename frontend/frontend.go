package frontend

import (
	"embed"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	in "github.com/zhekagigs/golang_todo/internal"
)

//go:embed templates/index.html
var templateFiles embed.FS

type TemplateData struct {
	Tasks []in.Task
}

func generateTemplate(indexFile embed.FS) (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 02, 2006 15:04")
		},
		"toLowerCase": strings.ToLower,
	}

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFS(indexFile, "templates/index.html")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func HandleTaskListRead(w http.ResponseWriter, r *http.Request, th *in.TaskHolder) {
	tmpl, err := generateTemplate(templateFiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks := th.Read()

	data := TemplateData{
		Tasks: tasks,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleTaskListCreate(w http.ResponseWriter, r *http.Request, th *in.TaskHolder) {
	tmpl, err := generateTemplate(templateFiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, _ := io.ReadAll(r.Body)

	task, err := in.UnmarshalTask(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	th.Add(*task)

	tasks := th.Read()
	data := TemplateData{
		Tasks: tasks,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
