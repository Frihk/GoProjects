# lem-in

## Description

lem-in is a digital ant farm simulation written in Go. The program reads a colony map from a file: A graph of rooms connected by tunnels and determines the most efficient way to move a given number of ants from the designated start room to the end room in as few turns as possible.

The goal is to solve a flow and pathfinding problem: find the optimal set of paths through the colony and coordinate ant movement so that no two ants occupy the same room at the same time, and the total number of turns is minimized.

---

## Project Structure

```text
lem-in/
├── cmd/
│   └── main.go              # Entry point: reads args and orchestrates the program
├── internal/
│   ├── parser.go            # Parses and validates the input file
│   ├── pathfinder.go        # Pathfinding logic: finds and selects optimal disjoint paths
│   ├── distributor.go       # Distribution logic: balances load of ants across paths
│   └── simulation.go        # Core engine: strictly enforces collision rules and movement timing
├── tests/
│   ├── parser_test.go       # Unit tests for the parser
│   └── distributor_test.go  # Unit tests for the distributor
└── visualiser/
    ├── server.go            # Lightweight Go web server to host the front-end
    └── front-end/           # Pure JS, HTML, and CSS (Zero Node/NPM dependencies)
```

---

## Features & Algorithms

### Edmonds-Karp & BFS Pathfinding
The program uses a variation of the **Edmonds-Karp** algorithm combined with Breadth-First Search (BFS) to traverse the network of rooms and tunnels. It systematically explores the graph to identify the most optimal set of disjoint paths from `##start` to `##end` to prevent network bottlenecks.

### Ant Distribution & Load Balancing
To determine which path an ant should take, the program calculates the relative "cost" (arrival delay) of each path as `Length of Path + Ants Assigned To Path So Far`. It directs ants to the path that yields the lowest cost, gracefully spilling overflow into alternative, slightly longer paths if it calculates that doing so reduces the total turns.

### Strict Collision Engine
The simulation logic adheres perfectly to the "one ant per room" rule. Our turn timeline iterates using forward lookahead synchronization, actively preventing room collisions while still allowing ants to simultaneously enter and leave the same room in a single turn.

---

## Graphical Visualiser

We have a lightweight, standalone 2D graphical dashboard built entirely in **Go** and standard **JavaScript**. Anyone cloning this repository can run it out of the box without installing any external dependencies or running build tools.

### How to Run

1. Ensure you have **Go** installed on your machine.
2. Feed any valid ant farm map file into the main program, and pipe the output into the visualiser web server:

```bash
go run ./cmd <path_to_map_file> | go run visualiser/server.go
```

3. The Go server will bind to port 3000 by default. Open your web browser of choice and navigate to: **[http://localhost:3000](http://localhost:3000)**

### Dashboard Controls

The Head-Up Display (HUD) tracks your population, arrivals, and progress dynamically. Use the Command Center at the bottom to control the playback:

| Button | Name | Functionality |
| :---: | :--- | :--- |
| `▶️` | **Play / Pause** | Automatically animates the turn-by-turn progression of the ants. Toggle to freeze. |
| `⏭️` | **Next Turn (Step)** | Manually increments the simulation forward by **exactly one turn**, letting you closely inspect collision-free room transfers. |
| `🔄` | **Reset** | Aborts the current run and instantly snaps the simulation back to Turn 0. |
