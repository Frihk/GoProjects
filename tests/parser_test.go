package internal

import "testing"

func TestParseInput_Valid(t *testing.T) {
	lines := []string{
		"4",
		"##start",
		"A 0 0",
		"B 1 0",
		"##end",
		"C 2 2",
		"A-B",
		"B-C",
		"",
		"L1-B",
		"L2-C",
	}

	data, err := ParseInput(lines)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.Ants != 4 {
		t.Fatalf("expected ant count 4, got %d", data.Ants)
	}
	if len(data.Nodes) != 3 {
		t.Fatalf("expected 3 rooms, got %d", len(data.Nodes))
	}
	if len(data.Links) != 2 {
		t.Fatalf("expected 2 tunnels, got %d", len(data.Links))
	}

	rooms := map[string]Room{}
	for _, r := range data.Nodes {
		rooms[r.ID] = r
	}
	if rooms["A"].Group != "start" {
		t.Fatalf("expected A to be start, got %q", rooms["A"].Group)
	}
	if rooms["C"].Group != "end" {
		t.Fatalf("expected C to be end, got %q", rooms["C"].Group)
	}
	if rooms["B"].Group != "room" {
		t.Fatalf("expected B to be room, got %q", rooms["B"].Group)
	}
	if rooms["B"].FX != 1 || rooms["B"].FY != 0 {
		t.Fatalf("expected B coords (1,0), got (%d,%d)", rooms["B"].FX, rooms["B"].FY)
	}
}

func TestParseInput_Errors(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{
			name: "missing ant count",
			lines: []string{
				"##start",
				"A 0 0",
				"##end",
				"B 1 1",
				"A-B",
			},
		},
		{
			name: "invalid ant count",
			lines: []string{
				"0",
				"##start",
				"A 0 0",
				"##end",
				"B 1 1",
				"A-B",
			},
		},
		{
			name: "missing start",
			lines: []string{
				"3",
				"A 0 0",
				"##end",
				"B 1 1",
				"A-B",
			},
		},
		{
			name: "missing end",
			lines: []string{
				"3",
				"##start",
				"A 0 0",
				"B 1 1",
				"A-B",
			},
		},
		{
			name: "duplicate room",
			lines: []string{
				"##start",
				"A 0 0",
				"A 1 1",
				"##end",
				"B 2 2",
				"A-B",
			},
		},
		{
			name: "unknown link",
			lines: []string{
				"##start",
				"A 0 0",
				"##end",
				"C 2 2",
				"A-B",
			},
		},
		{
			name: "invalid coordinates",
			lines: []string{
				"##start",
				"A x 0",
				"##end",
				"B 1 1",
			},
		},
		{
			name: "invalid format",
			lines: []string{
				"##start",
				"A 0 0",
				"##end",
				"B 1 1",
				"A 1",
			},
		},
	}

	for _, tc := range tests {
		_, err := ParseInput(tc.lines)
		if err == nil {
			t.Fatalf("%s: expected error, got nil", tc.name)
		}
	}
}
