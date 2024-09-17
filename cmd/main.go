package main

import (
	"fmt"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/fileUtils"
	"github.com/goPirateBay/server"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {

	directory := constants.InitDirFiles
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		createFilesTest(directory)
	}

	cacheFiles := fileUtils.NewFileCache(time.Minute)

	cacheFiles.StartPeriodicCacheUpdate(constants.InitDirFiles, 2*time.Minute)

	go server.StartServer(cacheFiles)

	time.Sleep(time.Second / 2)
	for _, file := range cacheFiles.GetAllFiles() {
		fmt.Printf("File: %s, Size: %d bytes HASH: %s\n", file.Name, file.Size, file.SHA1Hash)
	}
	select {}
}

func createFilesTest(dirPath string) {

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Println("Directory created:", dirPath)

	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	for _, fileName := range files {
		filePath := filepath.Join(dirPath, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("Failed to create fileUtils: %v", err)
		}

		_, err = file.WriteString("This is some netUtils content for " + fileName)
		if err != nil {
			log.Fatalf("Failed to write to fileUtils: %v", err)
		}

		err = file.Close()
		if err != nil {
			log.Fatalf("Failed to close fileUtils: %v", err)
		}

		fmt.Println("File created:", filePath)
	}

	fmt.Println("All files created successfully!")
}
