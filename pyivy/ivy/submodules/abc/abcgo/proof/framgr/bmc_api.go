package framgr

import "github.com/glycerine/abcgo/aig/aig"

// BMC is the shared FRA bounded-model-checking state.
type BMC = Fra_Bmc_t_

// UnrollInitialized builds initialized BMC frames and returns the node mapping.
func UnrollInitialized(graph *aig.Graph, frames int, keepOutputs bool) (*aig.Graph, []aig.Edge) {
	return fraUnrollInitialized(graph, frames, keepOutputs)
}
