package analytics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func driveCoordinator(wg *sync.WaitGroup) {
	filesPaths, err := getJSONFiles("resources/")
	if err != nil {
		panic(err)
	}
	coord := MakeCoordinator(filesPaths, 5)
	for coord.Done() == false {
		time.Sleep(time.Second)
	}
	if coord.Done() {
		fmt.Println("coordinator done")
	}

	time.Sleep(time.Second)
	wg.Done()
}

func driveWorkers(wg *sync.WaitGroup) {

	mapf := WcMap
	reducef := WcReduce

	Worker(mapf, reducef)
	wg.Done()
}

func getJSONFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	var jsonFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			jsonFiles = append(jsonFiles, filepath.Join(dir, entry.Name()))
		}
	}

	return jsonFiles, nil
}
