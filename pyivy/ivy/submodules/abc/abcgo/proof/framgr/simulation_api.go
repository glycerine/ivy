package framgr

import "github.com/glycerine/abcgo/aig/aig"

// Simulation is the shared bit-parallel FRA simulation state.
type Simulation = fraSimulation

// NewSimulation constructs the shared FRA simulator.
func NewSimulation(graph *aig.Graph, prefix, frames, wordsPerFrame int) *Simulation {
	return newFraSimulation(graph, prefix, frames, wordsPerFrame)
}
