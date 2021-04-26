package main

import (
	//"fmt"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	//"errors"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
)

var logger = log.New()

const (
	WORKER_THREADS = 4
)

func main() {

	logFile, err := os.OpenFile("logger.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = logFile
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.WithFields(log.Fields{"agent": "main"})
	logger.Info("Program starting")

	// Initializations
	photoDirectory := "/home/michael/Pictures"
	// photoDirectory := "photos/"

	// Initialize photomap
	var photoMap *SafeMap = NewSafeMap()

	photoList, err := GetPhotos(photoDirectory)

	if err != nil {
		logger.Fatal("Error getting photos list")
		panic(err)
	}

	photoChannel := make(chan string)
	var waitGroup sync.WaitGroup
	waitGroup.Add(WORKER_THREADS)

	// Spawn some go routines
	for i := 0; i < WORKER_THREADS; i++ {
		go ProcessPhoto(i, photoChannel, &waitGroup, photoMap)
	}

	// Iterate through all the photos
	logger.Info("Printing Files")
	for _, photo := range photoList {
		photoChannel <- photo
	}
	close(photoChannel)
	logger.Info("Photo channel closed")

	waitGroup.Wait()
}

// Get a photo of of the channel, check the photo map, write if possible
func ProcessPhoto(routineId int, inputChannel chan string, waitGroup *sync.WaitGroup, photoMap *SafeMap) {
	logger.Info("Starting Go Routine ", routineId)
	for fileName := range inputChannel {
		//logger.Info("Routine ", routineId, ": Processing ", fileName)
		// Open file
		file, err := os.Open(fileName)
		if err != nil {
			logger.Error("Issue opening ", fileName)
			logger.Error(err)
		}

		// Hash file
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			logger.Error("Issue copying file ", fileName)
			logger.Error(err)
		}
		// Turn the hash into a string
		sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
		//logger.Info("Routine ", routineId, ": ", sha)
		// Close the file, not needed anymore
		file.Close()

		// Check if hash exists
		collidedFile := photoMap.WriteUnique(sha, fileName)

		if collidedFile != "" {
			// collission
			logger.Info("Routine ", routineId, " Collision: ", fileName, " == ", collidedFile)
		}

	}
	logger.Info("Worker ", routineId, " done")
	waitGroup.Done()
	return
}

func GetPhotos(directory string) ([]string, error) {
	var photos []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		// Check errors
		if err != nil {
			return err
		}

		// Not going to include directories
		if info.IsDir() {
			return nil
		}

		// TODO: Weed out file types
		photos = append(photos, path)
		return nil
	})

	return photos, err
}

// Compare two photos
// TODO: Compare the file themselves, not their pahts
func PhotoCompare(photo1, photo2 string) (bool, error) {
	// Split the paths
	_, file1 := filepath.Split(photo1)
	_, file2 := filepath.Split(photo2)

	return file1 == file2, nil
}

// Thread safe map structure
type SafeMap struct {
	_map    map[string]string
	mapLock *sync.RWMutex
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		_map:    make(map[string]string),
		mapLock: &sync.RWMutex{},
	}
}

func (sm *SafeMap) Read(key string) string {
	sm.mapLock.RLock()
	defer sm.mapLock.RUnlock()
	return sm._map[key]

}

func (sm *SafeMap) Write(key, value string) {
	sm.mapLock.Lock()
	defer sm.mapLock.Unlock()
	sm._map[key] = value
}

// Write only succeeds if the key has never been seen before
// Returns collided value in case of collision. Otherwise empty string returned
func (sm *SafeMap) WriteUnique(key, value string) string {
	sm.mapLock.Lock()
	defer sm.mapLock.Unlock()
	existingValue := sm._map[key]
	// If default value, we're good to write
	if existingValue == "" {
		sm._map[key] = value
		return ""
	} else {
		// Found something, write fails
		return sm._map[key]
	}
}
