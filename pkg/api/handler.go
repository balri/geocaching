package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"geocaching/pkg/cacheodon"
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	radius := r.URL.Query().Get("radius")
	config, err := cacheodon.NewDatastore("config.toml")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	api, err := cacheodon.NewGeocachingAPI(config.Store.Configuration)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if err = api.Auth(os.Getenv("GEOCACHING_CLIENT_ID"), os.Getenv("GEOCACHING_CLIENT_SECRET")); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	searchTerms := config.Store.SearchTerms
	rad, err := strconv.Atoi(radius)
	if err == nil {
		log.Printf("Using radius %d", rad)
		searchTerms.RadiusMeters = rad
	}

	caches, err := api.Search(searchTerms)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Printf("Found %d caches", len(caches))
	payload, err := json.Marshal(caches)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	sendResponse(w, 200, []byte(payload))
}

func sendResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
