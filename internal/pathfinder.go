package internal

import (
	"math"

	"lem-in/internal/list"
)

type (
	empty     struct{}
	roomType  int
	linkState int8
)

type node struct {
	// name is the room identifier.
	name string
	// group is the category of the room (normal|alpha|omega).
	group roomType
	// links is a map of all the rooms this node is connected to,
	// the value indicates the accessability of the link.
	links map[*node]linkState
}

type graph map[string]*node

const (
	// normal is the basic room type.
	normal roomType = iota
	// alpha is the start room.
	alpha
	// omega is the end room.
	omega
)

const (
	open linkState = iota
	closed
)

// initGraph initialises a graph from the set of rooms and tunnels.
func initGraph(rooms []Room, tunnels []Tunnel) (start, end *node, farm graph) {
	farm = graph{}
	for _, r := range rooms {
		nd := &node{name: r.ID, links: make(map[*node]linkState)}

		switch r.Group {
		case "start":
			nd.group = alpha
			start = nd
		case "end":
			nd.group = omega
			end = nd
		}

		farm[nd.name] = nd
	}

	for _, t := range tunnels {
		entrance := farm[t.Source]
		exit := farm[t.Target]

		// links are by directional.
		entrance.links[exit] = open
		exit.links[entrance] = open
	}

	return start, end, farm
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

	queue.PushBack(start)
outerLoop:
	for queue.Len() > 0 {
		// pop node that is at the front of the queue.
		current := queue.Remove(queue.Front())

		// mark current node as visited.
		// add all of current's neighbouring nodes to the end of the queue.
		for neighbour, linkStatus := range current.links {
			// skip visited rooms and rooms with closed links.
			if _, seen := visited[neighbour]; linkStatus == closed || seen {
				continue
			}

			parent[neighbour] = current
			if neighbour == end {
				break outerLoop
			}

			visited[neighbour] = empty{}
			queue.PushBack(neighbour)
		}
	}

	return parent
}

// closeRoute marks the route traced from `end` to `start` in the parent map as closed.
func closeRoute(parent map[*node]*node, start, end *node) {
	for current := end; current != start; current = parent[current] {
		previous := parent[current]

		previous.links[current] = closed
	}
}

// findExploredRoutes returns the set of discovered routes in a graph from
// `start` to `end`.
func findExploredRoutes(start, end *node) (routes [][]string) {
	for child, linkStatus := range start.links {
		if linkStatus == open {
			continue
		}

		path := []string{start.name}
		current := child

		for {
			path = append(path, current.name)

			if current == end {
				break
			}

			for child, linkStatus = range current.links {
				if linkStatus == closed {
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
// `start` room to the `end` room.
func FindPaths(ants int, rooms []Room, tunnels []Tunnel) (optimalRoutes [][]string) {
	start, end, _ := initGraph(rooms, tunnels)
	bestTurns := math.MaxInt

	for {
		parent := bfs(start, end)

		if _, exists := parent[end]; !exists {
			// There are no more routes to the `end` node to explore.
			break
		}

		closeRoute(parent, start, end)
		routes := findExploredRoutes(start, end)

		// Formula for calculating number of turns.
		// <total turns> = ceil(
		// 		(<number of ants> + <sum of all route lengths>) /
		// 		<total number of routes>
		// ) ​- 1

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
