package deduplicator

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
)

type PhotoDeduplicator struct {
	directory       string
	photoMap        map[string]string
	hashingRoutines int
}

type DedupeFileMetadata struct {
	Path          string
	DuplicatePath string
}

// Holds key value pairs
type pair struct {
	key, val string
}

// Create a new photo deduplicator
func New(directory string, hashingRoutines int) *PhotoDeduplicator {

	return &PhotoDeduplicator{
		directory:       directory,
		photoMap:        make(map[string]string),
		hashingRoutines: hashingRoutines,
	}
}

// Run the deduplication
// a channel is passed to the function which will serve details about the photos being processed
// waitgroup will notify when all photos have been processed
func (deduplicator *PhotoDeduplicator) Serve(dedupedPhotoChannel chan<- DedupeFileMetadata, dedupedPhotoWaitGroup *sync.WaitGroup) {

	photoList, err := getPhotos(deduplicator.directory)

	if err != nil {
		log.Fatal("Error getting photos list")
		panic(err)
	}

	// Channel file names are pushed onto this channel
	photoChannel := make(chan string)
	// Wait group to verify all photos have been collected
	var photoWaitGroup sync.WaitGroup
	photoWaitGroup.Add(deduplicator.hashingRoutines)

	// Hashed files and correpsonding file name pushed onto this channel
	keyValueChannel := make(chan pair)
	// Wait group to verify all photos have been hashed
	var hashingWaitGroup sync.WaitGroup
	hashingWaitGroup.Add(1)

	// Add a waiter to the photo WaitGroup being provided so we can signal that all photos have been
	// processed
	dedupedPhotoWaitGroup.Add(1)

	// Spawn some go routines to do the hashing
	for i := 0; i < deduplicator.hashingRoutines; i++ {
		go processPhoto(i, photoChannel, keyValueChannel, &photoWaitGroup)
	}

	// Spawn the go routine to store the hashes
	go checkCollision(keyValueChannel, dedupedPhotoChannel, &hashingWaitGroup, &deduplicator.photoMap)

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

	// Close the output channel
	close(dedupedPhotoChannel)
	// Close the final waitgroup to signal all photos have been processed
	dedupedPhotoWaitGroup.Done()
}

// Receives a photo hashes it, and places it on a channel for further actions
func processPhoto(routineId int, inputChannel chan string, outputChannel chan pair, photoWaitGroup *sync.WaitGroup) {
	log.Info("Starting Go Routine ", routineId)
	for fileName := range inputChannel {

		hashedValue := hashPhoto(fileName)

		var keyValue pair = pair{hashedValue, fileName}

		outputChannel <- keyValue

	}
	log.Info("Hashing worker ", routineId, " done")
	photoWaitGroup.Done()
	return
}

// Read pairs off of a channel, add them to the map if they don't already exist
// Identify when a collision has occured
func checkCollision(inputChannel chan pair, outputChannel chan<- DedupeFileMetadata, hashingWaitGroup *sync.WaitGroup, photoMap *map[string]string) {
	for keyValuePair := range inputChannel {

		fileMetadata := DedupeFileMetadata{
			Path:          keyValuePair.val,
			DuplicatePath: "",
		}

		collidedFile := (*photoMap)[keyValuePair.key]

		if collidedFile == "" {
			// Write to map
			(*photoMap)[keyValuePair.key] = keyValuePair.val
		} else {
			log.Info("Collision: ", keyValuePair.val, " == ", collidedFile)

			// Mark as duplicate
			fileMetadata.DuplicatePath = collidedFile
		}

		outputChannel <- fileMetadata
	}

	hashingWaitGroup.Done()
	return
}

func getPhotos(directory string) ([]string, error) {
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

// Helper function to hash a file, return hased value
func hashPhoto(fileName string) string {

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
	return sha
}
