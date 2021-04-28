package main

import (
	//"fmt"
	"crypto/sha256"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var logger = log.New()

const (
	WORKER_THREADS = 4
)

// Holds key value pairs
type pair struct {
	key, val string
}

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
	//photoDirectory := "/home/michael/Pictures"
	photoDirectory := "photos/"

	photoList, err := GetPhotos(photoDirectory)

	if err != nil {
		logger.Fatal("Error getting photos list")
		panic(err)
	}

	// Channel file names are pushed onto this channel
	photoChannel := make(chan string)
	var photoWaitGroup sync.WaitGroup
	photoWaitGroup.Add(WORKER_THREADS)

	// Hashed files and correpsonding file name pushed onto this channel
	keyValueChannel := make(chan pair)
	var hashingWaitGroup sync.WaitGroup
	hashingWaitGroup.Add(1)

	// Spawn some go routines to do the hashing
	for i := 0; i < WORKER_THREADS; i++ {
		go ProcessPhoto(i, photoChannel, keyValueChannel, &photoWaitGroup)
	}

	// Create Map
	photoMap := make(map[string]string)

	// Spawn the go routine to store the hashes
	go AddToMap(keyValueChannel, &hashingWaitGroup, &photoMap)

	// Iterate through all the photos
	logger.Info("Printing Files")
	for _, photo := range photoList {
		photoChannel <- photo
	}
	close(photoChannel)
	logger.Info("Photo channel closed")

	// Wait for all the photos to be processed
	photoWaitGroup.Wait()
	// Close the channel
	close(keyValueChannel)
	// Wait for all the hashing
	hashingWaitGroup.Wait()

}

// Get a photo of of the channel, check the photo map, write if possible
func ProcessPhoto(routineId int, inputChannel chan string, outputChannel chan pair, photoWaitGroup *sync.WaitGroup) {
	logger.Info("Starting Go Routine ", routineId)
	for fileName := range inputChannel {
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
		// Close the file, not needed anymore
		file.Close()

		// Check if hash exists

		var keyValue pair = pair{sha, fileName}

		outputChannel <- keyValue

	}
	logger.Info("Worker ", routineId, " done")
	photoWaitGroup.Done()
	return
}

// Read pairs off of a channel, add them to the map if they don't already exist
// Identify when a collision has occured
func AddToMap(inputChannel chan pair, hashingWaitGroup *sync.WaitGroup, photoMap *map[string]string) {
	for keyValuePair := range inputChannel {

		collidedFile := (*photoMap)[keyValuePair.key]

		if collidedFile == "" {
			// Write to map
			(*photoMap)[keyValuePair.key] = keyValuePair.val
		} else {
			logger.Info("Collision: ", keyValuePair.val, " == ", collidedFile)
		}
	}

	hashingWaitGroup.Done()
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
