package internal

import (
	"fmt"
	"strings"
)

// Move represents a single ant movement in a turn.
type Move struct {
	AntID    int    `json:"antId"`
	RoomName string `json:"roomName"`
}

// Step represents all ant movements in a single turn.
type Step struct {
	Moves []Move `json:"moves"`
}

type Ant struct {
	ID       int
	Path     []string
	Position int
}

// SimulateSteps runs the simulation and returns the structured steps.
func SimulateSteps(ants int, paths [][]string, endRoom string) []Step {
	assignments := DistributeAnts(ants, paths)

	var allAnts []Ant
	antID := 1
	for _, a := range assignments {
		for range a.Ants {
			allAnts = append(allAnts, Ant{ID: antID, Path: a.Path})
			antID++
		}
	}

	var steps []Step
	capacity := make(map[string]int)
	allFinished := false

	for !allFinished {
		var moves []Move
		allFinished = true

		for i := range allAnts {
			ant := &allAnts[i]

			if ant.Position >= len(ant.Path) {
				continue
			}

			if ant.Position > 0 {
				prevRoom := ant.Path[ant.Position-1]
				capacity[prevRoom]--
			}

			allFinished = false
			room := ant.Path[ant.Position]

			if capacity[room] < 1 {
				if room != endRoom {
					capacity[room]++
				}

				ant.Position++
				moves = append(moves, Move{AntID: ant.ID, RoomName: room})
			}
		}

		if len(moves) > 0 {
			steps = append(steps, Step{Moves: moves})
		}
	}

	return steps
}

// Simulate prints the simulation output to stdout (original behavior).
func Simulate(ants int, paths [][]string, endRoom string) {
	steps := SimulateSteps(ants, paths, endRoom)
	for _, step := range steps {
		var parts []string
		for _, m := range step.Moves {
			parts = append(parts, fmt.Sprintf("L%d-%s", m.AntID, m.RoomName))
		}
		fmt.Println(strings.Join(parts, " "))
	}
}
