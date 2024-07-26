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



func RealMain() int {
	taskHolder, checkExit, exitCode := InitialMain()
	if checkExit {
		return exitCode
	}

	// runs main CLI routine
	returnCode := RunTaskManagmentCLI(taskHolder)

	return returnCode
}

func InitialMain() (*TaskHolder, bool, int) {
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
