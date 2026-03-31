package internal

import (
	"math"

	"lem-in/internal/list"
)

type empty struct{}

// node represents a room (or half a room for In/Out splitting)
type node struct {
	name     string
	isOut    bool    // Identifies if this is the 'Out' door of a split room
	outgoing []*edge // Forward physical paths leaving this node
	incoming []*edge // Reverse residual "undo" paths leaving this node
}

// edge represents a directional tunnel or the internal capacity of a room
type edge struct {
	from     *node
	to       *node
	capacity int
	reverse  *edge // Pointer to the twin residual edge
}

// addEdge creates a forward edge and its residual reverse edge, linking them.
func addEdge(from, to *node, capacity int) {
	forward := &edge{from: from, to: to, capacity: capacity}
	backward := &edge{from: to, to: from, capacity: 0}

	forward.reverse = backward
	backward.reverse = forward

	// The forward edge goes into the physical outgoing slice
	from.outgoing = append(from.outgoing, forward)
	// The backward edge goes into the residual incoming slice of the target node
	to.incoming = append(to.incoming, backward)
}

// newGraph initializes the graph using In/Out node splitting to enforce 1 ant per room.
func newGraph(rooms []Room, tunnels []Tunnel) (start, end *node) {
	lookup := make(map[string]*node)

	// Step 1: Create nodes and split normal rooms
	for _, r := range rooms {
		if r.Group == "start" {
			start = &node{name: r.ID, isOut: true}
			lookup[r.ID+"_in"] = start
			lookup[r.ID+"_out"] = start
		} else if r.Group == "end" {
			end = &node{name: r.ID, isOut: false}
			lookup[r.ID+"_in"] = end
			lookup[r.ID+"_out"] = end
		} else {
			// Normal rooms are split to enforce the bottleneck
			inNode := &node{name: r.ID, isOut: false}
			outNode := &node{name: r.ID, isOut: true}

			// Internal room capacity: 1 ant per room
			addEdge(inNode, outNode, 1)

			lookup[r.ID+"_in"] = inNode
			lookup[r.ID+"_out"] = outNode
		}
	}

	// Step 2: Connect the tunnels
	for _, t := range tunnels {
		sourceOut := lookup[t.Source+"_out"]
		sourceIn := lookup[t.Source+"_in"]
		targetOut := lookup[t.Target+"_out"]
		targetIn := lookup[t.Target+"_in"]

		// Tunnels are bidirectional, connecting Out doors to In doors
		addEdge(sourceOut, targetIn, 1)
		addEdge(targetOut, sourceIn, 1)
	}

	return start, end
}

// bfs finds the shortest path in the residual graph.
// It returns a map linking a node to the *edge* used to reach it.
func bfs(start, end *node) map[*node]*edge {
	parent := make(map[*node]*edge)
	visited := make(map[*node]empty)

	queue := list.New[*node]()
	queue.PushBack(start)
	visited[start] = empty{}

	for queue.Len() > 0 {
		current := queue.Remove(queue.Front())

		if current == end {
			break
		}

		// 1. Explore forward physical edges
		for _, e := range current.outgoing {
			if e.capacity > 0 {
				if _, seen := visited[e.to]; !seen {
					visited[e.to] = empty{}
					parent[e.to] = e
					queue.PushBack(e.to)
				}
			}
		}

		// 2. Explore reverse residual edges (The "Undo" paths)
		for _, e := range current.incoming {
			if e.capacity > 0 {
				if _, seen := visited[e.to]; !seen {
					visited[e.to] = empty{}
					parent[e.to] = e
					queue.PushBack(e.to)
				}
			}
		}
	}

	return parent
}

// extractPaths walks the graph from start to end, strictly following
// original physical edges that have been fully consumed (capacity == 0).
func extractPaths(start, end *node) [][]string {
	var paths [][]string

	// Look at all physical outgoing connections from the Start room
	for _, startEdge := range start.outgoing {
		// If flow went through this forward edge
		if startEdge.capacity == 0 {
			var path []string
			currNode := startEdge.to

			// Walk the flow until we hit the End room
			for currNode != end {
				// Only record the name when passing the 'Out' door to avoid duplicates
				if currNode.isOut {
					path = append(path, currNode.name)
				}

				// Find the next consumed physical edge
				for _, nextEdge := range currNode.outgoing {
					if nextEdge.capacity == 0 {
						currNode = nextEdge.to
						break
					}
				}
			}
			// Finally, append the end room
			path = append(path, end.name)
			paths = append(paths, path)
		}
	}

	return paths
}

// FindPaths runs the Edmonds-Karp loop to find the optimal set of routes.
func FindPaths(ants int, rooms []Room, tunnels []Tunnel) [][]string {
	start, end := newGraph(rooms, tunnels)

	bestTurns := math.MaxInt
	var optimalRoutes [][]string

	for {
		parentMap := bfs(start, end)

		// If BFS cannot reach the end, no more paths exist
		if _, exists := parentMap[end]; !exists {
			break
		}

		// Backtrack to update residual capacities (The Undo Mechanism)
		curr := end
		for curr != start {
			incomingEdge := parentMap[curr]

			// Consume forward capacity, create reverse capacity
			incomingEdge.capacity--
			incomingEdge.reverse.capacity++

			curr = incomingEdge.from
		}

		// Extract current clean paths
		currentRoutes := extractPaths(start, end)

		// Calculate turns required for this configuration
		sumOfRouteLengths := 0
		for _, r := range currentRoutes {
			sumOfRouteLengths += len(r)
		}

		numberOfRoutes := len(currentRoutes)
		turns := (ants+sumOfRouteLengths+numberOfRoutes-1)/numberOfRoutes - 1

		// Optimization Check: Stop if adding this path makes us slower
		if turns >= bestTurns {
			break
		}

		bestTurns = turns

		// Create a deep copy of the current routes to save as the best state
		optimalRoutes = make([][]string, len(currentRoutes))
		for i, route := range currentRoutes {
			optimalRoutes[i] = make([]string, len(route))
			copy(optimalRoutes[i], route)
		}
	}

	return optimalRoutes
}
