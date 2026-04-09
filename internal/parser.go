package internal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Room represents a node in the graph.
type Room struct {
	ID    string `json:"id"`
	Group string `json:"group"`
	FX    int    `json:"fx"`
	FY    int    `json:"fy"`
}

// Tunnel represents an edge between two rooms.
type Tunnel struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// GraphData is the full payload returned to the frontend.
type GraphData struct {
	Ants  int      `json:"ants"`
	Start string   `json:"start,omitempty"`
	End   string   `json:"end,omitempty"`
	Nodes []Room   `json:"nodes"`
	Links []Tunnel `json:"links"`
	Steps []Step   `json:"steps,omitempty"`
}

// ReadAllLines reads all lines from r into a slice of strings.
func ReadAllLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// ParseInput parses the lem-in output lines into GraphData.
func ParseInput(lines []string) (GraphData, error) {
	var data GraphData
	rooms := make(map[string]Room)
	seenTunnels := make(map[string]struct{})
	var pendingGroup string
	var haveStart, haveEnd bool
	haveAntCount := false

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			switch line {
			case "##start":
				if pendingGroup != "" {
					return GraphData{}, invalidDataError("multiple start/end markers without a room")
				}

				pendingGroup = "start"
			case "##end":
				if pendingGroup != "" {
					return GraphData{}, invalidDataError("multiple start/end markers without a room")
				}

				pendingGroup = "end"
			}

			continue
		}

		// Movement lines appear after the map; ignore the rest.
		if strings.HasPrefix(line, "L") {
			break
		}

		// Ant count must be the first non-comment, non-empty data line.
		if !haveAntCount {
			if !isAntCountLine(line) {
				return GraphData{}, invalidDataError("invalid number of ants")
			}
			ants, err := strconv.Atoi(line)
			if err != nil || ants <= 0 {
				return GraphData{}, invalidDataError("invalid number of ants")
			}
			data.Ants = ants
			haveAntCount = true
			continue
		}

		if isRoomLine(line) {
			fields := strings.Fields(line)
			name := fields[0]
			x, err := strconv.Atoi(fields[1])
			if err != nil {
				return GraphData{}, invalidDataError("invalid room coordinates")
			}
			y, err := strconv.Atoi(fields[2])
			if err != nil {
				return GraphData{}, invalidDataError("invalid room coordinates")
			}

			if _, exists := rooms[name]; exists {
				return GraphData{}, invalidDataError(fmt.Sprintf("duplicate room %q", name))
			}

			room := Room{ID: name, Group: "room", FX: x, FY: y}

			if pendingGroup != "" {
				switch pendingGroup {
				case "start":
					if haveStart {
						return GraphData{}, invalidDataError("duplicate start room")
					}

					haveStart = true
					data.Start = name
				case "end":
					if haveEnd {
						return GraphData{}, invalidDataError("duplicate end room")
					}

					haveEnd = true
					data.End = name
				}

				room.Group = pendingGroup
				pendingGroup = ""
			}

			data.Nodes = append(data.Nodes, room)
			rooms[name] = room
			continue
		}

		if pendingGroup != "" {
			return GraphData{}, invalidDataError("start/end marker must be followed by a room")
		}

		if isTunnelLine(line) {
			parts := strings.Split(line, "-")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return GraphData{}, invalidDataError("invalid tunnel format")
			}
			src, dst := parts[0], parts[1]
			if _, ok := rooms[src]; !ok {
				return GraphData{}, invalidDataError(fmt.Sprintf("link to unknown room %q", src))
			}
			if _, ok := rooms[dst]; !ok {
				return GraphData{}, invalidDataError(fmt.Sprintf("link to unknown room %q", dst))
			}
			if src == dst {
				return GraphData{}, invalidDataError(fmt.Sprintf("room %q links to itself", src))
			}
			tunnelKey := normalizeTunnel(src, dst)
			if _, exists := seenTunnels[tunnelKey]; exists {
				return GraphData{}, invalidDataError(fmt.Sprintf("duplicate tunnel between %q and %q", src, dst))
			}
			seenTunnels[tunnelKey] = struct{}{}
			data.Links = append(data.Links, Tunnel{Source: src, Target: dst})
			continue
		}

		return GraphData{}, invalidDataError("invalid format")
	}

	if pendingGroup != "" {
		return GraphData{}, invalidDataError("start/end marker without a room")
	}
	if !haveAntCount {
		return GraphData{}, invalidDataError("invalid number of ants")
	}
	if !haveStart {
		return GraphData{}, invalidDataError("no start room found")
	}
	if !haveEnd {
		return GraphData{}, invalidDataError("no end room found")
	}

	return data, nil
}

func invalidDataError(reason string) error {
	return fmt.Errorf("ERROR: invalid data format, %s", reason)
}

func normalizeTunnel(src, dst string) string {
	if src > dst {
		src, dst = dst, src
	}
	return src + "-" + dst
}

func isRoomLine(line string) bool {
	fields := strings.Fields(line)
	return len(fields) == 3
}

func isTunnelLine(line string) bool {
	if strings.Contains(line, " ") {
		return false
	}
	return strings.Count(line, "-") == 1
}

func isAntCountLine(line string) bool {
	if strings.Contains(line, " ") {
		return false
	}
	_, err := strconv.Atoi(line)
	return err == nil
}
