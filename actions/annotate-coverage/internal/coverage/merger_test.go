package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeProfiles(t *testing.T) {
	tests := []struct {
		name        string
		profiles    []*Profile
		wantErr     bool
		errContains string
		validate    func(t *testing.T, merged []*Profile)
	}{
		{
			name: "merge two profiles with different files",
			profiles: []*Profile{
				{
					FileName: "file1.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
					},
				},
				{
					FileName: "file2.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 5, StartCol: 20, EndLine: 7, EndCol: 2, NumStmt: 1, Count: 0},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 2)
				// Results should be sorted by filename
				assert.Equal(t, "file1.go", merged[0].FileName)
				assert.Equal(t, "file2.go", merged[1].FileName)
				assert.Len(t, merged[0].Blocks, 1)
				assert.Len(t, merged[1].Blocks, 1)
			},
		},
		{
			name: "merge two profiles with overlapping blocks in set mode",
			profiles: []*Profile{
				{
					FileName: "main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
						{StartLine: 14, StartCol: 15, EndLine: 16, EndCol: 2, NumStmt: 1, Count: 0},
					},
				},
				{
					FileName: "main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
						{StartLine: 14, StartCol: 15, EndLine: 16, EndCol: 2, NumStmt: 1, Count: 1},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 1)
				assert.Equal(t, "main.go", merged[0].FileName)
				assert.Equal(t, "set", merged[0].Mode)
				require.Len(t, merged[0].Blocks, 2)

				// In set mode, blocks with same position should have max count
				assert.Equal(t, 1, merged[0].Blocks[0].Count) // max(1, 1)
				assert.Equal(t, 1, merged[0].Blocks[1].Count) // max(0, 1)
			},
		},
		{
			name: "merge two profiles with overlapping blocks in count mode",
			profiles: []*Profile{
				{
					FileName: "utils.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 5, StartCol: 25, EndLine: 7, EndCol: 2, NumStmt: 1, Count: 5},
						{StartLine: 9, StartCol: 30, EndLine: 11, EndCol: 16, NumStmt: 2, Count: 3},
					},
				},
				{
					FileName: "utils.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 5, StartCol: 25, EndLine: 7, EndCol: 2, NumStmt: 1, Count: 3},
						{StartLine: 9, StartCol: 30, EndLine: 11, EndCol: 16, NumStmt: 2, Count: 2},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 1)
				assert.Equal(t, "utils.go", merged[0].FileName)
				assert.Equal(t, "count", merged[0].Mode)
				require.Len(t, merged[0].Blocks, 2)

				// In count mode, blocks with same position should sum counts
				assert.Equal(t, 8, merged[0].Blocks[0].Count) // 5 + 3
				assert.Equal(t, 5, merged[0].Blocks[1].Count) // 3 + 2
			},
		},
		{
			name: "merge profiles with atomic mode",
			profiles: []*Profile{
				{
					FileName: "concurrent.go",
					Mode:     "atomic",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 100},
					},
				},
				{
					FileName: "concurrent.go",
					Mode:     "atomic",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 50},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 1)
				assert.Equal(t, "atomic", merged[0].Mode)
				require.Len(t, merged[0].Blocks, 1)
				assert.Equal(t, 150, merged[0].Blocks[0].Count) // 100 + 50
			},
		},
		{
			name: "merge profiles with multiple files and overlapping blocks",
			profiles: []*Profile{
				{
					FileName: "file1.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 2},
					},
				},
				{
					FileName: "file2.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 5, StartCol: 20, EndLine: 7, EndCol: 2, NumStmt: 1, Count: 3},
					},
				},
				{
					FileName: "file1.go",
					Mode:     "count",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 3},
						{StartLine: 20, StartCol: 5, EndLine: 22, EndCol: 2, NumStmt: 2, Count: 1},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 2)

				// Find each file (order is deterministic but alphabetical)
				var file1, file2 *Profile
				for _, p := range merged {
					if p.FileName == "file1.go" {
						file1 = p
					} else if p.FileName == "file2.go" {
						file2 = p
					}
				}

				require.NotNil(t, file1)
				require.NotNil(t, file2)

				// file1.go should have 2 blocks (one merged, one unique)
				require.Len(t, file1.Blocks, 2)
				assert.Equal(t, 5, file1.Blocks[0].Count) // 2 + 3
				assert.Equal(t, 1, file1.Blocks[1].Count)

				// file2.go should have 1 block (no overlap)
				require.Len(t, file2.Blocks, 1)
				assert.Equal(t, 3, file2.Blocks[0].Count)
			},
		},
		{
			name: "single profile returns copy",
			profiles: []*Profile{
				{
					FileName: "single.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 1)
				assert.Equal(t, "single.go", merged[0].FileName)
				assert.Len(t, merged[0].Blocks, 1)
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
			name: "mismatched modes",
			profiles: []*Profile{
				{
					FileName: "file1.go",
					Mode:     "set",
					Blocks:   []ProfileBlock{{StartLine: 1, StartCol: 1, EndLine: 2, EndCol: 2, NumStmt: 1, Count: 1}},
				},
				{
					FileName: "file2.go",
					Mode:     "count",
					Blocks:   []ProfileBlock{{StartLine: 1, StartCol: 1, EndLine: 2, EndCol: 2, NumStmt: 1, Count: 5}},
				},
			},
			wantErr:     true,
			errContains: "mode",
		},
		{
			name: "blocks are sorted by position",
			profiles: []*Profile{
				{
					FileName: "main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 20, StartCol: 5, EndLine: 22, EndCol: 2, NumStmt: 1, Count: 1},
						{StartLine: 10, StartCol: 13, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 1},
						{StartLine: 10, StartCol: 5, EndLine: 10, EndCol: 20, NumStmt: 1, Count: 0},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, merged []*Profile) {
				require.Len(t, merged, 1)
				require.Len(t, merged[0].Blocks, 3)

				// Verify sorting: line 10 col 5, line 10 col 13, line 20 col 5
				assert.Equal(t, 10, merged[0].Blocks[0].StartLine)
				assert.Equal(t, 5, merged[0].Blocks[0].StartCol)

				assert.Equal(t, 10, merged[0].Blocks[1].StartLine)
				assert.Equal(t, 13, merged[0].Blocks[1].StartCol)

				assert.Equal(t, 20, merged[0].Blocks[2].StartLine)
				assert.Equal(t, 5, merged[0].Blocks[2].StartCol)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := MergeProfiles(tt.profiles)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, merged)

			if tt.validate != nil {
				tt.validate(t, merged)
			}
		})
	}
}

func TestMergeProfilesWithTestFixtures(t *testing.T) {
	t.Run("merge multiple coverage files", func(t *testing.T) {
		// Load two separate coverage files
		data1 := loadTestFixture(t, "multiple_files_1.out")
		data2 := loadTestFixture(t, "multiple_files_2.out")

		profiles1, err := ParseProfiles(data1)
		require.NoError(t, err)

		profiles2, err := ParseProfiles(data2)
		require.NoError(t, err)

		// Merge them
		allProfiles := append(profiles1, profiles2...)
		merged, err := MergeProfiles(allProfiles)
		require.NoError(t, err)

		// Should have profiles for all unique files
		// multiple_files_1.out has file1.go and file2.go
		// multiple_files_2.out has file3.go and file4.go
		require.Len(t, merged, 4)

		// Verify files are sorted alphabetically
		assert.Equal(t, "github.com/example/project/file1.go", merged[0].FileName)
		assert.Equal(t, "github.com/example/project/file2.go", merged[1].FileName)
		assert.Equal(t, "github.com/example/project/file3.go", merged[2].FileName)
		assert.Equal(t, "github.com/example/project/file4.go", merged[3].FileName)
	})

	t.Run("merge then serialize", func(t *testing.T) {
		// Load multiple coverage files
		data1 := loadTestFixture(t, "multiple_files_1.out")
		data2 := loadTestFixture(t, "multiple_files_2.out")

		profiles1, err := ParseProfiles(data1)
		require.NoError(t, err)

		profiles2, err := ParseProfiles(data2)
		require.NoError(t, err)

		// Merge
		allProfiles := append(profiles1, profiles2...)
		merged, err := MergeProfiles(allProfiles)
		require.NoError(t, err)

		// Serialize
		serialized, err := SerializeProfiles(merged)
		require.NoError(t, err)

		// Parse back
		reparsed, err := ParseProfiles(serialized)
		require.NoError(t, err)

		// Should have same structure
		require.Len(t, reparsed, len(merged))
		for i := range merged {
			assert.Equal(t, merged[i].FileName, reparsed[i].FileName)
			assert.Equal(t, merged[i].Mode, reparsed[i].Mode)
			assert.Equal(t, len(merged[i].Blocks), len(reparsed[i].Blocks))
		}
	})
}

func TestMergeCount(t *testing.T) {
	tests := []struct {
		name string
		mode string
		a    int
		b    int
		want int
	}{
		{
			name: "set mode takes max (both 1)",
			mode: "set",
			a:    1,
			b:    1,
			want: 1,
		},
		{
			name: "set mode takes max (0 and 1)",
			mode: "set",
			a:    0,
			b:    1,
			want: 1,
		},
		{
			name: "set mode takes max (1 and 0)",
			mode: "set",
			a:    1,
			b:    0,
			want: 1,
		},
		{
			name: "count mode sums",
			mode: "count",
			a:    5,
			b:    3,
			want: 8,
		},
		{
			name: "count mode with zero",
			mode: "count",
			a:    5,
			b:    0,
			want: 5,
		},
		{
			name: "atomic mode sums",
			mode: "atomic",
			a:    100,
			b:    50,
			want: 150,
		},
		{
			name: "unknown mode defaults to sum",
			mode: "unknown",
			a:    10,
			b:    20,
			want: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCount(tt.mode, tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMakeBlockKey(t *testing.T) {
	tests := []struct {
		name        string
		block1      ProfileBlock
		block2      ProfileBlock
		shouldMatch bool
	}{
		{
			name: "identical blocks have same key",
			block1: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			block2: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 10, // Different count shouldn't matter
			},
			shouldMatch: true,
		},
		{
			name: "different start line",
			block1: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			block2: ProfileBlock{
				StartLine: 11, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			shouldMatch: false,
		},
		{
			name: "different start column",
			block1: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			block2: ProfileBlock{
				StartLine: 10, StartCol: 10, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			shouldMatch: false,
		},
		{
			name: "different end position",
			block1: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			block2: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 13, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			shouldMatch: false,
		},
		{
			name: "different num statements",
			block1: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 1, Count: 5,
			},
			block2: ProfileBlock{
				StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2,
				NumStmt: 2, Count: 5,
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := makeBlockKey(tt.block1)
			key2 := makeBlockKey(tt.block2)

			if tt.shouldMatch {
				assert.Equal(t, key1, key2)
			} else {
				assert.NotEqual(t, key1, key2)
			}
		})
	}
}

func TestCopyProfile(t *testing.T) {
	original := &Profile{
		FileName: "test.go",
		Mode:     "count",
		Blocks: []ProfileBlock{
			{StartLine: 10, StartCol: 5, EndLine: 12, EndCol: 2, NumStmt: 1, Count: 5},
			{StartLine: 20, StartCol: 10, EndLine: 22, EndCol: 3, NumStmt: 2, Count: 3},
		},
	}

	copied := copyProfile(original)

	// Verify copy is equal
	assert.Equal(t, original.FileName, copied.FileName)
	assert.Equal(t, original.Mode, copied.Mode)
	assert.Equal(t, original.Blocks, copied.Blocks)

	// Verify it's a deep copy by modifying the copy
	copied.FileName = "modified.go"
	copied.Blocks[0].Count = 999

	// Original should be unchanged
	assert.Equal(t, "test.go", original.FileName)
	assert.Equal(t, 5, original.Blocks[0].Count)
}

func TestMergeEmptyCoverage(t *testing.T) {
	t.Run("merge profile with empty blocks", func(t *testing.T) {
		profiles := []*Profile{
			{
				FileName: "empty.go",
				Mode:     "set",
				Blocks:   []ProfileBlock{},
			},
		}

		merged, err := MergeProfiles(profiles)
		require.NoError(t, err)
		require.Len(t, merged, 1)
		assert.Equal(t, "empty.go", merged[0].FileName)
		assert.Empty(t, merged[0].Blocks)
	})
}
