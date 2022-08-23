package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"photo-deduplicator/internal/deduplicator"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {

	// Default values
	var (
		help                bool
		verbose             bool
		hashingRoutineCount = 4
		inputDirectory      = "photos/"
		outputDirectory     = ""
		logFileName         = ""
		purge               = false
	)

	// Take in arguments
	getopt.FlagLong(&help, "help", 'h', "Help")
	getopt.FlagLong(&verbose, "verbose", 'v', "Verbose printing")
	getopt.FlagLong(&hashingRoutineCount, "hashingRoutineCount", 'c', "Number of routines hashing the files.")
	getopt.FlagLong(&inputDirectory, "input", 'i', "Directory to deduplicate.")
	getopt.FlagLong(&outputDirectory, "output", 'o', "Directory to store deduplicated files")
	getopt.FlagLong(&logFileName, "logFile", 'L', "Log file")
	getopt.FlagLong(&purge, "purge", 'p', "Purge deduplicated files")

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

	// Set to verbose log level if turned on
	if verbose {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("Debug level set")

	}

	// List out the arguments
	log.Info("**Application Configuration**")
	log.Info("Hashing Routines: ", hashingRoutineCount)
	log.Info("Input Directory: ", inputDirectory)
	log.Info("Output Directory: ", outputDirectory)
	log.Info("Purge: ", strconv.FormatBool(purge))
	log.Info("Log file: ", logFileName)

	// Data validation

	// Verify info
	inputDirectoryInfo, err := os.Stat(inputDirectory)
	if err != nil {
		// Error trying to read directory
		log.Errorf("Error reading the input directory %s\n", inputDirectory)
		log.Errorf("%s\n", err.Error())
		fmt.Printf("Error reading from input directory %s\n", inputDirectory)
		return
	}
	if !inputDirectoryInfo.IsDir() {
		// Not valid directory
		log.Errorf("Input directory (%s) is not a directory\n", inputDirectory)
		fmt.Printf("Looks like %s is not a directory. Exiting\n", inputDirectory)
		return
	}

	if outputDirectory != "" {
		// validate output directory if it exists
		outputDirectoryInfo, err := os.Stat(outputDirectory)
		if errors.Is(err, os.ErrNotExist) {
			// Create a new directory
			err := os.Mkdir(outputDirectory, 0750)
			if err != nil && !os.IsExist(err) {
				log.Errorf("%s\n", err.Error())
				fmt.Printf("Unable to create the output directory %s", outputDirectory)
			}
		} else if err != nil {
			// Error trying to read directory
			log.Errorf("Error reading the output directory %s\n", outputDirectory)
			log.Errorf("%s\n", err.Error())
			fmt.Printf("Error reading from output directory %s\n", outputDirectory)
			return
		} else if !outputDirectoryInfo.IsDir() {
			// Not a directory
			log.Fatalf("Output directory (%s) is not a directory\n", outputDirectory)
			fmt.Printf("Looks like output directory %s is not a directory. Exiting\n", outputDirectory)
			return
		}
	}

	// Start deduplication
	deduper := deduplicator.New(inputDirectory, hashingRoutineCount)
	deduper.SetBufferSize(50)
	photoChannel := make(chan deduplicator.DedupeFileMetadata, 100)
	var photoWaitGroup sync.WaitGroup

	deduper.Serve(photoChannel, &photoWaitGroup)

	totalDuplicates := 0

	failedCopies := []deduplicator.DedupeFileMetadata{}

	// Process photo channel
	for photoMetadata := range photoChannel {

		if photoMetadata.DuplicatePath != "" {
			totalDuplicates += 1
			fmt.Printf("%s is a duplicate of %s\n", photoMetadata.Path, photoMetadata.DuplicatePath)
			continue
		}

		if outputDirectory == "" {
			continue
		}

		// Copy the file to the new directory
		sourceFile, err := os.OpenFile(photoMetadata.Path, os.O_RDWR, 0666)

		if err != nil {
			// Error opening the file
			failedCopies = append(failedCopies, photoMetadata)
			log.Errorf("Error opening source photo (%s) (%s)\n", photoMetadata.Path, err.Error())
			continue
		}

		uuid, err := uuid.NewRandom()

		if err != nil {
			// Error generating UUID
			log.Errorf("Error generating uuid for photo name (%s)\n", err.Error())
			failedCopies = append(failedCopies, photoMetadata)
			continue
		}

		destinationFileName := filepath.Join(outputDirectory, (uuid.String() + ".jpg"))

		destinationFile, err := os.OpenFile(destinationFileName, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Errorf("Unable to open destination file %s(%s)\n", destinationFileName, err.Error())
			failedCopies = append(failedCopies, photoMetadata)
			continue
		}

		// Copy the photo
		if _, err := io.Copy(destinationFile, sourceFile); err != nil {
			log.Errorf("Unable to copy %s to %s (%s)\n", photoMetadata.Path, destinationFile, err.Error())
			failedCopies = append(failedCopies, photoMetadata)
			continue
		}

		// Flush to disk
		if err := destinationFile.Sync(); err != nil {
			log.Errorf("Unable to flush %s to disk (%s)\n", destinationFileName, err.Error())
			failedCopies = append(failedCopies, photoMetadata)
			continue
		}

		// Close file

		if err := destinationFile.Close(); err != nil {
			log.Errorf("Unable to close %s (%s)\n", destinationFileName, err.Error())
			continue
		}

		if purge {
			// delete the old file
		}

	}

	fmt.Println("Deduplicated", totalDuplicates, "photos")

	photoWaitGroup.Wait()

}
