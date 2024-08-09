package analytics

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

/*
The workers will talk to the coordinator via RPC.
Each worker process will ask the coordinator for a task,
read the task's input from one or more files, execute the task,
and write the task's output to one or more files.
The coordinator should notice
if a worker hasn't completed its task in a reasonable amount of time
(for this lab, use ten seconds),
and give the same task to a different worker.
*/

type Coordinator struct {
	// Your definitions here.
	Tasks          []Task
	NumReduce      int
	MapTaskCode    int
	ReduceTaskCode int
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) GiveMeAMapTask(request *TaskRequest, reply *TaskReply) error {
	log.Println("Coordinator::Map requested")

	for _, task := range c.Tasks {
		if !task.isMapped {
			reply.Filename = task.Filename
			reply.NumReducers = c.NumReduce
			reply.TaskNumber = task.TaskNum
			task.State = InProgress
		}
	}
	return nil
}

func (c *Coordinator) TaskDone(request *TaskRequest, reply *TaskReply) error {
	log.Println("Coordinator::TaskDone requested")
	log.Println("Coordinator::received", request)

	for _, task := range c.Tasks {
		if task.TaskNum == request.TaskNumber {
			task.State = request.State
			if request.State == Completed && request.TaskType == MapTask {
				task.isMapped = true

			}
			log.Println("Coordinator::updated task state", task.Filename, task.State)
		}

		// if !task.isMapped {
		// 	reply.Filename = task.Filename
		// 	reply.NumReducers = c.NumReduce
		// 	reply.TaskNumber = task.TaskNum
		// 	task.State = InProgress
		// }
	}
	return nil
}

// If no re- sponse is received from a worker in a certain amount of time, the master marks the worker as failed. Any map tasks completed by the worker are reset back to their ini- tial idle state, and therefore become eligible for schedul- ing on other workers. Similarly, any map task or reduce task in progress on a failed worker is also reset to idle and becomes eligible for rescheduling.
func (c *Coordinator) PingWorker() error {
	return nil
}

func (c *Coordinator) giveMeAReduceTask(request *TaskRequest, reply *TaskReply) error {
	fmt.Println("Coordinator::Reduce requested")
	for _, task := range c.Tasks {
		if !task.isReduced {
			reply.Filename = task.Filename
		}
	}
	return nil
}

// an example RPC handler.
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	fmt.Println("Coordinator::Example hande")
	reply.Y = args.X + 11
	reply.World = " It's world, my man "
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	log.Println("server:: start serving")
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Coordinator::Server listens on sockname")
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	result := true
	for _, task := range c.Tasks {
		if !task.isMapped || !task.isReduced {
			result = false
		}
	}
	return result
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	var tasks []Task
	for i, filename := range files {
		tasks = append(tasks, Task{filename, i, false, false, MapTask})
	}
	c := Coordinator{
		Tasks:          tasks,
		NumReduce:      nReduce,
		MapTaskCode:    0,
		ReduceTaskCode: 0,
	}

	log.Println("Coordinator:: files to work", tasks)
	log.Println("Coordinator:: new coordinator made v.01")

	c.server()
	return &c
}
