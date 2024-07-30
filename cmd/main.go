package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zhekagigs/golang_todo/cli"
	"github.com/zhekagigs/golang_todo/frontend"
	"github.com/zhekagigs/golang_todo/internal"
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

	go startHTTPServer(taskHolder, server)

	returnCode := cliApp.RunTaskManagmentCLI(taskHolder)
	time.Sleep(100 * time.Millisecond) // waiting for startHttpGoroutine
	return returnCode
}

func startHTTPServer(taskHolder *internal.TaskHolder, server HTTPServer) {
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			frontend.HandleTaskDelete(w, r, taskHolder)
		} else {
			frontend.HandleTaskListRead(w, r, taskHolder)
		}
	})

	http.HandleFunc("/tasks/create", func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling create")
		frontend.HandleTaskListCreate(w, r, taskHolder)
	})

	log.Println("Starting server on :8080")
	if err := server.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
