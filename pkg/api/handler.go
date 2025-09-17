package api

import (
	"encoding/json"
	"fmt"
	"geocaching/pkg/sheets"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	cacheodon "github.com/balri/cacheodon/pkg/geocaching"
)

const GEOCACHE_URL_PREFIX = "https://coord.info/"

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

func getDefaultSearchTerms(rad int) cacheodon.SearchTerms {
	params := cacheodon.SearchTerms{
		Latitude:      -27.4705,
		Longitude:     153.0260,
		RadiusMeters:  1000,
		IgnorePremium: false,
	}
	params.ShowDisabled = BoolPtr(false)
	params.SortAsc = true
	params.Sort = "distance"
	params.OperationType = "query"
	params.HideOwned = BoolPtr(true)
	params.NotFoundBy = []string{os.Getenv("GEOCACHING_CLIENT_ID")}

	if rad > 0 {
		log.Printf("Using radius %d", rad)
		params.RadiusMeters = rad
	}

	// By default get all standard types minus events
	params.CacheType = []cacheodon.CacheType{
		cacheodon.Traditional,
		cacheodon.Multi,
		cacheodon.Virtual,
		cacheodon.Letterbox,
		cacheodon.Unknown,
		cacheodon.Webcam,
		cacheodon.Earthcache,
		cacheodon.Wherigo,
	}

	return params
}

func getUnsolvedSearchTerms(rad int) cacheodon.SearchTerms {
	params := getDefaultSearchTerms(rad)
	params.CacheType = []cacheodon.CacheType{cacheodon.Unknown}
	params.Corrected = BoolPtr(false)

	return params
}

func getClient() (*cacheodon.GeocachingAPI, error) {
	config := cacheodon.APIConfig{
		GeocachingAPIURL: "https://www.geocaching.com",
	}

	client, err := cacheodon.NewGeocachingAPI(config)
	if err != nil {
		return nil, err
	}

	err = client.Auth(
		os.Getenv("GEOCACHING_CLIENT_ID"),
		os.Getenv("GEOCACHING_CLIENT_SECRET"),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getCaches(searchTerms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
	api, err := getClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return api.Search(searchTerms)
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	rad, err := strconv.Atoi(r.URL.Query().Get("radius"))
	if err != nil || rad <= 0 {
		rad = 25000 // 25km
	}
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

	// Build a map of excluded attribute IDs for fast lookup
	excludedIDs := map[int]bool{
		int(cacheodon.ChallengeCache): true,
		int(cacheodon.FieldPuzzle):    true,
		int(cacheodon.BonusCache):     true,
		int(cacheodon.WirelessBeacon): true,
	}

	for _, att := range cache.Attributes {
		if excludedIDs[att.ID] && att.IsApplicable {
			return false
		}
	}

	return true
}

func getUnsolved(w http.ResponseWriter, r *http.Request) {
	rad, err := strconv.Atoi(r.URL.Query().Get("radius"))
	if err != nil || rad <= 0 {
		rad = 100000 // 100km
	}
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

func getSolved(w http.ResponseWriter, r *http.Request) {
	rad, err := strconv.Atoi(r.URL.Query().Get("radius"))
	if err != nil || rad <= 0 {
		rad = 2500 // 2.5km @todo increase this once we're happy with it
	}
	params := getDefaultSearchTerms(rad)
	params.CacheType = []cacheodon.CacheType{cacheodon.Unknown}
	params.Corrected = BoolPtr(true)
	params.NotFoundBy = []string{}

	caches, err := getCaches(params)

	sheet := sheets.NewSheetClient(
		os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		os.Getenv("SPREADSHEET_ID"),
		os.Getenv("SPREADSHEET_SHEET_NAME"),
	)
	if err := sheet.EnsureSheetExists(); err != nil {
		log.Fatalf("Failed to ensure sheet exists: %v", err)
	}

	existingCodes := sheet.GetExistingCodes()

	for _, cache := range caches {
		if existingCodes[cache.Code] {
			continue // Skip if already present
		}
		row := []interface{}{
			cache.Code,
			GEOCACHE_URL_PREFIX + cache.Code,
			cache.Name,
			formatCoords(
				cache.UserCorrectedCoordinates.Latitude,
				cache.UserCorrectedCoordinates.Longitude,
			),
			cache.Difficulty,
			cache.Terrain,
		}
		sheet.AppendRow(row)
	}

	log.Printf("Found %d solved caches", len(caches))
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

func formatCoords(lat, lon float64) string {
	latDir := "N"
	if lat < 0 {
		latDir = "S"
		lat = -lat
	}
	lonDir := "E"
	if lon < 0 {
		lonDir = "W"
		lon = -lon
	}
	latDeg := int(lat)
	latMin := (lat - float64(latDeg)) * 60
	lonDeg := int(lon)
	lonMin := (lon - float64(lonDeg)) * 60
	return fmt.Sprintf("%s%d %06.3f %s%d %06.3f", latDir, latDeg, latMin, lonDir, lonDeg, lonMin)
}
