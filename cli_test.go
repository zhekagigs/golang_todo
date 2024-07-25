package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// captureOutput captures stdout and returns it as a string
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestReadTasks(t *testing.T) {
	tests := []struct {
		name       string
		setupTasks func(*TaskHolder)
		expected   string
	}{
		{
			name:       "No tasks",
			setupTasks: func(th *TaskHolder) {},
			expected:   "No tasks found.\n",
		},
		{
			name: "Single task",
			setupTasks: func(th *TaskHolder) {
				th.CreateTask("Test task 1", Brewing, mockTime)
			},
			expected: "\nList of tasks:\n\nid:1,[Brewing] Test task 1",
		},
		{
			name: "Multiple tasks",
			setupTasks: func(th *TaskHolder) {
				th.CreateTask("Task 1", Brewing, mockTime)
				th.CreateTask("Task 2", Marketing, mockTime)
			},
			expected: "\nList of tasks:\n\nid:1,[Brewing] Task 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := NewTaskHolder()
			tt.setupTasks(taskHolder)

			output := captureOutput(func() {
				readTasks(taskHolder)
			})

			// fmt.Println(output)

			if !strings.HasPrefix(output, tt.expected) {
				t.Errorf("got %v, want %v", output, tt.expected)
			}

			tasks := taskHolder.Read()
			if len(tasks) > 0 {
				for _, task := range tasks {
					if !strings.Contains(output, task.Msg) {
						t.Errorf("output doesn't contain task message: %v", task.Msg)
					}
				}
			}
		})
	}
}

func TestCreateCLITask(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError bool
		expectedTask  func(*testing.T, *Task)
	}{
		{
			name:          "Valid input",
			input:         "Finish brewing IPA, 0, 2024-08-29 14:27\n",
			expectedError: false,
			expectedTask: func(t *testing.T, task *Task) {
				if task == nil {
					t.Fatal("Expected task to be created, but it was nil")
				}
				if task.Msg != "Finish brewing IPA" {
					t.Errorf("Expected task message to be 'Finish brewing IPA', got '%s'", task.Msg)
				}
				if task.Category != Brewing {
					t.Errorf("Expected task category to be Brewing, got %v", task.Category)
				}
				expectedTime, _ := time.Parse(TASK_TIME_FORMAT, "2024-08-29 14:27")
				if !task.PlannedAt.Equal(expectedTime) {
					t.Errorf("Expected planned time to be %v, got %v", expectedTime, task.PlannedAt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := NewTaskHolder()
			reader := bufio.NewReader(strings.NewReader(tt.input))

			err := createTask(taskHolder, reader)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectedError && tt.expectedTask != nil {
				tasks := taskHolder.Read()
				if len(tasks) == 0 {
					t.Fatal("Expected a task to be created, but none was found")
				}
				tt.expectedTask(t, &tasks[0])
			}
		})
	}
}

func TestDeleteCLITask(t *testing.T) {
	setupTaskHolder := func() *TaskHolder {
		th := NewTaskHolder()
		th.CreateTask("Task 1", Brewing, time.Now().Add(24*time.Hour))
		th.CreateTask("Task 2", Marketing, time.Now().Add(48*time.Hour))
		th.CreateTask("Task 3", Logistics, time.Now().Add(72*time.Hour))
		return th
	}

	tests := []struct {
		name          string
		taskId        int
		setupHolder   func() *TaskHolder
		expectedError bool
		expectedTasks int
	}{
		{
			name:          "Delete existing task",
			taskId:        2,
			setupHolder:   setupTaskHolder,
			expectedError: false,
			expectedTasks: 2,
		},
		{
			name:          "Delete non-existent task",
			taskId:        99,
			setupHolder:   setupTaskHolder,
			expectedError: true,
			expectedTasks: 3,
		},
		{
			name:   "Delete from empty TaskHolder",
			taskId: 1,
			setupHolder: func() *TaskHolder {
				return NewTaskHolder()
			},
			expectedError: true,
			expectedTasks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := tt.setupHolder()

			err := deleteTask(taskHolder, tt.taskId)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			remainingTasks := taskHolder.Read()
			if len(remainingTasks) != tt.expectedTasks {
				t.Errorf("Expected %d tasks after deletion, but got %d", tt.expectedTasks, len(remainingTasks))
			}

			if !tt.expectedError {
				for _, task := range remainingTasks {
					if task.Id == tt.taskId {
						t.Errorf("Task with ID %d should have been deleted, but it still exists", tt.taskId)
					}
				}
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	setupTaskHolder := func() *TaskHolder {
		th := NewTaskHolder()
		th.CreateTask("Initial Task", Brewing, time.Now().Add(24*time.Hour))
		return th
	}

	tests := []struct {
		name           string
		taskId         int
		input          string
		expectedOutput string
		expectedError  bool
		validateTask   func(*testing.T, *Task)
	}{
		{
			name:           "Update all fields",
			taskId:         1,
			input:          "New Task Description\ny\ntrue\ny\n2\ny\n2025-07-01 10:00\n",
			expectedOutput: "Task updated successfully.",
			expectedError:  false,
			validateTask: func(t *testing.T, task *Task) {
				if task.Msg != "New Task Description" {
					t.Errorf("Expected task message to be 'New Task Description', got '%s'", task.Msg)
				}
				if !task.Done {
					t.Errorf("Expected task to be done")
				}
				if task.Category != Logistics {
					t.Errorf("Expected task category to be Logistics, got %v", task.Category)
				}
				expectedTime, _ := time.Parse(TASK_TIME_FORMAT, "2025-07-01 10:00")
				if !task.PlannedAt.Equal(expectedTime) {
					t.Errorf("Expected planned time to be %v, got %v", expectedTime, task.PlannedAt)
				}
			},
		},
		{
			name:           "Skip all updates",
			taskId:         1,
			input:          "\nn\nn\nn\n",
			expectedOutput: "Task updated successfully.",
			expectedError:  false,
			validateTask: func(t *testing.T, task *Task) {
				// Task should remain unchanged
				if task.Msg != "Initial Task" {
					t.Errorf("Expected task message to be 'Initial Task', got '%s'", task.Msg)
				}
				if task.Done {
					t.Errorf("Expected task to be not done")
				}
				if task.Category != Brewing {
					t.Errorf("Expected task category to be Brewing, got %v", task.Category)
				}
			},
		},
		// {
		// 	name:   "Invalid task ID",
		// 	taskId: 999,
		// 	input:  "\nn\nn\nn\n",
		// 	expectedOutput: "Error updating task:",
		// 	expectedError:  true,
		// },
		// {
		// 	name:   "Invalid boolean input",
		// 	taskId: 1,
		// 	input:  "\ny\ninvalid\n",
		// 	expectedOutput: "Error updating task:",
		// 	expectedError:  true,
		// },
		// {
		// 	name:   "Invalid category input",
		// 	taskId: 1,
		// 	input:  "\nn\ny\n5\n",
		// 	expectedOutput: "Error updating task:",
		// 	expectedError:  true,
		// },
		// {
		// 	name:   "Invalid date input",
		// 	taskId: 1,
		// 	input:  "\nn\nn\ny\ninvalid date\n",
		// 	expectedOutput: "Error updating task:",
		// 	expectedError:  true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := setupTaskHolder()

			// Prepare input and output
			input := strings.NewReader(tt.input)
			reader := bufio.NewReader(input)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the function
			err := updateTask(taskHolder, tt.taskId, reader)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check for expected error
			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', but got '%s'", tt.expectedOutput, output)
			}

			// Validate task state if no error was expected
			if !tt.expectedError && tt.validateTask != nil {
				tasks := taskHolder.Read()
				if len(tasks) == 0 {
					t.Fatal("No tasks found in TaskHolder")
				}
				tt.validateTask(t, &tasks[0])
			}
		})
	}
}

var mockExit func()

// Mock WriteToJson function
var mockWriteToJson func(string, ...Task) error

func init() {
	WriteToJson = func(filePath string, tasks ...Task) error {
		return mockWriteToJson(filePath, tasks...)
	}
}

func TestRunTaskManagmentCLI(t *testing.T) {
	tests := []struct {
		name           string
		inputs         []string
		expectedOutput []string
		setup          func(*TaskHolder)
	}{
		{
			name: "Read empty task list",
			inputs: []string{
				"read\n",
				"exit\n",
			},
			expectedOutput: []string{
				"No tasks found.",
				"Available Commands: read, create, update, delete",
			},
			setup: func(th *TaskHolder) {},
		},
		{
			name: "Create a task",
			inputs: []string{
				"create\n",
				"New Task, 0, 2023-07-01 10:00\n",
				"read\n",
				"exit\n",
			},
			expectedOutput: []string{
				"Enter new task on one line in a format 'task, category, planned to finish date'",
				"New Task",
				"Available Commands: read, create, update, delete",
			},
			setup: func(th *TaskHolder) {},
		},
		{
			name: "Update a task",
			inputs: []string{
				"update 1\n",
				"Updated Task\n",
				"y\n",
				"true\n",
				"n\n",
				"n\n",
				"read\n",
				"exit\n",
			},
			expectedOutput: []string{
				"Updating task. Press Enter to skip a field if you don't want to update it.",
				"Task updated successfully.",
				"Updated Task",
				"Available Commands: read, create, update, delete",
			},
			setup: func(th *TaskHolder) {
				th.CreateTask("Initial Task", Brewing, time.Now().Add(24*time.Hour))
			},
		},
		{
			name: "Delete a task",
			inputs: []string{
				"delete 1\n",
				"read\n",
				"exit\n",
			},
			expectedOutput: []string{
				"No tasks found.",
				"Available Commands: read, create, update, delete",
			},
			setup: func(th *TaskHolder) {
				th.CreateTask("Task to Delete", Brewing, time.Now().Add(24*time.Hour))
			},
		},
		{
			name: "Invalid command",
			inputs: []string{
				"invalid\n",
				"exit\n",
			},
			expectedOutput: []string{
				"Invalid command. Please try again.",
				"Available Commands: read, create, update, delete",
			},
			setup: func(th *TaskHolder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			taskHolder := NewTaskHolder()
			tt.setup(taskHolder)

			// Prepare input
			input := strings.NewReader(strings.Join(tt.inputs, ""))

			// Capture stdout
			oldStdin, oldStdout := os.Stdin, os.Stdout
			readPipe, writePipe, _ := os.Pipe()
			os.Stdout = writePipe
			os.Stdin = readPipe

			// Prepare to read output
			outputChan := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, readPipe)
				outputChan <- buf.String()
			}()

			// Mock exit function
			exited := false
			mockExit = func() {
				exited = true
				writePipe.Close()
			}

			// Run the function in a goroutine
			go RunTaskManagmentCLI(taskHolder)

			// Write inputs
			io.Copy(writePipe, input)

			// Wait for exit
			output := <-outputChan

			// Restore stdin and stdout
			os.Stdin, os.Stdout = oldStdin, oldStdout

			// Check if exited
			if !exited {
				t.Error("RunTaskManagmentCLI did not exit")
			}

			// Check output
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nGot: %s", expected, output)
				}
			}
		})
	}
}
