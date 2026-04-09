# lem-in

## Description

lem-in is a digital ant farm simulation written in Go.
The program reads a colony map from a file; a graph of rooms connected by tunnels
and determines the most efficient way to move a given number of ants from the
designated start room to the end room in as few turns as possible.

The goal is to solve a flow and pathfinding problem; find the optimal set of
paths through the colony and coordinate ant movement so that no two ants occupy
the same room at the same time, and the total number of turns is minimized.

---

## Project Structure

```text
lem-in/
├── Makefile                    # Build targets for lem-in and visualiser
├── cmd/
│   └── main.go                 # CLI entry point: reads a map file and prints the simulation
├── internal/
│   ├── parser.go               # Parses and validates the input file
│   ├── parser_test.go          # Unit tests for the parser
│   ├── pathfinder.go           # Edmonds-Karp pathfinding with node-disjoint splitting
│   ├── pathfinder_test.go      # Unit tests for the pathfinder
│   ├── distributor.go          # Balances ant load across paths
│   ├── distributor_test.go     # Unit tests for the distributor
│   ├── simulation.go           # Simulation engine: produces structured move steps
│   └── list/
│       └── list.go             # Generic linked-list used by BFS
├── front-end/
│   ├── package.json            # npm project (esbuild, TypeScript, force-graph)
│   ├── tsconfig.json           # TypeScript configuration
│   ├── src/
│   │   ├── server.go           # Go HTTP server that serves the visualiser
│   │   └── index.ts            # TypeScript front-end: graph rendering & ant animation
│   └── public/
│       ├── index.html          # Visualiser HTML shell
│       └── index.css           # Dashboard & HUD styles
└── tests/
    ├── e2e_test.go             # End-to-end tests
    ├── samples/                # Valid example map files
    ├── bad_samples/            # Invalid map files for error-path testing
    └── json/                   # JSON representations of sample maps
```

---

## Features & Algorithms

### Edmonds-Karp

The program uses a variation of the **Edmonds-Karp** algorithm to traverse the network
of rooms and tunnels. It systematically explores the graph using a breadth-first-search
to identify the most optimal set of disjoint paths from `##start` to `##end` to prevent
network bottlenecks.

### Ant Distribution & Load Balancing

To determine which path an ant should take, the program calculates the relative "cost"
(arrival delay) of each path as `Length of Path + Ants Assigned To Path So Far`.
It directs ants to the path that yields the lowest cost, gracefully spilling overflow into
alternative, slightly longer paths if it calculates that doing so reduces the total turns.

### Strict Collision Engine

The simulation logic adheres to the "one ant per room" rule.
The turn timeline iterates while actively preventing room collisions by keeping track of the
maximum capacity of a room.

## Building

### Prerequisites

- **Go** 1.22+
- **Node.js** and **npm** (required for the visualiser front-end)

### Makefile Targets

```bash
make            # Build both lem-in and visualiser
make lem-in     # Build the CLI binary only
make visualiser # Install npm deps, bundle the front-end, then build the server binary
make clean      # Remove binaries, front-end/dist, and front-end/node_modules
```

## Usage

### Lem-in simulation

```bash
make lem-in
./lem-in <map_file>
```

`<map_file>` is a path to the file containing map data. The file must be in the correct format.

The program prints the input map followed by the turn-by-turn ant movements to stdout.

### Visualiser

The visualiser is a 2D interactive dashboard powered by [force-graph](https://github.com/vasturiano/force-graph),
served by a lightweight Go HTTP server.
The server reads the map and steps data from a file passed in via the command line or if no
file is passed in it reads the data from the Standard input.

```bash
make visualiser
./visualiser [-port N] [map_file]
```

| Argument | Description |
| :---: | :--- |
| `-port N` | specifies the port number `N` that the server will be using. Defaults to **3000**. |
| `map_file` | specifies a text file with the farm and ant steps data. If not given the server will read from standard input. |

If successful the server runs at **[http://localhost:3000](http://localhost:3000)** by default.

You can therefore use the simulator and visualiser together:

```bash
make
./lem-in ./tests/samples/example05.txt | ./visualiser
```

### Dashboard Controls

The Head-Up Display (HUD) tracks your population, arrivals, and progress dynamically.
Use the Command Center at the bottom to control the playback:

| Button | Name | Functionality |
| :----: | :--- | :------------ |
| ▶️ | **Play / Pause** | Automatically animates the turn-by-turn progression of the ants. Toggle to freeze. |
| ⏭️ | **Next Turn** | Manually increments the simulation forward by exactly one turn for close inspection. |
| 🔄 | **Reset** | Aborts the current run and snaps the simulation back to Turn 0. |
