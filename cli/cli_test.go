package cli

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"os"
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
				updt := in.TaskOptional{
					Done:      nil,
					Msg:       in.StringPtr("Test task 1"),
					Category:  in.CategoryPtr(in.TaskCategory(in.Brewing)),
					PlannedAt: in.TimePtr(in.MockTime),
				}
				th.CreateTask(updt)
			},
			expected: "\nList of tasks:\n\nid:",
		},
		{
			name: "Multiple tasks",
			setupTasks: func(th *in.TaskHolder) {
				th.CreateTask(in.TaskOptional{
					Done:      nil,
					Msg:       in.StringPtr("Task 1"),
					Category:  in.CategoryPtr(in.TaskCategory(in.Brewing)),
					PlannedAt: in.TimePtr(in.MockTime),
				})
				th.CreateTask(in.TaskOptional{
					Done:      nil,
					Msg:       in.StringPtr("Test task 2"),
					Category:  in.CategoryPtr(in.TaskCategory(in.Marketing)),
					PlannedAt: in.TimePtr(in.MockTime),
				})
			},
			expected: "\nList of tasks:\n\nid:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskHolder := in.NewTaskHolder("../internal/resources/cli_disk_test.json")
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
			taskHolder := in.NewTaskHolder("../internal/resources/cli_disk_test.json")
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
		th := in.NewTaskHolder("..internal/resources/cli_disk_test.json")
		th.CreateTask(in.TaskOptional{
			Done:      nil,
			Msg:       in.StringPtr("Test task 1"),
			Category:  in.CategoryPtr(in.TaskCategory(in.Brewing)),
			PlannedAt: in.TimePtr(in.MockTime),
		})
		th.CreateTask(in.TaskOptional{
			Done:      nil,
			Msg:       in.StringPtr("Test task 2"),
			Category:  in.CategoryPtr(in.TaskCategory(in.Marketing)),
			PlannedAt: in.TimePtr(in.MockTime),
		})
		th.CreateTask(in.TaskOptional{
			Done:      nil,
			Msg:       in.StringPtr("Test task 3"),
			Category:  in.CategoryPtr(in.TaskCategory(in.Logistics)),
			PlannedAt: in.TimePtr(in.MockTime),
		})
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
				return in.NewTaskHolder("../internal/resources/cli_disk_test.json")
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
	t.Skip("Formatting issues")

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
			input:          "Initial Task\ny\ntrue\ny\n2\ny\n2025-07-01 10:00\n",
			expectedOutput: "Updating task. Press Enter to skip a field if you don't want to update it.",
			expectedError:  false,
			validateTask: func(t *testing.T, task *in.Task) {
				if task.Msg != "Initial Task" {
					t.Errorf("Expected task message to be 'Initial Task', got '%s'", task.Msg)
				}
				if task.Done {
					t.Errorf("Expected task to be not done")
				}
				if task.Category != in.Brewing {
					t.Errorf("Expected task category to be Logistics, got %v", task.Category)
				}
				//TODO fix time
				// expectedTime, _ := time.Parse(in.TASK_TIME_FORMAT, "2025-07-01 10:00")
				// if !task.PlannedAt.Equal(expectedTime) {
				// 	t.Errorf("Expected planned time to be %v, got %v", expectedTime, task.PlannedAt)
				// }
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
		expectedCmd    command
		expectedTaskId int
		expectError    bool
	}{
		{"Valid read command", "read\n", READ, 0, false},
		{"Valid create command", "create\n", CREATE, 0, false},
		{"Valid update command", "update 5\n", UPDATE, 5, false},
		{"Valid delete command", "delete 3\n", DELETE, 3, false},
		{"Valid exit command", "exit\n", EXIT, 0, false},
		{"Empty input", "\n", "", -1, true},
		{"Invalid command", "invalid\n", "invalid", 0, false},
		{"Update without ID", "update\n", UPDATE, 0, false},
		{"Update with invalid ID", "update abc\n", "", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			cmd, taskId, word, err := parseCommand(reader)
			if word != "" {
				t.Errorf("parseCommand() word is not empty, %s", word)
			}
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
		cmd         command
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
				th.CreateTask(in.TaskOptional{
					Done:      nil,
					Msg:       in.StringPtr("Update task 1"),
					Category:  in.CategoryPtr(in.TaskCategory(in.Brewing)),
					PlannedAt: in.TimePtr(in.MockTime),
				})
			},
			input:      "Updated Task\ny\ntrue\nn\nn\n",
			expectExit: -1,
		},
		{
			name:   "Delete command",
			cmd:    DELETE,
			taskId: 1,
			setup: func(th *in.TaskHolder) {
				th.CreateTask(in.TaskOptional{
					Done:      nil,
					Msg:       in.StringPtr("Task to delete"),
					Category:  in.CategoryPtr(in.TaskCategory(in.Brewing)),
					PlannedAt: in.TimePtr(in.MockTime),
				})
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
			taskHolder := in.NewTaskHolder("../internal/resources/cli_disk_test.json")
			tt.setup(taskHolder)

			reader := bufio.NewReader(strings.NewReader(tt.input))

			exitCode := executeCommand(tt.cmd, tt.taskId, "", taskHolder, reader)

			if exitCode != tt.expectExit {
				t.Errorf("executeCommand() exitCode = %v, want %v", exitCode, tt.expectExit)
			}

			// Add more specific checks based on the command executed
			// For example, check if a task was created, updated, or deleted
		})
	}
}

func TestRunCLI(t *testing.T) {
	taskHolder := in.ProvideTaskHolderWithPath("../internal/resources/cli_disk_test.json")

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
	cliApp := &RealCLIApp{}
	go func() {
		cliApp.RunTaskManagmentCLI(taskHolder)
	}()

	in.WriteToCapturedStdin(inWrite, cmnds)

	in.RestoreStdout(write, oldstd)
	in.RestoreStdin(inRead, oldstdIn)
	output := in.ReadCapturedStdout(read)
	expectedOutputs := []string{
		"Available Commands: read, create, update, delete, exit, search, find",

		// "id:1,[Brewing] Initial Task",
		// "id:2,[Marketing] New task",
		"Thank you for using the Task Management CLI. Tasks are saved to",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nGot: %s", expected, output)
		}
	}

}

func TestMainAndPrintHelp(t *testing.T) {

	// Save original stdout, args, and flag.CommandLine
	oldStdout := os.Stdout
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		// Restore original stdout, args, and flag.CommandLine after all tests
		os.Stdout = oldStdout
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		expectExit     int
	}{
		{
			name:           "No arguments",
			args:           []string{"cmd"},
			expectedOutput: []string{"Error: JSON file path is required", "Usage: microbrewery-tasks"},
			expectExit:     1,
		},
		{
			name:           "Wrong file name",
			args:           []string{"cmd", "wrong_file.garbage"},
			expectedOutput: []string{"Error while reading json file: invalid file extension: wrong_file.garbage. Expected a .json file", "Usage: microbrewery-tasks"},
			expectExit:     1,
		},
		{
			name:           "Help flag",
			args:           []string{"cmd", "-h"},
			expectedOutput: []string{"Usage: microbrewery-tasks", "Options:", "Description:"},
			expectExit:     0,
		},
		{
			name:           "Valid file argument",
			args:           []string{"cmd", "../internal/resources/tasks.json"},
			expectedOutput: []string{"Microbrewery Tasks Application"},
			expectExit:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture stdout
			read, write, err := os.Pipe()
			defer read.Close()

			if err != nil {
				panic(err)
			}
			os.Stdout = write
			// Set up a new flag set for each test
			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ContinueOnError)
			os.Args = tt.args
			cliApp := &RealCLIApp{}
			_, _, actualExit, _ := cliApp.AppStarter(in.MockNewTaskHolder)

			// Close the write end of the pipe
			write.Close()

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, read)
			output := buf.String()

			// Check if exit was called when expected
			if tt.expectExit != actualExit {
				t.Errorf("Expected exit: %v, but got: %v", tt.expectExit, actualExit)
			}

			// Check for expected output
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nGot: %s", expected, output)
				}
			}
		})
	}
}
