package coverage

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FileDiff represents the changes to a single file in a diff.
type FileDiff struct {
	// OldName is the original filename (a/path/to/file)
	OldName string
	// NewName is the new filename (b/path/to/file)
	NewName string
	// AddedLines contains the line numbers that were added in this diff
	AddedLines []int
	// IsBinary indicates if this is a binary file
	IsBinary bool
	// IsRenamed indicates if the file was renamed
	IsRenamed bool
	// IsDeleted indicates if the file was deleted
	IsDeleted bool
}

// ParseDiff parses a unified diff format and returns the files and their added lines.
// The diff should be in the format produced by `git diff` or GitHub's diff output.
func ParseDiff(diffData []byte) ([]*FileDiff, error) {
	if len(diffData) == 0 {
		return nil, fmt.Errorf("diff data is empty")
	}

	scanner := bufio.NewScanner(bytes.NewReader(diffData))
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	var fileDiffs []*FileDiff
	var currentDiff *FileDiff
	var currentLine int // Track the current line number in the new file

	// Regex patterns for parsing diff format
	diffHeaderRe := regexp.MustCompile(`^diff --git a/(.+) b/(.+)$`)
	oldFileRe := regexp.MustCompile(`^--- a/(.+)$`)
	newFileRe := regexp.MustCompile(`^\+\+\+ b/(.+)$`)
	hunkHeaderRe := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	binaryFileRe := regexp.MustCompile(`^Binary files .+ differ$`)
	deletedFileRe := regexp.MustCompile(`^deleted file mode`)
	renamedFileRe := regexp.MustCompile(`^rename from (.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for diff header (start of new file)
		if matches := diffHeaderRe.FindStringSubmatch(line); matches != nil {
			// Save previous diff if exists
			if currentDiff != nil {
				fileDiffs = append(fileDiffs, currentDiff)
			}

			// Start new file diff
			currentDiff = &FileDiff{
				OldName:    matches[1],
				NewName:    matches[2],
				AddedLines: []int{},
			}
			currentLine = 0
			continue
		}

		if currentDiff == nil {
			continue
		}

		// Check for binary file marker
		if binaryFileRe.MatchString(line) {
			currentDiff.IsBinary = true
			continue
		}

		// Check for deleted file
		if deletedFileRe.MatchString(line) {
			currentDiff.IsDeleted = true
			continue
		}

		// Check for renamed file
		if matches := renamedFileRe.FindStringSubmatch(line); matches != nil {
			currentDiff.IsRenamed = true
			continue
		}

		// Check for old file name line
		if matches := oldFileRe.FindStringSubmatch(line); matches != nil {
			currentDiff.OldName = matches[1]
			continue
		}

		// Check for new file name line
		if matches := newFileRe.FindStringSubmatch(line); matches != nil {
			currentDiff.NewName = matches[1]
			continue
		}

		// Check for hunk header (@@ -old_start,old_lines +new_start,new_lines @@)
		if matches := hunkHeaderRe.FindStringSubmatch(line); matches != nil {
			// Extract the new file starting line number
			newStart, err := strconv.Atoi(matches[3])
			if err != nil {
				return nil, fmt.Errorf("invalid hunk header new start: %s", matches[3])
			}
			currentLine = newStart
			continue
		}

		// Parse diff content lines
		// Empty lines in diffs are context lines (they exist in both versions)
		if len(line) == 0 {
			// Empty line is a context line
			currentLine++
			continue
		}

		switch line[0] {
		case '+':
			// This is an added line (but not the +++ file header)
			if !strings.HasPrefix(line, "+++") {
				currentDiff.AddedLines = append(currentDiff.AddedLines, currentLine)
				currentLine++
			}
		case '-':
			// Removed line: don't advance currentLine (which tracks lines in the
			// *new* file). The `--- a/<path>` header line is also a '-' here, but
			// it has no impact since we never increment for this case anyway.
		case ' ':
			// This is a context line (present in both old and new)
			currentLine++
		case '\\':
			// This is a "\ No newline at end of file" marker, ignore
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading diff: %w", err)
	}

	// Don't forget to add the last file diff
	if currentDiff != nil {
		fileDiffs = append(fileDiffs, currentDiff)
	}

	return fileDiffs, nil
}

// GetAddedLinesByFile returns a map of filename to added line numbers.
// The filename is normalized to use the new name (after any renames).
// Binary files, deleted files, and non-Go files are excluded.
func GetAddedLinesByFile(fileDiffs []*FileDiff) map[string][]int {
	result := make(map[string][]int)

	for _, diff := range fileDiffs {
		// Skip binary files, deleted files, and files with no additions
		if diff.IsBinary || diff.IsDeleted || len(diff.AddedLines) == 0 {
			continue
		}

		// Use the new filename (normalized without a/ or b/ prefix)
		filename := diff.NewName

		// Only include Go files
		if !strings.HasSuffix(filename, ".go") {
			continue
		}

		result[filename] = diff.AddedLines
	}

	return result
}

// NormalizeFilename removes the a/ or b/ prefix from diff filenames.
func NormalizeFilename(filename string) string {
	if strings.HasPrefix(filename, "a/") || strings.HasPrefix(filename, "b/") {
		return filename[2:]
	}
	return filename
}
