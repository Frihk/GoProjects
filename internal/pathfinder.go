package internal

import (
	"math"

	"lem-in/internal/list"
)

type empty struct{}

type node struct {
	// name is the room identifier.
	name string
	// twin is a duplicate of this node that is used in the residual
	// graph of Edmond-Karp algorithm.
	twin *node
	// links is a map of all the rooms this node is connected to,
	// the value indicates the flow capacity of the link.
	links map[*node]int
}

type graph map[string]*node

// newGraph initialises a graph from the set of rooms and tunnels.
func newGraph(rooms []Room, tunnels []Tunnel) (start, end *node) {
	lookup := make(graph, len(rooms))

	for _, r := range rooms {
		nd := &node{name: r.ID, links: make(map[*node]int)}
		twin := nd

		switch r.Group {
		// the start and end nodes are not duplicated.
		case "start":
			start = nd
		case "end":
			end = nd
		default:
			twin = &node{name: r.ID, links: make(map[*node]int)}
		}

		nd.twin = twin
		twin.twin = nd
		lookup[nd.name] = nd
	}

	for _, t := range tunnels {
		entrance := lookup[t.Source]
		exit := lookup[t.Target]

		// links are bi-directional.
		entrance.links[exit] = 1
		exit.links[entrance] = 1
		// the links of the residual graph have a capacity of 0 to indicate they
		// are closed.
		entrance.twin.links[exit.twin] = 0
		exit.twin.links[entrance.twin] = 0
	}

	return start, end
}

// bfs performs a Breadth First Search on a graph for the `end` node.
// It returns a map of nodes' parent with `start` as the root node.
func bfs(start, end *node) map[*node]*node {
	// parent is a map of nodes' parent, beginning from the start node.
	// It allows us to trace the shortest path from the `start` node to
	// the `end` node.
	parent := map[*node]*node{start: nil}
	visited := map[*node]empty{start: {}}
	queue := list.New[*node]()
	validateAndPush := func(from, target *node, linkCapacity int) *node {
		// skip visited rooms and rooms with closed links.
		if _, seen := visited[target]; linkCapacity <= 0 || seen {
			return nil
		}

		parent[target] = from
		visited[target] = empty{}
		queue.PushBack(target)
		return target
	}

	queue.PushBack(start)
outerLoop:
	for queue.Len() > 0 {
		// pop node that is at the front of the queue.
		current := queue.Remove(queue.Front())

		// mark current node as visited.
		// add all of current's neighbouring nodes to the end of the queue.
		for neighbour, capacity := range current.links {
			if end == validateAndPush(current, neighbour, capacity) {
				break outerLoop
			}

			// also add the residual edge to the queue in order to explore alternate routes.
			capacity = current.twin.links[neighbour.twin]
			if end == validateAndPush(current.twin, neighbour.twin, capacity) {
				break outerLoop
			}
		}
	}

	return parent
}

// closeRoute marks the route traced from `end` to `start` in the parent map as closed.
func closeRoute(parent map[*node]*node, start, end *node) {
	for current := end; current != start; current = parent[current] {
		previous := parent[current]

		// close the forward edge
		previous.links[current]--
		// open the reverse edge
		current.twin.links[previous.twin]++
	}
}

// findExploredRoutes returns the set of discovered routes in a graph from
// `start` to `end`.
func findExploredRoutes(start, end *node) (routes [][]string) {
	for child, capacity := range start.links {
		if capacity >= 1 {
			continue
		}

		path := []string{start.name}
		current := child

		for {
			path = append(path, current.name)

			if current == end {
				break
			}

			for child, capacity = range current.links {
				if capacity <= 0 {
					current = child
					break
				}
			}
		}

		routes = append(routes, path)
	}

	return routes
}

// funcMap maps a given function onto every element of a slice and returns the
// results of each operation in a slice.
func funcMap[T1, T2 any](list []T1, f func(T1) T2) []T2 {
	output := make([]T2, 0, len(list))

	for _, v := range list {
		output = append(output, f(v))
	}

	return output
}

// sum returns the sum of all ints in a slice.
func sum(slc []int) int {
	res := 0

	for _, v := range slc {
		res += v
	}

	return res
}

// FindPaths returns the optimal set of routes to move a given number of ants from the
// `start` room to the `end` room using the Edmond-Karp algorithm.
func FindPaths(ants int, rooms []Room, tunnels []Tunnel) (optimalRoutes [][]string) {
	start, end := newGraph(rooms, tunnels)
	bestTurns := math.MaxInt

	for {
		parent := bfs(start, end)

		if _, exists := parent[end]; !exists {
			// There are no more routes to the `end` node to explore.
			break
		}

		closeRoute(parent, start, end)
		routes := findExploredRoutes(start, end)

		// Formula for calculating number of turns:
		// <total turns> = ceil((<number of ants> + <sum of all route lengths>) / <total number of routes>) ​- 1

		sumOfRouteLengths := sum(funcMap(routes, func(s []string) int { return len(s) }))
		turns := ((ants + sumOfRouteLengths + len(routes) - 1) / len(routes)) - 1

		if turns >= bestTurns {
			break
		}

		bestTurns = turns
		optimalRoutes = routes
	}

	return optimalRoutes
}
