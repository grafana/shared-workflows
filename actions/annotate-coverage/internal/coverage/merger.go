package coverage

import (
	"fmt"
	"sort"
)

// MergeProfiles merges multiple coverage profiles into a single profile.
// It implements the gocovmerge algorithm, which merges coverage blocks
// at the block level, handling overlapping coverage from multiple test runs.
//
// The algorithm:
// 1. Groups profiles by file name
// 2. For each file, merges all blocks using additive merging
// 3. Uses the mode from the first profile (all profiles must use same mode)
// 4. Returns merged profiles sorted by file name
func MergeProfiles(profiles []*Profile) ([]*Profile, error) {
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles to merge")
	}

	// If only one profile, still need to sort blocks
	if len(profiles) == 1 {
		sorted := copyProfile(profiles[0])
		// Sort blocks by position
		sort.Slice(sorted.Blocks, func(i, j int) bool {
			if sorted.Blocks[i].StartLine != sorted.Blocks[j].StartLine {
				return sorted.Blocks[i].StartLine < sorted.Blocks[j].StartLine
			}
			return sorted.Blocks[i].StartCol < sorted.Blocks[j].StartCol
		})
		return []*Profile{sorted}, nil
	}

	// Validate that all profiles use the same mode
	mode := profiles[0].Mode
	for i, p := range profiles {
		if p.Mode != mode {
			return nil, fmt.Errorf("profile %d has mode %q, expected %q", i, p.Mode, mode)
		}
	}

	// Group profiles by file name
	fileMap := make(map[string][]*Profile)
	for _, p := range profiles {
		fileMap[p.FileName] = append(fileMap[p.FileName], p)
	}

	// Merge blocks for each file
	var result []*Profile
	for fileName, fileProfiles := range fileMap {
		merged := mergeFileProfiles(fileName, mode, fileProfiles)
		result = append(result, merged)
	}

	// Sort by file name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].FileName < result[j].FileName
	})

	return result, nil
}

// mergeFileProfiles merges all profiles for a single file.
// This implements the core gocovmerge algorithm: combine all blocks
// and add their counts for identical blocks.
func mergeFileProfiles(fileName, mode string, profiles []*Profile) *Profile {
	// Collect all blocks
	var allBlocks []ProfileBlock
	for _, p := range profiles {
		allBlocks = append(allBlocks, p.Blocks...)
	}

	// If only one profile, return it directly
	if len(profiles) == 1 {
		return &Profile{
			FileName: fileName,
			Mode:     mode,
			Blocks:   allBlocks,
		}
	}

	// Group identical blocks and sum their counts
	blockMap := make(map[blockKey]*ProfileBlock)
	for _, block := range allBlocks {
		key := makeBlockKey(block)
		if existing, ok := blockMap[key]; ok {
			// Merge counts based on mode
			existing.Count = mergeCount(mode, existing.Count, block.Count)
		} else {
			// Create a copy of the block
			blockCopy := block
			blockMap[key] = &blockCopy
		}
	}

	// Convert map to slice and sort by position
	var mergedBlocks []ProfileBlock
	for _, block := range blockMap {
		mergedBlocks = append(mergedBlocks, *block)
	}

	// Sort blocks by start position (line, then column)
	sort.Slice(mergedBlocks, func(i, j int) bool {
		if mergedBlocks[i].StartLine != mergedBlocks[j].StartLine {
			return mergedBlocks[i].StartLine < mergedBlocks[j].StartLine
		}
		return mergedBlocks[i].StartCol < mergedBlocks[j].StartCol
	})

	return &Profile{
		FileName: fileName,
		Mode:     mode,
		Blocks:   mergedBlocks,
	}
}

// blockKey uniquely identifies a coverage block by its position.
type blockKey struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
}

// makeBlockKey creates a unique key for a coverage block.
func makeBlockKey(block ProfileBlock) blockKey {
	return blockKey{
		StartLine: block.StartLine,
		StartCol:  block.StartCol,
		EndLine:   block.EndLine,
		EndCol:    block.EndCol,
		NumStmt:   block.NumStmt,
	}
}

// mergeCount combines two block counts based on the coverage mode.
// - "set": max(a, b) - any execution counts as covered
// - "count": a + b - sum execution counts
// - "atomic": a + b - sum execution counts (from atomic operations)
func mergeCount(mode string, a, b int) int {
	switch mode {
	case "set":
		// In set mode, any execution means the block is covered
		// We take the max, but for set mode this is typically 0 or 1
		if a > b {
			return a
		}
		return b
	case "count", "atomic":
		// In count/atomic mode, we sum the execution counts
		return a + b
	default:
		// Unknown mode, default to sum
		return a + b
	}
}

// copyProfile creates a deep copy of a profile.
func copyProfile(p *Profile) *Profile {
	blocks := make([]ProfileBlock, len(p.Blocks))
	copy(blocks, p.Blocks)
	return &Profile{
		FileName: p.FileName,
		Mode:     p.Mode,
		Blocks:   blocks,
	}
}
