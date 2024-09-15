package file

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	cache     *FileIndex
	cacheTime time.Time
	cacheTTL  = time.Minute
	mutex     sync.Mutex
)

type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	SHA1Hash string `json:"sha1_hash,omitempty"`
}

type File struct {
	FilePath   string
	buffer     *bytes.Buffer
	OutputFile *os.File
}

type FileIndex struct {
	Files []FileInfo `json:"files"`
}

func (f *File) SetFile(fileName, path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	f.FilePath = filepath.Join(path, fileName)
	file, err := os.Create(f.FilePath)
	if err != nil {
		return err
	}
	f.OutputFile = file
	return nil
}

func (f *File) Write(chunk []byte) error {
	if f.OutputFile == nil {
		return nil
	}
	_, err := f.OutputFile.Write(chunk)
	return err
}

func (f *File) Close() error {
	return f.OutputFile.Close()
}

func calculateSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func FindFileByHash(files []FileInfo, hash string) *FileInfo {
	for _, file := range files {
		if file.SHA1Hash == hash {
			return &file
		}
	}
	return nil
}

func ListFilesInDirectory(directory string) (FileIndex, error) {
	var index FileIndex

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {

			sha1Hash, err := calculateSHA1(path) // Calcula o SHA-1 do arquivo
			if err != nil {
				return err
			}

			file := FileInfo{
				Name:     info.Name(),
				Size:     info.Size(),
				SHA1Hash: sha1Hash,
			}
			index.Files = append(index.Files, file)
		}

		return nil
	})

	return index, err
}
