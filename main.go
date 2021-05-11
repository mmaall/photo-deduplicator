package main

import (
	//"fmt"
	"crypto/sha256"
	"encoding/base64"
	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sync"
	// "time"
)

// Holds key value pairs
type pair struct {
	key, val string
}

var log = logrus.New()

func main() {

	// Default values
	var (
		help                bool
		hashingRoutineCount = 4
		directory           = "photos/"
		logFileName         = ""
	)

	// Take in arguments
	getopt.FlagLong(&help, "help", 'h', "Help")
	getopt.FlagLong(&hashingRoutineCount, "hashingRoutineCount", 'c', "Number of routines hashing the files.")
	getopt.FlagLong(&directory, "directory", 'd', "Directory to deduplicate.")
	getopt.FlagLong(&logFileName, "logFile", 'L', "Log file")

	// Parse arguments
	getopt.Parse()

	// Print help and exit if help exists
	if help {
		getopt.Usage()
		os.Exit(0)
	}

	// Initialize logging

	// See if a log file was provided
	if logFileName != "" {
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.Out = logFile
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}
	log.WithFields(logrus.Fields{"agent": "main"})

	// List out the arguments
	log.Info("**Application Configuration**")
	log.Info("Hashing Routines: ", hashingRoutineCount)
	log.Info("Directory: ", directory)
	log.Info("Log file: ", logFileName)

	photoList, err := GetPhotos(directory)

	if err != nil {
		log.Fatal("Error getting photos list")
		panic(err)
	}

	// Channel file names are pushed onto this channel
	photoChannel := make(chan string)
	// Wait group to verify all photos have been collected
	var photoWaitGroup sync.WaitGroup
	photoWaitGroup.Add(hashingRoutineCount)

	// Hashed files and correpsonding file name pushed onto this channel
	keyValueChannel := make(chan pair)
	// Wait group to verify all photos have been hashed
	var hashingWaitGroup sync.WaitGroup
	hashingWaitGroup.Add(1)

	// Spawn some go routines to do the hashing
	for i := 0; i < hashingRoutineCount; i++ {
		go ProcessPhoto(i, photoChannel, keyValueChannel, &photoWaitGroup)
	}

	// Create holding file hashes
	photoMap := make(map[string]string)

	// Spawn the go routine to store the hashes
	go AddToMap(keyValueChannel, &hashingWaitGroup, &photoMap)

	// Iterate through all the photos
	log.Info("Iterate through photos")
	for _, photo := range photoList {
		photoChannel <- photo
	}
	close(photoChannel)
	log.Info("Photo channel closed")

	// Wait for all the photos to be processed
	photoWaitGroup.Wait()
	// Close the channel
	close(keyValueChannel)
	// Wait for all the hashing
	hashingWaitGroup.Wait()

}

// Get a photo of of the channel, check the photo map, write if possible
func ProcessPhoto(routineId int, inputChannel chan string, outputChannel chan pair, photoWaitGroup *sync.WaitGroup) {
	log.Info("Starting Go Routine ", routineId)
	for fileName := range inputChannel {
		// Open file
		file, err := os.Open(fileName)
		if err != nil {
			log.Error("Issue opening ", fileName)
			log.Error(err)
		}

		// Hash file
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			log.Error("Issue copying file ", fileName)
			log.Error(err)
		}
		// Turn the hash into a string
		sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
		// Close the file, not needed anymore
		file.Close()

		// Check if hash exists

		var keyValue pair = pair{sha, fileName}

		outputChannel <- keyValue

	}
	log.Info("Worker ", routineId, " done")
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
			log.Info("Collision: ", keyValuePair.val, " == ", collidedFile)
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
