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
		expectExit     bool
	}{
		// {
		// 	name:           "No arguments",
		// 	args:           []string{"cmd", ""},
		// 	expectedOutput: []string{"Error: JSON file path is required", "Usage: microbrewery-tasks"},
		// 	expectExit:     true,
		// },
		{
			name:           "Help flag",
			args:           []string{"cmd", "-h"},
			expectedOutput: []string{"Usage: microbrewery-tasks", "Options:", "Description:"},
			expectExit:     false,
		},
		{
			name:           "Valid file argument",
			args:           []string{"cmd", "resources/tasks.json"},
			expectedOutput: []string{"Microbrewery Tasks Application", "List of tasks:"},
			expectExit:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set up a new flag set for each test
			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ContinueOnError)
			os.Args = tt.args

			// Capture panics and exits
			exitCalled := false
			oldOsExit := osExit
			osExit = func(code int) {
				exitCalled = true
				panic("os.Exit called")
			}
			defer func() {
				osExit = oldOsExit
				if r := recover(); r != nil {
					if r != "os.Exit called" {
						t.Fatalf("Unexpected panic: %v", r)
					}
				}
			}()
			// Run main
			func() {
				defer func() {
					if r := recover(); r != nil {
						if r != "os.Exit called" {
							t.Fatalf("Unexpected panic: %v", r)
						}
					}
				}()
				main()
			}()
			// Close the write end of the pipe
			w.Close()

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check if exit was called when expected
			if tt.expectExit != exitCalled {
				t.Errorf("Expected exit: %v, but got: %v", tt.expectExit, exitCalled)
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

// Mock os.Exit
var osExit = os.Exit
