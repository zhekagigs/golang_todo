package main

import (
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/zhekagigs/golang_todo/cli"
	"github.com/zhekagigs/golang_todo/controller"
	"github.com/zhekagigs/golang_todo/internal"
	"github.com/zhekagigs/golang_todo/logger"
	mid "github.com/zhekagigs/golang_todo/middleware"
	"github.com/zhekagigs/golang_todo/users"
	"github.com/zhekagigs/golang_todo/view"
)

func main() {
	os.Exit(RealMain(internal.NewTaskHolder, &controller.RealHTTPServer{}, &cli.RealCLIApp{}))
}

func RealMain(newTaskHolder func(diskPath string) *internal.TaskHolder, server controller.HTTPServer, cliApp cli.CLIApp) int {
	// ctx := context.Background()
	taskHolder, checkExit, exitCode := cliApp.AppStarter(newTaskHolder)
	if checkExit {
		return exitCode
	}

	renderer, err := view.NewRenderer()
	if err != nil {
		logger.Error.Printf("error starting view renderer")
	}
	taskConcurrentService := internal.NewConcurrentTaskService(taskHolder)
	taskRenderHandler := controller.NewTaskRenderHandler(taskHolder, renderer)
	userStore, err := users.NewUserStore("users.json")

	api := controller.NewApiService(taskConcurrentService, userStore)
	if err != nil {
		logger.Error.Printf("error loading user store file")
		return cli.ExitCodeError
	}
	authHandler := controller.NewAuthHandler(userStore)
	go startHTTPServer(taskRenderHandler, server, api, authHandler)

	returnCode := cliApp.RunTaskManagmentCLI(taskHolder)
	time.Sleep(100 * time.Millisecond) // waiting for startHttpGoroutine
	return returnCode
}

func startHTTPServer(taskHandler *controller.TaskRenderHandler, server controller.HTTPServer, api *controller.ApiService, authHandler *controller.AuthHandler) {

	router := http.NewServeMux()

	router.Handle("/api/", api) //why?
	router.HandleFunc("GET /api/tasks", (api.GetAllPosts))
	router.HandleFunc("GET /api/tasks/{id}", (api.GetTaskById))
	router.HandleFunc("POST /api/tasks", mid.AuthMiddleware(api.CreateTask))
	router.HandleFunc("PUT /api/tasks/{id}", mid.AuthMiddleware(api.UpdateTask))
	router.HandleFunc("DELETE /api/tasks/{id}", mid.AuthMiddleware(api.DeleteTask))

	router.HandleFunc("POST /login", (authHandler.LoginHandler))
	router.HandleFunc("/logout", authHandler.LogoutHandler)
	router.HandleFunc("GET /tasks/create", (mid.AuthMiddleware(taskHandler.HandleTaskCreate)))
	router.HandleFunc("GET /tasks", (taskHandler.HandleTaskListRead))
	router.HandleFunc("DELETE /tasks", (mid.AuthMiddleware(taskHandler.HandleTaskDelete)))
	router.HandleFunc("/tasks/update", (mid.AuthMiddleware(taskHandler.HandleTaskUpdate)))

	loggingHandler := mid.LoggingMiddleware{Next: router}

	logger.Info.Println("Starting server on :8080")
	if err := server.ListenAndServe(":8080", loggingHandler); err != nil {
		logger.Error.Fatalf("Failed to start server: %v", err)
	}
	// defer server.Shutdown()
}
