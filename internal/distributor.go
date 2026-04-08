package internal

type PathAssignment struct {
	Path []string
	Ants int
}

func DistributeAnts(numAnts int, paths [][]string) []PathAssignment {
	if len(paths) == 0 {
		return []PathAssignment{}
	}

	assignments := make([]PathAssignment, len(paths))
	for i, p := range paths {
		assignments[i] = PathAssignment{Path: p, Ants: 0}
	}

	for ant := 0; ant < numAnts; ant++ {
		bestPathIdx := 0
		bestArrivalCost := len(paths[0]) + assignments[0].Ants

		for i := 1; i < len(paths); i++ {
			arrivalCost := len(paths[i]) + assignments[i].Ants
			if arrivalCost < bestArrivalCost {
				bestArrivalCost = arrivalCost
				bestPathIdx = i
			}
		}

		assignments[bestPathIdx].Ants++
	}

	return assignments
}
