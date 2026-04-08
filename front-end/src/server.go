package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"lem-in/internal"
)

func readStdin() ([]string, error) {
	lines := []string{}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter text (Ctrl+D to finish):")
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err := scanner.Err()

	return lines, err
}

func main() {
	lines, err := readStdin()
	if err != nil {
		fmt.Println("Error reading from stdin:", err)
		os.Exit(1)
	}

	graphData, err := internal.ParseInput(lines)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	graphDataAsJson, err := json.Marshal(graphData)
	if err != nil {
		fmt.Println("ERROR: could not convert to json:", err)
		os.Exit(1)
	}

	port := "3000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	http.Handle("/", http.FileServer(http.Dir("front-end/public")))
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("front-end/dist"))))
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(graphDataAsJson); err != nil {
			fmt.Println("ERROR: error writing response")
		}
	})

	fmt.Printf("Server started at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
