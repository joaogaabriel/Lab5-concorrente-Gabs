package main

import (
	"fmt"
	"github.com/goPirateBay/file"
	"log"
)

func main() {

	directory := "/tmp/goPirateBay"

	fileIndex, err := file.ListFilesInDirectory(directory)
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	for _, file := range fileIndex.Files {
		fmt.Println(file.SHA1Hash)
		fmt.Printf("File: %s, Size: %d bytes\n", file.Name, file.Size)
	}
}
