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

_To be added._