package frontend

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	in "github.com/zhekagigs/golang_todo/internal"
)

//go:embed templates/*.html
var templateFiles embed.FS

type TemplateData struct {
	Tasks []in.Task
}

func generateTemplate(name string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 02, 2006 15:04")
		},
		"toLowerCase": strings.ToLower,
	}

	tmpl, err := template.New(name).Funcs(funcMap).ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func HandleTaskListRead(w http.ResponseWriter, r *http.Request, th *in.TaskHolder) {
	log.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
	tmpl, err := generateTemplate("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks := th.Read()

	data := TemplateData{
		Tasks: tasks,
	}

	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func HandleTaskListCreate(w http.ResponseWriter, r *http.Request, th *in.TaskHolder) {
	log.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
	if r.Method == http.MethodGet {
		// Render the create form
		tmpl, err := generateTemplate("create.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "create.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Post method call")
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Extract form values
		msg := r.FormValue("msg")
		category, err := strconv.Atoi(r.FormValue("category"))
		if err != nil {
			http.Error(w, "Invalid category", http.StatusBadRequest)
			return
		}
		plannedAt, err := time.Parse("2006-01-02T15:04", r.FormValue("plannedAt"))
		if err != nil {
			http.Error(w, "Invalid planned time", http.StatusBadRequest)
			return
		}

		// Create the task
		task := th.CreateTask(msg, in.TaskCategory(category), plannedAt)
		if task == nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		// Redirect to the task list page
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}

	// If neither GET nor POST, return method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func HandleTaskDelete(w http.ResponseWriter, r *http.Request, th *in.TaskHolder) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the task ID from the query parameter
	taskIDStr := r.URL.Query().Get("id")
	if taskIDStr == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Delete the task
	err = th.DeleteTask(taskID)
	if err != nil {
		if err == in.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Respond with a success status
	w.WriteHeader(http.StatusOK)
}
