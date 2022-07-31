package main

import (
	"fmt"
	"os"
	"photo-deduplicator/internal/deduplicator"
	"sync"

	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {

	// Default values
	// TODO: Maybe move these to parameter store and then this can all be stored in a config
	// Maybe an extract parameters function that either uses parameter store during ci/cd
	var (
		help                bool
		verbose             bool
		hashingRoutineCount = 4
		directory           = "photos/"
		logFileName         = ""
	)

	// Take in arguments
	getopt.FlagLong(&help, "help", 'h', "Help")
	getopt.FlagLong(&verbose, "verbose", 'v', "Verbose printing")
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

	// Set to verbose log level if turned on
	if verbose {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("Debug level set")

	}

	// List out the arguments
	log.Info("**Application Configuration**")
	log.Info("Hashing Routines: ", hashingRoutineCount)
	log.Info("Directory: ", directory)
	log.Info("Log file: ", logFileName)

	deduper := deduplicator.New(directory, hashingRoutineCount)
	photoChannel := make(chan deduplicator.DedupeFileMetadata, 15)
	var photoWaitGroup sync.WaitGroup

	deduper.Serve(photoChannel, &photoWaitGroup)

	totalDuplicates := 0

	for photoMetadata := range photoChannel {

		if photoMetadata.DuplicatePath != "" {
			totalDuplicates += 1
			fmt.Printf("%s is a duplicate of %s\n", photoMetadata.Path, photoMetadata.DuplicatePath)
		}

	}

	photoWaitGroup.Wait()

}
