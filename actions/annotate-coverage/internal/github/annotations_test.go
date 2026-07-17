package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupIntoRanges(t *testing.T) {
	tests := []struct {
		name     string
		lines    []int
		expected []LineRange
	}{
		{
			name:     "empty list",
			lines:    []int{},
			expected: nil,
		},
		{
			name:  "single line",
			lines: []int{5},
			expected: []LineRange{
				{Start: 5, End: 5},
			},
		},
		{
			name:  "two consecutive lines",
			lines: []int{5, 6},
			expected: []LineRange{
				{Start: 5, End: 6},
			},
		},
		{
			name:  "three consecutive lines",
			lines: []int{5, 6, 7},
			expected: []LineRange{
				{Start: 5, End: 7},
			},
		},
		{
			name:  "non-consecutive lines",
			lines: []int{5, 10, 15},
			expected: []LineRange{
				{Start: 5, End: 5},
				{Start: 10, End: 10},
				{Start: 15, End: 15},
			},
		},
		{
			name:  "mixed consecutive and non-consecutive",
			lines: []int{5, 6, 7, 10, 15, 16},
			expected: []LineRange{
				{Start: 5, End: 7},
				{Start: 10, End: 10},
				{Start: 15, End: 16},
			},
		},
		{
			name:  "all consecutive",
			lines: []int{1, 2, 3, 4, 5},
			expected: []LineRange{
				{Start: 1, End: 5},
			},
		},
		{
			name:  "complex pattern",
			lines: []int{1, 3, 4, 5, 8, 9, 12},
			expected: []LineRange{
				{Start: 1, End: 1},
				{Start: 3, End: 5},
				{Start: 8, End: 9},
				{Start: 12, End: 12},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupIntoRanges(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortAndGroupLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []int
		expected []LineRange
	}{
		{
			name:     "empty list",
			lines:    []int{},
			expected: nil,
		},
		{
			name:  "already sorted",
			lines: []int{1, 2, 3, 5, 7, 8},
			expected: []LineRange{
				{Start: 1, End: 3},
				{Start: 5, End: 5},
				{Start: 7, End: 8},
			},
		},
		{
			name:  "unsorted lines",
			lines: []int{8, 1, 5, 3, 2, 7},
			expected: []LineRange{
				{Start: 1, End: 3},
				{Start: 5, End: 5},
				{Start: 7, End: 8},
			},
		},
		{
			name:  "reverse sorted",
			lines: []int{15, 10, 7, 6, 5, 1},
			expected: []LineRange{
				{Start: 1, End: 1},
				{Start: 5, End: 7},
				{Start: 10, End: 10},
				{Start: 15, End: 15},
			},
		},
		{
			name:  "duplicates create separate single-line ranges",
			lines: []int{5, 5, 6, 6, 10, 10},
			expected: []LineRange{
				{Start: 5, End: 5},
				{Start: 5, End: 6},
				{Start: 6, End: 6},
				{Start: 10, End: 10},
				{Start: 10, End: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortAndGroupLines(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortAndGroupLines_DoesNotMutateInput(t *testing.T) {
	original := []int{10, 5, 8, 1, 3}
	input := make([]int, len(original))
	copy(input, original)

	_ = SortAndGroupLines(input)

	// Verify input wasn't mutated
	assert.Equal(t, original, input, "SortAndGroupLines should not mutate the input slice")
}
