package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/framgr"
)

// fraSimulation is shared with the combinational FRA manager.
type fraSimulation = framgr.Simulation

func newFraSimulation(graph *aig.Graph, prefix, frames, wordsPerFrame int) *fraSimulation {
	return framgr.NewSimulation(graph, prefix, frames, wordsPerFrame)
}
