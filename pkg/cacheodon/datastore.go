package cacheodon

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Most options are omitted if empty
// Bool values are 1, 0 or omitted to show both (as string)
// Found by and not found by use multiple keys for multiple values
// Matrix is comma separated list of 'd-t'
// Difficulty and terrain are comma separated list
// Dates in format YYYY-MM-DD
// Start and end date go together, all other dates are mutually exclusive
type SearchTerms struct {
	Latitude      float32
	Longitude     float32
	RadiusMeters  int
	AreaName      string
	IgnorePremium bool
	// Following added by me
	CacheTypes              []int    `json:"ct,omitempty"`
	CacheSizes              []int    `json:"cs,omitempty"`
	SearchTerm              string   `json:"st"`
	OperationType           string   `json:"ot"`
	Radius                  int      `json:"r"`
	CacheName               string   `json:"cn,omitempty"`
	HideFound               bool     `json:"hf,omitempty"`
	NotFoundBy              string   `json:"nfb,omitempty"`
	HideOwned               string   `json:"ho,omitempty"`
	FoundBy                 string   `json:"fb,omitempty"`
	Matrix                  []string `json:"m,omitempty"`
	SortAsc                 bool     `json:"asc"`
	Sort                    string   `json:"sort"`
	ShowDisabled            string   `json:"sd,omitempty"`
	PremiumOnly             bool     `json:"sp,omitempty"`
	ShowCorrectedCoordsOnly string   `json:"cc,omitempty"`
	ShowCacheNotesOnly      bool     `json:"pn,omitempty"`
	FavouritePoints         int      `json:"fp,omitempty"`
	Difficulty              []string `json:"d,omitempty"`
	Terrain                 []string `json:"t,omitempty"`
	FoundBeforeDate         string   `json:"fbd,omitempty"`
	FoundStartDate          string   `json:"fsd,omitempty"`
	FoundEndDate            string   `json:"fed,omitempty"`
	FoundOnDate             string   `json:"fod,omitempty"`
	FoundAfterDate          string   `json:"fad,omitempty"`
	PlacedBeforeDate        string   `json:"pbd,omitempty"`
	PlacedStartDate         string   `json:"psd,omitempty"`
	PlacedEndDate           string   `json:"ped,omitempty"`
	PlacedOnDate            string   `json:"pod,omitempty"`
	PlacedAfterDate         string   `json:"pad,omitempty"`
	Attributes              []int    `json:"att,omitempty"`
}

type APIConfig struct {
	// The URL of the Geocaching API.
	GeocachingAPIURL string
	HTTPProxyURL     string
	UnThrottle       bool // Should we disable rate-limiting for this API?
}

type configStore struct {
	Configuration APIConfig
	SearchTerms   SearchTerms
	DBFilename    string
}

type config struct {
	Filename string
	Store    configStore
}

// Write the current config out to a toml file.
func (c *config) Save() error {
	b, err := toml.Marshal(c.Store)
	if err != nil {
		return err
	}
	return os.WriteFile(c.Filename, b, 0644)
}

// Load the current config from a toml file.
func (c *config) Load() error {
	b, err := os.ReadFile(c.Filename)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, &c.Store)
}

func NewDatastore(filename string) (*config, error) {
	c := &config{
		Filename: filename,
	}
	if err := c.Load(); err != nil {
		if os.IsNotExist(err) {
			if err := c.Save(); err != nil {
				return nil, err
			}
		}
	}
	// Set some defaults
	if c.Store.DBFilename == "" {
		c.Store.DBFilename = "cacheodon.sqlite3"
	}
	return c, nil
}
