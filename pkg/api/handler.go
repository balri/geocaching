package api

import (
	"encoding/json"
	"fmt"
	"geocaching/pkg/sheets"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cacheodon "github.com/balri/cacheodon/pkg/geocaching"
)

const GEOCACHE_URL_PREFIX = "https://coord.info/"

const searchLat = -27.4705
const searchLon = 153.0260

var (
	regions = map[string]string{
		"52": "New South Wales",
		"53": "Victoria",
		"54": "Queensland",
		"55": "South Australia",
		"56": "Western Australia",
		"57": "Tasmania",
		"58": "Northern Territory",
		"59": "Australian Capital Territory",
		"82": "North Island NZ",
		"86": "South Island NZ",
	}
)

var cacheTypes = map[cacheodon.CacheType]string{
	cacheodon.Traditional:    "Traditional",
	cacheodon.Multi:          "Multi",
	cacheodon.Virtual:        "Virtual",
	cacheodon.Letterbox:      "Letterbox",
	cacheodon.Event:          "Event",
	cacheodon.Unknown:        "Unknown",
	cacheodon.APE:            "A.P.E. Cache",
	cacheodon.Webcam:         "Webcam",
	cacheodon.Locationless:   "Locationless",
	cacheodon.CITO:           "CITO",
	cacheodon.Earthcache:     "Earthcache",
	cacheodon.Mega:           "Mega",
	cacheodon.GPSMaze:        "GPS Maze",
	cacheodon.Wherigo:        "Wherigo",
	cacheodon.CommunityEvent: "Community Event",
	cacheodon.HQCache:        "HQ Cache",
	cacheodon.HQCelebration:  "HQ Celebration",
	cacheodon.BlockParty:     "Block Party",
	cacheodon.Giga:           "Giga",
}

var cacheSizes = map[cacheodon.CacheSize]string{
	cacheodon.NotChosen:   "Not chosen",
	cacheodon.Micro:       "Micro",
	cacheodon.Regular:     "Regular",
	cacheodon.Large:       "Large",
	cacheodon.VirtualSize: "Virtual",
	cacheodon.Other:       "Other",
	cacheodon.Small:       "Small",
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

func getDefaultSearchTerms(rad int) cacheodon.SearchTerms {
	params := cacheodon.SearchTerms{
		Latitude:      searchLat,
		Longitude:     searchLon,
		RadiusMeters:  1000,
		IgnorePremium: false,
	}
	params.ShowDisabled = BoolPtr(false)
	params.SortAsc = BoolPtr(true)
	params.Sort = "distance"
	params.OriginType = "query"
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

func RunSolvedSync() error {
	for regionID, region := range regions {
		log.Printf("Syncing solved caches for region: %s", region)
		if err := runSolved(regionID, region); err != nil {
			return fmt.Errorf("failed to sync region %s: %w", region, err)
		}
	}
	return nil
}

func runSolved(regionID, region string) error {
	params := cacheodon.SearchTerms{
		CacheType: []cacheodon.CacheType{
			cacheodon.Unknown,
			cacheodon.Multi,
			cacheodon.Letterbox,
			cacheodon.Wherigo,
		},
		IgnorePremium: false,
		Corrected:     BoolPtr(true),
		HideOwned:     BoolPtr(true),
		SortAsc:       BoolPtr(true),
		Sort:          "distance",
		OriginType:    cacheodon.Region,
		OriginID:      regionID,
	}

	caches, err := getCaches(params)
	if err != nil {
		return err
	}
	log.Printf("Found %d solved caches", len(caches))

	sheet := sheets.NewSheetClient(
		os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		os.Getenv("SPREADSHEET_ID"),
		region,
	)
	if err := sheet.EnsureSheetExists(); err != nil {
		return err
	}

	existingCodes := sheet.GetExistingCodes()

	numCaches := 0
	var rows [][]interface{}
	for _, cache := range caches {
		if existingCodes[cache.Code] {
			continue
		}
		postedCoords := formatCoords(
			cache.PostedCoordinates.Latitude,
			cache.PostedCoordinates.Longitude,
		)
		correctedCoords := formatCoords(
			cache.UserCorrectedCoordinates.Latitude,
			cache.UserCorrectedCoordinates.Longitude,
		)
		if postedCoords == correctedCoords {
			continue
		}
		cacheType, ok := cacheTypes[cacheodon.CacheType(cache.GeocacheType)]
		if !ok {
			cacheType = ""
		}
		cacheSize, ok := cacheSizes[cacheodon.CacheSize(cache.ContainerType)]
		if !ok {
			cacheSize = ""
		}
		cacheFound := ""
		if cache.UserFound {
			cacheFound = "Yes"
		}
		link := fmt.Sprintf(`=HYPERLINK("%s%s", "%s")`, GEOCACHE_URL_PREFIX, cache.Code, cache.Code)
		distance := math.Round(haversine(searchLat, searchLon, cache.PostedCoordinates.Latitude, cache.PostedCoordinates.Longitude)*100) / 100
		row := []interface{}{
			link,
			cache.Name,
			cache.FavoritePoints,
			postedCoords,
			correctedCoords,
			distance,
			formatDateForSheets(cache.PlacedDate),
			cacheType,
			cacheSize,
			cache.Difficulty,
			cache.Terrain,
			cache.Owner.Username,
			cache.Region,
			cache.Country,
			cacheFound,
		}
		rows = append(rows, row)
		numCaches++
	}

	if len(rows) > 0 {
		const batchSize = 1000
		for i := 0; i < len(rows); i += batchSize {
			end := i + batchSize
			if end > len(rows) {
				end = len(rows)
			}
			sheet.AppendRows(rows[i:end])
		}
	}

	log.Printf("Added %d new solved caches to the sheet", numCaches)

	return nil
}

func RunSolvedSyncForRegion(regionID string) error {
	region, ok := regions[regionID]
	if !ok {
		return fmt.Errorf("unknown region ID: %s", regionID)
	}
	log.Printf("Syncing solved caches for region: %s", region)
	return runSolved(regionID, region)
}

func sendResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func formatCoords(lat, lon float64) string {
	if lat == 0 && lon == 0 {
		return ""
	}
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

func formatDateForSheets(dateStr string) interface{} {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		return dateStr // fallback to original if parsing fails
	}
	// Google Sheets serial date: days since 1899-12-30
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	days := t.Sub(base).Hours() / 24
	return days
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
