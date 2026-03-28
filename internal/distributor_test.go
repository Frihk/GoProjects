package internal

import (
	"reflect"
	"testing"
)

func TestDistributeAnts(t *testing.T) {
	tests := []struct {
		name     string
		numAnts  int
		paths    [][]string
		expected []int
	}{
		{
			name:    "10 Ants across 1 short path and 1 long path",
			numAnts: 10,
			paths: [][]string{
				{"A", "end"},
				{"B", "C", "D", "E", "end"},
			},
			expected: []int{7, 3},
		},
		{
			name:    "5 Ants, 3 equal paths",
			numAnts: 5,
			paths: [][]string{
				{"A"},
				{"B"},
				{"C"},
			},
			expected: []int{2, 2, 1},
		},
		{
			name:    "Only long paths to absorb ants",
			numAnts: 3,
			paths: [][]string{
				{"A", "end"},
				{"B", "C", "D", "E", "end"},
			},
			expected: []int{3, 0},
		},
		{
			name:    "No ants",
			numAnts: 0,
			paths: [][]string{
				{"A", "end"},
			},
			expected: []int{0},
		},
		{
			name:     "No paths",
			numAnts:  10,
			paths:    [][]string{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignments := DistributeAnts(tt.numAnts, tt.paths)
			
			var actual []int
			
			if len(tt.paths) == 0 {
				if len(assignments) != 0 {
					t.Fatalf("Expected empty output for no paths, got %v", assignments)
				}
				return
			}
			
			if assignments == nil && len(tt.paths) > 0 {
				t.Fatalf("DistributeAnts returned nil but expected %v", tt.expected)
			}
			
			for _, expPath := range tt.paths {
				found := false
				for _, a := range assignments {
					if reflect.DeepEqual(a.Path, expPath) {
						actual = append(actual, a.Ants)
						found = true
						break
					}
				}
				if !found && assignments != nil {
					t.Fatalf("Path %v not found in assignments", expPath)
				}
			}

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("expected %v, actual %v", tt.expected, actual)
			}
		})
	}
}
