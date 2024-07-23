package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(BeerAscii())
	fmt.Printf("\n>>>>>>>>>>Microbrewery Tasks Application<<<<<<<<<<<<<\n\n")
	fmt.Printf("List of tasks: \n\n")
	all_tasks := generateRandomTasks(10)
	fileName := "tasks.json"
	SaveToJson(fileName, all_tasks...)
	savedTasks := ReadFromJson(fileName)
	fmt.Println()
	fmt.Println()
	PrintTasks(os.Stdout, savedTasks...)

}
