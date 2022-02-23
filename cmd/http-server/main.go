package main

import (
	"os"
	"photo-deduplicator/internal/restapi"

	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {

	var (
		verbose     bool
		logFileName = ""
		help        bool
	)

	// Take in arguments
	getopt.FlagLong(&help, "help", 'h', "Help")
	getopt.FlagLong(&verbose, "verbose", 'v', "Verbose printing")
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
	log.Infof("**Application Configuration**")
	log.Infof("Log file: %s", logFileName)

	log.Infof("Starting Http Server")

	httpServer, _ := restapi.NewDeduplicatorAPI()

	httpServer.Start()

}
