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
	bin := mustBuild(t)

	cmd := exec.Command(bin, filename)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
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

func validateOutputFormat(t *testing.T, inputFile, stdout string) [][]string {
	t.Helper()

	// 1. Verify echoed input.
	inputBytes, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("read input file: %v", err)
	}
	inputContent := strings.TrimRight(string(inputBytes), "\n")

	idx := strings.Index(stdout, "\n\n")
	if idx < 0 {
		t.Fatal("output missing blank-line separator between echoed input and simulation")
	}
	echoed := strings.TrimRight(stdout[:idx], "\n")
	if echoed != inputContent {
		t.Errorf("echoed input does not match the original file")
	}

	// 2. Parse turn lines.
	_, turns, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parseTurns: %v", err)
	}

	startRoom, endRoom, _ := parseStartEnd(inputFile)
	endMarkers := map[string]bool{
		endRoom: true,
		"0":     true,
	}

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

		roomOccupants := make(map[string][]int)
		for id, rm := range antRoom {
			roomOccupants[rm] = append(roomOccupants[rm], id)
		}
		for rm, ants := range roomOccupants {
			if len(ants) > 1 && !endMarkers[rm] && rm != startRoom {
				t.Errorf("turn %d: room %q occupied by multiple ants: %v",
					turnIdx+1, rm, ants)
			}
		}
	}

	return turns
}

// ---------------------------------------------------------------------------
// 1. Output Formatting Rules
// ---------------------------------------------------------------------------

func TestOutputFormat(t *testing.T) {
	t.Parallel()

	files := []string{
		"example00.txt",
		"example01.txt",
		"example03.txt",
		"example04.txt",
	}
	for _, f := range files {
		f := f
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
		tc := tc
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

func TestPerformance_Example06(t *testing.T) {
	stdout, _, _ := runLemInWithTimeout(t, samplePath("example06.txt"), 90*time.Second)

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if numTurns == 0 {
		t.Error("expected simulation output, got zero turns")
	}
	t.Logf("example06 (100 ants) completed with %d turns", numTurns)
}

func TestPerformance_Example07(t *testing.T) {
	stdout, _, _ := runLemInWithTimeout(t, samplePath("example07.txt"), 150*time.Second)

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if numTurns == 0 {
		t.Error("expected simulation output, got zero turns")
	}
	t.Logf("example07 (1000 ants) completed with %d turns", numTurns)
}

// ---------------------------------------------------------------------------
// 4. Failure Cases
// ---------------------------------------------------------------------------

func assertErrorOutput(t *testing.T, stdout, stderr string, exitCode int, label string) {
	t.Helper()
	combined := stdout + stderr
	if !strings.Contains(combined, "ERROR: invalid data format") {
		t.Errorf("%s: expected 'ERROR: invalid data format' in output, got:\nstdout: %s\nstderr: %s",
			label, stdout, stderr)
	}
	if exitCode != 0 && exitCode != 1 {
		t.Errorf("%s: expected exit code 0 or 1, got %d", label, exitCode)
	}
}

func TestFailureCases_BadSamples(t *testing.T) {
	t.Parallel()

	tests := []string{
		"badexample00.txt",
		"badexample01.txt",
	}
	for _, f := range tests {
		f := f
		t.Run(f, func(t *testing.T) {
			t.Parallel()
			stdout, stderr, exitCode := runLemIn(t, badSamplePath(f))
			assertErrorOutput(t, stdout, stderr, exitCode, f)
		})
	}
}

func TestFailureCases_Dynamic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "missing_start_room",
			content: "5\n##end\nend 1 1\nroom1 2 2\nroom1-end\n",
		},
		{
			name:    "missing_end_room",
			content: "5\n##start\nstart 0 0\nroom1 2 2\nstart-room1\n",
		},
		{
			name:    "invalid_coordinates",
			content: "3\n##start\nstart abc def\n##end\nend 1 1\nstart-end\n",
		},
		{
			name:    "room_links_to_itself",
			content: "3\n##start\nstart 0 0\n##end\nend 5 5\nmid 2 2\nstart-mid\nmid-mid\nmid-end\n",
		},
		{
			name:    "zero_ants",
			content: "0\n##start\nstart 0 0\n##end\nend 1 1\nstart-end\n",
		},
		{
			name:    "negative_ants",
			content: "-5\n##start\nstart 0 0\n##end\nend 1 1\nstart-end\n",
		},
		{
			name:    "duplicate_room",
			content: "3\n##start\nstart 0 0\nstart 1 1\n##end\nend 5 5\nstart-end\n",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp("", "lemin-bad-*.txt")
			if err != nil {
				t.Fatalf("create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tc.content); err != nil {
				t.Fatalf("write temp file: %v", err)
			}
			tmpFile.Close()

			stdout, stderr, exitCode := runLemIn(t, tmpFile.Name())
			assertErrorOutput(t, stdout, stderr, exitCode, tc.name)
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

	const maxOptimalTurns = 9
	if numTurns > maxOptimalTurns {
		t.Errorf("greedy trap: expected at most %d turns (optimal two-path), got %d – "+
			"program may be taking the greedy shortcut", maxOptimalTurns, numTurns)
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

	const target = "bottleneck"
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
			if rm == target {
				count++
			}
		}
		if count > 1 {
			t.Errorf("turn %d: %d ants in room %q – vertex capacity violated",
				turnIdx+1, count, target)
		}
	}
}

// TestResidualEdgeReversal uses residual.txt where the initial BFS pushes
// flow through the "junction" room but the optimal 4-path solution requires
// reversing that flow. With 13 ants and 4 optimal paths the theoretical
// maximum is 6 turns (with a small margin of 7 allowed).
func TestResidualEdgeReversal(t *testing.T) {
	t.Parallel()

	stdout, _, _ := runLemIn(t, samplePath("residual.txt"))

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}

	const maxTurns = 7
	if numTurns > maxTurns {
		t.Errorf("residual reversal: expected at most %d turns, got %d – "+
			"residual edge handling may be incorrect", maxTurns, numTurns)
	}
}

// TestResidualJunctionDisjoint verifies that in residual.txt the "junction"
// room is never occupied by more than one ant at a time.
func TestResidualJunctionDisjoint(t *testing.T) {
	t.Parallel()

	stdout, _, _ := runLemIn(t, samplePath("residual.txt"))

	_, turns, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}

	const target = "junction"
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
			if rm == target {
				count++
			}
		}
		if count > 1 {
			t.Errorf("turn %d: %d ants in room %q – vertex capacity violated",
				turnIdx+1, count, target)
		}
	}
}

// TestDuplicateLink verifies that a file with duplicate tunnels is handled.
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
// Programmatic edge-case graphs
// ---------------------------------------------------------------------------

// TestProgrammatic_DiamondGraph verifies two node-disjoint paths are found:
//
//	Start -> A -> End
//	Start -> B -> End
//
// With 10 ants and 2 paths of 2 hops each, optimal = 6 turns.
func TestProgrammatic_DiamondGraph(t *testing.T) {
	t.Parallel()

	content := "10\n##start\nstart 0 1\nA 1 2\nB 1 0\n##end\nend 2 1\nstart-A\nstart-B\nA-end\nB-end\n"

	tmpFile, err := os.CreateTemp("", "lemin-diamond-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	stdout, _, _ := runLemIn(t, tmpFile.Name())

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}

	const maxTurns = 6
	if numTurns > maxTurns {
		t.Errorf("diamond graph: expected at most %d turns, got %d", maxTurns, numTurns)
	}
}

// TestProgrammatic_SinglePath verifies a single linear path:
//
//	Start -> A -> B -> End   (3 hops)
//	With 3 ants: 3 + 3 - 1 = 5 turns.
func TestProgrammatic_SinglePath(t *testing.T) {
	t.Parallel()

	content := "3\n##start\nstart 0 0\nA 1 0\nB 2 0\n##end\nend 3 0\nstart-A\nA-B\nB-end\n"

	tmpFile, err := os.CreateTemp("", "lemin-linear-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	stdout, _, _ := runLemIn(t, tmpFile.Name())

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}

	const maxTurns = 5
	if numTurns > maxTurns {
		t.Errorf("single path: expected at most %d turns, got %d", maxTurns, numTurns)
	}
}

// TestProgrammatic_DirectConnection verifies the trivial Start -> End case.
//
//	With 5 ants and 1 hop: 5 turns.
func TestProgrammatic_DirectConnection(t *testing.T) {
	t.Parallel()

	content := "5\n##start\nstart 0 0\n##end\nend 1 0\nstart-end\n"

	tmpFile, err := os.CreateTemp("", "lemin-direct-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	stdout, _, _ := runLemIn(t, tmpFile.Name())

	numTurns, _, err := parseTurns(stdout)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}

	const maxTurns = 5
	if numTurns > maxTurns {
		t.Errorf("direct connection: expected at most %d turns, got %d", maxTurns, numTurns)
	}
}
