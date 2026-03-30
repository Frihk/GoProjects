package internal

type roomType int

type empty struct{}

const (
	normal roomType = iota
	start
	end
)

type link struct {
	target *node
	isOpen bool
}

type node struct {
	name  string
	typ   roomType
	links map[link]empty
}

func initGraph(rooms []Room, tunnels []Tunnel) (alpha node, omega node) {panic("not implemented")}

// FindPaths returns a slice of routes, which is a slice of rooms from start to end.
func FindPaths(ants int, rooms []Room, tunnels []Tunnel) [][]string {panic("not implemented")}
