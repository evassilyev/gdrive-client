package gdclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	return config.Client(context.Background(), getToken(config))
}

func getToken(config *oauth2.Config) *oauth2.Token {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return tok
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func NewDriveClient() *DriveService {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, getToken(config))))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	return &DriveService{
		Service: srv,
	}
}

type DriveService struct {
	*drive.Service
}

func (ds *DriveService) CreateFolderIfNotExist(name, parentId string) (fid string, err error) {
	var f *drive.FileList
	if parentId == "" {
		parentId = "root"
	}
	f, err = ds.Files.List().Q(fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents", name, parentId)).Do()
	if err != nil {
		return
	}
	if len(f.Files) != 0 {
		fid = f.Files[0].Id
		return
	}
	newFile := &drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:     name,
		Parents:  []string{parentId},
	}
	newFile, err = ds.Files.Create(newFile).Do()
	if err != nil {
		return
	}
	fid = newFile.Id
	return
}

func (ds *DriveService) SaveImage(name, parentId, link string) (fid string, err error) {
	if parentId == "" {
		parentId = "root"
	}
	file := &drive.File{
		MimeType: "image/jpeg",
		Name:     name,
		Parents:  []string{parentId},
	}
	var resp *http.Response
	resp, err = http.Get(link)
	if err != nil {
		return
	}
	defer func() {
		err = resp.Body.Close()
	}()
	file, err = ds.Files.Create(file).Media(resp.Body).Do()
	if err != nil {
		return
	}
	fid = file.Id
	return
}

func (ds *DriveService) FileExists(name, parentId string) (exists bool, err error) {
	var f *drive.FileList
	if parentId == "" {
		parentId = "root"
	}
	f, err = ds.Files.List().Q(fmt.Sprintf("name = '%s' and '%s' in parents", name, parentId)).Do()
	if err != nil {
		return
	}
	exists = len(f.Files) != 0
	return
}
