package api

import (
	"fmt"
	"geocaching/pkg/sheets"

	"strconv"
	"time"
)

type CacheRow struct {
	Index           int
	Code            string
	Name            string
	Favorite        string
	PostedCoords    string
	CorrectedCoords string
	Distance        string
	PlacedDate      string
	CacheType       string
	CacheSize       string
	Difficulty      string
	Terrain         string
	Owner           string
	Region          string
	Country         string
	Found           string
	Note            string
	DateUpdated     string
}

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
		Code:            get(0),
		Name:            get(1),
		Favorite:        get(2),
		PostedCoords:    get(3),
		CorrectedCoords: get(4),
		Distance:        getDistance(5),
		PlacedDate:      convertDateToISO(get(6)),
		CacheType:       get(7),
		CacheSize:       get(8),
		Difficulty:      get(9),
		Terrain:         get(10),
		Owner:           get(11),
		Region:          get(12),
		Country:         get(13),
		Found:           get(14),
		Note:            get(15),
	}
}

func rowsToCacheRows(rows map[string]sheets.RowWithIndex) map[string]CacheRow {
	cacheRows := make(map[string]CacheRow)
	for code, row := range rows {
		cacheRows[code] = rowToCacheRow(row)
	}
	return cacheRows
}

func (c CacheRow) ToRow() []interface{} {
	return []interface{}{
		c.Code,
		"'" + c.Name,
		c.Favorite,
		c.PostedCoords,
		c.CorrectedCoords,
		c.Distance,
		c.PlacedDate,
		c.CacheType,
		c.CacheSize,
		c.Difficulty,
		c.Terrain,
		"'" + c.Owner,
		c.Region,
		c.Country,
		c.Found,
		"'" + c.Note,
		c.DateUpdated,
	}
}

// CacheRowSlice is a named type for a slice of CacheRow
type CacheRows []CacheRow

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
