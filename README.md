# lem-in

## Description

lem-in is a digital ant farm simulation written in Go. The program reads a colony map from a file: A graph of rooms connected by tunnels and determines the most efficient way to move a given number of ants from the designated start room to the end room in as few turns as possible.

The goal is to solve a flow and pathfinding problem: find the optimal set of paths through the colony and coordinate ant movement so that no two ants occupy the same room at the same time, and the total number of turns is minimized.

---

## Project Structure

```
lem-in/
├── cmd/
│   └── main.go              # Entry point: reads args and orchestrates the program
├── internal/
│   ├── parser.go            # Parses and validates the input file
│   ├── pathfinder.go        # Pathfinding logic:  finds and selects optimal paths
│   └── distributor.go       # Distribution logic: balances load of ants across paths
├── tests/
│   ├── parser_test.go       # Unit tests for the parser
│   └── distributor_test.go  # Unit tests for the distributor
└── visualizer/
    └── frontend/            # Visual representation of the ant farm simulation
```

---

## Usage

_To be added._

---

## Features & Algorithms

### Breadth-First Search (BFS) Pathfinding
The program uses a **Breadth-First Search (BFS)** algorithm to traverse the network of rooms and tunnels. It systematically explores the graph level by level to identify the shortest and most optimal paths from the `##start` room to the `##end` room, ensuring that disjoint paths are selected to prevent traffic jams.

### Ant Distribution & Load Balancing
To determine which path an ant should take, the program calculates the relative "cost" (arrival delay) of each path as `Length of Path + Ants Assigned To Path So Far`. For every ant, it evaluates all available disjoint paths and directs the ant down the one that yields the lowest cost. If a shorter path becomes overcrowded with queued ants, the algorithm naturally starts spilling ants over into longer paths, ensuring optimal efficiency overall.