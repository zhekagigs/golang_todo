package main

import (
	"errors"
	"os"
	"testing"
	"time"
)

var TimeExample = time.Date(2024, 7, 14, 12, 45, 00, 00, time.Local)

func TestFormatDatetime(t *testing.T) {
	got := formatDatetime(TimeExample)
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

func TestBeerAscii(t *testing.T) {
	got := BeerAscii()
	want, err := os.ReadFile("resources/beer.txt")
	check(err)
	if string(want) != got {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestCheck(t *testing.T) {
	t.Run("No Panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				t.Errorf("Code panicked unexpectedly")
			}
		}()
		check(nil)
	})
	t.Run("Panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("Code didn't panicked as it should")
			}
		}()
		check(errors.New("can't work with 42"))
	})
}
