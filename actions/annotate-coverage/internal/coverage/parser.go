package coverage

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/tools/cover"
)

// Profile represents a single coverage profile for a file.
type Profile struct {
	FileName string
	Mode     string
	Blocks   []ProfileBlock
}

// ProfileBlock represents a single block of code coverage.
type ProfileBlock struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}

// ParseProfiles parses coverage data in standard Go coverage format.
// It returns a slice of Profile structs representing the coverage data.
func ParseProfiles(data []byte) ([]*Profile, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("coverage data is empty")
	}

	// Use golang.org/x/tools/cover to parse the standard format
	profiles, err := cover.ParseProfilesFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage profiles: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no coverage profiles found in data")
	}

	// Convert to our internal format
	result := make([]*Profile, len(profiles))
	for i, p := range profiles {
		profile := &Profile{
			FileName: p.FileName,
			Mode:     p.Mode,
			Blocks:   make([]ProfileBlock, len(p.Blocks)),
		}
		for j, b := range p.Blocks {
			profile.Blocks[j] = ProfileBlock{
				StartLine: b.StartLine,
				StartCol:  b.StartCol,
				EndLine:   b.EndLine,
				EndCol:    b.EndCol,
				NumStmt:   b.NumStmt,
				Count:     b.Count,
			}
		}
		result[i] = profile
	}

	return result, nil
}

// ParseProfilesFromFile parses coverage data from a file path.
// This is primarily used when reading coverage files directly.
func ParseProfilesFromFile(path string, data []byte) ([]*Profile, error) {
	return ParseProfiles(data)
}

// ParseProfilesFromZip extracts and parses all coverage files from a zip archive.
// It looks for files matching common coverage patterns (*.out, *.cov, coverage.txt).
// Returns all parsed profiles from all coverage files found in the archive.
func ParseProfilesFromZip(zipData []byte) ([]*Profile, error) {
	if len(zipData) == 0 {
		return nil, fmt.Errorf("zip data is empty")
	}

	// Create a reader for the zip data
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip archive: %w", err)
	}

	var allProfiles []*Profile

	// Iterate through files in the archive
	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Check if this is a coverage file based on naming patterns
		name := strings.ToLower(file.Name)
		if !isCoverageFile(name) {
			continue
		}

		// Open the file
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s in archive: %w", file.Name, err)
		}

		// Read the contents
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s from archive: %w", file.Name, err)
		}

		// Skip empty files
		if len(data) == 0 {
			continue
		}

		// Parse the coverage data
		profiles, err := ParseProfiles(data)
		if err != nil {
			// Log but don't fail on individual file parse errors
			// This allows partial success if some files are malformed
			continue
		}

		allProfiles = append(allProfiles, profiles...)
	}

	if len(allProfiles) == 0 {
		return nil, fmt.Errorf("no valid coverage files found in archive")
	}

	return allProfiles, nil
}

// isCoverageFile checks if a filename matches common coverage file patterns.
func isCoverageFile(name string) bool {
	return strings.HasSuffix(name, ".out") ||
		strings.HasSuffix(name, ".cov") ||
		strings.HasSuffix(name, "coverage.txt") ||
		strings.Contains(name, "coverage") && strings.HasSuffix(name, ".txt")
}

// ValidateProfile checks if a coverage profile is well-formed.
// It returns an error if the profile has invalid block data.
func ValidateProfile(p *Profile) error {
	if p == nil {
		return fmt.Errorf("profile is nil")
	}

	if p.FileName == "" {
		return fmt.Errorf("profile has empty filename")
	}

	if p.Mode == "" {
		return fmt.Errorf("profile has empty mode")
	}

	// Validate each block
	for i, block := range p.Blocks {
		if err := validateBlock(block, i); err != nil {
			return fmt.Errorf("invalid block in %s: %w", p.FileName, err)
		}
	}

	return nil
}

// validateBlock checks if a single coverage block is valid.
func validateBlock(b ProfileBlock, index int) error {
	if b.StartLine <= 0 {
		return fmt.Errorf("block %d has invalid start line: %d", index, b.StartLine)
	}

	if b.EndLine <= 0 {
		return fmt.Errorf("block %d has invalid end line: %d", index, b.EndLine)
	}

	if b.EndLine < b.StartLine {
		return fmt.Errorf("block %d has end line (%d) before start line (%d)", index, b.EndLine, b.StartLine)
	}

	if b.StartLine == b.EndLine && b.EndCol < b.StartCol {
		return fmt.Errorf("block %d has end column (%d) before start column (%d) on same line", index, b.EndCol, b.StartCol)
	}

	if b.NumStmt < 0 {
		return fmt.Errorf("block %d has negative statement count: %d", index, b.NumStmt)
	}

	if b.Count < 0 {
		return fmt.Errorf("block %d has negative count: %d", index, b.Count)
	}

	return nil
}

// SerializeProfiles converts profiles back to standard Go coverage format.
// This is used when saving merged coverage data.
func SerializeProfiles(profiles []*Profile) ([]byte, error) {
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles to serialize")
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	// Write header with mode from first profile
	mode := profiles[0].Mode
	if mode == "" {
		mode = "set" // default mode
	}
	fmt.Fprintf(writer, "mode: %s\n", mode)

	// Write each profile
	for _, p := range profiles {
		for _, b := range p.Blocks {
			fmt.Fprintf(writer, "%s:%d.%d,%d.%d %d %d\n",
				p.FileName,
				b.StartLine, b.StartCol,
				b.EndLine, b.EndCol,
				b.NumStmt, b.Count)
		}
	}

	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush writer: %w", err)
	}

	return buf.Bytes(), nil
}
