package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

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
	// Initialize with environment variables or defaults
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	usersFile := os.Getenv("USERS_FILE")
	if usersFile == "" {
		usersFile = "users.json"
	}

	taskHolder, checkExit, exitCode, isWeb := cliApp.AppStarter(newTaskHolder)
	if checkExit {
		return exitCode
	}

	renderer, err := view.NewRenderer()
	if err != nil {
		logger.Error.Printf("error starting view renderer")
		return cli.ExitCodeError
	}

	taskConcurrentService := internal.NewConcurrentTaskService(taskHolder)
	taskRenderHandler := controller.NewTaskRenderHandler(taskHolder, renderer)
	userStore, err := users.NewUserStore(usersFile)
	if err != nil {
		logger.Error.Printf("error loading user store file")
		return cli.ExitCodeError
	}

	api := controller.NewApiService(taskConcurrentService, userStore)
	authHandler := controller.NewAuthHandler(userStore)

	// Setup shutdown channel
	shutdownChan := make(chan struct{})
	errChan := make(chan error, 1)
	// Start HTTP server in goroutine
	go func() {
		if err := startHTTPServer(port, taskRenderHandler, server, api, authHandler); err != nil {
			logger.Error.Printf("Failed to start server: %v", err)
			errChan <- err
		}
	}()

	// Handle graceful shutdown
	go handleShutdown(shutdownChan)

	// Run service
	if isWeb {
		logger.Info.Println("web flag provide")
		select {
		case err := <-errChan:
			logger.Error.Println(err)
			return cli.ExitCodeError
		case <-shutdownChan:
			return cli.ExitCodeSuccess
		}
	} else {
		logger.Info.Println("web flag not provided, cli and web app started")
		cliDoneChan := make(chan int)

		go func() {
			returnCode := cliApp.RunTaskManagmentCLI(taskHolder)
			cliDoneChan <- returnCode
		}()
		select {
		case returnCode := <-cliDoneChan:
			close(shutdownChan)
			return returnCode
		case err := <-errChan:
			logger.Error.Println(err)
			return cli.ExitCodeError
		case <-shutdownChan:
			return cli.ExitCodeSuccess
		}
	}
}

func startHTTPServer(port string, taskHandler *controller.TaskRenderHandler, server controller.HTTPServer, api *controller.ApiService, authHandler *controller.AuthHandler) error {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/tasks", api.GetAllPosts)
	router.HandleFunc("GET /api/tasks/{id}", api.GetTaskById)
	router.HandleFunc("POST /api/tasks", mid.AuthMiddleware(api.CreateTask))
	router.HandleFunc("PUT /api/tasks/{id}", mid.AuthMiddleware(api.UpdateTask))
	router.HandleFunc("DELETE /api/tasks/{id}", mid.AuthMiddleware(api.DeleteTask))

	router.HandleFunc("POST /login", authHandler.LoginHandler)
	router.HandleFunc("POST /logout", authHandler.LogoutHandler)

	router.HandleFunc("GET /tasks/create", mid.AuthMiddleware(taskHandler.HandleTaskCreate))
	router.HandleFunc("GET /tasks", taskHandler.HandleTaskListRead)
	router.HandleFunc("DELETE /tasks", mid.AuthMiddleware(taskHandler.HandleTaskDelete))
	router.HandleFunc("GET /tasks/update", mid.AuthMiddleware(taskHandler.HandleTaskUpdate))
	router.HandleFunc("POST /tasks/update", mid.AuthMiddleware(taskHandler.HandleTaskUpdate))

	router.HandleFunc("GET /", taskHandler.HandleTaskListRead)

	// Health check
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggingHandler := mid.LoggingMiddleware{Next: router}

	logger.Info.Printf("Starting server on :%s", port)
	return server.ListenAndServe(":"+port, loggingHandler)
}

func handleShutdown(done chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigChan:
		logger.Info.Println("Received shutdown signal")
		done <- struct{}{}
	case <-done:
		logger.Info.Println("Server stopped")
	}
}
