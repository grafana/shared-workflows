package coverage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProfiles(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, profiles []*Profile)
	}{
		{
			name:    "valid single file coverage",
			fixture: "valid_single.out",
			wantErr: false,
			validate: func(t *testing.T, profiles []*Profile) {
				require.Len(t, profiles, 2, "should have 2 files")

				// Profiles are grouped by filename (alphabetically)
				// So handler.go comes before main.go

				// Check first profile (handler.go)
				assert.Equal(t, "github.com/example/project/handler.go", profiles[0].FileName)
				assert.Equal(t, "set", profiles[0].Mode)
				assert.Len(t, profiles[0].Blocks, 3)

				// Check first block of handler.go
				block := profiles[0].Blocks[0]
				assert.Equal(t, 20, block.StartLine)
				assert.Equal(t, 30, block.StartCol)
				assert.Equal(t, 22, block.EndLine)
				assert.Equal(t, 16, block.EndCol)
				assert.Equal(t, 2, block.NumStmt)
				assert.Equal(t, 1, block.Count)

				// Check second profile (main.go)
				assert.Equal(t, "github.com/example/project/main.go", profiles[1].FileName)
				assert.Len(t, profiles[1].Blocks, 2)
			},
		},
		{
			name:    "valid count mode coverage",
			fixture: "valid_count.out",
			wantErr: false,
			validate: func(t *testing.T, profiles []*Profile) {
				require.Len(t, profiles, 1)
				assert.Equal(t, "count", profiles[0].Mode)
				assert.Len(t, profiles[0].Blocks, 4)

				// Verify counts are preserved
				assert.Equal(t, 5, profiles[0].Blocks[0].Count)
				assert.Equal(t, 3, profiles[0].Blocks[1].Count)
			},
		},
		{
			name:    "valid atomic mode coverage",
			fixture: "valid_atomic.out",
			wantErr: false,
			validate: func(t *testing.T, profiles []*Profile) {
				require.Len(t, profiles, 1)
				assert.Equal(t, "atomic", profiles[0].Mode)
				assert.Len(t, profiles[0].Blocks, 3)

				// Verify high counts from concurrent execution
				assert.Equal(t, 100, profiles[0].Blocks[0].Count)
				assert.Equal(t, 50, profiles[0].Blocks[1].Count)
			},
		},
		{
			name:        "empty coverage data",
			fixture:     "empty.out",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "malformed coverage data",
			fixture:     "malformed.out",
			wantErr:     true,
			errContains: "failed to parse",
		},
		{
			name:        "missing mode header",
			fixture:     "no_mode.out",
			wantErr:     true,
			errContains: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadTestFixture(t, tt.fixture)

			profiles, err := ParseProfiles(data)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, profiles)

			if tt.validate != nil {
				tt.validate(t, profiles)
			}
		})
	}
}

func TestParseProfilesFromFile(t *testing.T) {
	t.Run("parse from file path", func(t *testing.T) {
		data := loadTestFixture(t, "valid_single.out")

		profiles, err := ParseProfilesFromFile("testfile.out", data)

		require.NoError(t, err)
		require.NotNil(t, profiles)
		assert.Len(t, profiles, 2)
	})
}

func TestParseProfilesFromZip(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, profiles []*Profile)
	}{
		{
			name:    "valid archive with coverage files",
			fixture: "valid_archive.zip",
			wantErr: false,
			validate: func(t *testing.T, profiles []*Profile) {
				// Should have profiles from both coverage files in the archive
				require.NotEmpty(t, profiles)

				// Verify we got data from multiple files
				fileNames := make(map[string]bool)
				for _, p := range profiles {
					fileNames[p.FileName] = true
				}
				assert.True(t, len(fileNames) >= 2, "should have profiles from multiple source files")
			},
		},
		{
			name:        "empty archive",
			fixture:     "empty_archive.zip",
			wantErr:     true,
			errContains: "no valid coverage files",
		},
		{
			name:        "archive with no coverage files",
			fixture:     "no_coverage_archive.zip",
			wantErr:     true,
			errContains: "no valid coverage files",
		},
		{
			name:        "empty zip data",
			fixture:     "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "invalid zip data",
			fixture:     "valid_single.out", // Not a zip file
			wantErr:     true,
			errContains: "failed to read zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			if tt.fixture != "" {
				data = loadTestFixture(t, tt.fixture)
			}

			profiles, err := ParseProfilesFromZip(data)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, profiles)

			if tt.validate != nil {
				tt.validate(t, profiles)
			}
		})
	}
}

func TestIsCoverageFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"out extension", "coverage.out", true},
		{"cov extension", "coverage.cov", true},
		{"coverage.txt", "coverage.txt", true},
		{"coverage in name with txt", "mycoverage.txt", true},
		{"random txt file", "readme.txt", false},
		{"go source file", "main.go", false},
		{"no extension", "coverage", false},
		{"json file", "coverage.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCoverageFile(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		name        string
		profile     *Profile
		wantErr     bool
		errContains string
	}{
		{
			name: "valid profile",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
				},
			},
			wantErr: false,
		},
		{
			name:        "nil profile",
			profile:     nil,
			wantErr:     true,
			errContains: "nil",
		},
		{
			name: "empty filename",
			profile: &Profile{
				FileName: "",
				Mode:     "set",
				Blocks:   []ProfileBlock{},
			},
			wantErr:     true,
			errContains: "empty filename",
		},
		{
			name: "empty mode",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "",
				Blocks:   []ProfileBlock{},
			},
			wantErr:     true,
			errContains: "empty mode",
		},
		{
			name: "invalid start line",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 0, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
				},
			},
			wantErr:     true,
			errContains: "invalid start line",
		},
		{
			name: "invalid end line",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 10, StartCol: 5, EndLine: 0, EndCol: 2, NumStmt: 1, Count: 1},
				},
			},
			wantErr:     true,
			errContains: "invalid end line",
		},
		{
			name: "end line before start line",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 15, StartCol: 5, EndLine: 10, EndCol: 2, NumStmt: 1, Count: 1},
				},
			},
			wantErr:     true,
			errContains: "end line",
		},
		{
			name: "end col before start col on same line",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 10, StartCol: 20, EndLine: 10, EndCol: 10, NumStmt: 1, Count: 1},
				},
			},
			wantErr:     true,
			errContains: "end column",
		},
		{
			name: "negative statement count",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: -1, Count: 1},
				},
			},
			wantErr:     true,
			errContains: "negative statement count",
		},
		{
			name: "negative count",
			profile: &Profile{
				FileName: "main.go",
				Mode:     "set",
				Blocks: []ProfileBlock{
					{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: -1},
				},
			},
			wantErr:     true,
			errContains: "negative count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfile(tt.profile)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSerializeProfiles(t *testing.T) {
	tests := []struct {
		name        string
		profiles    []*Profile
		wantErr     bool
		errContains string
		validate    func(t *testing.T, data []byte)
	}{
		{
			name: "serialize single profile",
			profiles: []*Profile{
				{
					FileName: "github.com/example/project/main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
						{StartLine: 14, StartCol: 15, EndLine: 16, EndCol: 2, NumStmt: 1, Count: 0},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				// Parse it back to verify correctness
				profiles, err := ParseProfiles(data)
				require.NoError(t, err)
				require.Len(t, profiles, 1)
				assert.Equal(t, "github.com/example/project/main.go", profiles[0].FileName)
				assert.Equal(t, "set", profiles[0].Mode)
				assert.Len(t, profiles[0].Blocks, 2)
			},
		},
		{
			name: "serialize multiple profiles",
			profiles: []*Profile{
				{
					FileName: "file1.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 5, StartCol: 10, EndLine: 7, EndCol: 2, NumStmt: 2, Count: 5},
					},
				},
				{
					FileName: "file2.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 3, Count: 10},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				profiles, err := ParseProfiles(data)
				require.NoError(t, err)
				require.Len(t, profiles, 2)
			},
		},
		{
			name:        "empty profiles slice",
			profiles:    []*Profile{},
			wantErr:     true,
			errContains: "no profiles",
		},
		{
			name:        "nil profiles",
			profiles:    nil,
			wantErr:     true,
			errContains: "no profiles",
		},
		{
			name: "profile with empty mode uses default",
			profiles: []*Profile{
				{
					FileName: "file.go",
					Mode:     "", // Empty mode
					Blocks: []ProfileBlock{
						{StartLine: 1, StartCol: 1, EndLine: 2, EndCol: 2, NumStmt: 1, Count: 1},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				// Should default to "set" mode
				profiles, err := ParseProfiles(data)
				require.NoError(t, err)
				require.Len(t, profiles, 1)
				assert.Equal(t, "set", profiles[0].Mode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := SerializeProfiles(tt.profiles)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, data)

			if tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestParseAndSerializeRoundTrip(t *testing.T) {
	// Load a valid coverage file
	data := loadTestFixture(t, "valid_single.out")

	// Parse it
	profiles, err := ParseProfiles(data)
	require.NoError(t, err)

	// Serialize it back
	serialized, err := SerializeProfiles(profiles)
	require.NoError(t, err)

	// Parse the serialized version
	reparsed, err := ParseProfiles(serialized)
	require.NoError(t, err)

	// Compare the two parsed versions
	require.Len(t, reparsed, len(profiles))
	for i := range profiles {
		assert.Equal(t, profiles[i].FileName, reparsed[i].FileName)
		assert.Equal(t, profiles[i].Mode, reparsed[i].Mode)
		assert.Equal(t, len(profiles[i].Blocks), len(reparsed[i].Blocks))

		for j := range profiles[i].Blocks {
			assert.Equal(t, profiles[i].Blocks[j], reparsed[i].Blocks[j])
		}
	}
}

func TestValidateBlock(t *testing.T) {
	tests := []struct {
		name        string
		block       ProfileBlock
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid block",
			block:   ProfileBlock{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 2, Count: 5},
			wantErr: false,
		},
		{
			name:    "valid single line block",
			block:   ProfileBlock{StartLine: 10, StartCol: 5, EndLine: 10, EndCol: 20, NumStmt: 1, Count: 1},
			wantErr: false,
		},
		{
			name:    "valid block with zero count",
			block:   ProfileBlock{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBlock(tt.block, 0)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// loadTestFixture loads a test fixture file from testdata/coverage/
func loadTestFixture(t *testing.T, filename string) []byte {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "coverage", filename)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load test fixture %s", filename)

	return data
}
