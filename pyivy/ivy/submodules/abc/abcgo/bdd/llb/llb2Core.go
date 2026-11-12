package llb

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_CoreComputeCube constructs a minterm over the selected BDD variables.
// A missing assignment value denotes false, matching zero-initialized C
// counterexample buffers.
func Llb_CoreComputeCube(dd *DdManager, variables *Vec_Int_t, useIndices int32, values string) *DdNode {
	if dd == nil || variables == nil {
		return nil
	}
	cube := core.Cudd_ReadOne(dd)
	for i, entry := range vec.Vec_IntArray(variables) {
		index := i
		if useIndices != 0 {
			index = int(entry)
		}
		if index < 0 || index >= len(dd.Vars) {
			return nil
		}
		variable := core.Cudd_bddIthVar(dd, index)
		if values == "" || i >= len(values) || values[i] != 1 {
			variable = core.Cudd_Not(variable)
		}
		cube = core.Cudd_bddAnd(dd, cube, variable)
		if cube == nil {
			return nil
		}
	}
	return cube
}

// Llb_CoreExperiment uses the exact native symbolic engine. The legacy flow
// partitioner remains available to the caller for partition discovery, while
// image computation shares one canonical BDD manager.
func Llb_CoreExperiment(initial, working *Aig_Man_t, params *Gia_ParLlb_t, result *Vec_Ptr_t, timeout abctime) int32 {
	if params == nil {
		return -1
	}
	useFlow := params.fUseFlow
	params.fUseFlow = 0
	defer func() { params.fUseFlow = useFlow }()
	return Llb_ManModelCheckAig(initial, params, nil, nil)
}
