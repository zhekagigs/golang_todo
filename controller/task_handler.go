package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/zhekagigs/golang_todo/internal"
	"github.com/zhekagigs/golang_todo/logger"
	"github.com/zhekagigs/golang_todo/view"
)

type TaskRenderHandler struct {
	service  internal.TaskServiceInterface
	renderer view.Renderer
}

func handleError(w http.ResponseWriter, err error, status int, message string) bool {
	if err != nil {
		if err == internal.ErrNotFound {
			logger.Error.Printf("Task not found: %s", message)
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			logger.Error.Printf("%s: %v", message, err)
			if message == "" {
				message = err.Error()
			}
			http.Error(w, message, status)
		}

		return true
	}
	return false
}

func getTaskIdFromQuery(r *http.Request) (int, error) {
	taskIDStr := r.URL.Query().Get("id")
	if taskIDStr == "" {
		return 0, fmt.Errorf("missing Task ID")
	}
	return strconv.Atoi(taskIDStr)
}

func NewTaskRenderHandler(service internal.TaskServiceInterface, renderer view.Renderer) *TaskRenderHandler {
	return &TaskRenderHandler{service: service, renderer: renderer}
}

func (h *TaskRenderHandler) HandleTaskListRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := h.service.Read()

	err := h.renderer.RenderTaskList(w, tasks)
	if handleError(w, err, http.StatusInternalServerError, "") {
		return
	}
}

func (h *TaskRenderHandler) HandleTaskCreate(w http.ResponseWriter, r *http.Request) {
	err := h.renderer.RenderCreateForm(w)
	handleError(w, err, http.StatusInternalServerError, "")
	return
}

func (h *TaskRenderHandler) HandleTaskUpdate(w http.ResponseWriter, r *http.Request) {
	// logger.Info.Printf("Handling %s request for task update from %s", r.Method, r.RemoteAddr)

	taskID, err := getTaskIdFromQuery(r)
	if handleError(w, err, http.StatusBadRequest, "Invalid task ID") {
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.handleGetTaskUpdate(w, taskID)
	case http.MethodPost:
		h.handlePostTaskUpdate(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TaskRenderHandler) handleGetTaskUpdate(w http.ResponseWriter, taskID int) {
	task, err := h.service.FindTaskById(taskID)
	if handleError(w, err, http.StatusNotFound, "Task not found") {
		return
	}

	err = h.renderer.RenderTaskUpdate(w, task)
	handleError(w, err, http.StatusInternalServerError, "Error rendering update form")
}

func (h *TaskRenderHandler) handlePostTaskUpdate(w http.ResponseWriter, r *http.Request, taskID int) {
	update, err := ExtractFormValues(r)
	if handleError(w, err, http.StatusBadRequest, "Invalid form data") {
		return
	}

	// logger.Info.Printf("Updating task with ID: %d", taskID)
	err = h.service.PartialUpdateTask(taskID, update)
	if handleError(w, err, http.StatusInternalServerError, "Failed to update task") {
		return
	}

	// logger.Info.Printf("Successfully updated task with ID: %d", taskID)
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *TaskRenderHandler) HandleTaskDelete(w http.ResponseWriter, r *http.Request) {
	// logger.Info.Printf("Handling %s request for task deletion from %s", r.Method, r.RemoteAddr)
	if r.Method != http.MethodDelete {
		logger.Error.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	taskID, err := getTaskIdFromQuery(r)
	err = h.service.DeleteTask(taskID)
	if handleError(w, err, http.StatusInternalServerError, fmt.Sprint(taskID)) {
		return
	}

	// logger.Info.Printf("Successfully deleted task with ID: %d", taskID)
	w.WriteHeader(http.StatusOK)
}

func ExtractFormValues(r *http.Request) (*internal.TaskOptional, error) {
	// logger.Info.Println("Extracting form values")

	var errs []error
	addErr := func(err error, msg string) {
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", msg, err))
			logger.Error.Printf("%s: %v", msg, err)
		}
	}
	checkErr := func(err error, msg string) {
		if err != nil {
			addErr(err, msg)
		}
	}
	msg := r.FormValue("msg")

	category, err := strconv.Atoi(r.FormValue("category"))
	checkErr(err, "Invalid category")

	dateString := r.FormValue("plannedAt")
	var plannedAt *internal.CustomTime
	if dateString != "" {
		plannedData, err := time.Parse("2006-01-02T15:04", dateString)
		plannedAt = internal.TimePtr(plannedData)
		checkErr(err, "invalid planned time")
	} else {
		plannedAt = nil
	}

	var done *bool
	doneValue := r.FormValue("done")
	if doneValue != "" {
		doneBool, err := strconv.ParseBool(doneValue)
		done = internal.BoolPtr(doneBool)
		checkErr(err, "invalid done value")
	} else {
		done = nil
	}

	if len(errs) > 0 {
		logger.Error.Printf("Form value errors: %v", errs)
		return nil, fmt.Errorf("form value errors: %v", errs)
	}
	update := &internal.TaskOptional{
		Done:      done,
		Msg:       internal.StringPtr(msg),
		Category:  (*internal.TaskCategory)(&category),
		PlannedAt: plannedAt,
	}
	return update, nil
}
