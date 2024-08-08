package internal

import (
	"fmt"

	"github.com/google/uuid"
)

type TaskRequest struct {
	Operation  string
	Task       TaskOptional
	TaskId     int
	trackingId uuid.UUID
}

type TaskResult struct {
	Task  *Task
	Error error
}

type Worker struct {
	id         int
	taskHolder *TaskHolder
	request    <-chan TaskRequest
	result     chan<- TaskResult
	quit       chan bool
}

type WorkerPool struct {
	workers []Worker
	// wg      sync.WaitGroup
}

func (w *Worker) Start() {
	for {
		select {
		case req, ok := <-w.request:
			if !ok {
				return
			}
			result := w.processRequest(req)
			w.result <- result
		case <-w.quit:
			return
		}
	}
}

func (w *Worker) Stop() {
	close(w.quit)
}

func (w *Worker) processRequest(req TaskRequest) TaskResult {
	switch req.Operation {
	case "CREATE":
		// fmt.Println("Worker calls holder to create task", w.id)
		task := w.taskHolder.CreateTask(req.Task)
		return TaskResult{Task: task}
	default:
		return TaskResult{Error: fmt.Errorf("unknown operation: %s", req.Operation)}
	}
}

func NewWorkerPool(numWorkers int, taskHolder *TaskHolder, taskRequests <-chan TaskRequest, taskResponse chan<- TaskResult) WorkerPool {
	// fmt.Println("Init NewWorkerPool")
	var workers []Worker
	for i := 0; i < numWorkers; i++ {
		workers = append(workers, Worker{
			id:         i,
			taskHolder: taskHolder,
			request:    taskRequests,
			result:     taskResponse,
			quit:       make(chan bool),
		})
	}
	return WorkerPool{ // composite literal
		workers: workers,
	}
}

func (wp *WorkerPool) Start() {
	for i := range wp.workers {
		go wp.workers[i].Start()
	}
}

func (wp *WorkerPool) Close() {
	for _, worker := range wp.workers {
		worker.Stop()
	}
}
