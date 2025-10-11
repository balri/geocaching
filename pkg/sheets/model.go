package sheets

type SheetWriter interface {
	AppendRows(rows [][]interface{}) error
	UpdateRows(rows []RowWithIndex) error
	GetExistingRows() map[string]RowWithIndex
	EnsureSheetExistsWithHeaderAndFilter(header []interface{}) error
	ExtendFilterToAllRows(colCount int64) error
}
