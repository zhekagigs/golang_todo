package repository

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zhekagigs/golang_todo/internal"
)

var (
	// Global test variables
	testBucketName string
	testObjectName string
	credsPath      string
)

func TestMain(m *testing.M) {
	// Setup
	log.Println("Setting up test environment...")

	// Set up GCS environment variables for tests
	testBucketName = "go-todo-app-json-storage"
	testObjectName = "test-tasks.json"

	os.Setenv("GCS_BUCKET_NAME", testBucketName)
	os.Setenv("GCS_OBJECT_NAME", testObjectName)

	// Get credentials path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get home directory: %v", err)
		os.Exit(1)
	}

	credsPath = filepath.Join(homeDir, ".config", "gcloud", "application_default_credentials.json")
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		log.Printf("Credentials file not found at %s - run 'gcloud auth application-default login' first", credsPath)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	log.Println("Cleaning up test environment...")
	// You could add cleanup code here, like deleting test files from GCS

	os.Exit(code)
}

func TestRepository(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Create repository
	repo, err := NewGCSRepository(ctx, testBucketName, testObjectName, credsPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	// Test saving tasks
	t.Run("Save and Load Tasks", func(t *testing.T) {
		// Create test tasks
		testTasks := []internal.Task{
			{
				Id:        1,
				Msg:       "Test task 1",
				Category:  internal.Brewing,
				Done:      false,
				CreatedAt: time.Now().Round(0),
				PlannedAt: time.Now().Add(24 * time.Hour).Round(0),
			},
			{
				Id:        2,
				Msg:       "Test task 2",
				Category:  internal.Marketing,
				Done:      true,
				CreatedAt: time.Now().Round(0),
				PlannedAt: time.Now().Add(48 * time.Hour).Round(0),
			},
		}

		// Save tasks
		err := repo.SaveTasks(testTasks)
		if err != nil {
			t.Fatalf("Failed to save tasks: %v", err)
		}

		// Load tasks
		loadedTasks, err := repo.LoadTasks()
		if err != nil {
			t.Fatalf("Failed to load tasks: %v", err)
		}

		// Verify loaded tasks
		if len(loadedTasks) != len(testTasks) {
			t.Errorf("Expected %d tasks, got %d", len(testTasks), len(loadedTasks))
		}

		// Compare tasks
		for i, task := range testTasks {
			if loadedTasks[i].Id != task.Id {
				t.Errorf("Task %d: expected ID %d, got %d", i, task.Id, loadedTasks[i].Id)
			}
			if loadedTasks[i].Msg != task.Msg {
				t.Errorf("Task %d: expected message %q, got %q", i, task.Msg, loadedTasks[i].Msg)
			}
			if loadedTasks[i].Category != task.Category {
				t.Errorf("Task %d: expected category %v, got %v", i, task.Category, loadedTasks[i].Category)
			}
			if loadedTasks[i].Done != task.Done {
				t.Errorf("Task %d: expected done %v, got %v", i, task.Done, loadedTasks[i].Done)
			}
		}
	})

	t.Run("Load Non-Existent File", func(t *testing.T) {
		// Create new repo with non-existent file
		nonExistentRepo, err := NewGCSRepository(ctx, "go-todo-app-json-storage", "non-existent.json", credsPath)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}
		defer nonExistentRepo.Close()

		// Try to load tasks
		tasks, err := nonExistentRepo.LoadTasks()
		if err != nil {
			t.Fatalf("Expected no error for non-existent file, got: %v", err)
		}

		// Should get empty slice, not error
		if len(tasks) != 0 {
			t.Errorf("Expected empty task list, got %d tasks", len(tasks))
		}
	})

	// Cleanup
	t.Cleanup(func() {
		// You might want to delete the test file from the bucket here
		// but be careful with cleanup in integration tests
	})
}

// TestNewGCSRepository tests the creation of a new repository
func TestNewGCSRepository(t *testing.T) {
	ctx := context.Background()
	repo, err := NewGCSRepository(ctx, testBucketName, testObjectName, credsPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	if repo.bucketName != testBucketName {
		t.Errorf("Expected bucket name 'test-bucket', got %q", repo.bucketName)
	}

	if repo.objectName != testObjectName {
		t.Errorf("Expected object name 'test.json', got %q", repo.objectName)
	}
}

// Helper function to verify the GCS configuration
func TestGetGCSConfig(t *testing.T) {
	// Save current env
	oldBucket := os.Getenv("GCS_BUCKET_NAME")
	oldObject := os.Getenv("GCS_OBJECT_NAME")
	defer func() {
		os.Setenv("GCS_BUCKET_NAME", oldBucket)
		os.Setenv("GCS_OBJECT_NAME", oldObject)
	}()

	// Test with both variables set
	os.Setenv("GCS_BUCKET_NAME", "test-bucket")
	os.Setenv("GCS_OBJECT_NAME", "test.json")

	bucket, object, err := GetGCSConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bucket != "test-bucket" {
		t.Errorf("Expected bucket 'test-bucket', got %q", bucket)
	}
	if object != "test.json" {
		t.Errorf("Expected object 'test.json', got %q", object)
	}

	// Test with neither set (should error)
	os.Unsetenv("GCS_BUCKET_NAME")
	_, _, err = GetGCSConfig()
	if err == nil {
		t.Error("Expected error when GCS_BUCKET_NAME not set")
	}
}
