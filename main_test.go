package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

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
			args:           []string{"cmd", "resources/tasks.json"},
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

			_, _, actualExit := InitialMain()

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
