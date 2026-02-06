package changedetector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kit/log"
)

// MockGitOps is a mock implementation of GitOperations for testing
type MockGitOps struct {
	GetChangedFilesFunc func(fromRef, toRef string) ([]string, error)
	RefExistsFunc       func(ref string) (bool, error)
}

func (m *MockGitOps) GetChangedFiles(fromRef, toRef string) ([]string, error) {
	if m.GetChangedFilesFunc != nil {
		return m.GetChangedFilesFunc(fromRef, toRef)
	}
	return []string{}, nil
}

func (m *MockGitOps) RefExists(ref string) (bool, error) {
	if m.RefExistsFunc != nil {
		return m.RefExistsFunc(ref)
	}
	return true, nil
}

func TestLoadTags_Errors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "file does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.json")
			},
			wantErr: true,
		},
		{
			name: "invalid JSON",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "invalid.json")
				if err := os.WriteFile(path, []byte("not valid json {{{"), 0644); err != nil {
					t.Fatalf("failed to write invalid json: %v", err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "valid JSON",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "tags.json")
				tags := map[string]string{"comp1": "b893b86"}
				data, _ := json.Marshal(tags)
				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatalf("failed to write tags json: %v", err)
				}
				return path
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			_, err := LoadTags(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteChanges(t *testing.T) {
	tests := []struct {
		name    string
		changes map[string]bool
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name:    "successful write",
			changes: map[string]bool{"comp1": true, "comp2": false},
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "changes.json")
			},
			wantErr: false,
		},
		{
			name:    "write to invalid directory",
			changes: map[string]bool{"comp1": true},
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent", "subdir", "changes.json")
			},
			wantErr: true,
		},
		{
			name:    "empty changes",
			changes: map[string]bool{},
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "empty.json")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			err := WriteChanges(tt.changes, path)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteChanges() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify contents if write was successful
			if !tt.wantErr {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read written file: %v", err)
				}

				var result map[string]bool
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("Failed to parse written JSON: %v", err)
				}

				if len(result) != len(tt.changes) {
					t.Errorf("Written changes count = %d, want %d", len(result), len(tt.changes))
				}
			}
		})
	}
}

func TestDetector_WithMockGit(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		tags       Tags
		mockGit    *MockGitOps
		wantErr    bool
		wantChange bool
	}{
		{
			name: "tag does not exist - triggers rebuild",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "nonexistent-tag"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return false, nil
				},
			},
			wantErr:    false,
			wantChange: true,
		},
		{
			name: "git ref check error",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "b893b86"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return false, fmt.Errorf("git error")
				},
			},
			wantErr: true,
		},
		{
			name: "git diff error - triggers rebuild",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "b893b86"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
				GetChangedFilesFunc: func(fromRef, toRef string) ([]string, error) {
					return nil, fmt.Errorf("diff failed")
				},
			},
			wantErr:    false,
			wantChange: true,
		},
		{
			name: "file matches pattern",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "b893b86"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
				GetChangedFilesFunc: func(fromRef, toRef string) ([]string, error) {
					return []string{"pkg/app/main.go"}, nil
				},
			},
			wantErr:    false,
			wantChange: true,
		},
		{
			name: "file does not match pattern",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "b893b86"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
				GetChangedFilesFunc: func(fromRef, toRef string) ([]string, error) {
					return []string{"docs/README.md"}, nil
				},
			},
			wantErr:    false,
			wantChange: false,
		},
		{
			name: "invalid pattern in config",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"[invalid"}},
				},
			},
			tags: Tags{"comp1": "b893b86"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
				GetChangedFilesFunc: func(fromRef, toRef string) ([]string, error) {
					return []string{"pkg/app/main.go"}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "no tag defaults to changed",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
			},
			wantErr:    false,
			wantChange: true,
		},
		{
			name: "empty tag defaults to changed",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": ""},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
			},
			wantErr:    false,
			wantChange: true,
		},
		{
			name: "none tag defaults to changed",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "none"},
			mockGit: &MockGitOps{
				RefExistsFunc: func(ref string) (bool, error) {
					return true, nil
				},
			},
			wantErr:    false,
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &Detector{
				config: tt.config,
				git:    tt.mockGit,
				tags:   tt.tags,
				target: "HEAD",
				logger: log.NewNopLogger(),
			}

			changes, err := detector.DetectChanges()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got := changes["comp1"]; got != tt.wantChange {
					t.Errorf("DetectChanges() comp1 = %v, want %v", got, tt.wantChange)
				}
			}
		})
	}
}

func TestDetectChanges_DependencyGraphError(t *testing.T) {
	// Test with circular dependency to trigger graph error
	config := &Config{
		Components: map[string]ComponentConfig{
			"a": {
				Paths:        []string{"a/**"},
				Dependencies: []string{"b"},
			},
			"b": {
				Paths:        []string{"b/**"},
				Dependencies: []string{"a"},
			},
		},
	}

	tags := Tags{"a": "b893b86", "b": "b893b86"}
	detector := NewDetector(config, tags, "HEAD")

	_, err := detector.DetectChanges()
	if err == nil {
		t.Error("DetectChanges() expected error for circular dependency, got nil")
	}
}

func TestMatchPaths_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		includes []string
		excludes []string
		wantErr  bool
	}{
		{
			name:     "invalid include pattern",
			file:     "test.go",
			includes: []string{"[invalid"},
			excludes: []string{},
			wantErr:  true,
		},
		{
			name:     "invalid exclude pattern",
			file:     "test.go",
			includes: []string{"*.go"},
			excludes: []string{"[invalid"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MatchPaths(tt.file, tt.includes, tt.excludes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
