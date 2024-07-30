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

type HTTPServer interface {
	ListenAndServe(addr string, handler http.Handler) error
}

type RealHTTPServer struct{}

func (s *RealHTTPServer) ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}

func main() {
	os.Exit(RealMain(internal.NewTaskHolder, &RealHTTPServer{}, &cli.RealCLIApp{}))
}

func RealMain(newTaskHolder func(diskPath string) *internal.TaskHolder, server HTTPServer, cliApp cli.CLIApp) int {
	taskHolder, checkExit, exitCode := cliApp.AppStarter(newTaskHolder)
	if checkExit {
		return exitCode
	}

	renderer, err := view.NewRenderer()
	if err != nil {
		logger.Error.Printf("error starting view renderer")
	}
	taskHandler := handlers.NewTaskHandler(taskHolder, renderer)

	go startHTTPServer(taskHandler, server)

	returnCode := cliApp.RunTaskManagmentCLI(taskHolder)
	time.Sleep(100 * time.Millisecond) // waiting for startHttpGoroutine
	return returnCode
}

func startHTTPServer(taskHandler *handlers.TaskHandler, server HTTPServer) {
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
	if err := server.ListenAndServe(":8080", nil); err != nil {
		logger.Error.Fatalf("Failed to start server: %v", err)
	}
}
