package main

import (
	"fmt"
	"os"

	"lem-in/internal"
)

func main() {
	f, err := os.Open("tests/parser_example.txt")
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

	fmt.Printf("Ants: %d\n", data.Ants)
	fmt.Printf("Rooms: %+v\n", data.Nodes)
	fmt.Printf("Links: %+v\n", data.Links)
}
