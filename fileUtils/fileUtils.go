package fileUtils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
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

func (fc *FileCache) StartPeriodicCacheUpdate(dir string, interval time.Duration) {
	go func() {
		for {

			log.Printf("Update caching...")
			err := fc.LoadFiles(dir)
			if err != nil {
				log.Fatalf("Falied to update chache: %v", err)
			}
			time.Sleep(interval)
		}
	}()
}

func (c *FileCache) GetAllFiles() []FileInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var files []FileInfo
	for _, file := range c.cache {
		if time.Now().Before(c.expiration[file.SHA1Hash]) {
			files = append(files, file)
		}
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
	defer fc.mu.RUnlock()
	file, found := fc.cache[sha1_hash]
	if !found {
		return FileInfo{}, false
	}

	if time.Now().After(fc.expiration[sha1_hash]) {

		delete(fc.cache, sha1_hash)
		delete(fc.expiration, sha1_hash)
		return FileInfo{}, false
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
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
