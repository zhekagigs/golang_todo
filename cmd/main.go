package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zhekagigs/golang_todo/cli"
	"github.com/zhekagigs/golang_todo/handlers"
	"github.com/zhekagigs/golang_todo/internal"
	"github.com/zhekagigs/golang_todo/logger"
	"github.com/zhekagigs/golang_todo/view"
)

func main() {
	os.Exit(RealMain(internal.NewTaskHolder, &handlers.RealHTTPServer{}, &cli.RealCLIApp{}))
}

func RealMain(newTaskHolder func(diskPath string) *internal.TaskHolder, server handlers.HTTPServer, cliApp cli.CLIApp) int {
	taskHolder, checkExit, exitCode := cliApp.AppStarter(newTaskHolder)
	if checkExit {
		return exitCode
	}

	renderer, err := view.NewRenderer()
	if err != nil {
		logger.Error.Printf("error starting view renderer")
	}
	taskConcurrentService := internal.NewConcurrentTaskService(taskHolder)
	taskRenderHandler := handlers.NewTaskRenderHandler(taskHolder, renderer)

	api := handlers.NewApiService(taskConcurrentService)

	go startHTTPServer(taskRenderHandler, server, api)

	returnCode := cliApp.RunTaskManagmentCLI(taskHolder)
	time.Sleep(100 * time.Millisecond) // waiting for startHttpGoroutine
	return returnCode
}

func startHTTPServer(taskHandler *handlers.TaskRenderHandler, server handlers.HTTPServer, api *handlers.ApiService) {

	router := http.NewServeMux()
	router.Handle("/api/", api) //why?
	router.HandleFunc("GET /api/tasks", api.GetAllPosts)
	router.HandleFunc("GET /api/tasks/{id}", api.GetTaskById)
	router.HandleFunc("POST /api/tasks", api.CreateTask)

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		logger.Info.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			taskHandler.HandleTaskListRead(w, r)
		case http.MethodDelete:
			taskHandler.HandleTaskDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tasks/update", func(w http.ResponseWriter, r *http.Request) {
		logger.Info.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
		taskHandler.HandleTaskUpdate(w, r)
	})

	http.HandleFunc("/tasks/create", func(w http.ResponseWriter, r *http.Request) {
		logger.Info.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
		taskHandler.HandleTaskCreate(w, r)
	})

	log.Println("Starting server on :8080")
	if err := server.ListenAndServe(":8080", router); err != nil {
		logger.Error.Fatalf("Failed to start server: %v", err)
	}
}

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		crw := &customResponseWriter{ResponseWriter: w}

		next.ServeHTTP(crw, r)

		duration := time.Since(start)

		log.Printf(
			"Method: %s, Path: %s, Status: %d, Duration: %v",
			r.Method,
			r.URL.Path,
			crw.status,
			duration,
		)
	}
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
}
