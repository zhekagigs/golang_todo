package analytics

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func ScanFile(f *os.File) (map[string]string, error) {
	var (
		kmap map[string]string
		err  error
	)
	// Read the entire file into a byte slice
	b, err := io.ReadAll(f)
	if err != nil {
		return kmap, err
	}
	// Convert the byte slice to a string
	str := string(b[:])
	// Split the string into lines
	lines := strings.Split(str, "\n")
	// Iterate over the lines and split them into key-value pairs
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		key := parts[0]
		value := parts[1]
		kmap[key] = value
	}
	return kmap, nil
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
// The map phase should divide the intermediate keys into buckets for nReduce reduce
// tasks, where nReduce is the number of reduce tasks -- the argument that
// main/mrcoordinator.go passes to MakeCoordinator(). Each mapper should create
// nReduce intermediate files for consumption by the reduce tasks.
// The worker implementation should put the output of the X'th reduce task in the file mr-out-X.
// A mr-out-X file should contain one line per Reduce function output. The line should be generated with the Go "%v %v" format, called with the key and value.
// When the job is completely finished, the worker processes should exit. A simple way to implement this is to use the return value from call(): if the worker fails to contact the coordinator, it can assume that the coordinator has exited because the job is done, so the worker can terminate too. Depending on your design, you might also find it helpful to have a "please exit" pseudo-task that the coordinator can give to workers.
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	log.Println("Worker:: worker started")
	for {
		reply := CallCoordinatorForTask("GiveMeAMapTask")
		log.Println("task received", MessageType(reply.TaskType))
		if reply.TaskType == MapTask || reply.TaskState == Idle {
			tmpFileNames := doMap(mapf, reply)
			CallMapTaskDone(tmpFileNames, reply.TaskNumber)
		}
		if reply.TaskType == ReduceTask {
			doReduce(reply, reply.Filename, reducef)
		} else {
			log.Panicf("unknown task type, %v", reply.TaskType)
		}
	}
}

func CallMapTaskDone(tmpFileNames []string, tasknumber int) TaskReply {
	log.Println("Worker::callMapTaskDone initialized")
	args := TaskRequest{Completed, tmpFileNames, tasknumber, MapTask}
	reply := TaskReply{}
	handlerName := "TaskDone"
	ok := call("Coordinator."+handlerName, &args, &reply)
	if ok {
		log.Printf("Worker::coordinator replied with %s\n", reply)
	} else {
		fmt.Println("MyCall failed!")
	}

	return reply
}

func doReduce(replied TaskReply, intermediate string, reducef func(string, []string) string) {
	oname := "mr-out-" + strconv.Itoa(replied.TaskNumber)
	ofile, _ := os.Create(oname)
	produceReducedOutput(intermediate, reducef, ofile)
	log.Println("Worker::finished printing out file " + oname)
	ofile.Close()
}

func doMap(mapf func(string, string) []KeyValue, replied TaskReply) []string {
	log.Printf("Worker::doMap\n")
	intermediate := collectIntermediate(mapf, replied.Filename, replied.NumReducers)
	// debug := 0
	// for k, v := range intermediate {
	// 	fmt.Println(k, v)
	// 	if debug > 10 {
	// 		break
	// 	}
	// 	debug++
	// }
	var interNames []string
	interName := "map-inter-"

	for key, interGroupings := range intermediate {
		sort.Sort(ByKey(interGroupings))
		tmpIntername := interName + strconv.Itoa(key) + "-" + replied.Filename
		interFile, _ := os.Create(tmpIntername)
		for _, kvpair := range interGroupings {
			fmt.Fprintf(interFile, "%v %v\n", kvpair.Key, kvpair.Value)
		}
		log.Println("Worker::created temp file ", interName+strconv.Itoa(key))
		interFile.Close()
		interNames = append(interNames, tmpIntername)
	}
	log.Println("Worker::finished printing intermeditea file " + interName)
	return interNames
}

func collectIntermediate(mapf func(string, string) []KeyValue, filename string, nReduce int) map[int][]KeyValue {
	log.Printf("Worker::doMap\n")

	partitions := make(map[int][]KeyValue)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()
	keyValuePairs := mapf(filename, string(content))
	for _, kv := range keyValuePairs {
		partitionKey := ihash(kv.Key) % nReduce

		if _, ok := partitions[partitionKey]; !ok {
			partitions[partitionKey] = []KeyValue{kv}
		} else {
			partitions[partitionKey] = append(partitions[partitionKey], kv)
		}
		if (len(partitions)) > nReduce {
			panic("Worker:: partition unexpected")
		}
	}

	return partitions
}

func produceReducedOutput(intermediateFile string, reducef func(string, []string) string, ofile *os.File) {
	i := 0
	f, err := os.Open(intermediateFile)
	if err != nil {
		panic("Error reduce task reading inter file")
	}
	defer f.Close()

	// // Create a scanner to scan the file
	// scanner := Scanner{
	// 	F:        f,
	// 	Buffer:   make([]byte, 1024),
	// 	Capacity: 1024,
	// }

	scanner := bufio.NewScanner(f)

	var kvCollector []KeyValue

	// Iterate over each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		// Check for errors while scanning
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}
		// Trim the newline character from the end of the line
		line = strings.TrimRight(line, "\r\n")
		// Parse the line into a KeyValue struct
		var kv KeyValue

		_, err = fmt.Sscanf(line, "%s %s", &kv.Key, &kv.Value)
		if err != nil {
			fmt.Println(err)
			continue
		}
		kvCollector = append(kvCollector, kv)
		// Do something with the KeyValue struct
	}

	for i < len(kvCollector) {
		j := i + 1
		for j < len(kvCollector) && kvCollector[j].Key == kvCollector[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kvCollector[k].Value)
		}
		output := reducef(kvCollector[i].Key, values)

		fmt.Fprintf(ofile, "%v %v\n", kvCollector[i].Key, output)

		i = j
	}
}

func CallCoordinatorForTask(handlerName string) TaskReply {
	log.Println("Worker::CallForMapTask")
	args := TaskRequest{}
	reply := TaskReply{}
	ok := call("Coordinator."+handlerName, &args, &reply)
	if ok {
		log.Printf("Worker::coordinator replied with %s\n", reply)
	} else {
		fmt.Println("MyCall failed!")
	}
	return reply
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {
	log.Println()
	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99
	args.Hello = "Hello"

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
		fmt.Printf("coordinator replyed with %s\n", reply.World)
	} else {
		fmt.Printf("Example call failed!\n")
	}

}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func SplitSlice(array []int, numberOfChunks int) [][]int {
	if len(array) == 0 {
		return nil
	}
	if numberOfChunks <= 0 {
		return nil
	}
	if numberOfChunks == 1 {
		return [][]int{array}
	}
	result := make([][]int, numberOfChunks)
	// we have more splits than elements in the input array.
	if numberOfChunks > len(array) {
		for i := 0; i < len(array); i++ {
			result[i] = []int{array[i]}
		}
		return result
	}
	for i := 0; i < numberOfChunks; i++ {
		min := (i * len(array) / numberOfChunks)
		max := ((i + 1) * len(array)) / numberOfChunks
		result[i] = array[min:max]
	}
	return result
}
