package bmc

import "github.com/glycerine/abcgo/aig/aig"

// UnrollAIGFrames creates frame POs in time order and then appends the last
// frame's register-input outputs. Initial registers are fixed to zero.
func UnrollAIGFrames(graph *aig.Graph, frames int32) *aig.Graph {
	if graph == nil || frames <= 0 {
		return nil
	}
	return aig.Aig_ManFrames(graph, frames, 1, 1, 1, 0, nil)
}
