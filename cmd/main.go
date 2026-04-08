package main

import (
	"fmt"
	"os"

	"lem-in/internal"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "ERROR: No input file provided")
		os.Exit(1)
	}

	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer f.Close()

	lines, err := internal.ReadAllLines(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	data, err := internal.ParseInput(lines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	paths := internal.FindPaths(data.Ants, data.Nodes, data.Links)

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "No valid paths found")
		return
	}

	internal.Simulate(data.Ants, paths, data.End)
}
