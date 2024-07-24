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
	helpFlag := flag.Bool("h", false, "Display help information")

	flag.Usage = printHelp

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return 0
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: JSON file path is required")
		flag.Usage()
		return 1
	}

	fileName := flag.Arg(0)
	fmt.Println(BeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	fmt.Printf("List of tasks:\n\n")

	savedTasks, err := ReadFromJson(fileName)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			fmt.Println("Error: Wrong file path")
		default:
			fmt.Println("Error reading json file &v", err)
		}
		flag.Usage()
		return 1
	}
	PrintTasks(os.Stdout, savedTasks...)
	return 0
}

func printHelp() {
	fmt.Println("Usage: microbrewery-tasks [options] <json-file-path>")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nDescription:")
	fmt.Println("  This CLI application reads a JSON file containing microbrewery tasks and displays them.")
	fmt.Println("  Provide the path to the JSON file as an argument.")
}
