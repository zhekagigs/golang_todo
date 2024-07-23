package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	helpFlag := flag.Bool("h", false, "Only flag there is and its help")

	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: JSON file path is required")
		printHelp()
		os.Exit(1)
		return
	}

	fileName := flag.Arg(0)
	fmt.Println(BeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	fmt.Printf("List of tasks:\n\n")

	savedTasks := ReadFromJson(fileName)
	PrintTasks(os.Stdout, savedTasks...)
}

func printHelp() {
	fmt.Println("Usage: microbrewery-tasks [options] <json-file-path>")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nDescription:")
	fmt.Println("  This CLI application reads a JSON file containing microbrewery tasks and displays them.")
	fmt.Println("  Provide the path to the JSON file as an argument.")
}
