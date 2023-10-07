package cacheodon

import (
	"os"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/microcosm-cc/bluemonday"
)

type GeocachingAPI struct {
	config           APIConfig
	client           *RLHTTPClient
	cookieJar        *cookiejar.Jar
	blueMondayPolicy *bluemonday.Policy
}

func NewGeocachingAPI(c APIConfig) (*GeocachingAPI, error) {
	var err error
	g := &GeocachingAPI{config: c}
	g.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	var proxyUrl *url.URL
	if c.HTTPProxyURL != "" {
		log.Println("Connecting with proxy:", c.HTTPProxyURL)
		proxyUrl, err = url.Parse(c.HTTPProxyURL)
		if err != nil {
			return nil, err
		}
	}

	var limiter *rate.Limiter

	if !c.UnThrottle {
		limiter = rate.NewLimiter(rate.Every(1*time.Second), 1)
	} else {
		limiter = rate.NewLimiter(rate.Inf, 1)
	}

	g.client = &RLHTTPClient{
		client: &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
			Jar:       g.cookieJar,
		},
		Ratelimiter: limiter,
	}
	g.blueMondayPolicy = bluemonday.StrictPolicy()
	return g, nil
}

// Broadly copying from https://github.com/btittelbach/gctools/blob/master/geocachingsitelib.py and
// https://github.com/cgeo/cgeo/blob/master/main/src/main/java/cgeo/geocaching/connector/gc/GCWebAPI.java
func (g *GeocachingAPI) Auth(clientID, clientSecret string) error {
	var err error

	// First we have to initiate a request to https://www.geocaching.com/account/signin
	// to obtain a "__RequestVerificationToken" value.
	RVTReq, err := http.NewRequest("GET", g.config.GeocachingAPIURL+"/account/signin", nil)
	if err != nil {
		return err
	}
	log.Debug("Request: __RequestVerificationToken")
	RVTResp, err := g.client.Do(RVTReq)
	if err != nil {
		return err
	}
	defer RVTResp.Body.Close()

	RVT := ""

	RVTBody, err := io.ReadAll(RVTResp.Body)
	if err != nil {
		return err
	}

	// These bastards hide the token in a hidden field in the page. There's a cookie by the same name,
	// but it isn't used for authentication, as far as I ca tell.
	rgx := regexp.MustCompile("name=\"__RequestVerificationToken\"\\s+type=\"hidden\"\\s+value=\"([^\"]+)\"")
	matches := rgx.FindStringSubmatch(string(RVTBody))
	if len(matches) < 1 {
		return fmt.Errorf("could not find __RequestVerificationToken")
	}
	RVT = matches[1]

	params := url.Values{}
	params.Add("__RequestVerificationToken", RVT)
	params.Add("ReturnUrl", `/play`)
	params.Add("UsernameOrEmail", clientID)
	params.Add("Password", clientSecret)
	body := strings.NewReader(params.Encode())

	POSTReq, err := http.NewRequest("POST", g.config.GeocachingAPIURL+"/account/signin", body)
	if err != nil {
		return err
	}
	POSTReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0")
	POSTReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	POSTReq.Header.Set("Accept-Language", "en-US,en;q=0.5")
	POSTReq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	POSTReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	POSTReq.Header.Set("Origin", g.config.GeocachingAPIURL+"")
	POSTReq.Header.Set("Dnt", "1")
	POSTReq.Header.Set("Connection", "keep-alive")
	POSTReq.Header.Set("Referer", g.config.GeocachingAPIURL+"/account/signin?returnUrl=%2fplay")
	POSTReq.Header.Set("Upgrade-Insecure-Requests", "1")
	POSTReq.Header.Set("Sec-Fetch-Dest", "document")
	POSTReq.Header.Set("Sec-Fetch-Mode", "navigate")
	POSTReq.Header.Set("Sec-Fetch-Site", "same-origin")
	POSTReq.Header.Set("Sec-Fetch-User", "?1")

	log.Debug("Request: Authentication")
	POSTResp, err := g.client.Do(POSTReq)
	if err != nil {
		return err
	}
	defer POSTResp.Body.Close()
	if body, err := io.ReadAll(POSTResp.Body); err == nil {
		if match, err := regexp.Match("It seems your Anti-Forgery Token is invalid", body); err == nil && match {
			return fmt.Errorf("Anti-Forgery Token is invalid")
		}
		if match, err := regexp.Match(`"isLoggedIn": true,`, body); err != nil || !match {
			// Print the body
			// log.Println(string(body))
			return fmt.Errorf("login failed")
		}

	} else {
		return fmt.Errorf("couldn't read body")
	}

	log.Println("Authenticated to", g.config.GeocachingAPIURL)
	return nil
}

type GocachePostedCoordinates struct {
	Latitude  float64 `json:"latitude" fake:"{number:1,180}"`
	Longitude float64 `json:"longitude" fake:"{number:1,180}"`
}

type GeocacheOwner struct {
	Code     string `json:"code" fake:"{regex:PR[1-9]{5}}"`
	Username string `json:"username" fake:"{username}"`
}

type GeocacheAttributes struct {
	ID           int    `json:"id" fake:"{number:1,100}"`
	Name         string `json:"name" fake:"{loremipsumword:1}"`
	IsApplicable bool   `json:"isApplicable" fake:"{bool}"`
}

type Geocache struct {
	ID                int                      `json:"id" fake:"{number:1,100000}"`
	Name              string                   `json:"name" fake:"{loremipsumword:1}"`
	Code              string                   `json:"code" fake:"{regex:GC[1-9]{5}}"` // GC12345
	PremiumOnly       bool                     `json:"premiumOnly" fake:"{bool}"`
	FavoritePoints    int                      `json:"favoritePoints" fake:"{number:1,1000}"`
	GeocacheType      int                      `json:"geocacheType" fake:"{number:1,10}"`
	ContainerType     int                      `json:"containerType" fake:"{number:1,10}"`
	Difficulty        float64                  `json:"difficulty" fake:"{number:1,5}"`
	Terrain           float64                  `json:"terrain" fake:"{number:1,5}"`
	UserFound         bool                     `json:"userFound" fake:"bool"`
	CacheStatus       int                      `json:"cacheStatus" fake:"{number:1,10}"`
	PostedCoordinates GocachePostedCoordinates `json:"postedCoordinates"`
	DetailsURL        string                   `json:"detailsUrl" fake:"{url}"`
	HasGeotour        bool                     `json:"hasGeotour" fake:"{bool}"`
	PlacedDate        string                   `json:"placedDate" fake:"{date}"`
	Owner             GeocacheOwner            `json:"owner"`
	LastFoundDate     string                   `json:"lastFoundDate" fake:"{date}"`
	TrackableCount    int                      `json:"trackableCount" fake:"{number:1,100}"`
	Region            string                   `json:"region" fake:"{city}"`
	Country           string                   `json:"country" fake:"{country}"`
	Attributes        []GeocacheAttributes     `json:"attributes"`
	HasCallerNote     bool                     `json:"hasCallerNote" fake:"bool"`
	Distance          string                   `json:"distance" fake:"{number:1,100}"`
	Bearing           string                   `json:"bearing" fake:"{number:1,100}"`

	LastFoundTime time.Time // This is a parsed version of LastFoundDate
	GUID          string    `fake:"{UUID}"` // We read this ourselves from the geocache's page
}

type GeocacheSearchResponse struct {
	Results []Geocache `json:"results"`
	Total   int        `json:"total"`
}

type GeocacheLog struct {
	LogID               int    `json:"LogID"`
	CacheID             int    `json:"CacheID"`
	LogGUID             string `json:"LogGuid"`
	Latitude            any    `json:"Latitude"`
	Longitude           any    `json:"Longitude"`
	LatLonString        string `json:"LatLonString"`
	LogTypeID           int    `json:"LogTypeID"`
	LogType             string `json:"LogType"`
	LogTypeImage        string `json:"LogTypeImage"`
	LogText             string `json:"LogText"`
	Created             string `json:"Created"`
	Visited             string `json:"Visited"`
	UserName            string `json:"UserName"`
	MembershipLevel     int    `json:"MembershipLevel"`
	AccountID           int    `json:"AccountID"`
	AccountGUID         string `json:"AccountGuid"`
	Email               string `json:"Email"`
	AvatarImage         string `json:"AvatarImage"`
	GeocacheFindCount   int    `json:"GeocacheFindCount"`
	GeocacheHideCount   int    `json:"GeocacheHideCount"`
	ChallengesCompleted int    `json:"ChallengesCompleted"`
	IsEncoded           bool   `json:"IsEncoded"`
	Creator             struct {
		GroupTitle    string `json:"GroupTitle"`
		GroupImageURL string `json:"GroupImageUrl"`
	} `json:"creator"`
	Images []any `json:"Images"`
}

type GeocacheLogSearchResponse struct {
	Status   string        `json:"status"`
	Data     []GeocacheLog `json:"data"`
	PageInfo struct {
		Idx        int `json:"idx"`
		Size       int `json:"size"`
		TotalRows  int `json:"totalRows"`
		TotalPages int `json:"totalPages"`
		Rows       int `json:"rows"`
	} `json:"pageInfo"`
}

// This runs the query against the geocaching API and returns a slice of up to `take` geocaches,
// and the total number of geocaches matching that query
func (g *GeocachingAPI) searchQuery(st searchTerms, skip, take int) ([]Geocache, int, error) {
	var err error
	req, err := http.NewRequest("GET", g.config.GeocachingAPIURL+"/api/proxy/web/search/v2", nil)
	if err != nil {
		return nil, 0, err
	}
	query := req.URL.Query()
	query.Add("skip", fmt.Sprint(skip))
	query.Add("take", fmt.Sprint(take))
	query.Add("asc", "true")
	// Note: Sorting by anything other than distance is a "premium feature." This means we
	// have to query all pages of results and sort them ourselves.
	query.Add("sort", "distance")
	// query.Add("properties", "callernote")
	query.Add("origin", fmt.Sprintf("%f,%f", st.Latitude, st.Longitude))
	query.Add("rad", fmt.Sprint(st.RadiusMeters))
	// query.Add("oid", "3356")
	// query.Add("ot", "city")
	query.Add("ot", "query")

	// Exclude caches found by me
	query.Add("nfb", os.Getenv("GEOCACHING_CLIENT_ID"))

	// Exclude all event types
	query.Add("ct","2,3,8,137,5,11,1858,4")

	// Exclude my caches
	query.Add("ho", "1")
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", g.config.GeocachingAPIURL+"/play")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "BMItemsPerPage=1000;-H Sec-Fetch-Dest:")

	log.Debug("Request: Search")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = io.ReadAll(resp.Body); err != nil {
		return nil, 0, err
	}

	// Check if the Content-Encoding is gzip, abort if not.
	if resp.Header.Get("Content-Encoding") == "gzip" {
		var r io.Reader
		if r, err = gzip.NewReader(bytes.NewReader(body)); err != nil {
			return nil, 0, err
		}
		if body, err = io.ReadAll(r); err != nil {
			return nil, 0, err
		}
	}

	// Unmarshal body into a GeocacheSearchResponse
	var searchResponse GeocacheSearchResponse
	if err = json.Unmarshal(body, &searchResponse); err != nil {
		return nil, 0, err
	}
	// Iterate over the results and parse the LastFoundDate
	for i := 0; i < len(searchResponse.Results); i++ {
		if searchResponse.Results[i].LastFoundDate != "" {
			searchResponse.Results[i].LastFoundTime, _ = parseTime(searchResponse.Results[i].LastFoundDate)
		}
	}

	return searchResponse.Results, searchResponse.Total, nil
}

// This returns a time.Time parsed from a LastFoundDate or PlacedDate as delivered by the
// gc.com api.
func parseTime(date string) (time.Time, error) {
	// Append my account's time zone to the date so it parses with timezone info
	// TODO Work out some way of querying a user's time zone use that here instead.
	tempTime := date + "+10:00"
	return time.Parse(time.RFC3339, tempTime)
}

// This sets the GUID field on the geocache.
func (g *GeocachingAPI) GetGUIDForGeocache(geocache *Geocache) error {
	url := fmt.Sprintf(g.config.GeocachingAPIURL+"/geocache/%s", geocache.Code)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", g.config.GeocachingAPIURL+"/play")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "BMItemsPerPage=1000;-H Sec-Fetch-Dest:")

	log.Debug("Request: GetGUIDForGeocache")
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rgx := regexp.MustCompile("guid='([a-f0-9-]*)';")
	matches := rgx.FindStringSubmatch(string(body))
	if len(matches) < 1 {
		return fmt.Errorf("could not find guid. This might be a premium geocache")
	}
	geocache.GUID = matches[1]
	return nil
}

func (g *GeocachingAPI) SanitiseLogText(text string) (result string) {
	result = text
	// Strip newlines
	result = g.blueMondayPolicy.Sanitize(result)
	result = strings.TrimSpace(result)
	result = html.UnescapeString(result)
	result = strings.ReplaceAll(result, "\n", " ")
	return
}

// This returns a slice of the logs associated with a given geocache ID
// Due to bastardy, we'll need to find the GUID of the geocache,
// from https://www.geocaching.com/geocache/<code>

// var lat=-27.485133, lng=152.959033, guid='85b9e86b-aa9e-4467-be4b-9591785cd114';
// Then hit geocache_logs.aspx and extract a token from a line like:
// userToken = 'poopppoopppo';
// Then we can use that

func (g *GeocachingAPI) GetLogs(geocache *Geocache) ([]GeocacheLog, error) {
	var err error

	if !g.config.UnThrottle {
		// Wait a random number of seconds between 3 and 8
		time.Sleep(time.Duration(rand.Intn(5)+3) * time.Second)
	}

	// Get the GUID for the geocache, if required
	if geocache.GUID == "" {
		err = g.GetGUIDForGeocache(geocache)
		if err != nil {
			return nil, err
		}
	}
	url := fmt.Sprintf(g.config.GeocachingAPIURL+"/seek/geocache_logs.aspx?guid=%s", geocache.GUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", g.config.GeocachingAPIURL+"/play")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "BMItemsPerPage=1000;-H Sec-Fetch-Dest:")

	log.Debug("Request: GetLogs userToken")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rgx := regexp.MustCompile("userToken = '([A-Z0-9]*)';")
	matches := rgx.FindStringSubmatch(string(body))
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not find the userToken required to request the logs")
	}
	userToken := matches[1]

	// Now we have the userToken, we can request the logs
	req, err = http.NewRequest("GET", g.config.GeocachingAPIURL+"/seek/geocache.logbook", nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	query.Add("tkn", userToken)
	query.Add("idx", "1")
	query.Add("num", "10")
	query.Add("sp", "false")
	query.Add("sf", "false")
	query.Add("decrypt", "false")
	req.URL.RawQuery = query.Encode()

	log.Debug("Request: GetLogs Logs")
	logResp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer logResp.Body.Close()
	logBody, err := io.ReadAll(logResp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal logBody into a struct
	var logresponse GeocacheLogSearchResponse
	err = json.Unmarshal(logBody, &logresponse)
	if err != nil {
		return nil, err
	}

	// Go through and sanitise all the log text.
	for i := 0; i < len(logresponse.Data); i++ {
		logresponse.Data[i].LogText = g.SanitiseLogText(logresponse.Data[i].LogText)
	}

	return logresponse.Data, nil

}

// This finds all geocaches
func (g *GeocachingAPI) Search(st searchTerms) ([]Geocache, error) {
	var err error
	var results []Geocache
	log.Println("Running a search")

	// Run the first query to get the total number of results
	var total int
	if results, total, err = g.searchQuery(st, 0, 500); err != nil {
		return nil, err
	}

	// This is to confirm that we don't spin endlessly spamming search
	// requests if something goes awry.
	var sanityCheck int

	// Run the rest of the queries to get the rest of the results
	for i := 500; i < total; i += 500 {
		sanityCheck++
		if sanityCheck > 10 {
			return nil, fmt.Errorf("sanity check failed")
		}
		if !g.config.UnThrottle {
			// Wait a random number of seconds between 2 and 5
			time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
		}
		var nextResults []Geocache
		if nextResults, _, err = g.searchQuery(st, i, 500); err != nil {
			return nil, err
		}
		results = append(results, nextResults...)
	}

	if !st.IgnorePremium {
		return results, nil
	}

	// Filter out the non-premium geocaches
	var nonPremiumGeocaches []Geocache
	for _, geocache := range results {
		if !geocache.PremiumOnly {
			nonPremiumGeocaches = append(nonPremiumGeocaches, geocache)
		}
	}

	return nonPremiumGeocaches, nil
}
