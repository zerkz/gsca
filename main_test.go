package main

import (
	"reflect"
	"testing"
)

func TestParseSelection(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []int
		max   int
	}{
		{
			name:  "single number",
			input: "1",
			want:  []int{0}, // 0-indexed
			max:   10,
		},
		{
			name:  "multiple numbers",
			input: "1,3,5",
			want:  []int{0, 2, 4},
			max:   10,
		},
		{
			name:  "range",
			input: "1-3",
			want:  []int{0, 1, 2},
			max:   10,
		},
		{
			name:  "mixed range and individual",
			input: "1,3-5,8",
			want:  []int{0, 2, 3, 4, 7},
			max:   10,
		},
		{
			name:  "with spaces",
			input: "1, 3-5 , 8",
			want:  []int{0, 2, 3, 4, 7},
			max:   10,
		},
		{
			name:  "duplicate entries",
			input: "1,1,3",
			want:  []int{0, 2},
			max:   10,
		},
		{
			name:  "wildcard select all",
			input: "*",
			want:  []int{0, 1, 2, 3, 4},
			max:   5,
		},
		{
			name:  "wildcard with spaces",
			input: " * ",
			want:  []int{0, 1, 2},
			max:   3,
		},
		{
			name:  "out of range",
			input: "1,15,3",
			want:  []int{0, 2},
			max:   10,
		},
		{
			name:  "invalid range (reversed)",
			input: "5-3",
			want:  nil, // Returns nil for invalid range
			max:   10,
		},
		{
			name:  "zero",
			input: "0,1",
			want:  []int{0},
			max:   10,
		},
		{
			name:  "empty string",
			input: "",
			want:  nil, // Returns nil for empty input
			max:   10,
		},
		{
			name:  "invalid characters",
			input: "abc,1,def",
			want:  []int{0},
			max:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSelection(tt.input, tt.max)

			// Handle nil vs empty slice comparison
			if tt.want == nil && len(got) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSelection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSelectionEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []int
		max   int
	}{
		{
			name:  "max boundary",
			input: "1-10",
			want:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			max:   10,
		},
		{
			name:  "beyond max",
			input: "1-15",
			want:  []int{},
			max:   10,
		},
		{
			name:  "reverse range",
			input: "10-1",
			want:  []int{},
			max:   10,
		},
		{
			name:  "single element range",
			input: "5-5",
			want:  []int{4},
			max:   10,
		},
		{
			name:  "complex mixed",
			input: "2,1,5-7,3,9,10",
			want:  []int{1, 0, 4, 5, 6, 2, 8, 9},
			max:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSelection(tt.input, tt.max)

			// For complex cases, just check length and containment
			if len(got) != len(tt.want) {
				t.Errorf("parseSelection() length = %v, want %v", len(got), len(tt.want))
				return
			}

			gotMap := make(map[int]bool)
			for _, idx := range got {
				gotMap[idx] = true
			}

			for _, idx := range tt.want {
				if !gotMap[idx] {
					t.Errorf("parseSelection() missing index %v", idx)
				}
			}
		})
	}
}
