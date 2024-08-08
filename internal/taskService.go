package internal

type ConcurrentTaskService struct {
	TaskHolder   *TaskHolder
	taskRequests chan<- TaskRequest
	results      <-chan TaskResult
	workerPool   WorkerPool
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
