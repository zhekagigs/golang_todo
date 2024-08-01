package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/zhekagigs/golang_todo/internal"
	"github.com/zhekagigs/golang_todo/logger"
)

var (
	ErrWrongRequest = errors.New("wrong request")
	ErrInternal     = errors.New("server error")
)

type ApiService struct {
	taskHolder *internal.TaskHolder
}

func (apiHandler ApiService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Info.Printf("ServeHttp Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
}

func NewApiService(internal *internal.TaskHolder) *ApiService {
	return &ApiService{
		taskHolder: internal,
	}
}

func (api *ApiService) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts := api.taskHolder.Read()

	postsJson, err := json.Marshal(posts)
	if isJsonErr(err, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(postsJson)
}

func (api *ApiService) GetTaskById(w http.ResponseWriter, r *http.Request) {
	taskId, err := getTaskIdFromPath(r)
	if handleError(w, err, http.StatusInternalServerError, "api: error processing taskId") {
		return
	}
	task, err := api.taskHolder.FindTaskById(taskId)
	if handleError(w, err, http.StatusInternalServerError, "api: task not found") {
		return
	}

	taskJson, err := json.Marshal(task)
	if isJsonErr(err, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader((http.StatusOK))
	w.Write(taskJson)
}

func (api *ApiService) CreateTask(w http.ResponseWriter, r *http.Request) {
	var taskRequest *internal.TaskOptional
	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if handleError(w, err, http.StatusBadRequest, "error decoding request body") {
		return
	}

	task := api.taskHolder.CreateTask(taskRequest)
	taskAsJson, err := json.Marshal(task)
	if err != nil {
		http.Error(w, ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(taskAsJson)
}

func (api *ApiService) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var taskRequest *internal.TaskOptional
	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if handleError(w, err, http.StatusBadRequest, "error decoding request body") {
		return
	}
	taskId, err := getTaskIdFromPath(r)
	if handleError(w, err, http.StatusBadRequest, "error parsing taskId") {
		return
	}
	task := api.taskHolder.PartialUpdateTask(taskId, taskRequest)
	taskAsJson, err := json.Marshal(task)
	if err != nil {
		http.Error(w, ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(taskAsJson)
}

func (api *ApiService) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskId, err := getTaskIdFromPath(r)
	if handleError(w, err, http.StatusBadRequest, "api: error processing taskId") {
		return
	}

	err = api.taskHolder.DeleteTask(taskId)
	if handleError(w, err, http.StatusBadRequest, "api: task not foound") {
		return
	}
	w.WriteHeader((http.StatusOK))
}

func isJsonErr(err error, w http.ResponseWriter) bool {
	return handleError(w, err, http.StatusBadRequest, "api: json serialization error")
}

func getTaskIdFromPath(r *http.Request) (int, error) {
	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		return -1, errors.New("task id is empty")
	}
	return strconv.Atoi(taskIdStr)
}
