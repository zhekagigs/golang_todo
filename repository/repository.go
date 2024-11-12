package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/zhekagigs/golang_todo/internal"
	"google.golang.org/api/option"
)

const (
	EnvGoogleAPIKey = "GOOGLE_API_KEY"
	EnvBucketName   = "GCS_BUCKET_NAME"
)

type TaskRepository interface {
	SaveTasks(tasks []internal.Task) error
	LoadTasks() ([]internal.Task, error)
}

type GCSRepository struct {
	bucketName   string
	objectName   string
	client       *storage.Client
	clientCtx    context.Context
	credentials  string
}

type GCSConfig struct {
	BucketName string
	ObjectName string
	APIKey string
}

func NewGCSRepository(ctx context.Context, bucketName, objectName, credentialsFile string) (*GCSRepository, error) {

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return &GCSRepository{
		bucketName:   bucketName,
		objectName:   objectName,
		client:       client,
		clientCtx:    ctx,
		credentials:  credentialsFile,
	}, nil
}

func (r * GCSRepository) Close() error {
	return r.client.Close()
}

// SaveTasks saves tasks to a JSON file in GCS bucket
func (r *GCSRepository) SaveTasks(tasks []internal.Task) error {
	bucket := r.client.Bucket(r.bucketName)
	obj := bucket.Object(r.objectName)

	// Create a new context with timeout for the upload operation
	ctx, cancel := context.WithTimeout(r.clientCtx, time.Minute)
	defer cancel()

	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/json"

	// Marshal tasks to JSON
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %v", err)
	}

	// Write data
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to GCS: %v", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	return nil
}

// LoadTasks loads tasks from a JSON file in GCS bucket
func (r *GCSRepository) LoadTasks() ([]internal.Task, error) {
	bucket := r.client.Bucket(r.bucketName)
	obj := bucket.Object(r.objectName)

	ctx, cancel := context.WithTimeout(r.clientCtx, time.Minute)
	defer cancel()

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			// If the file doesn't exist, return an empty task list
			return []internal.Task{}, nil
		}
		return nil, fmt.Errorf("failed to create reader: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %v", err)
	}

	var tasks []internal.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %v", err)
	}

	return tasks, nil
}


// GetGCSConfig returns bucket configuration from environment variables
func GetGCSConfig() (string, string, error) {
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return "", "", fmt.Errorf("GCS_BUCKET_NAME environment variable not set")
	}

	objectName := os.Getenv("GCS_OBJECT_NAME")
	if objectName == "" {
		objectName = "tasks.json" // Default value
	}

	return bucketName, objectName, nil
}
