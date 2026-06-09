package github

import "sort"

// LineRange represents a range of consecutive line numbers.
type LineRange struct {
	Start int
	End   int
}

// Annotation represents a GitHub Check Run annotation.
type Annotation struct {
	Path      string
	StartLine int
	EndLine   int
	Level     string // "notice", "warning", "failure"
	Title     string
	Message   string
}

// GroupIntoRanges groups consecutive line numbers into ranges.
// Lines must be sorted in ascending order.
func GroupIntoRanges(lines []int) []LineRange {
	if len(lines) == 0 {
		return nil
	}

	var ranges []LineRange
	rangeStart := lines[0]
	rangeEnd := lines[0]

	for i := 1; i < len(lines); i++ {
		if lines[i] == rangeEnd+1 {
			// Continue the range
			rangeEnd = lines[i]
		} else {
			// End the current range and start a new one
			ranges = append(ranges, LineRange{Start: rangeStart, End: rangeEnd})
			rangeStart = lines[i]
			rangeEnd = lines[i]
		}
	}

	// Append the final range
	ranges = append(ranges, LineRange{Start: rangeStart, End: rangeEnd})

	return ranges
}

// SortAndGroupLines sorts line numbers and groups them into ranges.
// This is a convenience function that combines sorting and grouping.
func SortAndGroupLines(lines []int) []LineRange {
	if len(lines) == 0 {
		return nil
	}

	// Make a copy to avoid mutating the input
	sorted := make([]int, len(lines))
	copy(sorted, lines)
	sort.Ints(sorted)

	return GroupIntoRanges(sorted)
}
