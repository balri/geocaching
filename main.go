package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	router := api.GetRouter()
	if router != nil {
		go startServer(router)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

mainloop:
	// In all cases, just exit and let the container restart from scratch.
	// There's less to get wrong doing it this way.
	for {
		select {
		case <-signalChan:
			log.Info("Signalled, breaking main loop")
			break mainloop
		}
	}
}

func startServer(router http.Handler) {
	server := http.Server{
		Addr:              listenAddress,
		Handler:           router,
		ReadHeaderTimeout: 2 * time.Second,
	}
	log.Info("listening for HTTP on: %s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServeError", err)
	}
}
