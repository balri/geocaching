package sheets

type SheetWriter interface {
	AppendRows(rows [][]interface{}) error
	UpdateRows(rows []RowWithIndex) error
	GetExistingRows() map[string]RowWithIndex
	EnsureSheetExistsWithHeaderAndFilter(header []interface{}) error
	ExtendFilterToAllRows(colCount int64) error
}

type colIdx int

const (
	ColumnCode            colIdx = 0
	ColumnName            colIdx = 1
	ColumnPostedCoords    colIdx = 2
	ColumnCorrectedCoords colIdx = 3
	ColumnDistance        colIdx = 4
	ColumnPlacedDate      colIdx = 5
	ColumnCacheType       colIdx = 6
	ColumnCacheSize       colIdx = 7
	ColumnDifficulty      colIdx = 8
	ColumnTerrain         colIdx = 9
	ColumnOwner           colIdx = 10
	ColumnFound           colIdx = 11
	ColumnNote            colIdx = 12
	ColumnDateUpdated     colIdx = 13
)
