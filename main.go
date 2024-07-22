package main

import (
	"fmt"
)

func main() {
	fmt.Println(printBeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	fmt.Printf("List of tasks: \n\n")
	all_tasks := generateRandomTasks(10)

	PrintTasks(all_tasks...)
}
