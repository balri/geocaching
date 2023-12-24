package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"geocaching/pkg/cacheodon"
)

func getCaches(params QueryParams) ([]cacheodon.Geocache, error) {
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
	if params.Radius > 0 {
		log.Printf("Using radius %d", params.Radius)
		searchTerms.RadiusMeters = params.Radius
	}

	return api.Search(searchTerms)
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	rad, _ := strconv.Atoi(r.URL.Query().Get("radius"))
	params := QueryParams{
		Radius: rad,
	}
	caches, err := getCaches(params)
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

func getUnsolved(w http.ResponseWriter, r *http.Request) {
	// TODO: only get mysteries without corrected coords
	params := QueryParams{
		Radius:                  250,
		CacheTypes:              []int{CacheTypes["Unknown"]},
		ShowCorrectedCoordsOnly: false,
	}
	caches, err := getCaches(params)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var filteredCaches []cacheodon.Geocache
	for _, cache := range caches {
		for _, att := range cache.Attributes {
			if att.ID == getAttributeID("Challenge Cache") && att.IsApplicable {
				continue
			}
			if att.ID == getAttributeID("Field Puzzle") && att.IsApplicable {
				continue
			}
			if att.ID == getAttributeID("Bonus cache") && att.IsApplicable {
				continue
			}
			filteredCaches = append(filteredCaches, cache)
		}
	}

	log.Printf("Found %d unsolved caches", len(filteredCaches))
	payload, err := json.Marshal(filteredCaches)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	sendResponse(w, 200, []byte(payload))
}

func getAttributeID(searchName string) int {
	for idx, name := range CacheAttributes {
		if name == searchName {
			return idx
		}
	}

	return -1
}

func sendResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
