package cacheodon

import (
	"os"

	"github.com/dustin/go-humanize"
	log "github.com/sirupsen/logrus"
)

type GeocachingAPIer interface {
	Auth(clientID, clientSecret string) error
	Search(st SearchTerms) ([]Geocache, error)
	GetLogs(geocache *Geocache) ([]GeocacheLog, error)
}

type Geocaching struct {
	api  GeocachingAPIer
	db   *FinderDB
	conf configStore
}

func NewGeocaching(conf configStore, api GeocachingAPIer) (*Geocaching, error) {
	var err error
	g := &Geocaching{}
	g.conf = conf
	g.api = api
	if err = g.api.Auth(os.Getenv("GEOCACHING_CLIENT_ID"), os.Getenv("GEOCACHING_CLIENT_SECRET")); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	g.db, err = NewFinderDB(conf.DBFilename)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return g, nil
}

func (g *Geocaching) Close() {
	g.db.Close()
}

// This polls the API for a list of geocaches and updates our database
// with the results. It returns a slice of postDetails containing the
// information necessary to produce a post about the cache.
func (g *Geocaching) Update() ([]postDetails, error) {
	var results []postDetails

	caches, err := g.api.Search(g.conf.SearchTerms)
	if err != nil {
		return results, err
	}
	log.Println("Found", len(caches), "geocaches")
	for _, cache := range caches {
		new, updated := g.db.UpdateCache(&cache)
		if !new && !updated {
			continue
		}
		if post, err := g.buildPostDetails(&cache, new, updated); err == nil {
			log.Printf("%v", post)
			results = append(results, post)
		} else {
			log.Error(err)
		}
	}
	return results, nil
}

// This truncates a string to the given maximum length and returns
// the result. If truncation was necessary, it adds an elipsis to
// the end of the string.
func truncate(s string, max int) string {
	if len(s) >= max {
		return s[:max-4] + "…\""
	}
	return s
}

type postDetails struct {
	AreaName        string
	UserName        string
	CacheName       string
	DetailsURL      string
	UsersFindsToday int
	LogText         string
	NewCache        bool
	PremiumOnly     bool
}

func (p *postDetails) ToString() string {
	message := ""
	if !p.NewCache {
		message += "In " + p.AreaName + ", \"" + p.UserName + "\""
		message += " just found the \"" + p.CacheName + "\""
		if p.PremiumOnly {
			message += " premium"
		}
		message += " geocache! " + p.DetailsURL
		if p.UsersFindsToday > 1 {
			message += " That's their " + humanize.Ordinal(p.UsersFindsToday) + " find today!"
		}
		message += " They wrote: \"" + p.LogText + "\""
	}
	geocachingHashtagString := " #geocaching"
	message = truncate(message, 500-len(geocachingHashtagString))
	message += geocachingHashtagString
	return message
}

func (g *Geocaching) buildPostDetails(gc *Geocache, new, updated bool) (postDetails, error) {
	var err error
	var result postDetails
	result.AreaName = g.conf.SearchTerms.AreaName
	result.CacheName = gc.Name
	result.DetailsURL = "https://www.geocaching.com" + gc.DetailsURL
	result.PremiumOnly = gc.PremiumOnly
	if new {
		// If the cache is new, don't bother trying to get the find logs for it.
		result.UserName = gc.Owner.Username
		result.UsersFindsToday = 0
		result.LogText = ""
		result.NewCache = true
	} else if updated {
		// If the cache was updated, get the latest log and add it to the database.
		var logs []GeocacheLog
		if logs, err = g.GetLogs(gc); err != nil {
			return result, err
		}
		g.db.AddLog(&logs[0], gc)

		result.UserName = logs[0].UserName
		result.UsersFindsToday = g.db.FindsSinceMidnight(logs[0].UserName)
		result.LogText = logs[0].LogText
		result.NewCache = false
	}
	return result, nil
}

func (g *Geocaching) GetLogs(geocache *Geocache) ([]GeocacheLog, error) {
	return g.api.GetLogs(geocache)
}
