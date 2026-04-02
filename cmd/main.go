package main

import (
	"fmt"
	"os"

	"lem-in/internal"
)

func main() {
	f, err := os.Open("../tests/parser_example.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	lines, err := internal.ReadAllLines(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := internal.ParseInput(lines)
	if err != nil {
		fmt.Println(err)
		return
	}
	paths := internal.FindPaths(data.Ants, data.Nodes, data.Links)

	if len(paths) == 0 {
		fmt.Println("No valid paths found")
		return
	}
	internal.Simulate(data.Ants, paths, "start", "end")
}