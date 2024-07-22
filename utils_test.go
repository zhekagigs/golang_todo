package main

import (
	"testing"
	"time"
)

var timeExample = time.Date(2024, 7, 14, 12, 45, 00, 00, time.Local)

func TestFormatDatetime(t *testing.T) {
	got := formatDatetime(timeExample)
	want := "Sunday, July 14, 2024 at 12:45"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestGenerateRandomTasks(t *testing.T) {
	want := 10
	got := len(generateRandomTasks(10))

	if want != got {
		t.Errorf("got %v want %v", got, want)
	}

}
