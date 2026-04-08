package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// jsonDir is the path (relative to the internal/ package) that holds all JSON test fixtures.
const jsonDir = "../tests/json"

// loadJSONTestCase loads a JSON file and parses it into Room and Tunnel slices.
func loadJSONTestCase(filename string) ([]Room, []Tunnel, error) {
	rooms, tunnels, _, err := loadJSONTestCaseFull(filename)
	return rooms, tunnels, err
}

// loadJSONTestCaseFull loads a JSON file and returns rooms, tunnels, and the ant count.
func loadJSONTestCaseFull(filename string) ([]Room, []Tunnel, int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var graphData GraphData
	if err := json.Unmarshal(data, &graphData); err != nil {
		return nil, nil, 0, fmt.Errorf("failed to unmarshal JSON from %s: %w", filename, err)
	}

	return graphData.Nodes, graphData.Links, graphData.Ants, nil
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
		{name: "Example00", filename: "example00.json", ants: 4, maxTurns: 6, description: "Simple linear graph with 4 rooms"},
		{name: "Example01", filename: "example01.json", ants: 10, maxTurns: 8, description: "Medium complexity graph"},
		{name: "Example02", filename: "example02.json", ants: 20, maxTurns: 11, description: "Complex graph with multiple routes"},
		{name: "Example03", filename: "example03.json", ants: 4, maxTurns: 6, description: "Alternative small graph"},
		{name: "Example04", filename: "example04.json", ants: 9, maxTurns: 6, description: "Medium graph with 9 ants"},
		{name: "Example05", filename: "example05.json", ants: 9, maxTurns: 8, description: "Complex graph with 9 ants"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms, tunnels, err := loadJSONTestCase(filepath.Join(jsonDir, tt.filename))
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
// Graph data is loaded from tests/samples/shortcut.txt (converted to tests/json/shortcut.json).
func TestFindPaths_GreedyShortcutTrap(t *testing.T) {
	// Graph structure:
	//        A ------- X ------- B
	//       /  \   (shortcut)     \
	//   Start   `-------------.    End
	//       \                  \  /
	//        C ------- Y ------- D
	//
	// Top Route: Start -> A -> X -> B -> End (Length 5, including start and end)
	// Bottom Route: Start -> C -> Y -> D -> End (Length 5)
	// Shortcut: Start -> A -> D -> End (Length 4)
	//
	// With 10 ants:
	// - Greedy approach: Takes shortcut (1 route of length 4) = 13 turns
	// - Optimal approach: Uses both top and bottom routes (2 routes) = 8 turns

	rooms, tunnels, ants, err := loadJSONTestCaseFull(filepath.Join(jsonDir, "shortcut.json"))
	if err != nil {
		t.Fatalf("Failed to load shortcut.json: %v", err)
	}

	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) == 0 {
		t.Fatal("No routes found. Algorithm failed to find any path.")
	}

	turns := calculateTurns(ants, routes)

	// The optimal solution should use 2 routes and complete in 9 turns or less
	expectedMaxTurns := 9

	if turns > expectedMaxTurns {
		t.Errorf("Greedy Shortcut Trap: Expected maximum %d turns (using 2 routes), got %d turns with %d route(s).\n"+
			"Routes found: %v", expectedMaxTurns, turns, len(routes), routes)
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
// Graph data is loaded from tests/samples/bottleneck.txt (converted to tests/json/bottleneck.json).
func TestFindPaths_VertexDisjointFlaw(t *testing.T) {
	// Graph structure:
	//         A
	//       /   \
	//   Start -- bottleneck --- End
	//       \   /              /
	//         C               /
	//          \             /
	//           B ----------'
	//
	// Despite three rooms (A, C, and start itself) funnelling into bottleneck,
	// node-splitting enforces that bottleneck can carry at most one path.
	// The only two vertex-disjoint paths are:
	//   Path 1: Start -> bottleneck -> End
	//   Path 2: Start -> C -> B -> End

	rooms, tunnels, ants, err := loadJSONTestCaseFull(filepath.Join(jsonDir, "bottleneck.json"))
	if err != nil {
		t.Fatalf("Failed to load bottleneck.json: %v", err)
	}

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
					"Routes: %v\n", room, routes)
			}
		}
	}

	// Given the graph structure, with proper node splitting, only 2 vertex-disjoint paths exist:
	// start -> Bottleneck -> end
	// start -> C -> B -> end
	expectedRoutes := 2
	if len(routes) != expectedRoutes {
		t.Errorf("Expected %d vertex-disjoint routes, found %d. Routes: %v",
			expectedRoutes, len(routes), routes)
	}
}

// TestFindPaths_ResidualEdgeReversal tests if the algorithm properly uses residual edges.
// This tests the core of Edmonds-Karp: flow cancellation via backward edges.
// Graph data is loaded from tests/samples/residual.txt (converted to tests/json/residual.json).
func TestFindPaths_ResidualEdgeReversal(t *testing.T) {
	// Graph structure:
	//       .---- A -------- E
	//      /      |           \
	//     /   B --|------- F   I
	//    /   / \  |          \ |
	//   Start   junction ---- End
	//    \   \ /  |         /  |
	//     \   C --|------- G   J
	//      \      |           /
	//       `---- D -------- H
	//
	// A naive BFS finds Start -> A -> junction -> End first, saturating the
	// junction bottleneck. Edmonds-Karp must then use residual (backward) edges
	// to cancel A's flow through junction and reroute A via E -> I -> End,
	// freeing junction for the D path.
	//
	// Optimal 4 vertex-disjoint paths:
	//   Path 1: Start -> A -> E -> I -> End
	//   Path 2: Start -> B -> F -> End
	//   Path 3: Start -> C -> G -> End
	//   Path 4: Start -> D -> junction -> End  (or D -> H -> J -> End)

	rooms, tunnels, ants, err := loadJSONTestCaseFull(filepath.Join(jsonDir, "residual.json"))
	if err != nil {
		t.Fatalf("Failed to load residual.json: %v", err)
	}

	routes := FindPaths(ants, rooms, tunnels)

	if len(routes) == 0 {
		t.Fatal("No routes found. Algorithm failed to find any path.")
	}

	// The graph has four independent exit paths from start (via A, B, C, D).
	// The optimal solution must discover all four vertex-disjoint routes.
	expectedRoutes := 4
	if len(routes) != expectedRoutes {
		t.Errorf("Residual Edge Reversal: Expected exactly %d routes, found %d.\n"+
			"Routes: %v\n"+
			"The algorithm may have failed to use residual edges to maximise flow.",
			expectedRoutes, len(routes), routes)
	}

	// Verify vertex-disjointness: no two routes may share an intermediate room.
	roomUsage := make(map[string]int)
	for _, route := range routes {
		for i := 0; i < len(route)-1; i++ { // exclude the final "end" element
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
