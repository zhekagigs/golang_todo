package frontend

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	in "github.com/zhekagigs/golang_todo/internal"
)

func TestHandleTaskList(t *testing.T) {

	th := in.ProvideTaskHolder()
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/tasks", nil)
	if err != nil {
		t.Errorf("Error createsting test request")
	}
	HandleTaskListRead(w, r, th)
	response := w.Result()
	body, _ := io.ReadAll(response.Body)
	if status := response.StatusCode; status != http.StatusOK {
		t.Errorf("want %v, got %v, status %s", http.StatusOK, response.StatusCode, response.Status)
		t.Errorf("%v", string(body))
	}
	if !strings.Contains(string(body), "Initial Task") {
		t.Errorf("want %v, got %v", "Initial Taks", string(body))
	}

}
