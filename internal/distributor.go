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
		assignments[i] = PathAssignment{
			Path: p,
			Ants: 0,
		}
	}

	for i := 0; i < numAnts; i++ {
		best := 0

		for j := 1; j < len(assignments); j++ {
			costBest := len(assignments[best].Path) + assignments[best].Ants
			costJ := len(assignments[j].Path) + assignments[j].Ants

			if costJ < costBest {
				best = j
			}
		}

		assignments[best].Ants++
	}

	return assignments
}