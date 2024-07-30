package main

import (
	"net/http"
	"testing"

	in "github.com/zhekagigs/golang_todo/internal"
)

type MockHTTPServer struct {
	ListenAndServeCalledWith string
}

func (s *MockHTTPServer) ListenAndServe(addr string, handler http.Handler) error {
	s.ListenAndServeCalledWith = addr
	return nil
}

type MockCLIApp struct {
	AppStarterCalled bool
	RunCLICalled     bool
}

func (cli *MockCLIApp) AppStarter(newTaskHolder func(diskPath string) *in.TaskHolder) (*in.TaskHolder, bool, int) {
	cli.AppStarterCalled = true
	taskHolder := in.MockNewTaskHolder("")
	return taskHolder, false, 0
}

func (cli *MockCLIApp) RunTaskManagmentCLI(taskHolder *in.TaskHolder) int {
	cli.RunCLICalled = true
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