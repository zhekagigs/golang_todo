package internal

import (
	"os"
	"reflect"
	"testing"
)

func TestProvideMocktimeNow(t *testing.T) {
	t.Run("Returns correct mock time", func(t *testing.T) {
		mockTimeNow := ProvideMocktimeNow(t)
		if got := mockTimeNow(); !got.Equal(MockTime) {
			t.Errorf("ProvideMocktimeNow() = %v, want %v", got, MockTime)
		}
	})

}

func TestProvideTask(t *testing.T) {
	got := ProvideTask(t)
	want := NewTask(1, "task_my_task", 1, MockTime)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ProvideTask() = %v, want %v", got, want)
	}
}

func TestProvideTaskHolder(t *testing.T) {
	th := ProvideTaskHolder()
	if th == nil {
		t.Fatalf("ProvideTaskHolder() returned nil")
	}
	if th.DiskPath != "resources/cli_disk_test.json" {
		t.Errorf("ProvideTaskHolder() DiskPath = %v, want %v", th.DiskPath, "resources/cli_disk_test.json")
	}
	if len(th.Tasks) != 1 {
		t.Errorf("ProvideTaskHolder() task count = %v, want 1", len(th.Tasks))
	}
	if th.Tasks[0].Msg != "Initial Task" {
		t.Errorf("ProvideTaskHolder() task message = %v, want 'Initial Task'", th.Tasks[0].Msg)
	}
}

func TestProvideTaskHolderWithPath(t *testing.T) {
	testPath := "test/path.json"
	th := ProvideTaskHolderWithPath(testPath)
	if th == nil {
		t.Fatalf("ProvideTaskHolderWithPath() returned nil")
	}
	if th.DiskPath != testPath {
		t.Errorf("ProvideTaskHolderWithPath() DiskPath = %v, want %v", th.DiskPath, testPath)
	}
	if len(th.Tasks) != 1 {
		t.Errorf("ProvideTaskHolderWithPath() task count = %v, want 1", len(th.Tasks))
	}
	if th.Tasks[0].Msg != "Initial Task" {
		t.Errorf("ProvideTaskHolderWithPath() task message = %v, want 'Initial Task'", th.Tasks[0].Msg)
	}
}

func TestMockNewTaskHolder(t *testing.T) {
	th := MockNewTaskHolder("some/path.json")
	if th == nil {
		t.Fatalf("MockNewTaskHolder() returned nil")
	}
	if th.DiskPath != "resources/cli_disk_test.json" {
		t.Errorf("MockNewTaskHolder() DiskPath = %v, want %v", th.DiskPath, "resources/cli_disk_test.json")
	}
	if len(th.Tasks) != 1 {
		t.Errorf("MockNewTaskHolder() task count = %v, want 1", len(th.Tasks))
	}
	if th.Tasks[0].Msg != "Initial Task" {
		t.Errorf("MockNewTaskHolder() task message = %v, want 'Initial Task'", th.Tasks[0].Msg)
	}
}

func TestReadCapturedStdout(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test data to the file
	testData := "Test output"
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in temp file: %v", err)
	}

	// Test ReadCapturedStdout
	got := ReadCapturedStdout(tmpfile)
	if got != testData {
		t.Errorf("ReadCapturedStdout() = %v, want %v", got, testData)
	}
}

func TestCaptureAndRestoreStdout(t *testing.T) {
	oldStdout, r, w := CaptureStdout()

	// Write something to stdout
	os.Stdout.Write([]byte("Test output"))

	RestoreStdout(w, oldStdout)

	// Read captured output
	captured := ReadCapturedStdout(r)
	if captured != "Test output" {
		t.Errorf("CaptureStdout() captured %v, want %v", captured, "Test output")
	}

	// Verify stdout is restored
	if os.Stdout != oldStdout {
		t.Errorf("RestoreStdout() did not restore original stdout")
	}
}

func TestCaptureAndRestoreStdin(t *testing.T) {
	oldStdin, r, w := CaptureStdin()

	// Write test input
	w.Write([]byte("Test input"))
	w.Close()

	RestoreStdin(r, oldStdin)

	// Verify stdin is restored
	if os.Stdin != oldStdin {
		t.Errorf("RestoreStdin() did not restore original stdin")
	}
}
