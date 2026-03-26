package int

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
	Nodes []Room   `json:"nodes"`
	Links []Tunnel `json:"links"`
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
	var pendingGroup string
	var haveStart, haveEnd bool

	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			switch line {
			case "##start":
				if pendingGroup != "" {
					return GraphData{}, fmt.Errorf("line %d: multiple start/end markers without a room", i+1)
				}
				pendingGroup = "start"
			case "##end":
				if pendingGroup != "" {
					return GraphData{}, fmt.Errorf("line %d: multiple start/end markers without a room", i+1)
				}
				pendingGroup = "end"
			default:
				// comment line, ignore
			}
			continue
		}

		// Movement lines appear after the map; ignore the rest.
		if strings.HasPrefix(line, "L") {
			break
		}

		// Ant count line (first line) can be ignored.
		if isAntCountLine(line) && len(data.Nodes) == 0 && len(data.Links) == 0 {
			continue
		}

		if isRoomLine(line) {
			fields := strings.Fields(line)
			name := fields[0]
			x, err := strconv.Atoi(fields[1])
			if err != nil {
				return GraphData{}, fmt.Errorf("line %d: invalid x coordinate", i+1)
			}
			y, err := strconv.Atoi(fields[2])
			if err != nil {
				return GraphData{}, fmt.Errorf("line %d: invalid y coordinate", i+1)
			}

			if _, exists := rooms[name]; exists {
				return GraphData{}, fmt.Errorf("line %d: duplicate room %q", i+1, name)
			}

			group := "room"
			if pendingGroup != "" {
				group = pendingGroup
				pendingGroup = ""
				switch group {
				case "start":
					if haveStart {
						return GraphData{}, fmt.Errorf("line %d: duplicate start room", i+1)
					}
					haveStart = true
				case "end":
					if haveEnd {
						return GraphData{}, fmt.Errorf("line %d: duplicate end room", i+1)
					}
					haveEnd = true
				}
			}

			room := Room{ID: name, Group: group, FX: x, FY: y}
			data.Nodes = append(data.Nodes, room)
			rooms[name] = room
			continue
		}

		if pendingGroup != "" {
			return GraphData{}, fmt.Errorf("line %d: start/end marker must be followed by a room", i+1)
		}

		if isTunnelLine(line) {
			parts := strings.Split(line, "-")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return GraphData{}, fmt.Errorf("line %d: invalid tunnel format", i+1)
			}
			src, dst := parts[0], parts[1]
			if _, ok := rooms[src]; !ok {
				return GraphData{}, fmt.Errorf("line %d: link to unknown room %q", i+1, src)
			}
			if _, ok := rooms[dst]; !ok {
				return GraphData{}, fmt.Errorf("line %d: link to unknown room %q", i+1, dst)
			}
			data.Links = append(data.Links, Tunnel{Source: src, Target: dst})
			continue
		}

		return GraphData{}, fmt.Errorf("line %d: invalid format", i+1)
	}

	if pendingGroup != "" {
		return GraphData{}, fmt.Errorf("end of input: start/end marker without a room")
	}
	if !haveStart || !haveEnd {
		return GraphData{}, fmt.Errorf("missing ##start or ##end")
	}

	return data, nil
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
