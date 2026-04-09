package internal

import (
	"fmt"
	"strings"
)

type Ant struct {
	ID       int
	Path     []string
	Position int
}

func Simulate(ants int, paths [][]string, endRoom string) {
	assignments := DistributeAnts(ants, paths)

	var allAnts []Ant
	for i, a := range assignments {
		for range a.Ants {
			allAnts = append(allAnts, Ant{ID: i + 1, Path: a.Path})
		}
	}

	var moves []string
	capacity := make(map[string]int)
	allFinished := false

	for !allFinished {
		moves = moves[:0]
		allFinished = true

		for i := range allAnts {
			ant := &allAnts[i]

			if ant.Position >= len(ant.Path) {
				continue
			}

			// Leave previous room
			if ant.Position > 0 {
				prevRoom := ant.Path[ant.Position-1]
				capacity[prevRoom]--
			}

			allFinished = false
			room := ant.Path[ant.Position]

			// Advance ant if room is not full
			if capacity[room] < 1 {
				if room != endRoom {
					capacity[room]++
				}

				ant.Position++
				moves = append(moves, fmt.Sprintf("L%d-%s", ant.ID, room))
			}
		}

		if len(moves) > 0 {
			fmt.Println(strings.Join(moves, " "))
		}
	}
}
