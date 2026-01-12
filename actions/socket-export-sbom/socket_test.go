package socket_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/shared-workflows/actions/socket-export-sbom"
)

func TestParseScanID_ReturnsScanID(t *testing.T) {
	t.Parallel()
	want := "0aa47a52-6fcd-4676-a92d-eabe4e75c819"
	data, err := os.ReadFile("testdata/scan_id.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := socket.ParseScanID(data)
	if err != nil {
		t.Fatal(err)
	}
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestParseSBOM_ReturnsExpectedSBOM(t *testing.T) {
	t.Parallel()
	want := getTestSBOM()
	data, err := os.ReadFile("testdata/sbom.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := socket.ParseSBOM(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseScanID_ReturnsErrorGivenInvalidJSON(t *testing.T) {
	t.Parallel()
	invalidJSON := []byte(`{invalid json}`)
	_, err := socket.ParseScanID(invalidJSON)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseSBOM_ReturnsErrorGivenInvalidJSON(t *testing.T) {
	t.Parallel()
	invalidJSON := []byte(`{invalid json}`)
	_, err := socket.ParseSBOM(invalidJSON)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestWithBaseURL_OverRidesDefault(t *testing.T) {
	t.Parallel()
	want := "https://example.api/v1"
	c := socket.NewClient("dummy-apiKey", "grafana", socket.WithBaseURL(want))
	got := c.BaseURL
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}

}
func TestGetSBOM_ReturnsExpectedSBOMGivenRepo(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/orgs/grafana/repos/loki":
				http.ServeFile(w, r, "testdata/scan_id.json")
			case r.URL.Path == "/orgs/grafana/export/spdx/0aa47a52-6fcd-4676-a92d-eabe4e75c819":
				http.ServeFile(w, r, "testdata/sbom.json")
			default:
				t.Errorf("unexpected request path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	defer ts.Close()
	c := socket.NewClient("dummyAPIKey", "grafana")
	c.BaseURL = ts.URL
	c.HTTPClient = ts.Client()
	want := getTestSBOM()
	got, err := c.GetSBOM("loki")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetSBOM_ReturnsErrorGivenInvalidRepo(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
	defer ts.Close()
	c := socket.NewClient("dummyAPIKey", "grafana")
	c.BaseURL = ts.URL
	c.HTTPClient = ts.Client()
	_, err := c.GetSBOM("invalid-repo")
	if err == nil {
		t.Error("expected error for invalid repo, got nil")
	}
}

func getTestSBOM() socket.SBOM {
	createdTime, err := time.Parse(time.RFC3339, "2026-01-03T21:18:34Z")
	if err != nil {
		panic(err)
	}
	return socket.SBOM{
		SpdxVersion:       "SPDX-2.3",
		DataLicense:       "CC0-1.0",
		SPDXID:            "SPDXRef-DOCUMENT",
		Name:              "github.com/grafana/loki/cmd/chunks-inspect",
		DocumentNamespace: "http://spdx.org/spdxdocs/github.com%2Fgrafana%2Floki%2Fcmd%2Fchunks-inspect-5ecf2cf3-25a7-41fe-8805-1a010802ff9f",
		CreationInfo: struct {
			Created  time.Time `json:"created"`
			Creators []string  `json:"creators"`
		}{
			Created:  createdTime,
			Creators: []string{"Tool: @cyclonedx/cdxgen-11.2.5"},
		},
		DocumentDescribes: []string{"SPDXRef-Package-github.com.grafana.loki.cmd.chunks-inspect"},
		Packages: []socket.Package{
			{
				Name:                  "github.com/grafana/loki/cmd/chunks-inspect",
				SPDXID:                "SPDXRef-Package-github.com.grafana.loki.cmd.chunks-inspect",
				VersionInfo:           "",
				PackageFileName:       "",
				PrimaryPackagePurpose: "APPLICATION",
				DownloadLocation:      "NOASSERTION",
				FilesAnalyzed:         false,
				Homepage:              "NOASSERTION",
				LicenseDeclared:       "NOASSERTION",
				ExternalRefs: []struct {
					ReferenceCategory string `json:"referenceCategory"`
					ReferenceType     string `json:"referenceType"`
					ReferenceLocator  string `json:"referenceLocator"`
				}{
					{
						ReferenceCategory: "PACKAGE-MANAGER",
						ReferenceType:     "purl",
						ReferenceLocator:  "pkg:golang/github.com/grafana/loki/cmd/chunks-inspect",
					},
				},
			},
		},
		Relationships: []struct {
			SpdxElementID      string `json:"spdxElementId"`
			RelationshipType   string `json:"relationshipType"`
			RelatedSpdxElement string `json:"relatedSpdxElement"`
		}{
			{
				SpdxElementID:      "SPDXRef-DOCUMENT",
				RelationshipType:   "DEPENDS_ON",
				RelatedSpdxElement: "SPDXRef-Package-github.com.grafana.loki.cmd.chunks-inspect",
			},
		},
	}
}
