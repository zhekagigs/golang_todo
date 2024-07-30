package view

import (
	"embed"
	"fmt"
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

		update, err := ExtractFormValues(r)

		if err != nil {
			http.Error(w, "Invalid category", http.StatusBadRequest)
			return
		}

		// Create the task
		task := th.CreateTask(update)
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

type FormAdapter struct {
}

func ExtractFormValues(r *http.Request) (*in.TaskOptional, error) {
	msg := r.FormValue("msg")
	var errs []error

	addErr := func(err error, msg string) {
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", msg, err))
		}
	}

	checkErr := func(err error, msg string) {
		if err != nil {
			addErr(err, msg)
		}
	}

	category, err := strconv.Atoi(r.FormValue("category"))
	checkErr(err, "Invalid category")

	plannedAt, err := time.Parse("2006-01-02T15:04", r.FormValue("plannedAt"))
	checkErr(err, "invalid planned time")

	done, err := strconv.ParseBool(r.FormValue("done"))
	checkErr(err, "invalid done value")

	if len(errs) > 0 {
		return nil, fmt.Errorf("form value errors: %v", errs)
	}

	update := &in.TaskOptional{
		Done:      &done,
		Msg:       in.StringPtr(msg),
		Category:  (*in.TaskCategory)(&category),
		PlannedAt: &plannedAt,
	}

	return update, nil
}

func HandleTaskUpdate(w http.ResponseWriter, r *http.Request, th *in.TaskHolder) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	taskIDStr := r.URL.Query().Get("id")
	if taskIDStr == "" {
		http.Error(w, "Missing Task Id", http.StatusBadRequest)
		return
	}
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	update, err := ExtractFormValues(r)
	if err != nil {
		http.Error(w, "Error processing form values", http.StatusBadRequest)
	}

	err = th.PartialUpdateTask(taskID, update)
	if err != nil {
		http.Error(w, "Error updating task", http.StatusInternalServerError)
	}

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
