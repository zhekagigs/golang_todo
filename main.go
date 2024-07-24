package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	os.Exit(RealMain())
}

type commands string

const (
	READ   commands = "read"
	CREATE commands = "create"
	UPDATE commands = "update"
	DELETE commands = "delete"
	EXIT   commands = "exit"
)

func RealMain() int {
	// readCmnd := flag.NewFlagSet(string(READ), flag.ExitOnError)
	// createCmnd := flag.NewFlagSet(string(CREATE), flag.ExitOnError)
	// updateCmnd := flag.NewFlagSet(string(UPDATE), flag.ExitOnError)
	// deleteCmnd := flag.NewFlagSet(string(DELETE), flag.ExitOnError)
	taskHolder, checkExit, exitCode := ConfigureMain()
	if checkExit {
		return exitCode
	}
	RunTaskManagmentCLI(taskHolder)

	return 0
}

func ConfigureMain() (*TaskHolder, bool, int) {
	helpFlag := flag.Bool("h", false, "Help is here")

	flag.Usage = printHelp

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return nil, true, 0
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: JSON file path is required")
		flag.Usage()
		return nil, true, 1
	}

	fileName := flag.Arg(0)

	savedTasks, err := ReadFromJson(fileName)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			fmt.Println("Error: Wrong file path")
		default:
			fmt.Printf("Error while reading json file: %v\n", err)
		}
		flag.Usage()
		return nil, true, 1
	}
	fmt.Println(BeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	fmt.Println(savedTasks)
	PrintTasks(os.Stdout, savedTasks...)
	taskHolder := NewTaskHolder()
	for _, task := range savedTasks {
		taskHolder.Add(task)
	}
	return taskHolder, false, 0
}

func printHelp() {
	fmt.Println("Usage: microbrewery-tasks [options] <json-file-path>")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nDescription:")
	fmt.Println("  This CLI application reads a JSON file containing microbrewery tasks and displays them.")
	fmt.Println("  Provide the path to the JSON file as an argument.")
}
