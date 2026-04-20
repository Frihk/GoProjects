package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"lem-in/internal"
)

func readLines(scanner *bufio.Scanner) ([]string, error) {
	lines := []string{}

	fmt.Println("Enter text (Ctrl+D to finish):")
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err := scanner.Err()

	return lines, err
}

func main() {
	var port int

	flag.IntVar(&port, "port", 3000, "Port to serve on")
	flag.Parse()

	var lines []string
	var err error

	if len(flag.Args()) < 1 {
		lines, err = readLines(bufio.NewScanner(os.Stdin))
	} else {
		filename := flag.Arg(0)
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not open file %q: %e\n", filename, err)
			os.Exit(1)
		}
		defer file.Close()

		lines, err = readLines(bufio.NewScanner(file))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not read file %q: %e\n", filename, err)
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not read data: %e\n", err)
		os.Exit(1)
	}

	graphData, err := internal.ParseInput(lines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	paths := internal.FindPaths(graphData.Ants, graphData.Nodes, graphData.Links)
	if len(paths) > 0 {
		graphData.Steps = internal.SimulateSteps(graphData.Ants, paths, graphData.End)
	}

	graphDataAsJson, err := json.Marshal(graphData)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: could not convert to json:", err)
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(http.Dir("front-end/public")))
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("front-end/dist"))))
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(graphDataAsJson); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: error writing response")
		}
	})

	fmt.Printf("Visualizer ready at http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
	}
}
