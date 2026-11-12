package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_DriverCountRefs counts edge references into each graph node.
func Llb_DriverCountRefs(g *Aig_Man_t) *Vec_Int_t {
	if g == nil {
		return nil
	}
	counts := vec.Vec_IntStart(int(aig.Aig_ManObjNumMax(g)))
	for _, node := range g.Nodes() {
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer && node.Kind != aig.Output {
			continue
		}
		faninCount := 1
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			faninCount = 2
		}
		for i := 0; i < faninCount; i++ {
			id := node.Fanins[i].NodeID()
			vec.Vec_IntAddToEntry(counts, id, 1)
		}
	}
	return counts
}

// Llb_DriverCollectNs returns IDs of latch-input drivers.
func Llb_DriverCollectNs(g *Aig_Man_t, refs *Vec_Int_t) *Vec_Int_t {
	if g == nil {
		return nil
	}
	registers := int(g.RegisterCount())
	outputs := g.Outputs()
	result := vec.Vec_IntAlloc(registers)
	for i := len(outputs) - registers; i < len(outputs); i++ {
		if i >= 0 {
			vec.Vec_IntPush(result, int32(outputs[i].NodeID()))
		}
	}
	return result
}

// Llb_DriverCollectCs returns IDs of latch outputs.
func Llb_DriverCollectCs(g *Aig_Man_t) *Vec_Int_t {
	if g == nil {
		return nil
	}
	registers := int(g.RegisterCount())
	inputs := g.Inputs()
	result := vec.Vec_IntAlloc(registers)
	for i := len(inputs) - registers; i < len(inputs); i++ {
		if i >= 0 {
			vec.Vec_IntPush(result, int32(inputs[i].NodeID()))
		}
	}
	return result
}

// Llb_DriverPhaseCube captures complemented latch-driver phases. The variable
// order follows the full node-ID variable map used by the image manager.
func Llb_DriverPhaseCube(g *Aig_Man_t, refs *Vec_Int_t, dd *DdManager) *DdNode {
	if g == nil || dd == nil {
		return nil
	}
	cube := core.Cudd_ReadOne(dd)
	outputs := g.Outputs()
	registers := int(g.RegisterCount())
	for i := len(outputs) - registers; i < len(outputs); i++ {
		if i < 0 {
			continue
		}
		edge := outputs[i]
		if !edge.Complemented() {
			continue
		}
		index := edge.NodeID()
		if index < len(dd.Vars) {
			cube = core.Cudd_bddAnd(dd, cube, core.Cudd_Not(core.Cudd_bddIthVar(dd, index)))
		}
		if cube == nil {
			return nil
		}
	}
	return cube
}

// Llb_DriverLastPartition creates a manager with one variable per CI.
func Llb_DriverLastPartition(g *Aig_Man_t, nextState *Vec_Int_t, timeout abctime) *DdManager {
	if g == nil {
		return nil
	}
	return core.NewBDDManager(len(g.Inputs()))
}
