package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	in "github.com/zhekagigs/golang_todo/internal"
)

type MockHTTPServer struct {
	ListenAndServeCalledWith string
	ShutdownCount            int
}

func (s *MockHTTPServer) ListenAndServe(addr string, handler http.Handler) error {
	s.ListenAndServeCalledWith = addr
	return nil
}

func (s *MockHTTPServer) Shutdown(ctx context.Context) error {
	s.ShutdownCount++
	return nil
}

type MockCLIApp struct {
	AppStarterCalled bool
	RunCLICalled     bool
}

func (cli *MockCLIApp) AppStarter(newTaskHolder func(diskPath string) *in.TaskHolder) (*in.TaskHolder, bool, int, bool) {
	cli.AppStarterCalled = true
	taskHolder := in.MockNewTaskHolder("")
	return taskHolder, false, 0, false
}

func (cli *MockCLIApp) RunTaskManagmentCLI(taskHolder *in.TaskHolder) int {
	cli.RunCLICalled = true
	time.Sleep(1 * time.Second) // TODO figure out a block for server
	return 0
}

func TestRealMain(t *testing.T) {
	mockServer := &MockHTTPServer{}
	mockCli := &MockCLIApp{}
	exitCode := RealMain(in.MockNewTaskHolder, mockServer, mockCli)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if mockServer.ListenAndServeCalledWith != ":8080" {
		t.Errorf("Expected server to listen on :8080, got %s", mockServer.ListenAndServeCalledWith)
	}
	if !mockCli.AppStarterCalled {
		t.Error("Expected AppStarter to be called")
	}
	if !mockCli.RunCLICalled {
		t.Error("Expected RunTaskManagmentCLI to be called")
	}
}
