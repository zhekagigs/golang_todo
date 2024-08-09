package analytics

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAnalytics(t *testing.T) {
	t.Skip("not ready")
	var wg sync.WaitGroup
	wg.Add(2)

	go driveCoordinator(&wg)
	time.Sleep(time.Second * 2)
	go driveWorkers(&wg)
	time.Sleep(time.Second * 5)
	wg.Wait()
}

func TestWordCount(t *testing.T) {
	// Example usage
	content := "Hello world hello Go programming Hello"

	// Use Map function
	mapped := WcMap("file1", content)

	// Group by key (this step is usually part of the MapReduce framework)
	grouped := make(map[string][]string)
	for _, kv := range mapped {
		grouped[kv.Key] = append(grouped[kv.Key], kv.Value)
	}
	fmt.Println(grouped)
	// Use Reduce function
	for key, values := range grouped {
		result := WcReduce(key, values)
		fmt.Printf("%s: %s\n", key, result)
	}
}
