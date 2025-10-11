package sheets

import (
	"context"
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetClient struct {
	service       *sheets.Service
	spreadsheetID string
	sheetName     string
}

func NewSheetClient(jsonPath, spreadsheetID, sheetName string) *SheetClient {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(jsonPath))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}
	return &SheetClient{
		service:       srv,
		spreadsheetID: spreadsheetID,
		sheetName:     sheetName,
	}
}

func (s *SheetClient) UpdateRows(updates []RowWithIndex) error {
	ctx := context.Background()
	ss, err := s.service.Spreadsheets.Get(s.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return err
	}
	var sheetID int64 = -1
	for _, sh := range ss.Sheets {
		if sh.Properties.Title == s.sheetName {
			sheetID = sh.Properties.SheetId
			break
		}
	}
	if sheetID == -1 {
		return fmt.Errorf("sheet not found")
	}

	var requests []*sheets.Request
	for _, upd := range updates {
		req := &sheets.Request{
			UpdateCells: &sheets.UpdateCellsRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    int64(upd.Index),
					EndRowIndex:      int64(upd.Index + 1),
					StartColumnIndex: 0,
					EndColumnIndex:   int64(len(upd.Row)),
				},
				Rows: []*sheets.RowData{
					{Values: toCellData(upd.Row)},
				},
				Fields: "*",
			},
		}
		requests = append(requests, req)
	}

	return withBackoff(func() error {
		_, err := s.service.Spreadsheets.BatchUpdate(s.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: requests,
		}).Context(ctx).Do()
		return err
	})
}

func (s *SheetClient) AppendRows(rows [][]interface{}) error {
	ctx := context.Background()
	return withBackoff(func() error {
		_, err := s.service.Spreadsheets.Values.Append(
			s.spreadsheetID,
			s.sheetName+"!A:Z",
			&sheets.ValueRange{Values: rows},
		).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
		return err
	})
}

type RowWithIndex struct {
	Index int // 0-based index (0 = header, 1 = first data row)
	Row   []interface{}
}

func (s *SheetClient) GetExistingRows() map[string]RowWithIndex {
	ctx := context.Background()
	resp, err := s.service.Spreadsheets.Values.Get(
		s.spreadsheetID,
		s.sheetName,
	).Context(ctx).Do()
	if err != nil {
		log.Printf("Failed to read existing rows: %v", err)
		return nil
	}
	rows := make(map[string]RowWithIndex)
	for i, row := range resp.Values {
		if i == 0 || len(row) == 0 { // skip header or empty
			continue
		}
		code, ok := row[0].(string)
		if ok {
			rows[code] = RowWithIndex{
				Index: i,
				Row:   row,
			}
		}
	}
	return rows
}

func (s *SheetClient) EnsureSheetExistsWithHeaderAndFilter(header []interface{}) error {
	ctx := context.Background()
	ss, err := s.service.Spreadsheets.Get(s.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return err
	}
	var sheetID int64 = -1
	for _, sh := range ss.Sheets {
		if sh.Properties.Title == s.sheetName {
			sheetID = sh.Properties.SheetId
			break
		}
	}
	if sheetID == -1 {
		// Sheet doesn't exist, create it
		addSheetReq := &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: s.sheetName,
				},
			},
		}
		resp, err := s.service.Spreadsheets.BatchUpdate(s.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{addSheetReq},
		}).Context(ctx).Do()
		if err != nil {
			return err
		}
		sheetID = resp.Replies[0].AddSheet.Properties.SheetId

		// Add header row
		_, err = s.service.Spreadsheets.Values.Update(
			s.spreadsheetID,
			s.sheetName+"!A1:Z1",
			&sheets.ValueRange{Values: [][]interface{}{header}},
		).ValueInputOption("RAW").Context(ctx).Do()
		if err != nil {
			return err
		}

		// Add filter to header row (covers header + first data row)
		filterReq := &sheets.Request{
			SetBasicFilter: &sheets.SetBasicFilterRequest{
				Filter: &sheets.BasicFilter{
					Range: &sheets.GridRange{
						SheetId:          sheetID,
						StartRowIndex:    0,
						EndRowIndex:      2, // header + first data row
						StartColumnIndex: 0,
						EndColumnIndex:   int64(len(header)), // <-- dynamic width
					},
				},
			},
		}

		// Freeze the header row
		freezeReq := &sheets.Request{
			UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
				Properties: &sheets.SheetProperties{
					SheetId: sheetID,
					GridProperties: &sheets.GridProperties{
						FrozenRowCount: 1,
					},
				},
				Fields: "gridProperties.frozenRowCount",
			},
		}

		_, err = s.service.Spreadsheets.BatchUpdate(s.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{filterReq, freezeReq},
		}).Context(ctx).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SheetClient) ExtendFilterToAllRows(colCount int64) error {
	ctx := context.Background()
	ss, err := s.service.Spreadsheets.Get(s.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return err
	}
	var sheetID int64 = -1
	for _, sh := range ss.Sheets {
		if sh.Properties.Title == s.sheetName {
			sheetID = sh.Properties.SheetId
			break
		}
	}
	if sheetID == -1 {
		return nil // Sheet not found
	}

	// Get all values in the sheet to determine the last non-empty row
	resp, err := s.service.Spreadsheets.Values.Get(
		s.spreadsheetID,
		s.sheetName,
	).Context(ctx).Do()
	if err != nil {
		return err
	}
	rowCount := int64(len(resp.Values))
	if rowCount < 2 {
		rowCount = 2 // Always cover at least header + one row for filter
	}

	filterReq := &sheets.Request{
		SetBasicFilter: &sheets.SetBasicFilterRequest{
			Filter: &sheets.BasicFilter{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      rowCount,
					StartColumnIndex: 0,
					EndColumnIndex:   colCount,
				},
			},
		},
	}

	placedDateReq := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    1, // skip header
				StartColumnIndex: 6, // Placed Date column (G = 6)
				EndColumnIndex:   7,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					NumberFormat: &sheets.NumberFormat{
						Type:    "DATE",
						Pattern: "dd/mm/yyyy",
					},
				},
			},
			Fields: "userEnteredFormat.numberFormat",
		},
	}

	_, err = s.service.Spreadsheets.BatchUpdate(s.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{filterReq, placedDateReq},
	}).Context(ctx).Do()
	return err
}

func toCellData(row []interface{}) []*sheets.CellData {
	cells := make([]*sheets.CellData, len(row))
	for i, v := range row {
		cells[i] = &sheets.CellData{UserEnteredValue: &sheets.ExtendedValue{}}
		switch val := v.(type) {
		case string:
			if len(val) > 0 && val[0] == '=' {
				cells[i].UserEnteredValue.FormulaValue = &val
			} else {
				cells[i].UserEnteredValue.StringValue = &val
			}
		case float64:
			cells[i].UserEnteredValue.NumberValue = &val
		}
	}
	return cells
}

func withBackoff(call func() error) error {
	maxRetries := 15
	maxBackoff := 60 * time.Second
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := call()
		if err == nil {
			return nil
		}
		if gErr, ok := err.(*googleapi.Error); ok && (gErr.Code == 429 || gErr.Code == 403) {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("Rate limited by Google Sheets API, retrying in %v...", backoff)
			time.Sleep(backoff)
			continue
		}
		return err
	}
	return fmt.Errorf("failed after %d retries", maxRetries)
}
