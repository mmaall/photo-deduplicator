package main

import (
	//"fmt"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	//"errors"
	"github.com/sirupsen/logrus"
	"path/filepath"
	//"photo-deduplicator/helper"
)

var log = logrus.New()

func main() {
	log.Out = os.Stdout

	// You could set this to any `io.Writer` such as a file
	file, err := os.OpenFile("photo-deduplicator.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	log.WithFields(logrus.Fields{"agent": "main"})
	log.Info("Program starting")

	// Initializations
	photoDirectory := "/home/michael/Pictures"
	var photoMap map[string][]string
	photoMap = make(map[string][]string)
	// Get a list of photos

	photoList, err := GetPhotos(photoDirectory)

	if err != nil {
		log.Fatal("Error getting photos list")
		panic(err)
	}

	// Iterate through all the photos
	log.Info("Printing Files")
	for _, photo := range photoList {
		//log.Info(photo)

		file, err := os.Open(photo)
		if err != nil {
			log.Fatal(err)
		}

		// Hash the file
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			log.Fatal(err)
		}
		// Turn the hash into a string
		sha := base64.URLEncoding.EncodeToString(h.Sum(nil))

		// Close the file, not needed anymore
		file.Close()

		// Check what is there
		photoList := photoMap[sha]

		// Empty list, include new one
		if len(photoList) == 0 {
			// Add photo to list
			photoList := append(photoList, photo)
			photoMap[sha] = photoList
		} else {
			// Collision in map, fully verify they are the same

			log.Info("Collision found with ", photo)

			duplicateFound := false
			var duplicateFile string

			// Iterate through photos and see if we have a match
			// Search for duplicates
			for _, foundPhoto := range photoList {
				log.Info("\t", foundPhoto)
				duplicateFound, _ = PhotoCompare(photo, foundPhoto)
				if duplicateFound {
					duplicateFile = foundPhoto
				}

			}

			// No collision found, add hash
			if !duplicateFound {
				photoList := append(photoList, photo)
				photoMap[sha] = photoList
			} else {
				log.Info("Duplicate confirmed: ", photo, " == ", duplicateFile)
			}
		}

	}

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
