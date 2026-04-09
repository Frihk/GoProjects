package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Build helpers – compile the binary once for all tests
// ---------------------------------------------------------------------------

var (
	binary    string
	buildOnce sync.Once
	buildErr  error
	rootDir   string
)

const defaultRunTimeout = 10 * time.Second

func mustBuild(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			buildErr = fmt.Errorf("getwd: %w", err)
			return
		}
		rootDir = filepath.Dir(wd)

		tmp, err := os.CreateTemp("", "lem-in-e2e-*")
		if err != nil {
			buildErr = fmt.Errorf("tempfile: %w", err)
			return
		}
		tmp.Close()
		binary = tmp.Name()

		cmd := exec.Command("go", "build", "-o", binary, "./cmd")
		cmd.Dir = rootDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("go build failed: %v\n%s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatalf("build binary: %v", buildErr)
	}
	return binary
}

func TestMain(m *testing.M) {
	code := m.Run()
	if binary != "" {
		os.Remove(binary)
	}
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Execution helpers
// ---------------------------------------------------------------------------

func samplePath(name string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "samples", name)
}

func badSamplePath(name string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "bad_samples", name)
}

func runLemIn(t *testing.T, filename string) (string, string, int) {
	t.Helper()
	return runLemInWithTimeout(t, filename, defaultRunTimeout)
}

func runLemInWithTimeout(t *testing.T, filename string, timeout time.Duration) (string, string, int) {
	t.Helper()
	bin := mustBuild(t)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, filename)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("program exceeded timeout of %v for %s", timeout, filepath.Base(filename))
	}

	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("exec error for %s: %v", filepath.Base(filename), err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// writeTempInput writes content to a temp file and returns its path.
// The caller should defer os.Remove on the returned path.
func writeTempInput(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "lemin-test-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func parseTurns(output string) (int, [][]string, error) {
	idx := strings.Index(output, "\n\n")
	if idx < 0 {
		return 0, nil, fmt.Errorf("no blank-line separator found in output")
	}

	turnSection := strings.TrimRight(output[idx+2:], "\n \t")
	if turnSection == "" {
		return 0, nil, nil
	}

	var turns [][]string
	for _, line := range strings.Split(turnSection, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		turns = append(turns, strings.Fields(line))
	}
	return len(turns), turns, nil
}

func parseStartEnd(filename string) (string, string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var pending, start, end string
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "##start" {
			pending = "start"
			continue
		}
		if line == "##end" {
			pending = "end"
			continue
		}
		if pending != "" && !strings.HasPrefix(line, "#") {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				if pending == "start" {
					start = fields[0]
				} else {
					end = fields[0]
				}
			}
			pending = ""
		}
	}
	if start == "" {
		return "", "", fmt.Errorf("no ##start room found")
	}
	if end == "" {
		return "", "", fmt.Errorf("no ##end room found")
	}
	return start, end, nil
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

var moveRe = regexp.MustCompile(`^L(\d+)-(\S+)$`)

// validateOutputFormat checks the full output for spec compliance:
//   - echoed input matches the original file
//   - every move has the L<id>-<room> format
//   - no ant moves twice in a single turn
//   - ant IDs appear in ascending order per turn
//   - no intermediate room is occupied by more than one ant at a time
func validateOutputFormat(t *testing.T, inputFile, stdout string) [][]string {
	t.Helper()

	// 1. Verify echoed input.
	inputBytes, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("read input file: %v", err)
	}

	inputContent := strings.TrimRight(string(inputBytes), "\n")
	echoed, _, found := strings.Cut(stdout, "\n\n")
	if !found {
		t.Fatal("output missing blank-line separator between echoed input and simulation")
	}
	if echoed != inputContent {
		t.Errorf("echoed input does not match the original file")
	}

	// 2. Parse turn lines.
	_, turns, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parseTurns: %v", err)
	}

	startRoom, endRoom, _ := parseStartEnd(inputFile)

	antRoom := make(map[int]string)

	for turnIdx, moves := range turns {
		var antIDs []int
		seen := make(map[int]bool)

		for _, mv := range moves {
			m := moveRe.FindStringSubmatch(mv)
			if m == nil {
				t.Errorf("turn %d: invalid move format %q", turnIdx+1, mv)
				continue
			}

			antID, _ := strconv.Atoi(m[1])
			room := m[2]

			if seen[antID] {
				t.Errorf("turn %d: ant %d moved more than once", turnIdx+1, antID)
			}
			seen[antID] = true
			antIDs = append(antIDs, antID)
			antRoom[antID] = room
		}

		if !sort.IntsAreSorted(antIDs) {
			t.Errorf("turn %d: ant IDs not in ascending order: %v", turnIdx+1, antIDs)
		}

		// Check room occupancy — start and end rooms have unlimited capacity.
		roomOccupants := make(map[string][]int)
		for id, rm := range antRoom {
			roomOccupants[rm] = append(roomOccupants[rm], id)
		}
		for rm, ants := range roomOccupants {
			if len(ants) > 1 && rm != endRoom && rm != startRoom {
				t.Errorf("turn %d: room %q occupied by multiple ants: %v",
					turnIdx+1, rm, ants)
			}
		}
	}

	return turns
}

// assertVertexCapacity verifies that a specific room is never occupied by
// more than one ant at any point during the simulation.
func assertVertexCapacity(t *testing.T, turns [][]string, room string) {
	t.Helper()

	antPos := make(map[int]string)
	for turnIdx, moves := range turns {
		for _, mv := range moves {
			m := moveRe.FindStringSubmatch(mv)
			if m == nil {
				continue
			}
			antID, _ := strconv.Atoi(m[1])
			antPos[antID] = m[2]
		}

		var count int
		for _, rm := range antPos {
			if rm == room {
				count++
			}
		}
		if count > 1 {
			t.Errorf("turn %d: %d ants in room %q – vertex capacity violated",
				turnIdx+1, count, room)
		}
	}
}

// assertErrorOutput checks that the program reported an error.
func assertErrorOutput(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	combined := stdout + stderr
	if !strings.Contains(combined, "ERROR: invalid data format") {
		t.Errorf("expected 'ERROR: invalid data format' in output, got:\nstdout: %s\nstderr: %s",
			stdout, stderr)
	}
	if exitCode != 0 && exitCode != 1 {
		t.Errorf("expected exit code 0 or 1, got %d", exitCode)
	}
}

// ---------------------------------------------------------------------------
// 1. Output Formatting Rules
// ---------------------------------------------------------------------------

func TestOutputFormat(t *testing.T) {
	t.Parallel()

	for _, f := range []string{"example00.txt", "example01.txt", "example03.txt", "example04.txt"} {
		t.Run(f, func(t *testing.T) {
			t.Parallel()
			stdout, _, _ := runLemIn(t, samplePath(f))
			validateOutputFormat(t, samplePath(f), stdout)
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Audit Standard Examples – Turn Limits
// ---------------------------------------------------------------------------

func TestAuditTurnLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		file     string
		maxTurns int
	}{
		{"example00.txt", 6},
		{"example01.txt", 8},
		{"example02.txt", 11},
		{"example03.txt", 6},
		{"example04.txt", 6},
		{"example05.txt", 8},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			t.Parallel()
			stdout, _, _ := runLemIn(t, samplePath(tc.file))

			numTurns, _, err := parseTurns(stdout)
			if err != nil {
				t.Fatalf("parse output: %v", err)
			}
			if numTurns > tc.maxTurns {
				t.Errorf("expected at most %d turns, got %d", tc.maxTurns, numTurns)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Performance / Timeout Requirements
// ---------------------------------------------------------------------------

func TestPerformance(t *testing.T) {
	tests := []struct {
		file    string
		timeout time.Duration
		ants    int
	}{
		{"example06.txt", 90 * time.Second, 100},
		{"example07.txt", 150 * time.Second, 1000},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			stdout, _, _ := runLemInWithTimeout(t, samplePath(tc.file), tc.timeout)

			numTurns, _, err := parseTurns(stdout)
			if err != nil {
				t.Fatalf("parse output: %v", err)
			}
			if numTurns == 0 {
				t.Error("expected simulation output, got zero turns")
			}
			t.Logf("%s (%d ants) completed with %d turns", tc.file, tc.ants, numTurns)
		})
	}
}

// ---------------------------------------------------------------------------
// 4. Failure Cases
// ---------------------------------------------------------------------------

func TestFailureCases_BadSamples(t *testing.T) {
	t.Parallel()

	for _, f := range []string{"badexample00.txt", "badexample01.txt"} {
		t.Run(f, func(t *testing.T) {
			t.Parallel()
			stdout, stderr, exitCode := runLemIn(t, badSamplePath(f))
			assertErrorOutput(t, stdout, stderr, exitCode)
		})
	}
}

func TestFailureCases_Dynamic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
	}{
		{"missing_start_room", "5\n##end\nend 1 1\nroom1 2 2\nroom1-end\n"},
		{"missing_end_room", "5\n##start\nstart 0 0\nroom1 2 2\nstart-room1\n"},
		{"invalid_coordinates", "3\n##start\nstart abc def\n##end\nend 1 1\nstart-end\n"},
		{"room_links_to_itself", "3\n##start\nstart 0 0\n##end\nend 5 5\nmid 2 2\nstart-mid\nmid-mid\nmid-end\n"},
		{"zero_ants", "0\n##start\nstart 0 0\n##end\nend 1 1\nstart-end\n"},
		{"negative_ants", "-5\n##start\nstart 0 0\n##end\nend 1 1\nstart-end\n"},
		{"duplicate_room", "3\n##start\nstart 0 0\nstart 1 1\n##end\nend 5 5\nstart-end\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeTempInput(t, tc.content)
			defer os.Remove(path)

			stdout, stderr, exitCode := runLemIn(t, path)
			assertErrorOutput(t, stdout, stderr, exitCode)
		})
	}
}

// ---------------------------------------------------------------------------
// 5. Algorithmic Edge Cases – Edmonds-Karp Verification
// ---------------------------------------------------------------------------

// TestGreedyTrap uses shortcut.txt where the greedy shortest path (length 4)
// blocks two longer disjoint paths (length 5 each). With 10 ants the two
// longer paths yield at most 9 turns vs 12+ for the greedy shortcut.
func TestGreedyTrap(t *testing.T) {
	t.Parallel()

	stdout, _, _ := runLemIn(t, samplePath("shortcut.txt"))

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if numTurns > 9 {
		t.Errorf("expected at most 9 turns (optimal two-path), got %d – "+
			"program may be taking the greedy shortcut", numTurns)
	}
}

// TestNodeDisjointEnforcement uses bottleneck.txt where multiple paths funnel
// through a single "bottleneck" room. At most one ant may occupy it per turn.
func TestNodeDisjointEnforcement(t *testing.T) {
	t.Parallel()

	stdout, _, _ := runLemIn(t, samplePath("bottleneck.txt"))

	_, turns, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	assertVertexCapacity(t, turns, "bottleneck")
}

// TestResidualEdgeReversal uses residual.txt where the initial BFS pushes
// flow through the "junction" room but the optimal 4-path solution requires
// reversing that flow. Checks both turn-count optimality (≤7) and that the
// "junction" room vertex capacity is never violated.
func TestResidualEdgeReversal(t *testing.T) {
	t.Parallel()

	stdout, _, _ := runLemIn(t, samplePath("residual.txt"))

	numTurns, turns, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if numTurns > 7 {
		t.Errorf("expected at most 7 turns, got %d – residual edge handling may be incorrect", numTurns)
	}
	assertVertexCapacity(t, turns, "junction")
}

// TestDuplicateLink verifies that a file with duplicate tunnels is handled
// either by reporting an error or by solving correctly.
func TestDuplicateLink(t *testing.T) {
	t.Parallel()

	file := badSamplePath("duplicate_link.txt")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Skip("duplicate_link.txt not present")
	}

	stdout, stderr, exitCode := runLemIn(t, file)

	combined := stdout + stderr
	if strings.Contains(combined, "ERROR: invalid data format") {
		if exitCode != 0 && exitCode != 1 {
			t.Errorf("expected exit code 0 or 1, got %d", exitCode)
		}
		return
	}

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if numTurns == 0 {
		t.Error("expected either an error message or simulation output")
	}
}

// ---------------------------------------------------------------------------
// 6. Programmatic Edge-Case Graphs
// ---------------------------------------------------------------------------

func TestProgrammaticGraphs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		maxTurns int
	}{
		{
			// Two node-disjoint paths: Start→A→End, Start→B→End.
			// 10 ants, 2 paths of 2 hops → optimal 6 turns.
			name:     "diamond",
			content:  "10\n##start\nstart 0 1\nA 1 2\nB 1 0\n##end\nend 2 1\nstart-A\nstart-B\nA-end\nB-end\n",
			maxTurns: 6,
		},
		{
			// Single linear path: Start→A→B→End (3 hops).
			// 3 ants → 3 + 3 - 1 = 5 turns.
			name:     "single_path",
			content:  "3\n##start\nstart 0 0\nA 1 0\nB 2 0\n##end\nend 3 0\nstart-A\nA-B\nB-end\n",
			maxTurns: 5,
		},
		{
			// Direct connection: Start→End (1 hop).
			// 5 ants → 5 turns.
			name:     "direct_connection",
			content:  "5\n##start\nstart 0 0\n##end\nend 1 0\nstart-end\n",
			maxTurns: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeTempInput(t, tc.content)
			defer os.Remove(path)

			stdout, _, _ := runLemIn(t, path)

			numTurns, _, err := parseTurns(stdout)
			if err != nil {
				t.Fatalf("parse output: %v", err)
			}
			if numTurns > tc.maxTurns {
				t.Errorf("expected at most %d turns, got %d", tc.maxTurns, numTurns)
			}
		})
	}
}
