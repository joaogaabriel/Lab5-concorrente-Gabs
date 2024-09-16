package main

import (
	"fmt"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/file"
	"github.com/goPirateBay/server"
	"log"
	"os"
	"path/filepath"
)

func main() {

	directory := constants.InitDirFiles
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		createFilesTest(directory)
	}
	go server.StartServer()

	fileIndex, err := file.ListFilesInDirectory(directory)
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	for _, file := range fileIndex.Files {
		fmt.Println(file.SHA1Hash)
		fmt.Printf("File: %s, Size: %d bytes\n", file.Name, file.Size)
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
			log.Fatalf("Failed to create file: %v", err)
		}

		_, err = file.WriteString("This is some test content for " + fileName)
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}

		err = file.Close()
		if err != nil {
			log.Fatalf("Failed to close file: %v", err)
		}

		fmt.Println("File created:", filePath)
	}

	fmt.Println("All files created successfully!")
}
