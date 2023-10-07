package cacheodon

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type searchTerms struct {
	Latitude      float32
	Longitude     float32
	RadiusMeters  int
	AreaName      string
	IgnorePremium bool
}

type APIConfig struct {
	// The URL of the Geocaching API.
	GeocachingAPIURL string
	HTTPProxyURL     string
	UnThrottle       bool // Should we disable rate-limiting for this API?
}

type configStore struct {
	Configuration APIConfig
	SearchTerms   searchTerms
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
