package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	in "github.com/zhekagigs/golang_todo/internal"
)

type commands string

const (
	READ   commands = "read"
	CREATE commands = "create"
	UPDATE commands = "update"
	DELETE commands = "delete"
	EXIT   commands = "exit"
)

const (
	ExitCodeSuccess = 0
	ExitCodeError   = 1
)

type CLIApp interface {
	AppStarter(newTaskHolder func(diskPath string) *in.TaskHolder) (*in.TaskHolder, bool, int)
	RunTaskManagmentCLI(taskHolder *in.TaskHolder) int
}

type RealCLIApp struct {
}

func (cli *RealCLIApp) RunTaskManagmentCLI(taskHolder *in.TaskHolder) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		displayCommands()
		cmd, taskId, err := parseCommand(reader)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if exitCode := executeCommand(cmd, taskId, taskHolder, reader); exitCode != -1 {
			return exitCode
		}
	}
}

func (cli *RealCLIApp) AppStarter(newTaskHolder func(diskPath string) *in.TaskHolder) (*in.TaskHolder, bool, int) {
	fileName, savedTasks, isHelp, isExit, exitCode := ParseUserArg()
	if isHelp {
		return nil, isExit, exitCode
	}
	PrintCLITitle(savedTasks)

	taskHolder, err := PopulateTaskHolder(fileName, savedTasks, newTaskHolder)
	if err != nil {
		fmt.Printf("Error populating task holder: %v\n", err)
		return nil, true, ExitCodeError
	}
	return taskHolder, false, ExitCodeSuccess
}

func PrintCLITitle(savedTasks []in.Task) {
	fmt.Println(in.BeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	in.PrintTasks(os.Stdout, savedTasks...)
}

func PopulateTaskHolder(fileName string, savedTasks []in.Task, newTaskHolder func(diskPath string) *in.TaskHolder) (*in.TaskHolder, error) {
	if fileName == "" {
		fileName = "resources/disk.json"
	}
	taskHolder := newTaskHolder(fileName)
	for _, task := range savedTasks {
		taskHolder.Add(task)
	}
	return taskHolder, nil
}

func ParseUserArg() (fileName string, savedTasks []in.Task, isHelp bool, isExit bool, exitCode int) {
	helpFlag := flag.Bool("h", false, "Help is here")

	flag.Usage = PrintHelp

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return "", nil, true, true, ExitCodeSuccess
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: JSON file path is required")
		flag.Usage()
		return "", nil, true, true, ExitCodeError
	}

	fileName = flag.Arg(0)
	savedTasks, err := in.ReadTasksFromJSON(fileName)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			fmt.Println("Error: Wrong file path")
		default:
			fmt.Printf("Error while reading json file: %v\n", err)
		}
		flag.Usage()
		return "", nil, true, true, ExitCodeError
	}
	return fileName, savedTasks, false, false, ExitCodeSuccess
}

func PrintHelp() {
	fmt.Println("Usage: microbrewery-tasks [options] <json-file-path>")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nDescription:")
	fmt.Println("  This CLI application reads a JSON file containing microbrewery tasks and displays them.")
	fmt.Println("  Provide the path to the JSON file as an argument.")
}

func displayCommands() {
	fmt.Println("\nAvailable Commands: read, create, update, delete, exit")
	fmt.Print("Enter Command: ")
}

func parseCommand(reader *bufio.Reader) (commands, int, error) {
	cmdString, _ := reader.ReadString('\n')
	parts := strings.Fields(cmdString)
	if len(parts) == 0 {
		return "", 0, fmt.Errorf("Please enter a command.")
	}

	cmd := commands(strings.TrimSpace(strings.ToLower(parts[0])))

	var taskId int
	var err error
	if len(parts) > 1 && (cmd == UPDATE || cmd == DELETE) {
		taskId, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("Invalid task ID. Please enter a number.")
		}
	}

	return cmd, taskId, nil
}

func executeCommand(cmd commands, taskId int, taskHolder *in.TaskHolder, reader *bufio.Reader) int {
	var err error
	switch cmd {
	case READ:
		readTasks(taskHolder)
	case CREATE:
		err = createTask(taskHolder, reader)
	case UPDATE:
		err = updateTask(taskHolder, taskId, reader)
	case DELETE:
		err = deleteTask(taskHolder, taskId)
	case EXIT:
		return exitApp(taskHolder)
	default:
		fmt.Println("Invalid command. Please try again.")
	}

	if err != nil {
		fmt.Println("Failed operation with error: ", err)
	}

	return -1 // Continue the loop
}

func exitApp(taskHolder *in.TaskHolder) int {
	fmt.Println("Thank you for using the Task Management CLI. Tasks are saved to ", taskHolder.DiskPath, " GoodBye!")
	err := in.WriteToJson(taskHolder.DiskPath, taskHolder.Tasks...)
	if err != nil {
		panic(err)
	}
	return 0
}

func deleteTask(taskHolder *in.TaskHolder, taskId int) error {
	err := taskHolder.DeleteTask(taskId)
	if err != nil {
		return err
	}
	return err
}

func updateTask(taskHolder *in.TaskHolder, taskId int, reader *bufio.Reader) error {
	fmt.Println("Updating task. Press Enter to skip a field if you don't want to update it.")

	// Update task message
	fmt.Print("Enter new task description (or press Enter to skip): ")
	msg, _ := reader.ReadString('\n')
	msg = strings.TrimSpace(msg)

	// Update task status
	var donePtr *bool
	fmt.Print("Update task status? (y/n): ")
	updateStatus, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(updateStatus)) == "y" {
		fmt.Print("Is the task done? (true/false): ")
		doneStr, _ := reader.ReadString('\n')
		doneStr = strings.TrimSpace(doneStr)
		if parsedDone, err := strconv.ParseBool(doneStr); err == nil {
			donePtr = &parsedDone
		} else {
			return err
		}
	}

	// Update task category
	var category in.TaskCategory
	var categoryPtr *in.TaskCategory
	fmt.Print("Update task category? (y/n): ")
	updateCategory, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(updateCategory)) == "y" {
		fmt.Println("Available categories:")
		fmt.Println("0: Brewing")
		fmt.Println("1: Marketing")
		fmt.Println("2: Logistics")
		fmt.Println("3: Quality")
		fmt.Print("Enter new category (0-3): ")
		categoryStr, _ := reader.ReadString('\n')
		if parsedCategory, err := strconv.Atoi(strings.TrimSpace(categoryStr)); err == nil && parsedCategory >= 0 && parsedCategory <= 3 {
			category = in.TaskCategory(parsedCategory)
			categoryPtr = &category
		} else {
			return err
		}
	}

	// Update planned time
	var plannedAtPtr *time.Time
	fmt.Print("Update planned time? (y/n): ")
	updatePlannedTime, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(updatePlannedTime)) == "y" {
		fmt.Print("Enter new planned time (YYYY-MM-DD HH:MM): ")
		timeStr, _ := reader.ReadString('\n')
		if parsedTime, err := time.Parse(in.TASK_TIME_FORMAT, strings.TrimSpace(timeStr)); err == nil {
			plannedAtPtr = &parsedTime
		} else {
			return err
		}
	}

	update := &in.TaskOptional{
		Done:      donePtr,
		Msg:       in.StringPtr(msg),
		Category:  categoryPtr,
		PlannedAt: plannedAtPtr,
	}

	err := taskHolder.PartialUpdateTask(taskId, update)
	if err != nil {
		fmt.Printf("Error updating task: %v\n", err)
		return err
	} else {
		fmt.Println("Task updated successfully.")
	}
	return nil
}

func createTask(taskHolder *in.TaskHolder, reader *bufio.Reader) error {

	fmt.Println("Enter new task on one line in a format 'task, category, planned to finish date'")
	fmt.Println("Available categories:")
	fmt.Println("0: Brewing")
	fmt.Println("1: Marketing")
	fmt.Println("2: Logistics")
	fmt.Println("3: Quality")
	fmt.Println("Format time (YYYY-MM-DD HH:MM)")
	fmt.Println("Example: `Finish brewing IPA, 0, 2024-08-29 14:27`")

	line, err := reader.ReadString('\n') //TODO unignore errors
	if err != nil {
		return err
	}

	lines := strings.Split(line, ",")

	taskValue := lines[0]
	categoryNum, err := strconv.Atoi(strings.TrimSpace(lines[1]))
	fmt.Println(categoryNum)
	if err != nil {
		return err
	}
	plannedAt := lines[2]
	var plannedParsedAt time.Time
	parsedTime, err := time.Parse(in.TASK_TIME_FORMAT, strings.TrimSpace(plannedAt))
	if err == nil {
		plannedParsedAt = parsedTime
	}
	updt := &in.TaskOptional{
		Done:      nil,
		Msg:       in.StringPtr(taskValue),
		Category:  in.CategoryPtr(in.TaskCategory(categoryNum)),
		PlannedAt: in.TimePtr(plannedParsedAt),
	}

	taskHolder.CreateTask(updt)
	return nil
}

func readTasks(taskHolder *in.TaskHolder) {
	all_tasks := taskHolder.Read()
	if len(all_tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Printf("\nList of tasks:\n\n")
	in.PrintTasks(os.Stdout, all_tasks...)
}
