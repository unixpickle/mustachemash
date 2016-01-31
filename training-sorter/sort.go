package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/unixpickle/mustachemash/mustacher"
)

func SortImages(basePath string, paths []string, dbs []*mustacher.Database) {
	log.Println("Sorting...")

	pathChan := make(chan string, len(paths))
	for _, path := range paths {
		pathChan <- path
	}
	close(pathChan)

	wg := &sync.WaitGroup{}
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go SortImageRoutine(basePath, dbs, pathChan, wg)
	}

	wg.Wait()
}

func SortImageRoutine(basePath string, dbs []*mustacher.Database, ch <-chan string,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range ch {
		category, err := CategorizeImage(path, dbs)
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		destPath := filepath.Join(basePath, string(category), filepath.Base(path))
		if path != destPath {
			log.Println("Re-categorizing", path, "to", category)
			if err := os.Rename(path, destPath); err != nil {
				log.Println("Error:", err)
			}
		}
	}
}
