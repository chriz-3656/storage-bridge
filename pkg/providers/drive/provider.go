package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/storage-bridge/core/pkg/storage"
)

type Provider struct {
	srv *drive.Service
}

func New(ctx context.Context, credentialsFile string, tokenFile string) (*Provider, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client, err := getClient(ctx, config, tokenFile)
	if err != nil {
		return nil, err
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return &Provider{srv: srv}, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config, tokenFile string) (*http.Client, error) {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		// If token file doesn't exist, we would normally run OAuth flow here.
		// For a CLI tool, this involves printing a URL and asking for auth code.
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		saveToken(tokenFile, tok)
	}
	return config.Client(ctx, tok), nil
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

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

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Unable to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (p *Provider) Name() string {
	return "drive"
}

// resolvePath traverses the path components to find the file ID.
func (p *Provider) resolvePath(path string) (string, error) {
	path = strings.Trim(path, "/")
	if path == "" || path == "." {
		return "root", nil
	}

	parts := strings.Split(path, "/")
	currentId := "root"

	for _, part := range parts {
		query := fmt.Sprintf("'%s' in parents and name='%s' and trashed=false", currentId, part)
		r, err := p.srv.Files.List().Q(query).Fields("files(id, name, mimeType)").Do()
		if err != nil {
			return "", err
		}
		if len(r.Files) == 0 {
			return "", fmt.Errorf("path not found: %s", part)
		}
		currentId = r.Files[0].Id
	}
	return currentId, nil
}

func (p *Provider) Stat(ctx context.Context, path string) (*storage.Entry, error) {
	id, err := p.resolvePath(path)
	if err != nil {
		return nil, err
	}
	if id == "root" {
		return &storage.Entry{Path: "/", IsDir: true}, nil
	}

	file, err := p.srv.Files.Get(id).Fields("id, name, mimeType, size, modifiedTime").Do()
	if err != nil {
		return nil, err
	}

	modTime, _ := time.Parse(time.RFC3339, file.ModifiedTime)
	isDir := file.MimeType == "application/vnd.google-apps.folder"
	return &storage.Entry{
		Path:    path,
		IsDir:   isDir,
		Size:    file.Size,
		ModTime: modTime,
	}, nil
}

type driveIterator struct {
	files []*drive.File
	index int
}

func (i *driveIterator) Next(ctx context.Context) (*storage.Entry, error) {
	if i.index >= len(i.files) {
		return nil, io.EOF
	}
	f := i.files[i.index]
	i.index++

	modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
	isDir := f.MimeType == "application/vnd.google-apps.folder"
	return &storage.Entry{
		Path:    f.Name,
		IsDir:   isDir,
		Size:    f.Size,
		ModTime: modTime,
	}, nil
}

func (p *Provider) List(ctx context.Context, path string) (storage.Iterator, error) {
	id, err := p.resolvePath(path)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("'%s' in parents and trashed=false", id)
	r, err := p.srv.Files.List().Q(query).Fields("files(id, name, mimeType, size, modifiedTime)").Do()
	if err != nil {
		return nil, err
	}

	return &driveIterator{files: r.Files}, nil
}

func (p *Provider) Get(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error) {
	id, err := p.resolvePath(path)
	if err != nil {
		return nil, err
	}

	call := p.srv.Files.Get(id)
	if offset > 0 || length > 0 {
		var rangeStr string
		if length > 0 {
			rangeStr = fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)
		} else {
			rangeStr = fmt.Sprintf("bytes=%d-", offset)
		}
		call.Header().Set("Range", rangeStr)
	}

	res, err := call.Download()
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func (p *Provider) ensurePath(path string) (string, string, error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "root", "", nil
	}
	dir, file := filepath.Split(path)
	dir = strings.Trim(dir, "/")
	
	if dir == "" {
		return "root", file, nil
	}

	parts := strings.Split(dir, "/")
	currentId := "root"
	
	for _, part := range parts {
		query := fmt.Sprintf("'%s' in parents and name='%s' and mimeType='application/vnd.google-apps.folder' and trashed=false", currentId, part)
		r, err := p.srv.Files.List().Q(query).Fields("files(id, name)").Do()
		if err != nil {
			return "", "", err
		}
		
		if len(r.Files) == 0 {
			// Create folder
			folder := &drive.File{
				Name:     part,
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{currentId},
			}
			f, err := p.srv.Files.Create(folder).Fields("id").Do()
			if err != nil {
				return "", "", err
			}
			currentId = f.Id
		} else {
			currentId = r.Files[0].Id
		}
	}
	
	return currentId, file, nil
}

func (p *Provider) Put(ctx context.Context, path string, in io.Reader, size int64, modTime time.Time) error {
	parentId, fileName, err := p.ensurePath(path)
	if err != nil {
		return err
	}

	// Check if file already exists to update it
	query := fmt.Sprintf("'%s' in parents and name='%s' and trashed=false", parentId, fileName)
	r, err := p.srv.Files.List().Q(query).Fields("files(id)").Do()
	if err != nil {
		return err
	}

	f := &drive.File{
		Name:    fileName,
		Parents: []string{parentId},
	}
	
	if !modTime.IsZero() {
		f.ModifiedTime = modTime.Format(time.RFC3339)
	}

	if len(r.Files) > 0 {
		// Update existing
		f.Parents = nil // Cannot update parents when updating a file
		_, err = p.srv.Files.Update(r.Files[0].Id, f).Media(in).Do()
	} else {
		// Create new
		_, err = p.srv.Files.Create(f).Media(in).Do()
	}
	return err
}

func (p *Provider) Remove(ctx context.Context, path string) error {
	id, err := p.resolvePath(path)
	if err != nil {
		return err
	}
	return p.srv.Files.Delete(id).Do()
}

func (p *Provider) SpaceUsed(ctx context.Context, path string) (int64, error) {
	about, err := p.srv.About.Get().Fields("storageQuota").Do()
	if err != nil {
		return 0, err
	}
	return about.StorageQuota.Usage, nil
}
