package main

import (
	"flag"
	"os"

	"geocaching/pkg/api"

	log "github.com/sirupsen/logrus"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose logging")
	region := flag.String("region", "", "Region ID to sync (required)")

	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if *region == "" {
		log.Error("You must specify a region ID with -region")
		flag.Usage()
		os.Exit(1)
	}

	if err := api.RunSolvedSyncForRegion(*region); err != nil {
		log.Fatalf("Failed to sync solved caches: %v", err)
	}
}
