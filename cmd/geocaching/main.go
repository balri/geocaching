package main

import (
	"flag"

	"geocaching/pkg/api"

	log "github.com/sirupsen/logrus"
)

var (
	listenAddress = ":80"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose logging")

	flag.Parse()
	if *verbose {
		// Set the log level to debug
		log.SetLevel(log.DebugLevel)
	}
	// Set the log format to include a leading timestamp in ISO8601 format
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if err := api.RunSolvedSync(); err != nil {
		log.Fatalf("Failed to sync solved caches: %v", err)
	}
}
