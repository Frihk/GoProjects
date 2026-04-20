// txtojson converts lem-in text input files into the JSON format used by tests.
//
// Usage:
//
//	go run ./cmd/txtojson -o <output-dir> <input.txt> [<input2.txt> ...]
//
// If -o is omitted the JSON files are written next to each input file.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lem-in/internal"
)

func main() {
	outDir := flag.String("o", "", "directory to write JSON files into (defaults to input file's directory)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: txtojson [-o <output-dir>] <input.txt> [...]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, inputPath := range flag.Args() {
		if err := convertFile(inputPath, *outDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", inputPath, err)
			os.Exit(1)
		}
	}
}

func convertFile(inputPath, outDir string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	lines, err := internal.ReadAllLines(f)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	graphData, err := internal.ParseInput(lines)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	out, err := json.MarshalIndent(graphData, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Determine output path
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	dir := outDir
	if dir == "" {
		dir = filepath.Dir(inputPath)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	outPath := filepath.Join(dir, base+".json")

	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	fmt.Printf("%-40s -> %s\n", inputPath, outPath)
	return nil
}
