package main

import (
	"sync"
	"testing"
)

func BenchmarkContextSwitch(b *testing.B) {

	var wg sync.WaitGroup
	begin := make(chan struct{})
	testChan := make(chan struct{})

	var empty struct{}
	sender := func() {
		defer wg.Done()
		<-begin
		for i := 0; i < b.N; i++ {
			testChan <- empty
		}
	}
	receiver := func() {
		defer wg.Done()
		<-begin
		for i := 0; i < b.N; i++ {
			<-testChan
		}
	}

	wg.Add(2)
	go sender()
	go receiver()
	b.StartTimer()
	close(begin)
	wg.Wait()
}
