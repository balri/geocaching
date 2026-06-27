package api

import (
	"fmt"
	"geocaching/pkg/sheets"
	"strconv"
	"strings"
	"time"

	cacheodon "github.com/balri/cacheodon/pkg/geocaching"
)

// GeocachingAPI defines the interface for interacting with the Geocaching API.
type GeocachingAPI interface {
	Search(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error)
	GetCacheNoteForGeocache(cache cacheodon.Geocache) (string, error)
}

// CacheRow represents a single geocache entry as a spreadsheet row.
type CacheRow struct {
	Index           int
	Code            string
	Name            string
	PostedCoords    string
	CorrectedCoords string
	Distance        string
	PlacedDate      string
	CacheType       string
	CacheSize       string
	Difficulty      string
	Terrain         string
	Owner           string
	Found           string
	Note            string
	DateUpdated     string
	ChangeLog       string
}

// CacheRows is a slice of CacheRow values.
type CacheRows []CacheRow

func rowToCacheRow(row sheets.RowWithIndex) CacheRow {
	get := func(i int) string {
		if i < len(row.Row) {
			return fmt.Sprint(row.Row[i])
		}
		return ""
	}
	getDistance := func(i int) string {
		if i < len(row.Row) {
			switch v := row.Row[i].(type) {
			case float64:
				return fmt.Sprintf("%.2f", v)
			case string:
				// Try to parse and format if it's a string number
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					return fmt.Sprintf("%.2f", f)
				}
				return v
			default:
				return fmt.Sprint(v)
			}
		}
		return ""
	}

	return CacheRow{
		Index:           row.Index,
		Code:            get(int(sheets.ColumnCode)),
		Name:            get(int(sheets.ColumnName)),
		PostedCoords:    get(int(sheets.ColumnPostedCoords)),
		CorrectedCoords: get(int(sheets.ColumnCorrectedCoords)),
		Distance:        getDistance(int(sheets.ColumnDistance)),
		PlacedDate:      convertDateToISO(get(int(sheets.ColumnPlacedDate))),
		CacheType:       get(int(sheets.ColumnCacheType)),
		CacheSize:       get(int(sheets.ColumnCacheSize)),
		Difficulty:      get(int(sheets.ColumnDifficulty)),
		Terrain:         get(int(sheets.ColumnTerrain)),
		Owner:           get(int(sheets.ColumnOwner)),
		Found:           get(int(sheets.ColumnFound)),
		Note:            get(int(sheets.ColumnNote)),
	}
}

func rowsToCacheRows(rows map[string]sheets.RowWithIndex) map[string]CacheRow {
	cacheRows := make(map[string]CacheRow)
	for code, row := range rows {
		cacheRows[code] = rowToCacheRow(row)
	}
	return cacheRows
}

// ToRow converts a CacheRow to a flat slice of cell values for appending to a sheet.
func (c CacheRow) ToRow() []interface{} {
	return []interface{}{
		c.Code,
		"'" + c.Name,
		c.PostedCoords,
		c.CorrectedCoords,
		c.Distance,
		c.PlacedDate,
		c.CacheType,
		c.CacheSize,
		c.Difficulty,
		c.Terrain,
		"'" + c.Owner,
		c.Found,
		"'" + c.Note,
		c.DateUpdated,
		"",
	}
}

// ToRowForUpdate converts a CacheRow to cell values for updating an existing sheet row.
func (c CacheRow) ToRowForUpdate(changeLog []string) []interface{} {
	dist, _ := strconv.ParseFloat(c.Distance, 64)
	diff, _ := strconv.ParseFloat(c.Difficulty, 64)
	terr, _ := strconv.ParseFloat(c.Terrain, 64)

	var placedDate interface{} = c.PlacedDate
	if t, err := time.Parse("2006-01-02", c.PlacedDate); err == nil {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		placedDate = t.Sub(base).Hours() / 24
	}

	var dateUpdated interface{} = c.DateUpdated
	if t, err := time.Parse("2006-01-02 15:04:05", c.DateUpdated); err == nil {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		dateUpdated = t.Sub(base).Hours() / 24
	}

	return []interface{}{
		c.Code,
		c.Name,
		c.PostedCoords,
		c.CorrectedCoords,
		dist,
		placedDate,
		c.CacheType,
		c.CacheSize,
		diff,
		terr,
		c.Owner,
		c.Found,
		c.Note,
		dateUpdated,
		strings.Join(changeLog, "; "),
	}
}

// ToRows converts all CacheRows to a 2D slice of cell values.
func (cs CacheRows) ToRows() [][]interface{} {
	rows := make([][]interface{}, len(cs))
	for i, c := range cs {
		rows[i] = c.ToRow()
	}
	return rows
}

func convertDateToISO(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	// Try parsing as dd/mm/yyyy
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return dateStr // fallback to original if parsing fails
	}
	return t.Format("2006-01-02")
}
