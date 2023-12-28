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

func getDefaultSearchTerms(rad int) cacheodon.SearchTerms {
	config, err := cacheodon.NewDatastore("config.toml")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	params := config.Store.SearchTerms
	params.IgnorePremium = false
	params.ShowDisabled = "0"
	params.SortAsc = true
	params.Sort = "distance"
	params.OperationType = "query"
	params.HideOwned = "1"
	params.NotFoundBy = os.Getenv("GEOCACHING_CLIENT_ID")

	params.RadiusMeters = 20000
	if rad > 0 {
		log.Printf("Using radius %d", rad)
		params.RadiusMeters = rad
	}

	// By default get all standard types minus events
	params.CacheTypes = []int{
		cacheTypes["Traditional"],
		cacheTypes["Multi"],
		cacheTypes["Virtual"],
		cacheTypes["Letterbox"],
		cacheTypes["Unknown"],
		cacheTypes["Webcam"],
		cacheTypes["Earthcache"],
		cacheTypes["Wherigo"],
	}

	return params
}

func getUnsolvedSearchTerms(rad int) cacheodon.SearchTerms {
	params := getDefaultSearchTerms(rad)
	params.CacheTypes = []int{cacheTypes["Unknown"]}
	params.ShowCorrectedCoordsOnly = "0"

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
	params := getDefaultSearchTerms(rad)
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

func filterUnsolved(cache cacheodon.Geocache) bool {
	if strings.Contains(strings.ToLower(cache.Name), "bonus") {
		return false
	}

	excludedAttributes := map[int]bool{
		cacheAttributes["Challenge Cache"]: true,
		cacheAttributes["Field Puzzle"]:    true,
		cacheAttributes["Bonus cache"]:     true,
		// Not sure about this but it seems logical right?
		cacheAttributes["Wireless Beacon"]: true,
	}

	for _, att := range cache.Attributes {
		if _, hasAttr := excludedAttributes[att.ID]; hasAttr && att.IsApplicable {
			return false
		}
	}

	return true
}

func getUnsolved(w http.ResponseWriter, r *http.Request) {
	rad, _ := strconv.Atoi(r.URL.Query().Get("radius"))
	params := getUnsolvedSearchTerms(rad)
	caches, err := getCaches(params)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var filteredCaches []cacheodon.Geocache
	for _, cache := range caches {
		if filterUnsolved(cache) {
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

func sendResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
