package changedetector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			yaml: `
components:
  migrator:
    paths:
      - "migrations/**"
    excludes: []
    dependencies: []
`,
			wantErr: false,
		},
		{
			name: "valid config with global excludes",
			yaml: `
global_excludes:
  - "**/*_test.go"
  - "docs/**"

components:
  migrator:
    paths:
      - "migrations/**"
    dependencies: []
`,
			wantErr: false,
		},
		{
			name: "empty config",
			yaml: `
components: {}
`,
			wantErr: true,
			errMsg:  "at least one component",
		},
		{
			name: "undefined dependency",
			yaml: `
components:
  apiserver:
    paths:
      - "api/**"
    dependencies:
      - "nonexistent"
`,
			wantErr: true,
			errMsg:  "undefined component",
		},
		{
			name: "component with no paths",
			yaml: `
components:
  apiserver:
    paths: []
    dependencies: []
`,
			wantErr: true,
			errMsg:  "at least one path",
		},
		{
			name: "invalid YAML syntax",
			yaml: `
components:
  apiserver:
    paths: [
      - invalid yaml here
`,
			wantErr: true,
			errMsg:  "parsing YAML",
		},
		{
			name:    "malformed YAML",
			yaml:    `{{{not valid yaml}}}`,
			wantErr: true,
			errMsg:  "parsing YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write temp config: %v", err)
			}

			_, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("LoadConfig() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Error("LoadConfig() should return error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "error reading config file") {
		t.Errorf("LoadConfig() error should mention file read failure, got: %v", err)
	}
}

func TestLoadConfig_InvalidGlobPatterns(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid path pattern",
			yaml: `
components:
  test:
    paths:
      - "[invalid"
    dependencies: []
`,
			wantErr: true,
			errMsg:  "invalid path pattern",
		},
		{
			name: "invalid exclude pattern",
			yaml: `
components:
  test:
    paths:
      - "pkg/**"
    excludes:
      - "[invalid"
    dependencies: []
`,
			wantErr: true,
			errMsg:  "invalid exclude pattern",
		},
		{
			name: "invalid global exclude pattern",
			yaml: `
global_excludes:
  - "[invalid"

components:
  test:
    paths:
      - "pkg/**"
    dependencies: []
`,
			wantErr: true,
			errMsg:  "global_excludes has invalid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write temp config: %v", err)
			}

			_, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("LoadConfig() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestGlobalExcludes_Integration(t *testing.T) {
	config := &Config{
		GlobalExcludes: []string{"**/*_test.go", "docs/**"},
		Components: map[string]ComponentConfig{
			"app": {
				Paths:    []string{"pkg/**"},
				Excludes: []string{"pkg/vendor/**"}, // Component-specific
			},
		},
	}

	tags := Tags{"app": "HEAD"}
	detector := &Detector{
		config: config,
		git: &MockGitOps{
			ExistingRefs: map[string]bool{
				"HEAD": true,
			},
			CommitsBetween: map[string]map[string][]string{
				"HEAD": {
					"HEAD": []string{"commit1"},
				},
			},
			FilesInCommit: map[string][]string{
				"commit1": {
					"pkg/app/main.go",      // Should match
					"pkg/app/main_test.go", // Excluded by global
					"docs/README.md",       // Excluded by global
					"pkg/vendor/lib.go",    // Excluded by component
				},
			},
		},
		tags:   tags,
		target: "HEAD",
	}

	changes, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges() error = %v", err)
	}

	// Only pkg/app/main.go should match (others excluded)
	if !changes["app"] {
		t.Error("Expected 'app' to be changed (pkg/app/main.go matches)")
	}
}

func TestGlobalExcludes(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		tags        Tags
		mockGit     *MockGitOps
		wantErr     bool
		wantChanged bool
	}{
		{
			name: "global exclude filters out test files",
			config: &Config{
				GlobalExcludes: []string{"**/*_test.go"},
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "abc123"},
			mockGit: &MockGitOps{
				ExistingRefs: map[string]bool{
					"abc123": true,
				},
				CommitsBetween: map[string]map[string][]string{
					"abc123": {
						"HEAD": []string{"commit1"},
					},
				},
				FilesInCommit: map[string][]string{
					"commit1": {
						"pkg/app/main_test.go", // Excluded by global
					},
				},
			},
			wantErr:     false,
			wantChanged: false, // Should not be marked as changed
		},
		{
			name: "global exclude filters out docs",
			config: &Config{
				GlobalExcludes: []string{"docs/**", "**/*.md"},
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"**"}},
				},
			},
			tags: Tags{"comp1": "abc123"},
			mockGit: &MockGitOps{
				ExistingRefs: map[string]bool{
					"abc123": true,
				},
				CommitsBetween: map[string]map[string][]string{
					"abc123": {
						"HEAD": []string{"commit1"},
					},
				},
				FilesInCommit: map[string][]string{
					"commit1": {
						"docs/api.md", // Excluded by global
						"README.md",   // Excluded by global
					},
				},
			},
			wantErr:     false,
			wantChanged: false,
		},
		{
			name: "component exclude overrides - both apply",
			config: &Config{
				GlobalExcludes: []string{"**/*_test.go"},
				Components: map[string]ComponentConfig{
					"comp1": {
						Paths:    []string{"pkg/**"},
						Excludes: []string{"pkg/vendor/**"},
					},
				},
			},
			tags: Tags{"comp1": "abc123"},
			mockGit: &MockGitOps{
				ExistingRefs: map[string]bool{
					"abc123": true,
				},
				CommitsBetween: map[string]map[string][]string{
					"abc123": {
						"HEAD": []string{"commit1"},
					},
				},
				FilesInCommit: map[string][]string{
					"commit1": {
						"pkg/app/main.go", // Should match
					},
				},
			},
			wantErr:     false,
			wantChanged: true,
		},
		{
			name: "no global excludes - all files considered",
			config: &Config{
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "abc123"},
			mockGit: &MockGitOps{
				ExistingRefs: map[string]bool{
					"abc123": true,
				},
				CommitsBetween: map[string]map[string][]string{
					"abc123": {
						"HEAD": []string{"commit1"},
					},
				},
				FilesInCommit: map[string][]string{
					"commit1": {
						"pkg/app/main_test.go", // Would normally be excluded, but no global excludes
					},
				},
			},
			wantErr:     false,
			wantChanged: true, // Should match since no excludes
		},
		{
			name: "multiple global excludes",
			config: &Config{
				GlobalExcludes: []string{
					"**/*_test.go",
					"docs/**",
					"scripts/**",
					"*.md",
				},
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"**"}},
				},
			},
			tags: Tags{"comp1": "abc123"},
			mockGit: &MockGitOps{
				ExistingRefs: map[string]bool{
					"abc123": true,
				},
				CommitsBetween: map[string]map[string][]string{
					"abc123": {
						"HEAD": []string{"commit1"},
					},
				},
				FilesInCommit: map[string][]string{
					"commit1": {
						"pkg/app/main_test.go", // Excluded
						"docs/guide.md",        // Excluded
						"scripts/deploy.sh",    // Excluded
						"CHANGELOG.md",         // Excluded
					},
				},
			},
			wantErr:     false,
			wantChanged: false, // All files excluded
		},
		{
			name: "global exclude with one non-excluded file",
			config: &Config{
				GlobalExcludes: []string{"**/*_test.go"},
				Components: map[string]ComponentConfig{
					"comp1": {Paths: []string{"pkg/**"}},
				},
			},
			tags: Tags{"comp1": "abc123"},
			mockGit: &MockGitOps{
				ExistingRefs: map[string]bool{
					"abc123": true,
				},
				CommitsBetween: map[string]map[string][]string{
					"abc123": {
						"HEAD": []string{"commit1"},
					},
				},
				FilesInCommit: map[string][]string{
					"commit1": {
						"pkg/app/main_test.go", // Excluded
						"pkg/app/main.go",      // Should match
					},
				},
			},
			wantErr:     false,
			wantChanged: true, // One file matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &Detector{
				config: tt.config,
				git:    tt.mockGit,
				tags:   tt.tags,
				target: "HEAD",
			}

			changes, err := detector.DetectChanges()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got := changes["comp1"]; got != tt.wantChanged {
					t.Errorf("DetectChanges() comp1 = %v, want %v", got, tt.wantChanged)
				}
			}
		})
	}
}
