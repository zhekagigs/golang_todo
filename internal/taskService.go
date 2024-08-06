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

type ConcurrentTaskService struct {
	TaskHolder   *TaskHolder
	taskRequests chan<- TaskRequest
	results      <-chan TaskResult
	workerPool   WorkerPool
}

func (w *Worker) Start() {
	// fmt.Printf("Start worker %d\n", w.id)
	for {
		select {
		case req, ok := <-w.request:
			// fmt.Println("Handling request by", w.id)
			if !ok {
				// fmt.Println("Channel not ok")
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

func NewConcurrentTaskService(t *TaskHolder) *ConcurrentTaskService {
	// fmt.Println("Init NewConcurrentTaskService")
	taskRequests := make(chan TaskRequest)
	results := make(chan TaskResult)
	pool := NewWorkerPool(5, t, taskRequests, results)
	service := &ConcurrentTaskService{
		TaskHolder:   t,
		taskRequests: taskRequests,
		results:      results,
		workerPool:   pool,
	}
	service.workerPool.Start()

	return service
}

func (t *ConcurrentTaskService) CloseAll() {
	// close(t.results)
	close(t.taskRequests)

}

func (t *ConcurrentTaskService) CreateTask(task TaskOptional) *Task {
	// fmt.Println("Create task pushed on taskrequest channel", task)
	// t.taskRequests <- TaskRequest{Operation: "CREATE", Task: task}
	// result := <-t.results

	return t.TaskHolder.CreateTask(task)
}

func (t *ConcurrentTaskService) Add(task Task) {
	t.TaskHolder.Add(task)
}

func (t *ConcurrentTaskService) FindTaskById(taskId int) (*Task, error) {
	return t.TaskHolder.FindTaskById(taskId)
}

func (t *ConcurrentTaskService) PartialUpdateTask(taskId int, update *TaskOptional) error {
	return t.TaskHolder.PartialUpdateTask(taskId, update)
}

func (t *ConcurrentTaskService) DeleteTask(taskId int) error {
	return t.TaskHolder.DeleteTask(taskId)
}

// returns latestId and len of tasks
func (t *ConcurrentTaskService) Count() (int, int) {
	return t.TaskHolder.Count()
}

func (t *ConcurrentTaskService) Read() []Task {
	return t.TaskHolder.Read()
}
