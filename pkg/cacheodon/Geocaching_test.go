package cacheodon

import (
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

type mockGeocachingApi struct {
	caches []Geocache
	logs   []GeocacheLog
}

// Populate some dummy data into the struct
func (m *mockGeocachingApi) populate() {
	m.caches = []Geocache{
		{
			ID:             123456,
			Name:           "Secret Hideout",
			Code:           "GC1234",
			PremiumOnly:    false,
			FavoritePoints: 5,
			GeocacheType:   1,
			ContainerType:  2,
			Difficulty:     3.5,
			Terrain:        2.5,
			CacheStatus:    1,
			PostedCoordinates: GocachePostedCoordinates{
				Latitude:  37.7749,
				Longitude: -122.4194,
			},
			DetailsURL: "/geocache/GC1234",
			HasGeotour: false,
			PlacedDate: "2020-01-01T00:00:00",
			Owner: GeocacheOwner{
				Code:     "ABC123",
				Username: "JimblyBimbly",
			},
			LastFoundDate:  "2022-05-01T12:00:00",
			TrackableCount: 2,
			Region:         "California",
			Country:        "United States",
			Attributes: []GeocacheAttributes{
				{ID: 24, Name: "Wheelchair accessible", IsApplicable: false},
				{ID: 8, Name: "Scenic view", IsApplicable: true},
			},
			Distance:      "2.3mi",
			Bearing:       "NW",
			LastFoundTime: time.Now(),
			GUID:          "a8cf16ab-5a5d-42a2-9a8e-2b33d431c758",
		},
		{
			ID:             789123,
			Name:           "Bingo Hall",
			Code:           "GC456798",
			PremiumOnly:    false,
			FavoritePoints: 5,
			GeocacheType:   1,
			ContainerType:  2,
			Difficulty:     3.5,
			Terrain:        2.5,
			CacheStatus:    1,
			PostedCoordinates: GocachePostedCoordinates{
				Latitude:  37.7749,
				Longitude: -122.4194,
			},
			DetailsURL: "/geocache/GC456798",
			HasGeotour: false,
			PlacedDate: "2020-02-03T10:00:00",
			Owner: GeocacheOwner{
				Code:     "ABC123",
				Username: "PrinceOfBingo",
			},
			LastFoundDate:  "2022-07-11T12:12:34",
			TrackableCount: 2,
			Region:         "California",
			Country:        "United States",
			Attributes: []GeocacheAttributes{
				{ID: 24, Name: "Wheelchair accessible", IsApplicable: false},
				{ID: 8, Name: "Scenic view", IsApplicable: true},
			},
			Distance:      "2.3mi",
			Bearing:       "NW",
			LastFoundTime: time.Now(),
			GUID:          "a8cf16ab-5a5d-42a2-9a8e-2b33d431c758",
		},
	}
	m.logs = []GeocacheLog{
		{
			LogID:        2150129950,
			CacheID:      123456,
			LogGUID:      "fc59d67c-ccda-45a7-ad5c-f9a09f040d60",
			Latitude:     37.7749,
			Longitude:    -122.4194,
			LatLonString: "37.7749,-122.4194",
			LogTypeID:    1,
			LogType:      "Found it",
			LogTypeImage: "1.png",
			LogText: `Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!
								Had a great time at this event. Thanks for hosting!`,
			Created:             "3/16/2023",
			Visited:             "3/16/2023",
			UserName:            "Amy",
			MembershipLevel:     3,
			AccountID:           16998345,
			AccountGUID:         "8fa1b05c-d5f5-4f9c-8f0b-634b88017772",
			Email:               "amy@example.com",
			AvatarImage:         "https://www.example.com/avatar.jpg",
			GeocacheFindCount:   1123,
			GeocacheHideCount:   5,
			ChallengesCompleted: 2,
			IsEncoded:           true,
			Creator: struct {
				GroupTitle    string `json:"GroupTitle"`
				GroupImageURL string `json:"GroupImageUrl"`
			}{
				GroupTitle:    "Premium Member",
				GroupImageURL: "/images/icons/prem_user.gif",
			},
			Images: []any{},
		},
		{
			LogID:               2150129950,
			CacheID:             789123,
			LogGUID:             "fc59d67c-ccda-45a7-ad5c-f9a09f040d60",
			Latitude:            37.7749,
			Longitude:           -122.4194,
			LatLonString:        "37.7749,-122.4194",
			LogTypeID:           1,
			LogType:             "Found it",
			LogTypeImage:        "1.png",
			LogText:             "dogs dogs dogs!\n",
			Created:             "3/16/2023",
			Visited:             "3/16/2023",
			UserName:            "Beepo",
			MembershipLevel:     3,
			AccountID:           16998345,
			AccountGUID:         "8fa1b05c-d5f5-4f9c-8f0b-634b88017772",
			Email:               "amy@example.com",
			AvatarImage:         "https://www.example.com/avatar.jpg",
			GeocacheFindCount:   1123,
			GeocacheHideCount:   5,
			ChallengesCompleted: 2,
			IsEncoded:           true,
			Creator: struct {
				GroupTitle    string `json:"GroupTitle"`
				GroupImageURL string `json:"GroupImageUrl"`
			}{
				GroupTitle:    "Premium Member",
				GroupImageURL: "/images/icons/prem_user.gif",
			},
			Images: []any{},
		},
	}
}

// Advance the last found date on the stored cache
func (m *mockGeocachingApi) advanceLastFoundDate(index int) {
	m.caches[index].LastFoundTime = m.caches[index].LastFoundTime.Add(time.Hour * 24)
}

func (m *mockGeocachingApi) Auth(clientID, clientSecret string) error {
	return nil
}

func (m *mockGeocachingApi) Search(st SearchTerms) ([]Geocache, error) {
	return m.caches, nil
}

func (m *mockGeocachingApi) GetLogs(geocache *Geocache) ([]GeocacheLog, error) {
	var logs []GeocacheLog
	for _, log := range m.logs {
		if log.CacheID == geocache.ID {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func TestUpdate(t *testing.T) {
	var err error
	tempdir := t.TempDir()
	conf := configStore{
		SearchTerms: SearchTerms{
			AreaName: "Blerpville",
		},
		DBFilename: tempdir + "/test.sqlite3",
	}
	var g *Geocaching
	api := &mockGeocachingApi{}
	api.populate()
	if g, err = NewGeocaching(conf, api); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer g.Close()
	var logs []postDetails
	logs, err = g.Update()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if want, got := 2, len(logs); want != got {
		t.Errorf("Expected %d logs, got %d", want, got)
	}

	// Check that this cache showed up as a new one
	if !logs[0].NewCache {
		t.Errorf("Expected this to be a new cache")
	}
	// Check that it has the owner's details, not the owner's
	if want, got := "JimblyBimbly", logs[0].UserName; want != got {
		t.Errorf("Expected the owner to be %s, got %s", want, got)
	}

	// Advance the last found date on the mock cache
	api.advanceLastFoundDate(0)
	logs, err = g.Update()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if want, got := 1, len(logs); want != got {
		t.Errorf("Expected %d logs, got %d", want, got)
	}
	// Check that it is no longer a new cache
	if logs[0].NewCache {
		t.Errorf("Expected this to not be a new cache")
	}
	// Check that it has the finder's details, not the owner's
	if want, got := "Amy", logs[0].UserName; want != got {
		t.Errorf("Expected the finder to be %s, got %s", want, got)
	}
	// Check the very long message was truncated properly.
	if want, got := 500, len(logs[0].toString()); want != got {
		t.Errorf("Expected the log to be %d characters, got %d", want, got)
	}
	// Check the truncated string ends with "" #geocaching"
	if want, got := `" #geocaching`, logs[0].toString()[len(logs[0].toString())-13:]; want != got {
		t.Errorf("Expected the log to end with %q, got %q", want, got)
	}

	// TODO Check that we get the "That's their second find for the day!" thing.
	api.advanceLastFoundDate(0)
	logs, err = g.Update()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if want, got := 1, len(logs); want != got {
		t.Errorf("Expected %d logs, got %d", want, got)
	}
	// Check that it has the finder's details, not the owner's
	if want, got := "Amy", logs[0].UserName; want != got {
		t.Errorf("Expected the finder to be %s, got %s", want, got)
	}
	if want, got := 2, logs[0].UsersFindsToday; want != got {
		t.Errorf("Expected the finder to have found %d caches, got %d", want, got)
	}

	logs, err = g.Update()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if want, got := 0, len(logs); want != got {
		t.Errorf("Expected %d logs, got %d", want, got)
	}

	api.advanceLastFoundDate(1)
	logs, err = g.Update()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if want, got := 1, len(logs); want != got {
		t.Errorf("Expected %d logs, got %d", want, got)
	}
	// Check that it has the finder's details, not the owner's
	if want, got := "Beepo", logs[0].UserName; want != got {
		t.Errorf("Expected the finder to be %s, got %s", want, got)
	}
	api.advanceLastFoundDate(0)
	api.advanceLastFoundDate(1)
	logs, err = g.Update()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if want, got := 2, len(logs); want != got {
		t.Errorf("Expected %d logs, got %d", want, got)
	}
}
