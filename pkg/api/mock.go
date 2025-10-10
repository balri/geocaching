package api

import (
	"geocaching/pkg/sheets"

	cacheodon "github.com/balri/cacheodon/pkg/geocaching"
)

type mockGeocachingAPI struct {
	SearchFunc                  func(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error)
	GetCacheNoteForGeocacheFunc func(cache cacheodon.Geocache) (string, error)
}

func (m *mockGeocachingAPI) Search(terms cacheodon.SearchTerms) ([]cacheodon.Geocache, error) {
	return m.SearchFunc(terms)
}
func (m *mockGeocachingAPI) GetCacheNoteForGeocache(cache cacheodon.Geocache) (string, error) {
	return m.GetCacheNoteForGeocacheFunc(cache)
}

type mockSheet struct {
	AppendRowsCalls         [][]interface{}
	UpdateRowCalls          []sheets.RowWithIndex
	ExistingRows            map[string]sheets.RowWithIndex
	EnsureSheetExistsCalled bool
	ExtendFilterCalled      bool
}

func (m *mockSheet) AppendRows(rows [][]interface{}) {
	m.AppendRowsCalls = append(m.AppendRowsCalls, rows...)
}
func (m *mockSheet) UpdateRow(rowIndex int, row []interface{}) error {
	m.UpdateRowCalls = append(m.UpdateRowCalls, sheets.RowWithIndex{Index: rowIndex, Row: row})
	return nil
}
func (m *mockSheet) GetExistingRows() map[string]sheets.RowWithIndex {
	return m.ExistingRows
}
func (m *mockSheet) EnsureSheetExistsWithHeaderAndFilter(header []interface{}) error {
	m.EnsureSheetExistsCalled = true
	return nil
}
func (m *mockSheet) ExtendFilterToAllRows(colCount int64) error {
	m.ExtendFilterCalled = true
	return nil
}
