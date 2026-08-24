package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInspectOutput_AllFixtures(t *testing.T) {
	fixturesDir := "../../test-data"

	testFiles := []string{
		filepath.Join(fixturesDir, "gar-manifest.txt"),
		filepath.Join(fixturesDir, "dockerhub-manifest.txt"),
	}

	for _, path := range testFiles {
		name := filepath.Base(path)

		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read fixture %s: %v", path, err)
			}

			report := ParseInspectOutput(data)

			// --- Basic invariants (should hold for all inspect outputs) ---

			if report.IndexDigest == "" {
				t.Fatalf("indexDigest is empty")
			}

			if report.Manifests == nil {
				t.Fatalf("manifests slice is nil")
			}

			// --- Validate manifests ---
			imageDigests := map[string]bool{}
			attestationRefs := map[string]bool{}

			for _, m := range report.Manifests {
				if m.Digest == "" {
					t.Fatalf("manifest missing digest: %+v", m)
				}

				switch m.Kind {
				case "image":
					if m.Platform == "" {
						t.Fatalf("image manifest missing platform: %+v", m)
					}
					imageDigests[m.Digest] = true

				case "attestation":
					if m.RefersTo == "" {
						t.Fatalf("attestation missing refersTo: %+v", m)
					}
					attestationRefs[m.RefersTo] = true

				default:
					t.Fatalf("unexpected manifest kind %q", m.Kind)
				}
			}

			// --- Attestations must refer to real images ---
			for ref := range attestationRefs {
				if !imageDigests[ref] {
					t.Fatalf("attestation refers to unknown image digest: %s", ref)
				}
			}
		})
	}
}
