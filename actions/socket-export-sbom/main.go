package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Client struct {
	APIKey     string
	BaseURL    string
	Org        string
	HTTPClient *http.Client
}

func NewClient(apiKey, org, baseURL string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Org:        org,
		HTTPClient: http.DefaultClient,
	}
}

type Repo struct {
	Slug       string
	LastScanID string `json:"head_full_scan_id"`
}

// doRequest handles the common HTTP request logic: building URL, creating request,
// adding auth header, executing request, and reading response body.
// Returns the response body bytes, status code, and error.
func (client *Client) doRequest(method, path string) ([]byte, int, error) {
	URL := fmt.Sprintf("%s%s", client.BaseURL, path)
	authHeaderValue := fmt.Sprintf("Bearer %s", client.APIKey)
	req, err := http.NewRequestWithContext(context.Background(), method, URL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Add("Authorization", authHeaderValue)
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

func (client *Client) GetRepoLastScanID(name string) (string, error) {
	path := fmt.Sprintf("/orgs/%s/repos/%s", client.Org, name)
	data, statusCode, err := client.doRequest("GET", path)
	if err != nil {
		return "", err
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code %d", statusCode)
	}

	var r Repo
	err = json.Unmarshal(data, &r)
	if err != nil {
		return "", err
	}
	return r.LastScanID, nil
}

// It creates the file if it doesn't exist and overwrites it if it does.
func saveDataToFile(data []byte, filepath string) error {
	err := os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}
	return nil
}

func (client *Client) ExportSBOM(scanID, filepath string) error {
	path := fmt.Sprintf("/orgs/%s/export/spdx/%s", client.Org, scanID)

	data, statusCode, err := client.doRequest("GET", path)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("invalid status code %d", statusCode)
	}

	err = saveDataToFile(data, filepath)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	Usage := `
    Usage: main.go <repo name> <output filepath>
    `
	key := os.Getenv("SOCKET_API_TOKEN")
	if key == "" {
		log.Fatal("SOCKET_API_TOKEN not provided")
		os.Exit(1)
	}
	baseURL := os.Getenv("SOCKET_BASE_URL")
	if baseURL == "" {
		log.Fatal("Please specify socket base url, e.g. 'api.socket.dev/v0'")
		os.Exit(1)
	}
	org := os.Getenv("SOCKET_ORG")
	if org == "" {
		log.Fatal("Please specify socket org name, e.g. 'grafana'")
		os.Exit(1)
	}
	if len(os.Args) < 3 {
		log.Println(Usage)
		os.Exit(0)
	}
	client := NewClient(key, org, baseURL)
	repo, output := os.Args[1], os.Args[2]
	id, err := client.GetRepoLastScanID(repo)
	if err != nil {
		log.Printf("ERROR: could not get scan id for %s: %s", repo, err)
		os.Exit(1)
	}
	if id == "" {
		log.Printf("ERROR: no valid scan found for repo: %s", repo)
		os.Exit(1)

	}
	log.Printf("Last scan id for %s is %s", repo, id)
	log.Printf("exporting sbom to %s", output)
	err = client.ExportSBOM(id, output)
	if err != nil {
		log.Printf("ERROR: failed to export SBOM: %s\n", err)
		os.Exit(1)
	}
}
