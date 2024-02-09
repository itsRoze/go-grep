package main

import (
	"fmt"
	"grepclone/worker"
	"grepclone/worklist"
	"os"
	"path/filepath"
	"sync"
)

// Given a directory, this function will add all files in the directory to the worklist
func getAllFiles(wl *worklist.Worklist, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			nextPath := filepath.Join(path, entry.Name())
			getAllFiles(wl, nextPath)
		} else {
			wl.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}
}

func main() {
	var wg sync.WaitGroup
	wl := worklist.New(100)
	results := make(chan worker.Result, 100)
	numWorkers := 10

	// 1. Add all files to worklist
	wg.Add(1)
	go func() {
		defer wg.Done()
		getAllFiles(&wl, os.Args[2])
		wl.Finalize(numWorkers)
	}()

	// 2. Find matches
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				workEntry := wl.Next()
				if workEntry.Path != "" {
					workerResult := worker.FindInFile(workEntry.Path, os.Args[1])
					if workerResult != nil {
						for _, r := range workerResult.Inner {
							results <- r
						}
					}
				} else {
					return
				}
			}
		}()
	}

	blockWg := make(chan struct{})
	go func() {
		wg.Wait()
		close(blockWg)
	}()

	var displayWg sync.WaitGroup

	// 3. Print results
	displayWg.Add(1)
	go func() {
		for {
			select {
			case r := <-results:
				fmt.Printf("%v[%v]:%v\n", r.Path, r.LineNum, r.Line)
			case <-blockWg:
				if len(results) == 0 {
					displayWg.Done()
					return
				}
			}
		}
	}()
	displayWg.Wait()
}
