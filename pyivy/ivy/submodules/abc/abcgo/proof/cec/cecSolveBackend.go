package cec

import "github.com/glycerine/abcgo/proof/cec/core"

// cecSolverPhases computes the all-zero input phase used by ABC's optional SAT polarity normalization.
func cecSolverPhases(graph *Gia_Man_t, enabled bool) []bool {
	phases := make([]bool, graph.ObjectCount())
	if !enabled {
		return phases
	}
	model, err := core.NewModel(graph)
	if err != nil {
		panic(err)
	}
	values, err := model.Values(make([]bool, len(graph.InputLiterals())))
	if err != nil {
		panic(err)
	}
	copy(phases, values)
	return phases
}
