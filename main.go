package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(printBeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	fmt.Printf("List of tasks: \n\n")
	all_tasks := generateRandomTasks(10)

	PrintTasks(os.Stdout, all_tasks...)
}
