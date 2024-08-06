package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/zhekagigs/golang_todo/internal"
)

func setupConfig(t *testing.T) (*httptest.Server, *internal.TaskHolder) {
	// Setup
	taskHolder := internal.NewTaskHolder("resources/concurrent_disk.json")
	// taskChan := make(chan internal.Task)
	// optionalsChan := make(chan internal.TaskOptional)
	taskService := internal.NewConcurrentTaskService(taskHolder)
	t.Logf("loaded %d Tasks", len(taskHolder.Tasks))
	apiService := NewApiService(taskService)

	// Create a test server

	server := httptest.NewServer(http.HandlerFunc(apiService.CreateTask))
	t.Cleanup(func() {
		internal.WriteToJson(taskHolder.DiskPath, taskHolder.Tasks...)
		taskService.CloseAll()
		server.Close()
	})

	return server, taskHolder
}

func TestCreateTaskIntegration(t *testing.T) {
	server, taskHolder := setupConfig(t)
	// Prepare test data
	testTask, payload := provideTestData("Test msg")

	// Send request -- caller code
	resp, _ := postRequest(server.URL, payload)

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", resp.Status)
	}

	// Parse response body
	createdTask := parseResponse(resp, t)

	// Verify the created task
	assertTaskFields(createdTask, testTask, t)

	// Verify task was actually added to the TaskHolder
	verifyTaskInHolder(taskHolder, createdTask, t)
}

func TestManyCreateTaskIntegrationSequentially(t *testing.T) {
	t.Log("Starting TestManyCreateTaskIntegration")
	NUM := 50000
	server, taskHolder := setupConfig(t)
	tasks := internal.GenerateRandomTasks(NUM)
	for _, task := range tasks {
		testTask := internal.AdapterTaskOptional(task)
		payload, err := provideJsonBody(testTask)
		if err != nil {
			panic(err)
		}
		// Send request -- caller code
		resp, _ := postRequest(server.URL, payload)

		// Check response status
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", resp.Status)
		}

		// Parse response body
		createdTask := parseResponse(resp, t)

		// Verify the created task
		assertTaskFields(createdTask, testTask, t)

		// // Verify task was actually added to the TaskHolder
		// verifyTaskInHolder(taskHolder, createdTask, t)
	}
	latestid, total := taskHolder.Count()
	t.Logf("taskHolder: latestId:%d total:%d\n", latestid, total)
	if latestid != NUM {
		t.Errorf("want %d, got %d", NUM, latestid)
	}
	if total != NUM {
		t.Errorf("want %d, got %d", NUM, total)
	}
	// if taskHolder.
}

func TestMultipleClientsPostRequest(t *testing.T) {
	// t.Skip("fails on purpose")
	server, taskHolder := setupConfig(t)
	numClients := 100
	numRequestsPerClient := 30

	var wg sync.WaitGroup
	results := make(chan string, numClients*numRequestsPerClient)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			for j := 0; j < numRequestsPerClient; j++ {
				tasks := internal.GenerateRandomTasks(1)
				taskOptional := internal.AdapterTaskOptional(tasks[0])
				payload, err := provideJsonBody(taskOptional)
				if err != nil {
					t.Errorf("error marshalling task to json")
				}

				resp, err := postRequest(server.URL, payload)
				if err != nil {
					t.Errorf("error posting task to server %q", err)
					panic(err)
				}
				if resp.StatusCode != http.StatusCreated {
					results <- fmt.Sprintf("Client %d, Request %d: Expected status Created, got %v", clientID, j, resp.Status)
				}
				// Parse response body
				createdTask := parseResponse(resp, t)

				// Verify the created task
				assertTaskFields(createdTask, taskOptional, t)
			}
		}(i)
	}

	// Close the results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	for result := range results {
		t.Log(result)
	}

	latestID, total := taskHolder.Count()
	t.Logf("taskHolder: latestId:%d total:%d", latestID, total)

	if total != numClients*numRequestsPerClient {
		t.Errorf("Expected %d tasks, got %d", numClients*numRequestsPerClient, total)
	}
}

func client(serverAddr string, taskCh <-chan internal.Task) {
	testTask := internal.AdapterTaskOptional(<-taskCh)
	payload, _ := provideJsonBody(testTask)

	resp, _ := postRequest(serverAddr, payload)

	if resp.StatusCode != http.StatusCreated {
		panic("Expected status Created")
	}
}

func parseResponse(resp *http.Response, t *testing.T) internal.Task {
	var createdTask internal.Task
	err := json.NewDecoder(resp.Body).Decode(&createdTask)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	return createdTask
}

func verifyTaskInHolder(taskHolder *internal.TaskHolder, createdTask internal.Task, t *testing.T) {
	tasks := taskHolder.Read()
	found := false
	for _, task := range tasks {
		if task.Id == createdTask.Id {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created task not found in TaskHolder")
	}
}

func assertTaskFields(createdTask internal.Task, testTask internal.TaskOptional, t *testing.T) {
	if createdTask.Msg != *testTask.Msg {
		t.Errorf("Expected task message %s, got %s", *testTask.Msg, createdTask.Msg)
	}
	if createdTask.Category != *testTask.Category {
		t.Errorf("Expected category %v, got %v", *testTask.Category, createdTask.Category)
	}
	if createdTask.Done != *testTask.Done {
		t.Errorf("Expected done status %v, got %v", *testTask.Done, createdTask.Done)
	}
	if !createdTask.PlannedAt.Equal(testTask.PlannedAt.Time.Truncate(time.Second)) {
		t.Errorf("Expected planned time %v, got %v", testTask.PlannedAt.Time.Truncate(time.Second), createdTask.PlannedAt)
	}
	if createdTask.Id == 0 {
		t.Error("Expected non-zero ID for created task")
	}
}

func postRequest(server string, payload []byte) (*http.Response, error) {
	resp, err := http.Post(server, "application/json", bytes.NewBuffer(payload))
	return resp, err
}

func provideTestData(msg string) (internal.TaskOptional, []byte) {
	testTask := provideTask(msg)

	payload, err := provideJsonBody(testTask)
	if err != nil {
		panic(err)
	}
	return testTask, payload
}

func provideTask(msg string) internal.TaskOptional {
	now := time.Now()
	testTask := internal.TaskOptional{
		Done:      internal.BoolPtr(false),
		Msg:       internal.StringPtr(msg),
		Category:  internal.CategoryPtr(internal.Brewing),
		PlannedAt: &internal.CustomTime{Time: now.Add(24 * time.Hour)},
	}
	return testTask
}

func provideJsonBody(testTask internal.TaskOptional) ([]byte, error) {
	payload, err := json.Marshal(testTask)
	return payload, err
}