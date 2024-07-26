package cli

import (
	"bufio"
	"strings"
	"testing"
	"time"

	in "github.com/zhekagigs/golang_todo/internal"
)

func TestReadTasks(t *testing.T) {
	tests := []struct {
		name       string
		setupTasks func(*in.TaskHolder)
		expected   string
	}{
		{
			name:       "No tasks",
			setupTasks: func(th *in.TaskHolder) {},
			expected:   "No tasks found.\n",
		},
		{
			name: "Single task",
			setupTasks: func(th *in.TaskHolder) {
				th.CreateTask("Test task 1", in.Brewing, in.MockTime)
			},
			expected: "\nList of tasks:\n\nid:1,[Brewing] Test task 1",
		},
		{
			name: "Multiple tasks",
			setupTasks: func(th *in.TaskHolder) {
				th.CreateTask("Task 1", in.Brewing, in.MockTime)
				th.CreateTask("Task 2", in.Marketing, in.MockTime)
			},
			expected: "\nList of tasks:\n\nid:1,[Brewing] Task 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := in.NewTaskHolder()
			tt.setupTasks(taskHolder)

			old, r, w := in.CaptureStdout()
			readTasks(taskHolder)
			in.RestoreStdout(w, old)
			output := in.ReadCapturedStdout(r)

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
		expectedTask  func(*testing.T, *in.Task)
	}{
		{
			name:          "Valid input",
			input:         "Finish brewing IPA, 0, 2024-08-29 14:27\n",
			expectedError: false,
			expectedTask: func(t *testing.T, task *in.Task) {
				if task == nil {
					t.Fatal("Expected task to be created, but it was nil")
				}
				if task.Msg != "Finish brewing IPA" {
					t.Errorf("Expected task message to be 'Finish brewing IPA', got '%s'", task.Msg)
				}
				if task.Category != in.Brewing {
					t.Errorf("Expected task category to be in.Brewing, got %v", task.Category)
				}
				expectedTime, _ := time.Parse(in.TASK_TIME_FORMAT, "2024-08-29 14:27")
				if !task.PlannedAt.Equal(expectedTime) {
					t.Errorf("Expected planned time to be %v, got %v", expectedTime, task.PlannedAt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := in.NewTaskHolder()
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
	setupTaskHolder := func() *in.TaskHolder {
		th := in.NewTaskHolder()
		th.CreateTask("in.Task 1", in.Brewing, time.Now().Add(24*time.Hour))
		th.CreateTask("in.Task 2", in.Marketing, time.Now().Add(48*time.Hour))
		th.CreateTask("in.Task 3", in.Logistics, time.Now().Add(72*time.Hour))
		return th
	}

	tests := []struct {
		name          string
		taskId        int
		setupHolder   func() *in.TaskHolder
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
			name:   "Delete from empty in.in.TaskHolder",
			taskId: 1,
			setupHolder: func() *in.TaskHolder {
				return in.NewTaskHolder()
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
						t.Errorf("in.Task with ID %d should have been deleted, but it still exists", tt.taskId)
					}
				}
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {

	tests := []struct {
		name           string
		taskId         int
		input          string
		expectedOutput string
		expectedError  bool
		validateTask   func(*testing.T, *in.Task)
	}{
		{
			name:           "Update all fields",
			taskId:         1,
			input:          "New Task Description\ny\ntrue\ny\n2\ny\n2025-07-01 10:00\n",
			expectedOutput: "Task updated successfully.",
			expectedError:  false,
			validateTask: func(t *testing.T, task *in.Task) {
				if task.Msg != "New Task Description" {
					t.Errorf("Expected task message to be 'New Task Description', got '%s'", task.Msg)
				}
				if !task.Done {
					t.Errorf("Expected task to be done")
				}
				if task.Category != in.Logistics {
					t.Errorf("Expected task category to be Logistics, got %v", task.Category)
				}
				expectedTime, _ := time.Parse(in.TASK_TIME_FORMAT, "2025-07-01 10:00")
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
			validateTask: func(t *testing.T, task *in.Task) {
				// in.Task should remain unchanged
				if task.Msg != "Initial Task" {
					t.Errorf("Expected task message to be 'Initial Task', got '%s'", task.Msg)
				}
				if task.Done {
					t.Errorf("Expected task to be not done")
				}
				if task.Category != in.Brewing {
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
			taskHolder := in.ProvideTaskHolder()

			// Prepare input and output
			input := strings.NewReader(tt.input)
			reader := bufio.NewReader(input)

			oldStdout, r, w := in.CaptureStdout()

			// Run the function
			err := updateTask(taskHolder, tt.taskId, reader)

			// Restore stdout
			in.RestoreStdout(w, oldStdout)

			// Read captured output
			output := in.ReadCapturedStdout(r)

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
					t.Fatal("No tasks found in in.TaskHolder")
				}
				tt.validateTask(t, &tasks[0])
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCmd    commands
		expectedTaskId int
		expectError    bool
	}{
		{"Valid read command", "read\n", READ, 0, false},
		{"Valid create command", "create\n", CREATE, 0, false},
		{"Valid update command", "update 5\n", UPDATE, 5, false},
		{"Valid delete command", "delete 3\n", DELETE, 3, false},
		{"Valid exit command", "exit\n", EXIT, 0, false},
		{"Empty input", "\n", "", 0, true},
		{"Invalid command", "invalid\n", "invalid", 0, false},
		{"Update without ID", "update\n", UPDATE, 0, false},
		{"Update with invalid ID", "update abc\n", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			cmd, taskId, err := parseCommand(reader)

			if (err != nil) != tt.expectError {
				t.Errorf("parseCommand() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if cmd != tt.expectedCmd {
				t.Errorf("parseCommand() cmd = %v, want %v", cmd, tt.expectedCmd)
			}
			if taskId != tt.expectedTaskId {
				t.Errorf("parseCommand() taskId = %v, want %v", taskId, tt.expectedTaskId)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmd         commands
		taskId      int
		setup       func(*in.TaskHolder)
		input       string
		expectExit  int
		expectError bool
	}{
		{
			name:       "Read command",
			cmd:        READ,
			taskId:     0,
			setup:      func(th *in.TaskHolder) {},
			expectExit: -1,
		},
		{
			name:       "Create command",
			cmd:        CREATE,
			taskId:     0,
			setup:      func(th *in.TaskHolder) {},
			input:      "New Task, 0, 2023-07-01 10:00\n",
			expectExit: -1,
		},
		{
			name:   "Update command",
			cmd:    UPDATE,
			taskId: 1,
			setup: func(th *in.TaskHolder) {
				th.CreateTask("Initial Task", in.Brewing, time.Now().Add(24*time.Hour))
			},
			input:      "Updated Task\ny\ntrue\nn\nn\n",
			expectExit: -1,
		},
		{
			name:   "Delete command",
			cmd:    DELETE,
			taskId: 1,
			setup: func(th *in.TaskHolder) {
				th.CreateTask("Task to Delete", in.Brewing, time.Now().Add(24*time.Hour))
			},
			expectExit: -1,
		},
		{
			name:       "Exit command",
			cmd:        EXIT,
			taskId:     0,
			setup:      func(th *in.TaskHolder) {},
			expectExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := in.NewTaskHolder()
			tt.setup(taskHolder)

			reader := bufio.NewReader(strings.NewReader(tt.input))

			exitCode := executeCommand(tt.cmd, tt.taskId, taskHolder, reader)

			if exitCode != tt.expectExit {
				t.Errorf("executeCommand() exitCode = %v, want %v", exitCode, tt.expectExit)
			}

			// Add more specific checks based on the command executed
			// For example, check if a task was created, updated, or deleted
		})
	}
}

func TestRunCLI(t *testing.T) {
	// t.Skip("Not ready")
	taskHolder := in.ProvideTaskHolder()

	oldstd, read, write := in.CaptureStdout()
	oldstdIn, inRead, inWrite := in.CaptureStdin()

	cmnds := []string{
		"read\n",
		"create\nNew task, 1, 2024-08-01 10:00\n",
		"read\n",
		"update 2\ny\nUpdated task\ny\n2\ny\n2024-08-02 11:00\n",
		"read\n",
		"delete 2\n",
		"read\n",
		"exit\n",
	}

	go func() {
		RunTaskManagmentCLI(taskHolder)
	}()

	in.WriteToCapturedStdin(inWrite, cmnds)

	in.RestoreStdout(write, oldstd)
	in.RestoreStdin(inRead, oldstdIn)
	output := in.ReadCapturedStdout(read)
	expectedOutputs := []string{
		"Available Commands: read, create, update, delete, exit",
		"id:1,[Brewing] Initial Task",
		"id:2,[Marketing] New task",
		"Enter Command: Thank you for using the Task Management CLI. Tasks are saved to",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nGot: %s", expected, output)
		}
	}

}
