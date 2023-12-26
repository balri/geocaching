package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"geocaching/pkg/cacheodon"
)

func getSearchTerms(rad int) cacheodon.SearchTerms {
	config, err := cacheodon.NewDatastore("config.toml")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	params := config.Store.SearchTerms
	if rad > 0 {
		log.Printf("Using radius %d", rad)
		params.RadiusMeters = rad
	}

	params.CacheTypes = []int{
		CacheTypes["Traditional"],
		CacheTypes["Multi"],
		CacheTypes["Virtual"],
		CacheTypes["Letterbox"],
		CacheTypes["Unknown"],
		CacheTypes["Webcam"],
		CacheTypes["Earthcache"],
		CacheTypes["Wherigo"],
	}

	return params
}

func getCaches(searchTerms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
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

	return api.Search(searchTerms)
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	rad, _ := strconv.Atoi(r.URL.Query().Get("radius"))
	params := getSearchTerms(rad)
	if rad > 0 {
		log.Printf("Using radius %d", rad)
		params.RadiusMeters = rad
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

func filterCaches(cache cacheodon.Geocache) bool {
	if strings.Contains(cache.Name, "Bonus") {
		return false
	}
	for _, att := range cache.Attributes {
		if att.ID == getAttributeID("Challenge Cache") && att.IsApplicable {
			return false
		}
		if att.ID == getAttributeID("Field Puzzle") && att.IsApplicable {
			return false
		}
		if att.ID == getAttributeID("Bonus cache") && att.IsApplicable {
			return false
		}
		if att.ID == getAttributeID("Wireless Beacon") && att.IsApplicable {
			// Not sure about this but it seems logical right?
			return false
		}
	}

	return true
}

func getUnsolved(w http.ResponseWriter, r *http.Request) {
	// TODO: only get mysteries without corrected coords
	rad, _ := strconv.Atoi(r.URL.Query().Get("radius"))
	params := getSearchTerms(rad)
	params.CacheTypes = []int{CacheTypes["Unknown"]}
	params.ShowCorrectedCoordsOnly = "0"
	caches, err := getCaches(params)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var filteredCaches []cacheodon.Geocache
	for _, cache := range caches {
		if filterCaches(cache) {
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
