package fileUtils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/goPirateBay/constants"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	SHA1Hash string `json:"sha1_hash,omitempty"`
}

type File struct {
	FilePath   string
	buffer     *bytes.Buffer
	OutputFile *os.File
}

type FileCache struct {
	mu         sync.RWMutex
	cache      map[string]FileInfo
	expiration map[string]time.Time
	ttl        time.Duration
}

func NewFileCache(ttl time.Duration) *FileCache {
	return &FileCache{
		cache:      make(map[string]FileInfo),
		expiration: make(map[string]time.Time),
		ttl:        ttl,
	}
}

func (fc *FileCache) GetAllFiles() []FileInfo {

	err := fc.LoadFiles(constants.InitDirFiles)
	if err != nil {
		return nil
	}

	var files []FileInfo
	for _, file := range fc.cache {
		files = append(files, file)
	}
	return files
}

func (fc *FileCache) LoadFiles(dir string) error {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	for _, file := range files {

		if !file.IsDir() {
			filePath := filepath.Join(dir, file.Name())
			sha1Hash, err := calculateSHA1(filePath)
			if err != nil {
				return err
			}

			fileInfo := FileInfo{
				Name:     file.Name(),
				Path:     filePath,
				Size:     file.Size(),
				SHA1Hash: sha1Hash,
			}

			fc.cache[sha1Hash] = fileInfo

			fc.expiration[sha1Hash] = time.Now().Add(fc.ttl)
		}
	}
	return nil
}

func (fc *FileCache) GetFile(sha1_hash string) (FileInfo, bool) {

	fc.mu.RLock()
	file, found := fc.cache[sha1_hash]
	expiration, exists := fc.expiration[sha1_hash]
	fc.mu.RUnlock()

	if !found || (exists && time.Now().After(expiration)) {

		if err := fc.LoadFiles(constants.InitDirFiles); err != nil {
			log.Printf("Error to update cache: %v", err)
			return FileInfo{}, false
		}

		file, found = fc.cache[sha1_hash]
		if !found {
			return FileInfo{}, false
		}
	}

	return file, true
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
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
