package main

import (
	"fmt"
	"os"

	"lem-in/internal"
)

func main() {
	if len(os.Args) != 2 { 
		fmt.Println("ERROR: No input file provided")
	}

	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	lines, err := internal.ReadAllLines(f)
	for _, l := range lines {
		fmt.Println(l)
	}

	fmt.Println()

	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := internal.ParseInput(lines)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 🔥 Find paths using your algorithm
	paths := internal.FindPaths(data.Ants, data.Nodes, data.Links)

	if len(paths) == 0 {
		fmt.Println("No valid paths found")
		return
	}
	internal.Simulate(data.Ants, paths, "start", "0")
}