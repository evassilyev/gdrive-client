package gdclient

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
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
}
