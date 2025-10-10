package sheets

type SheetWriter interface {
	AppendRows(rows [][]interface{})
	UpdateRow(rowIndex int, row []interface{}) error
	GetExistingRows() map[string]RowWithIndex
	EnsureSheetExistsWithHeaderAndFilter(header []interface{}) error
	ExtendFilterToAllRows(colCount int64) error
}
