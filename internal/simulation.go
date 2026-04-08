package internal

import (
	"fmt"
	"strings"
)

func Simulate(ants int, paths [][]string, start, end string) {
	assignments := DistributeAnts(ants, paths)

	var allAnts []Ant
	id := 1

	for _, a := range assignments {
		for i := 0; i < a.Ants; i++ {
			fullPath := append([]string{start}, append(a.Path, end)...)

			allAnts = append(allAnts, Ant{
				ID:       id,
				Path:     fullPath,
				Position: 0,
			})
			id++
		}
	}
	for {
		var moves []string
		occupied := make(map[string]bool)
		allFinished := true

		for i := len(allAnts) - 1; i >= 0; i-- {
			ant := &allAnts[i]

			if ant.Finished {
				continue
			}

			allFinished = false

			next := ant.Position + 1
			if next >= len(ant.Path) {
				continue
			}

			nextRoom := ant.Path[next]

			if nextRoom == end || !occupied[nextRoom] {

				// free current
				if ant.Position > 0 {
					curr := ant.Path[ant.Position]
					if curr != start {
						delete(occupied, curr)
					}
				}

				ant.Position++

				if nextRoom != end {
					occupied[nextRoom] = true
				}

				moves = append(moves,
					fmt.Sprintf("L%d-%s", ant.ID, nextRoom),
				)

				if nextRoom == end {
					ant.Finished = true
				}
			}
		}

		if allFinished {
			break
		}

		if len(moves) > 0 {
			fmt.Println(strings.Join(moves, " "))
		}
	}
}