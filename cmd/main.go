package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/zhekagigs/golang_todo/cli"
	"github.com/zhekagigs/golang_todo/controller"
	"github.com/zhekagigs/golang_todo/internal"
	"github.com/zhekagigs/golang_todo/logger"
	mid "github.com/zhekagigs/golang_todo/middleware"
	"github.com/zhekagigs/golang_todo/repository"
	"github.com/zhekagigs/golang_todo/users"
	"github.com/zhekagigs/golang_todo/view"
)

func main() {
	err, repo := configureRepo()
	if err != nil {
		logger.Error.Printf("Failed to create repository: %v", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Load initial tasks from GCS
	taskHolder := loadTasks(repo)

	// Start the application if GCP
	os.Exit(RealMain(
		func(string) *internal.TaskHolder { return taskHolder },
		&controller.RealHTTPServer{},
		&cli.RealCLIApp{},
	))

	// Start app if local JSON storage
	// os.Exit(RealMain(internal.NewTaskHolder, &controller.RealHTTPServer{}, &cli.RealCLIApp{}))
}

func loadTasks(repo *repository.GCSRepository) *internal.TaskHolder {
	tasks, err := repo.LoadTasks()
	if err != nil {
		logger.Error.Printf("Failed to load tasks: %v", err)
		os.Exit(1)
	}

	taskHolder := internal.NewTaskHolder("")
	for _, task := range tasks {
		taskHolder.Add(task)
	}
	return taskHolder
}

// Get GCS configuration.
// Get credentials path.
// Initialize GCS repository.
func configureRepo() (error, *repository.GCSRepository) {
	ctx := context.Background()
	// Set up GCS environment variables for tests
	bucketName := "go-todo-app-json-storage"
	objectName := "test-tasks.json"

	os.Setenv("GCS_BUCKET_NAME", bucketName)
	os.Setenv("GCS_OBJECT_NAME", objectName)

	bucketName, objectName, err := repository.GetGCSConfig()
	if err != nil {
		logger.Error.Printf("Failed to get GCS config: %v", err)
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error.Printf("Failed to get home directory: %v", err)
		os.Exit(1)
	}

	credsPath := filepath.Join(homeDir, ".config", "gcloud", "application_default_credentials.json")
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		log.Printf("Credentials file not found at %s - run 'gcloud auth application-default login' first", credsPath)
		os.Exit(1)
	}

	repo, err := repository.NewGCSRepository(ctx, bucketName, objectName, credsPath)
	return err, repo
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
