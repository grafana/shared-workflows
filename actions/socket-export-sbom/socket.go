package socket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BASEURL = "https://api.socket.dev/v0"

type Client struct {
	APIKey     string
	BaseURL    string
	Org        string
	HTTPClient *http.Client
}

type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.BaseURL = url
		}
	}
}

func NewClient(apiKey, org string, opts ...Option) *Client {
	c := &Client{
		APIKey:  apiKey,
		BaseURL: BASEURL,
		Org:     org,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Package struct {
	Name                  string `json:"name"`
	SPDXID                string `json:"SPDXID"`
	VersionInfo           string `json:"versionInfo"`
	PackageFileName       string `json:"packageFileName"`
	PrimaryPackagePurpose string `json:"primaryPackagePurpose"`
	DownloadLocation      string `json:"downloadLocation"`
	FilesAnalyzed         bool   `json:"filesAnalyzed"`
	Homepage              string `json:"homepage"`
	LicenseDeclared       string `json:"licenseDeclared"`
	ExternalRefs          []struct {
		ReferenceCategory string `json:"referenceCategory"`
		ReferenceType     string `json:"referenceType"`
		ReferenceLocator  string `json:"referenceLocator"`
	} `json:"externalRefs"`
}

type SBOM struct {
	SpdxVersion       string `json:"spdxVersion"`
	DataLicense       string `json:"dataLicense"`
	SPDXID            string `json:"SPDXID"`
	Name              string `json:"name"`
	DocumentNamespace string `json:"documentNamespace"`
	CreationInfo      struct {
		Created  time.Time `json:"created"`
		Creators []string  `json:"creators"`
	} `json:"creationInfo"`
	DocumentDescribes []string  `json:"documentDescribes"`
	Packages          []Package `json:"packages"`
	Relationships     []struct {
		SpdxElementID      string `json:"spdxElementId"`
		RelationshipType   string `json:"relationshipType"`
		RelatedSpdxElement string `json:"relatedSpdxElement"`
	} `json:"relationships"`
}

func (c *Client) GetSBOM(repo string) (SBOM, error) {
	scanID, err := c.GetLastScanID(repo)
	if err != nil {
		return SBOM{}, err
	}
	if scanID == "" {
		return SBOM{}, fmt.Errorf("no valid scan found for repo: %s", repo)
	}
	sbom, err := c.FetchSBOM(scanID)
	if err != nil {
		return SBOM{}, err
	}
	return sbom, nil
}

func (c *Client) GetLastScanID(repo string) (string, error) {
	path := fmt.Sprintf("/orgs/%s/repos/%s", c.Org, repo)
	data, err := c.makeAPIRequest(path)
	if err != nil {
		return "", err
	}
	return ParseScanID(data)
}

func ParseScanID(data []byte) (string, error) {
	r := &struct {
		LastScanID string `json:"head_full_scan_id"`
	}{}
	err := json.Unmarshal(data, r)
	if err != nil {
		return "", fmt.Errorf("%v in %q", err, data)
	}
	return r.LastScanID, nil
}

func ParseSBOM(data []byte) (SBOM, error) {
	s := SBOM{}
	err := json.Unmarshal(data, &s)
	if err != nil {
		return SBOM{}, err
	}
	return s, nil
}

func (c *Client) FetchSBOM(id string) (SBOM, error) {
	path := fmt.Sprintf("/orgs/%s/export/spdx/%s", c.Org, id)
	data, err := c.makeAPIRequest(path)
	if err != nil {
		return SBOM{}, err
	}
	return ParseSBOM(data)
}

func (c *Client) makeAPIRequest(URI string) ([]byte, error) {
	URL := fmt.Sprintf("%s%s", c.BaseURL, URI)
	authHeaderValue := fmt.Sprintf("Bearer %s", c.APIKey)
	req, err := http.NewRequestWithContext(context.Background(), "GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authHeaderValue)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
