package api

import (
	"fmt"
	"geocaching/pkg/sheets"
	"testing"
	"time"

	cacheodon "github.com/balri/cacheodon/pkg/geocaching"
	"github.com/stretchr/testify/assert"
)

func TestBoolPtr(t *testing.T) {
	b := true
	ptr := BoolPtr(b)
	if ptr == nil || *ptr != b {
		t.Errorf("BoolPtr(%v) = %v, want pointer to %v", b, ptr, b)
	}
}

func TestFormatCoords(t *testing.T) {
	tests := []struct {
		lat, lon float64
		want     string
	}{
		{0, 0, ""},
		{-27.5, 153.0, "S27 30.000 E153 00.000"},
		{27.25, -153.5, "N27 15.000 W153 30.000"},
		{-27.123456, 153.654321, "S27 07.407 E153 39.259"},
	}
	for _, tt := range tests {
		got := formatCoords(tt.lat, tt.lon)
		if got != tt.want {
			t.Errorf("formatCoords(%v, %v) = %q, want %q", tt.lat, tt.lon, got, tt.want)
		}
	}
}

func TestFormatDateForSheets(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"2024-07-01T00:00:00", "2024-07-01"},
		{"2023-12-31T23:59:59", "2023-12-31"},
		{"", ""},
		{"notadate", "notadate"},
	}
	for _, tt := range tests {
		got := formatDateForSheets(tt.in)
		if got != tt.want {
			t.Errorf("formatDateForSheets(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestHaversine(t *testing.T) {
	// Brisbane (-27.4705, 153.0260) to Sydney (-33.8688, 151.2093)
	dist := haversine(-27.4705, 153.0260, -33.8688, 151.2093)
	if dist < 730 || dist > 740 {
		t.Errorf("haversine(Brisbane, Sydney) = %v, want ~733", dist)
	}
}

func TestRowsEqual(t *testing.T) {
	now := time.Now().Format("2006-01-02 15:04:05")
	a := CacheRow{
		Code:            "A",
		Name:            "Test",
		PostedCoords:    "S27 07.407 E153 39.259",
		CorrectedCoords: "S27 07.407 E153 39.259",
		Distance:        "1.23",
		PlacedDate:      "2024-07-01",
		CacheType:       "Traditional",
		CacheSize:       "Regular",
		Difficulty:      "2",
		Terrain:         "1.5",
		Owner:           "Owner",
		Found:           "Yes",
		Note:            "Note",
		DateUpdated:     now,
	}
	tests := []struct {
		b    CacheRow
		want bool
	}{
		{func() CacheRow { b := a; return b }(), true},
		{func() CacheRow { b := a; b.Code = "Different"; return b }(), true},
		{func() CacheRow { b := a; b.Name = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.PostedCoords = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.CorrectedCoords = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.Distance = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.PlacedDate = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.CacheType = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.CacheSize = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.Difficulty = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.Terrain = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.Owner = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.Found = "Different"; return b }(), false},
		{func() CacheRow { b := a; b.Note = "Different"; return b }(), true},
		{func() CacheRow { b := a; b.DateUpdated = "Different"; return b }(), true},
	}
	for _, tt := range tests {
		got := rowsEqual(a, tt.b)
		if got != tt.want {
			t.Errorf("rowsEqual(a, b) = %v, want %v", got, tt.want)
		}
	}
}

func TestGetDefaultSearchTerms(t *testing.T) {
	terms := getDefaultSearchTerms(2000)
	if terms.RadiusMeters != 2000 {
		t.Errorf("Expected RadiusMeters=2000, got %d", terms.RadiusMeters)
	}
	if *terms.ShowDisabled {
		t.Error("Expected ShowDisabled to be false")
	}
	if !*terms.SortAsc {
		t.Error("Expected SortAsc to be true")
	}
	if terms.Sort != "distance" {
		t.Errorf("Expected Sort=distance, got %s", terms.Sort)
	}
	if terms.OriginType != "query" {
		t.Errorf("Expected OriginType=query, got %s", terms.OriginType)
	}
	if !*terms.HideOwned {
		t.Error("Expected HideOwned to be true")
	}
}

func TestGetUnsolvedSearchTerms(t *testing.T) {
	terms := getUnsolvedSearchTerms(1500)
	if terms.RadiusMeters != 1500 {
		t.Errorf("Expected RadiusMeters=1500, got %d", terms.RadiusMeters)
	}
	if terms.Corrected == nil || *terms.Corrected {
		t.Error("Expected Corrected to be false")
	}
	if len(terms.CacheType) != 1 || terms.CacheType[0] != cacheodon.Unknown {
		t.Error("Expected CacheType to be [8] (Unknown)")
	}
}

func TestRunSolved(t *testing.T) {
	fixedTime := time.Date(2025, 10, 11, 5, 58, 35, 0, time.UTC)
	oldNowFunc := nowFunc
	nowFunc = func() time.Time { return fixedTime }
	defer func() { nowFunc = oldNowFunc }()

	tests := []struct {
		name       string
		mockAPI    *mockGeocachingAPI
		mockSheet  *mockSheet
		wantErr    string
		wantUpdate []sheets.RowWithIndex
		wantAppend [][]interface{}
	}{
		{
			name: "search error",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{}, fmt.Errorf("Search error")
				},
			},
			wantErr: "Search error",
		},
		{
			name: "no caches found",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{}, nil
				},
			},
			mockSheet: &mockSheet{
				ExistingRows: map[string]sheets.RowWithIndex{},
			},
		},
		{
			name: "caches found and updated",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{
						{
							Code:           "GC12345",
							Name:           "Test Cache",
							FavoritePoints: 10,
							PostedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.123456,
								Longitude: 153.654321,
							},
							UserCorrectedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.234567,
								Longitude: 153.765432,
							},
							PlacedDate:    "2024-07-01T00:00:00",
							GeocacheType:  int(cacheodon.Traditional),
							ContainerType: int(cacheodon.Regular),
							Difficulty:    2.0,
							Terrain:       1.5,
							Owner: cacheodon.GeocacheOwner{
								Code:     "PRO123",
								Username: "Owner",
							},
							Region:    "Queensland",
							Country:   "Australia",
							UserFound: true,
						},
					}, nil
				},
			},
			mockSheet: &mockSheet{
				ExistingRows: map[string]sheets.RowWithIndex{
					"GC12345": {
						Index: 1,
						Row: []interface{}{
							"GC12345", "Old Name", "S27 07.000 E153 39.000", "S27 07.000 E153 39.000",
							"0.50", "2023-01-01", "Traditional", "Regular", "2", "1.5",
							"OldOwner", "", "Old Note", "2024-06-01 12:00:00",
						},
					},
				},
			},
			wantUpdate: []sheets.RowWithIndex{
				{
					Index: 1,
					Row: []interface{}{
						"=HYPERLINK(\"https://coord.info/GC12345\", \"GC12345\")",
						"Test Cache",
						"S27 07.407 E153 39.259",
						"S27 14.074 E153 45.926",
						77.6,
						float64(45474), // 2024-07-01 in Excel date format
						"Traditional",
						"Regular",
						float64(2),
						1.5,
						"Owner",
						"Yes",
						"Old Note",
						float64(45941.24901620371), // 2025-10-11 05:58:35 in Excel date format,
					},
				},
			},
		},
		{
			name: "caches found and not updated",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{
						{
							Code:           "GC12345",
							Name:           "Test Cache",
							FavoritePoints: 10,
							PostedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.123456,
								Longitude: 153.654321,
							},
							UserCorrectedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.234567,
								Longitude: 153.765432,
							},
							PlacedDate:    "2024-07-01T00:00:00",
							GeocacheType:  int(cacheodon.Traditional),
							ContainerType: int(cacheodon.Regular),
							Difficulty:    2.0,
							Terrain:       1.5,
							Owner: cacheodon.GeocacheOwner{
								Code:     "PRO123",
								Username: "Owner",
							},
							Region:    "Queensland",
							Country:   "Australia",
							UserFound: true,
						},
					}, nil
				},
			},
			mockSheet: &mockSheet{
				ExistingRows: map[string]sheets.RowWithIndex{
					"GC12345": {
						Index: 1,
						Row: []interface{}{
							"GC12345", "Test Cache", "S27 07.407 E153 39.259", "S27 14.074 E153 45.926",
							"77.60", "2024-07-01", "Traditional", "Regular", "2", "1.5",
							"Owner", "Yes", "Old Note", "2025-10-11 05:58:35",
						},
					},
				},
			},
		},
		{
			name: "caches found and appended",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{
						{
							Code:           "GC12345",
							Name:           "Test Cache",
							FavoritePoints: 10,
							PostedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.123456,
								Longitude: 153.654321,
							},
							UserCorrectedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.234567,
								Longitude: 153.765432,
							},
							PlacedDate:    "2024-07-01T00:00:00",
							GeocacheType:  int(cacheodon.Traditional),
							ContainerType: int(cacheodon.Regular),
							Difficulty:    2.0,
							Terrain:       1.5,
							Owner: cacheodon.GeocacheOwner{
								Code:     "PRO123",
								Username: "Owner",
							},
							Region:    "Queensland",
							Country:   "Australia",
							UserFound: true,
						},
					}, nil
				},
			},
			mockSheet: &mockSheet{
				ExistingRows: map[string]sheets.RowWithIndex{},
			},
			wantAppend: [][]interface{}{
				{
					"=HYPERLINK(\"https://coord.info/GC12345\", \"GC12345\")",
					"'Test Cache",
					"S27 07.407 E153 39.259",
					"S27 14.074 E153 45.926",
					"77.60",
					"2024-07-01",
					"Traditional",
					"Regular",
					"2",
					"1.5",
					"'Owner",
					"Yes",
					"'",
					"2025-10-11 05:58:35",
				},
			},
		},
		{
			name: "cache with note found and appended",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{
						{
							Code:           "GC12345",
							Name:           "Test Cache",
							FavoritePoints: 10,
							PostedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.123456,
								Longitude: 153.654321,
							},
							UserCorrectedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.234567,
								Longitude: 153.765432,
							},
							PlacedDate:    "2024-07-01T00:00:00",
							GeocacheType:  int(cacheodon.Traditional),
							ContainerType: int(cacheodon.Regular),
							Difficulty:    2.0,
							Terrain:       1.5,
							Owner: cacheodon.GeocacheOwner{
								Code:     "PRO123",
								Username: "Owner",
							},
							Region:        "Queensland",
							Country:       "Australia",
							UserFound:     true,
							HasCallerNote: true,
						},
					}, nil
				},
				GetCacheNoteForGeocacheFunc: func(cache cacheodon.Geocache) (string, error) {
					return "ROT-47 cipher", nil
				},
			},
			mockSheet: &mockSheet{
				ExistingRows: map[string]sheets.RowWithIndex{},
			},
			wantAppend: [][]interface{}{
				{
					"=HYPERLINK(\"https://coord.info/GC12345\", \"GC12345\")",
					"'Test Cache",
					"S27 07.407 E153 39.259",
					"S27 14.074 E153 45.926",
					"77.60",
					"2024-07-01",
					"Traditional",
					"Regular",
					"2",
					"1.5",
					"'Owner",
					"Yes",
					"'ROT-47 cipher",
					"2025-10-11 05:58:35",
				},
			},
		},
		{
			name: "caches found and not appended",
			mockAPI: &mockGeocachingAPI{
				SearchFunc: func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
					return []cacheodon.Geocache{
						{
							Code:           "GC12345",
							Name:           "Test Cache",
							FavoritePoints: 10,
							PostedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.123456,
								Longitude: 153.654321,
							},
							// Corrected coords same as posted coords
							UserCorrectedCoordinates: cacheodon.GeocacheCoordinates{
								Latitude:  -27.123456,
								Longitude: 153.654321,
							},
							PlacedDate:    "2024-07-01T00:00:00",
							GeocacheType:  int(cacheodon.Traditional),
							ContainerType: int(cacheodon.Regular),
							Difficulty:    2.0,
							Terrain:       1.5,
							Owner: cacheodon.GeocacheOwner{
								Code:     "PRO123",
								Username: "Owner",
							},
							Region:    "Queensland",
							Country:   "Australia",
							UserFound: true,
						},
					}, nil
				},
			},
			mockSheet: &mockSheet{
				ExistingRows: map[string]sheets.RowWithIndex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runSolved(tt.mockAPI, tt.mockSheet, "54", "Queensland")
			if tt.wantErr == "" {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err, tt.wantErr)
				return
			}

			if len(tt.wantUpdate) > 0 {
				assert.Equal(t, tt.wantUpdate, tt.mockSheet.UpdateRowCalls)
			} else {
				assert.Empty(t, tt.mockSheet.UpdateRowCalls)
			}

			if len(tt.wantAppend) > 0 {
				assert.Equal(t, tt.wantAppend, tt.mockSheet.AppendRowsCalls)
			} else {
				assert.Empty(t, tt.mockSheet.AppendRowsCalls)
			}

			assert.True(t, tt.mockSheet.ExtendFilterCalled)
		})
	}
}
