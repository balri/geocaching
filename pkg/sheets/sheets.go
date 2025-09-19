package sheets

import (
	"context"
	"log"
	"math"
	"time"

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

func (s *SheetClient) AppendRows(rows [][]interface{}) {
	ctx := context.Background()
	var err error
	maxRetries := 15
	maxBackoff := 60 * time.Second
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err = s.service.Spreadsheets.Values.Append(
			s.spreadsheetID,
			s.sheetName+"!A:Z",
			&sheets.ValueRange{Values: rows},
		).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
		if err == nil {
			return
		}
		// Check for rate limit error
		if gErr, ok := err.(*googleapi.Error); ok && (gErr.Code == 429 || gErr.Code == 403) {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("Rate limited by Google Sheets API, retrying in %v...", backoff)
			time.Sleep(backoff)
			continue
		}
		log.Printf("Failed to append rows: %v", err)
		return
	}
	log.Printf("Failed to append rows after %d retries: %v", maxRetries, err)
}

func (s *SheetClient) GetExistingCodes() map[string]bool {
	ctx := context.Background()
	var resp *sheets.ValueRange
	var err error
	maxRetries := 15
	maxBackoff := 60 * time.Second
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = s.service.Spreadsheets.Values.Get(
			s.spreadsheetID,
			s.sheetName+"!A:A",
		).Context(ctx).Do()
		if err == nil {
			break
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
		log.Printf("Failed to read existing codes: %v", err)
		return nil
	}
	if err != nil {
		log.Printf("Failed to read existing codes after %d retries: %v", maxRetries, err)
		return nil
	}
	codes := make(map[string]bool)
	for _, row := range resp.Values {
		if len(row) > 0 {
			code, ok := row[0].(string)
			if ok {
				codes[code] = true
			}
		}
	}
	return codes
}

func (s *SheetClient) EnsureSheetExists() error {
	ctx := context.Background()
	// 1. Get spreadsheet metadata
	ss, err := s.service.Spreadsheets.Get(s.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return err
	}
	// 2. Check if sheet exists
	for _, sh := range ss.Sheets {
		if sh.Properties.Title == s.sheetName {
			return nil // Sheet exists
		}
	}
	// 3. Add the sheet if not found
	addSheetReq := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: s.sheetName,
			},
		},
	}
	_, err = s.service.Spreadsheets.BatchUpdate(s.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{addSheetReq},
	}).Context(ctx).Do()
	return err
}
