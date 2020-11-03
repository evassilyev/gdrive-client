package gdclient

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"sync"
)

func NewSheetsClient() *SheetsService {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, sheets.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, getToken(config))))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	return &SheetsService{
		Service: srv,
	}
}

type SheetsService struct {
	*sheets.Service
	createMu sync.Mutex
}

func (ss *SheetsService) CreateSheetIfNotExists(name, spreadsheetId string) (sheetId int64, err error) {

	ss.createMu.Lock()
	defer ss.createMu.Unlock()

	var ssheet *sheets.Spreadsheet
	ssheet, err = ss.Spreadsheets.Get(spreadsheetId).Do()
	if err != nil {
		return
	}
	for _, sheet := range ssheet.Sheets {
		if sheet.Properties.Title == name {
			sheetId = sheet.Properties.SheetId
			return
		}
	}
	requests := []*sheets.Request{
		{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: name,
				},
			},
		},
	}

	var resp *sheets.BatchUpdateSpreadsheetResponse
	resp, err = ss.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		IncludeSpreadsheetInResponse: true,
		Requests:                     requests,
	}).Do()
	if err != nil {
		return
	}
	for _, sheet := range resp.UpdatedSpreadsheet.Sheets {
		if sheet.Properties.Title == name {
			sheetId = sheet.Properties.SheetId
			return
		}
	}
	sheetId = -1
	return
}
