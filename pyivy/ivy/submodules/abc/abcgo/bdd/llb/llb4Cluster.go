package llb

import (
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func llbOrderedInputs(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t) []*core.DdNode {
	if dd == nil || g == nil {
		return nil
	}
	inputs := make([]*core.DdNode, len(g.Inputs()))
	for i, edge := range g.Inputs() {
		index := int32(edge.NodeID())
		if order != nil {
			id := int(edge.NodeID())
			if id >= int(vec.Vec_IntSize(order)) {
				return nil
			}
			index = vec.Vec_IntEntry(order, id)
		}
		if index < 0 || int(index) >= len(dd.Vars) {
			return nil
		}
		inputs[i] = core.Cudd_bddIthVar(dd, int(index))
	}
	return inputs
}

// Llb_Nonlin4FindOrder returns a stable topological node-ID order.
func Llb_Nonlin4FindOrder(g *Aig_Man_t, counter *int32) *Vec_Int_t {
	if g == nil {
		return vec.Vec_IntAlloc(0)
	}
	order := vec.Vec_IntAlloc(int(aig.Aig_ManObjNumMax(g)))
	for _, edge := range g.Inputs() {
		vec.Vec_IntPush(order, int32(edge.NodeID()))
	}
	for _, edge := range g.LiveEdges() {
		id := edge.NodeID()
		if id != 0 && !aig.Aig_ObjIsCi(g, edge) {
			vec.Vec_IntPush(order, int32(id))
		}
	}
	if counter != nil {
		*counter = int32(vec.Vec_IntSize(order))
	}
	return order
}
func Llb_Nonlin4FindOrder_rec(g *Aig_Man_t, edge *Aig_Obj_t, order *Vec_Int_t, counter *int32) {
	if g == nil || edge == nil {
		return
	}
	vec.Vec_IntPush(order, int32((*edge).NodeID()))
	if counter != nil {
		*counter++
	}
}

// Llb_Nonlin4FindPartitions compiles each output function to a canonical BDD.
func Llb_Nonlin4FindPartitions(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, outputs int32) *Vec_Ptr_t {
	parts := vec.Vec_PtrAlloc(8)
	if dd == nil || g == nil {
		return parts
	}
	inputs := llbOrderedInputs(dd, g, order)
	if inputs == nil {
		return parts
	}
	functions := llbBuildBDDTable(g, dd, inputs)
	if functions == nil {
		return parts
	}
	limit := len(g.Outputs())
	if outputs == 0 && int(g.RegisterCount()) < limit {
		limit -= int(g.RegisterCount())
	}
	for i := 0; i < limit; i++ {
		vec.Vec_PtrPush(parts, llbBDDForEdge(g, functions, g.Outputs()[i]))
	}
	return parts
}
func Llb_Nonlin4FindPartitions_rec(dd *DdManager, g *Aig_Man_t, edge *Aig_Obj_t, order *Vec_Int_t, roots *Vec_Ptr_t) *DdNode {
	if dd == nil || g == nil || edge == nil {
		return nil
	}
	vars := llbOrderedInputs(dd, g, order)
	if vars == nil {
		return nil
	}
	return llbBDDForEdge(g, llbBuildBDDTable(g, dd, vars), *edge)
}
func Llb_Nonlin4FindVars2Q(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(16)
	if g == nil {
		return result
	}
	input := map[int]bool{}
	for _, e := range g.Inputs() {
		input[e.NodeID()] = true
	}
	outputs := map[int]bool{}
	for _, e := range g.Outputs() {
		outputs[e.NodeID()] = true
	}
	for _, e := range g.LiveEdges() {
		id := e.NodeID()
		if !input[id] && !outputs[id] && id < int(aig.Aig_ManObjNumMax(g)) {
			vec.Vec_IntPush(result, int32(id))
		}
	}
	return result
}
func Llb_Nonlin4CountTerms(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, f *DdNode, co, flop int32) int32 {
	if dd == nil || f == nil {
		return 0
	}
	return int32(len(dd.Support(f)))
}
func Llb_Nonlin4PrintGroups(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, groups *Vec_Ptr_t) {
	if groups == nil {
		return
	}
	fmt.Printf("%d BDD groups\n", vec.Vec_PtrSize(groups))
}
func Llb_Nonlin4PrintSuppProfile(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, groups *Vec_Ptr_t) {
	if groups == nil {
		return
	}
	sizes := make([]int, 0, vec.Vec_PtrSize(groups))
	for _, x := range vec.Vec_PtrArray(groups) {
		if f, ok := x.(*core.DdNode); ok {
			sizes = append(sizes, core.Cudd_SupportSize(dd, f))
		}
	}
	sort.Ints(sizes)
	fmt.Printf("BDD support sizes: %v\n", sizes)
}
func Llb_Nonlin4Cluster(g *Aig_Man_t, dd **DdManager, order **Vec_Int_t, groups **Vec_Ptr_t, maxBDD, verbose int32) {
	if g == nil || dd == nil || order == nil || groups == nil {
		return
	}
	if *dd == nil {
		*dd = core.NewBDDManager(int(aig.Aig_ManObjNumMax(g)))
	}
	var count int32
	*order = Llb_Nonlin4FindOrder(g, &count)
	*groups = Llb_Nonlin4FindPartitions(*dd, g, *order, 0)
}
