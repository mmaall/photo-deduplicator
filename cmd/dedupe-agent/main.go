package main

import (
	//"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
	"os"
	"photo-deduplicator/internal/deduplicator"
)

var log = logrus.New()

func main() {

	// Default values
	// TODO: Maybe move these to parameter store and then this can all be stored in a config
	// Maybe an extract parameters function that either uses parameter store during ci/cd
	var (
		help                bool
		hashingRoutineCount = 4
		directory           = "photos/"
		logFileName         = ""
		dynamoTableName     = "PhotoHashTable"
		region              = "us-east-1"
	)

	// Take in arguments
	getopt.FlagLong(&help, "help", 'h', "Help")
	getopt.FlagLong(&hashingRoutineCount, "hashingRoutineCount", 'c', "Number of routines hashing the files.")
	getopt.FlagLong(&directory, "directory", 'd', "Directory to deduplicate.")
	getopt.FlagLong(&logFileName, "logFile", 'L', "Log file")
	getopt.FlagLong(&dynamoTableName, "dynamoTable", 't', "DynamoDB Table")
	getopt.FlagLong(&region, "region", 'r', "AWS region")

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

	deduper := deduplicator.New(directory, hashingRoutineCount)

	deduper.Run()

}
