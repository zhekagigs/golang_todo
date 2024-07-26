package internal

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

var MockTime = time.Date(2023, 7, 23, 12, 0, 0, 0, time.UTC)

func ProvideMocktimeNow(t *testing.T) func() time.Time {
	originalTimeNow := timeNow
	t.Cleanup(func() {
		timeNow = originalTimeNow
	})
	return func() time.Time {
		return MockTime
	}
}

func ProvideTask(t *testing.T) Task {
	return NewTask(1, "task_my_task", 1, MockTime)
}

func ProvideTaskHolder() *TaskHolder {
	th := NewTaskHolder()
	th.DiskPath = "../resources/cli_disk_test.json"
	th.CreateTask("Initial Task", Brewing, time.Now().Add(24*time.Hour))
	return th
}

func ReadCapturedStdout(r *os.File) string {
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	return output
}

func WriteToCapturedStdin(write *os.File, cmnds []string) {
	time.Sleep(100 * time.Millisecond)
	for _, cmnd := range cmnds {
		write.Write([]byte(cmnd))
		time.Sleep(100 * time.Millisecond)
	}
}

// Restore after capturing
func RestoreStdout(w *os.File, oldStdout *os.File) {
	w.Close()
	os.Stdout = oldStdout
}

// Restore after capturing
func RestoreStdin(r *os.File, oldStdin *os.File) {
	r.Close()
	os.Stdin = oldStdin
}

// Capture stdout. DON'T FORGET TO RESTORE!
func CaptureStdout() (oldStdout *os.File, read *os.File, write *os.File) {
	oldStdout = os.Stdout
	read, write, _ = os.Pipe()
	os.Stdout = write
	return
}

// Capture stdin. DON'T FORGET TO RESTORE!
func CaptureStdin() (oldStdin *os.File, read *os.File, write *os.File) {
	oldStdin = os.Stdin
	read, write, _ = os.Pipe()
	os.Stdin = read
	return
}
