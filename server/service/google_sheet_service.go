package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetService struct{}

func NewGoogleSheetService() *GoogleSheetService {
	return &GoogleSheetService{}
}

func (s *GoogleSheetService) SyncSheet(sheetID string, organization_name, organization_id string, ownerFirstName string,
	ownerLastName string, ownerEmail string, createdAt time.Time) {
	fmt.Printf("Syncing Google Sheet with ID: %s\n", sheetID)

	ctx := context.Background()

	b, err := os.ReadFile("google_sheet_auth/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse service account credentials: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := "1QcqsVK2IhhSuM4rIDXnl_Fs-SnqnIyBjjP1fwmha_z4"
	tabName := "Class Data"

	fmt.Println("\n--- Reading Headers ---")
	headers, err := s.readHeaders(srv, spreadsheetId, tabName)
	if err != nil {
		log.Printf("Warning: Could not read headers: %v", err)
	} else {
		fmt.Printf("Headers Found: %v\n", headers)
	}

	fmt.Println("\n--- Appending New Row ---")
	newRow := []interface{}{organization_name, organization_id, ownerFirstName, ownerLastName, ownerEmail, createdAt}
	err = s.appendRow(srv, spreadsheetId, tabName, newRow)
	if err != nil {
		log.Fatalf("Unable to append row: %v", err)
	}

	fmt.Println("Successfully appended new row to the sheet.")
}

func (s *GoogleSheetService) readHeaders(srv *sheets.Service, spreadsheetId string, tabName string) ([]string, error) {
	readRange := fmt.Sprintf("%s!A1:Z1", tabName)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve headers: %v", err)
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no data found in header row")
	}

	var headers []string
	for _, cell := range resp.Values[0] {
		headers = append(headers, fmt.Sprintf("%v", cell))
	}

	return headers, nil
}

func (s *GoogleSheetService) appendRow(srv *sheets.Service, spreadsheetId string, tabName string, rowData []interface{}) error {
	appendRange := fmt.Sprintf("%s!A1", tabName)
	valuesToAppend := [][]interface{}{rowData}
	valueRange := &sheets.ValueRange{
		Values: valuesToAppend,
	}
	_, err := srv.Spreadsheets.Values.Append(spreadsheetId, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	return nil
}
