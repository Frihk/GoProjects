package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// loadJSONTestCase loads a JSON file and parses it into Room and Tunnel slices.
func loadJSONTestCase(filename string) ([]Room, []Tunnel, error) {
	path := filepath.Join("..", "tests", "data", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var graphData GraphData
	if err := json.Unmarshal(data, &graphData); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", filename, err)
	}

	return graphData.Nodes, graphData.Links, nil
}

// calculateTurns computes the number of turns required for a given set of routes.
// Formula: Turns = ceil((Ants + SumOfRouteLengths) / NumberOfRoutes) - 1
func calculateTurns(ants int, routes [][]string) int {
	if len(routes) == 0 {
		return 0
	}

	sumOfRouteLengths := 0
	for _, route := range routes {
		sumOfRouteLengths += len(route)
	}

	// Integer ceiling division: (a + b - 1) / b
	turns := ((ants + sumOfRouteLengths + len(routes) - 1) / len(routes)) - 1
	return turns
}

// TestFindPaths_AuditExamples tests the FindPaths function against standard audit cases.
func TestFindPaths_AuditExamples(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		ants        int
		maxTurns    int
		description string
	}{
		{
			name:        "Example00",
			filename:    "example00.json",
			ants:        4,
			maxTurns:    6,
			description: "Simple linear graph with 4 rooms",
		},
		{
			name:        "Example01",
			filename:    "example01.json",
			ants:        10,
			maxTurns:    8,
			description: "Medium complexity graph",
		},
		{
			name:        "Example02",
			filename:    "example02.json",
			ants:        20,
			maxTurns:    11,
			description: "Complex graph with multiple routes",
		},
		{
			name:        "Example03",
			filename:    "example03.json",
			ants:        4,
			maxTurns:    6,
			description: "Alternative small graph",
		},
		{
			name:        "Example04",
			filename:    "example04.json",
			ants:        9,
			maxTurns:    6,
			description: "Medium graph with 9 ants",
		},
		{
			name:        "Example05",
			filename:    "example05.json",
			ants:        9,
			maxTurns:    8,
			description: "Complex graph with 9 ants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms, tunnels, err := loadJSONTestCase(tt.filename)
			if err != nil {
				t.Fatalf("Failed to load test case %s: %v", tt.filename, err)
			}

			routes := FindPaths(tt.ants, rooms, tunnels)

			if len(routes) == 0 {
				t.Fatalf("Expected at least 1 route, got 0. No path found from start to end.")
			}

			turns := calculateTurns(tt.ants, routes)

			if turns > tt.maxTurns {
				t.Errorf("%s: Expected maximum %d turns, got %d turns\n"+
					"Routes found: %d\n"+
					"Route details: %v",
					tt.description, tt.maxTurns, turns, len(routes), routes)
			}
		})
	}
}

// TestFindPaths_GreedyShortcutTrap tests if the algorithm falls for the greedy shortest path trap.
// A truly optimal solution requires ignoring a short path that blocks two longer independent paths.
func TestFindPaths_GreedyShortcutTrap(t *testing.T) {
	// Graph structure:
	//        A ------- X ------- B
	//       /                     \
	//   Start                      End
	//       \                     /
	//        C ------- Y ------- D
	//         \                 /
	//          `----- (shortcut) ----'
	//
	// Top Route: Start -> A -> X -> B -> End (Length 5, including start and end)
	// Bottom Route: Start -> C -> Y -> D -> End (Length 5)
	// Shortcut: Start -> A -> D -> End (Length 4)
	//
	// With 10 ants:
	// - Greedy approach: Takes shortcut (1 route of length 4) = 13 turns
	// - Optimal approach: Uses both top and bottom routes (2 routes) = 9 turns

	rooms := []Room{
		{ID: "start", Group: "start", FX: 0, FY: 2},
		{ID: "end", Group: "end", FX: 10, FY: 2},
		{ID: "A", Group: "room", FX: 2, FY: 0},
		{ID: "X", Group: "room", FX: 5, FY: 0},
		{ID: "B", Group: "room", FX: 8, FY: 0},
		{ID: "C", Group: "room", FX: 2, FY: 4},
		{ID: "Y", Group: "room", FX: 5, FY: 4},
		{ID: "D", Group: "room", FX: 8, FY: 4},
	}

	tunnels := []Tunnel{
		// Top route
		{Source: "start", Target: "A"},
		{Source: "A", Target: "X"},
		{Source: "X", Target: "B"},
		{Source: "B", Target: "end"},
		// Bottom route
		{Source: "start", Target: "C"},
		{Source: "C", Target: "Y"},
		{Source: "Y", Target: "D"},
		{Source: "D", Target: "end"},
		// Shortcut (the trap!)
		{Source: "A", Target: "D"},
	}

	ants := 10
	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) == 0 {
		t.Fatal("No routes found. Algorithm failed to find any path.")
	}

	turns := calculateTurns(ants, routes)

	// The optimal solution should use 2 routes and complete in 9 turns or less
	expectedMaxTurns := 9

	if turns > expectedMaxTurns {
		t.Errorf("Greedy Shortcut Trap: Expected maximum %d turns (using 2 routes), got %d turns with %d route(s).\n"+
			"This indicates the algorithm took the greedy shortcut instead of finding vertex-disjoint paths.\n"+
			"Routes found: %v",
			expectedMaxTurns, turns, len(routes), routes)
	}

	// Verify we found at least 2 routes (the optimal solution)
	if len(routes) < 2 {
		t.Errorf("Expected at least 2 vertex-disjoint routes to avoid the greedy trap, found only %d route(s).\n"+
			"Routes: %v\n"+
			"A proper Edmonds-Karp implementation should find the top and bottom paths.",
			len(routes), routes)
	}
}

// TestFindPaths_VertexDisjointFlaw tests if the algorithm properly enforces vertex-disjointness.
// Two paths should NOT share intermediate rooms (only start and end can be shared).
func TestFindPaths_VertexDisjointFlaw(t *testing.T) {
	// Graph structure:
	//        A
	//       /   \
	//   Start    BottleneckRoom ---- End
	//       \   /              /
	//        C                /
	//         \              /
	//          B -----------'
	//
	// Path 1: Start -> A -> BottleneckRoom -> End
	// Path 2: Start -> C -> BottleneckRoom -> End (INVALID - shares BottleneckRoom)
	// Path 3: Start -> C -> B -> End (Valid alternative if exists)
	//
	// Because BottleneckRoom is a normal room (not start/end), it can only be used by ONE path.
	// The algorithm MUST enforce node splitting to prevent room overlap.

	rooms := []Room{
		{ID: "start", Group: "start", FX: 0, FY: 2},
		{ID: "end", Group: "end", FX: 10, FY: 2},
		{ID: "A", Group: "room", FX: 2, FY: 0},
		{ID: "BottleneckRoom", Group: "room", FX: 5, FY: 2},
		{ID: "C", Group: "room", FX: 2, FY: 4},
		{ID: "B", Group: "room", FX: 5, FY: 4},
	}

	tunnels := []Tunnel{
		{Source: "start", Target: "A"},
		{Source: "A", Target: "BottleneckRoom"},
		{Source: "BottleneckRoom", Target: "end"},
		{Source: "start", Target: "C"},
		{Source: "C", Target: "BottleneckRoom"}, // Both paths converge here!
		{Source: "C", Target: "B"},
		{Source: "B", Target: "end"},
	}

	ants := 4
	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) == 0 {
		t.Fatal("No routes found. Algorithm failed to find any path.")
	}

	// Verify vertex-disjointness: no two routes should share intermediate rooms
	roomUsage := make(map[string]int)
	for _, route := range routes {
		for i := 1; i < len(route)-1; i++ { // Skip start (index 0) and end (last index)
			room := route[i]
			roomUsage[room]++
			if roomUsage[room] > 1 {
				t.Errorf("Vertex-Disjoint Flaw: Room '%s' is shared by multiple routes!\n"+
					"Routes: %v\n"+
					"Each intermediate room can only appear in ONE path. Check node splitting implementation.",
					room, routes)
			}
		}
	}

	// Given the graph structure, with proper node splitting, only 2 vertex-disjoint paths exist:
	// Path 1: start -> A -> BottleneckRoom -> end
	// Path 2: start -> C -> B -> end
	expectedRoutes := 2
	if len(routes) != expectedRoutes {
		t.Errorf("Expected %d vertex-disjoint routes, found %d. Routes: %v",
			expectedRoutes, len(routes), routes)
	}
}

// TestFindPaths_ResidualEdgeReversal tests if the algorithm properly uses residual edges.
// This tests the core of Edmonds-Karp: flow cancellation via backward edges.
func TestFindPaths_ResidualEdgeReversal(t *testing.T) {
	// Graph structure:
	//         A ------- C ------- D
	//        /           |         \
	//   Start            |          End
	//        \           |         /
	//         B --------'     F --'
	//                        /
	//                   A --'
	//
	// Initial BFS might push flow: Start -> A -> C -> D -> End
	// But optimal flow requires:
	//   Path 1: Start -> B -> C -> D -> End
	//   Path 2: Start -> A -> F -> End
	//
	// To find Path 2, the algorithm must:
	// 1. Push flow Start -> A -> C -> D -> End
	// 2. On next iteration, travel the residual edge C -> A (backward)
	// 3. Realize A can redirect to F -> End
	// 4. Cancel the A -> C flow and establish A -> F flow

	rooms := []Room{
		{ID: "start", Group: "start", FX: 0, FY: 2},
		{ID: "end", Group: "end", FX: 10, FY: 2},
		{ID: "A", Group: "room", FX: 2, FY: 0},
		{ID: "B", Group: "room", FX: 2, FY: 4},
		{ID: "C", Group: "room", FX: 5, FY: 2},
		{ID: "D", Group: "room", FX: 8, FY: 0},
		{ID: "F", Group: "room", FX: 8, FY: 4},
	}

	tunnels := []Tunnel{
		{Source: "start", Target: "A"},
		{Source: "start", Target: "B"},
		{Source: "A", Target: "C"},
		{Source: "B", Target: "C"},
		{Source: "C", Target: "D"},
		{Source: "D", Target: "end"},
		{Source: "A", Target: "F"},
		{Source: "F", Target: "end"},
	}

	ants := 6
	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) == 0 {
		t.Fatal("No routes found. Algorithm failed to find any path.")
	}

	// The optimal solution should find 2 vertex-disjoint routes
	expectedRoutes := 2
	if len(routes) < expectedRoutes {
		t.Errorf("Residual Edge Reversal: Expected at least %d routes, found %d.\n"+
			"Routes: %v\n"+
			"This indicates the algorithm failed to use residual edges for flow cancellation.",
			expectedRoutes, len(routes), routes)
	}

	// Verify vertex-disjointness
	roomUsage := make(map[string]int)
	for _, route := range routes {
		for i := 1; i < len(route)-1; i++ {
			room := route[i]
			roomUsage[room]++
			if roomUsage[room] > 1 {
				t.Errorf("Residual Edge Reversal: Room '%s' is shared by multiple routes!\n"+
					"Routes: %v\n"+
					"The flow cancellation did not properly extract vertex-disjoint paths.",
					room, routes)
			}
		}
	}
}

// TestFindPaths_SinglePath tests a trivial case with only one possible path.
func TestFindPaths_SinglePath(t *testing.T) {
	rooms := []Room{
		{ID: "start", Group: "start", FX: 0, FY: 0},
		{ID: "A", Group: "room", FX: 1, FY: 0},
		{ID: "end", Group: "end", FX: 2, FY: 0},
	}

	tunnels := []Tunnel{
		{Source: "start", Target: "A"},
		{Source: "A", Target: "end"},
	}

	ants := 5
	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) != 1 {
		t.Errorf("Expected exactly 1 route, found %d", len(routes))
	}

	expectedRoute := []string{"A", "end"}
	if len(routes) > 0 {
		if len(routes[0]) != len(expectedRoute) {
			t.Errorf("Expected route length %d, got %d. Route: %v", len(expectedRoute), len(routes[0]), routes[0])
		}
	}

	turns := calculateTurns(ants, routes)
	expectedTurns := 6 // Formula: ((5 + 2 + 1 - 1) / 1) - 1 = 6
	if turns != expectedTurns {
		t.Errorf("Expected %d turns, got %d", expectedTurns, turns)
	}
}

// TestFindPaths_NoPath tests error handling when no path exists.
func TestFindPaths_NoPath(t *testing.T) {
	rooms := []Room{
		{ID: "start", Group: "start", FX: 0, FY: 0},
		{ID: "A", Group: "room", FX: 1, FY: 0},
		{ID: "end", Group: "end", FX: 2, FY: 0},
	}

	// No tunnel connecting start to end
	tunnels := []Tunnel{
		{Source: "start", Target: "A"},
		// Missing: A -> end connection
	}

	ants := 5
	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) != 0 {
		t.Errorf("Expected 0 routes when no path exists, found %d routes: %v", len(routes), routes)
	}
}
