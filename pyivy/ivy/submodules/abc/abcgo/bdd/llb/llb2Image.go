package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_ImgSupports reports each partition's live AIG-variable support using
// the supplied lower/upper node ranges.
func Llb_ImgSupports(g *Aig_Man_t, managers *Vec_Ptr_t, starts, stops *Vec_Int_t, addPis, verbose int32) *Vec_Ptr_t {
	supports := vec.Vec_PtrAlloc(4)
	if g == nil {
		return supports
	}
	width := int(aig.Aig_ManObjNumMax(g))
	initial := vec.Vec_IntStart(width)
	if starts != nil {
		for _, id := range vec.Vec_IntArray(starts) {
			if id >= 0 && int(id) < width {
				vec.Vec_IntWriteEntry(initial, int(id), 1)
			}
		}
	}
	vec.Vec_PtrPush(supports, initial)
	// The direct backend does not maintain temporary per-manager functions;
	// keep one empty intermediate support for each transition partition.
	if managers != nil {
		for range vec.Vec_PtrArray(managers) {
			vec.Vec_PtrPush(supports, vec.Vec_IntStart(width))
		}
	}
	final := vec.Vec_IntStart(width)
	if stops != nil {
		for _, id := range vec.Vec_IntArray(stops) {
			if id >= 0 && int(id) < width {
				vec.Vec_IntWriteEntry(final, int(id), 1)
			}
		}
	}
	if addPis != 0 {
		nPis := len(g.Inputs()) - int(g.RegisterCount())
		for i := 0; i < nPis; i++ {
			vec.Vec_IntWriteEntry(final, g.Inputs()[i].NodeID(), 1)
		}
	}
	vec.Vec_PtrPush(supports, final)
	return supports
}

// Llb_ImgSchedule initializes empty quantification schedules. Direct image
// computation quantifies from the transition relation's actual support.
func Llb_ImgSchedule(supports *Vec_Ptr_t, q0, q1 **Vec_Ptr_t, verbose int32) {
	if supports == nil || vec.Vec_PtrSize(supports) == 0 {
		return
	}
	if q0 != nil {
		*q0 = vec.Vec_PtrAlloc(vec.Vec_PtrSize(supports))
		for range vec.Vec_PtrArray(supports) {
			vec.Vec_PtrPush(*q0, vec.Vec_IntAlloc(4))
		}
	}
	if q1 != nil {
		*q1 = vec.Vec_PtrAlloc(vec.Vec_PtrSize(supports))
		for range vec.Vec_PtrArray(supports) {
			vec.Vec_PtrPush(*q1, vec.Vec_IntAlloc(4))
		}
	}
	first, ok := vec.Vec_PtrEntry(supports, 0).(*Vec_Int_t)
	if !ok {
		return
	}
	for variable := 0; variable < vec.Vec_IntSize(first); variable++ {
		count, last := 0, -1
		for i := 0; i < vec.Vec_PtrSize(supports); i++ {
			support, ok := vec.Vec_PtrEntry(supports, i).(*Vec_Int_t)
			if ok && vec.Vec_IntEntry(support, variable) != 0 {
				count++
				last = i
			}
		}
		if count == 0 || last < 0 {
			continue
		}
		if count == 1 && q0 != nil {
			vec.Vec_IntPush(vec.Vec_PtrEntry(*q0, last).(*Vec_Int_t), int32(variable))
		}
		if count > 1 && q1 != nil {
			vec.Vec_IntPush(vec.Vec_PtrEntry(*q1, last).(*Vec_Int_t), int32(variable))
		}
	}
}

// Llb_ImgPartition builds the local relation for gates between two cuts.
func Llb_ImgPartition(g *Aig_Man_t, lower, upper *Vec_Ptr_t, timeout abctime) *DdManager {
	if g == nil {
		return nil
	}
	dd := core.NewBDDManager(int(aig.Aig_ManObjNumMax(g)))
	if dd == nil {
		return nil
	}
	inputs := make([]*core.DdNode, len(g.Inputs()))
	for _, e := range g.Inputs() {
		if e.NodeID() < len(dd.Vars) {
			inputs[e.NodeID()] = core.Cudd_bddIthVar(dd, e.NodeID())
		}
	}
	// The node-ID table includes variables for lower-cut internal nodes.
	funcs := make([]*core.DdNode, int(aig.Aig_ManObjNumMax(g)))
	funcs[0] = core.Cudd_ReadOne(dd)
	for id, variable := range inputs {
		if variable != nil {
			funcs[g.Inputs()[id].NodeID()] = variable
		}
	}
	for _, item := range llbEdges(Llb_ManCutNodes(g, lower, upper)) {
		node := g.Node(item)
		left, right := llbBDDForEdge(g, funcs, node.Fanins[0]), llbBDDForEdge(g, funcs, node.Fanins[1])
		if node.Kind == aig.ExclusiveOr {
			funcs[item.NodeID()] = core.Cudd_bddXor(dd, left, right)
		} else {
			funcs[item.NodeID()] = core.Cudd_bddAnd(dd, left, right)
		}
		if funcs[item.NodeID()] == nil {
			core.Cudd_Quit(dd)
			return nil
		}
	}
	relation := core.Cudd_ReadOne(dd)
	for _, item := range llbEdges(Llb_ManCutRange(g, lower, upper)) {
		if item.NodeID() >= len(funcs) || funcs[item.NodeID()] == nil {
			continue
		}
		relation = core.Cudd_bddAnd(dd, relation, core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, item.NodeID()), funcs[item.NodeID()]))
		if relation == nil {
			core.Cudd_Quit(dd)
			return nil
		}
	}
	return dd
}

// Llb_ImgComputeCube creates the conjunction of variables named by node IDs.
func Llb_ImgComputeCube(g *Aig_Man_t, nodeIDs *Vec_Int_t, dd *DdManager) *DdNode {
	if g == nil || nodeIDs == nil || dd == nil {
		return nil
	}
	cube := core.Cudd_ReadOne(dd)
	for _, id := range vec.Vec_IntArray(nodeIDs) {
		if id < 0 || int(id) >= len(dd.Vars) {
			return nil
		}
		cube = core.Cudd_bddAnd(dd, cube, core.Cudd_bddIthVar(dd, int(id)))
		if cube == nil {
			return nil
		}
	}
	return cube
}

// Llb_ImgQuantifyFirst and Reset are compatibility hooks for the old
// partitioned engine; no BDD-manager state is mutated by the direct backend.
func Llb_ImgQuantifyFirst(g *Aig_Man_t, managers, quant *Vec_Ptr_t, verbose int32) {}
func Llb_ImgQuantifyReset(managers *Vec_Ptr_t)                                     {}

// Llb_ImgComputeImage computes the exact image over the graph's latch
// transition functions in a node-ID variable order.
func Llb_ImgComputeImage(g *Aig_Man_t, managers *Vec_Ptr_t, dd *DdManager, initial *DdNode, q0, q1, refs *Vec_Int_t, timeout abctime, backward, reorder, verbose int32) *DdNode {
	if g == nil || dd == nil || initial == nil {
		return nil
	}
	count := int(aig.Aig_ManObjNumMax(g))
	ci := g.Inputs()
	inputs := make([]*core.DdNode, len(ci))
	for i, edge := range ci {
		inputs[i] = core.Cudd_bddIthVar(dd, edge.NodeID())
	}
	functions := llbBuildBDDTable(g, dd, inputs)
	if functions == nil {
		return nil
	}
	registers := int(g.RegisterCount())
	outputs := g.Outputs()
	if registers < 0 || registers > len(outputs) {
		return nil
	}
	relation := initial
	for i := 0; i < registers; i++ {
		next := outputs[len(outputs)-registers+i]
		nextFunc := llbBDDForEdge(g, functions, next)
		id := next.NodeID()
		if id < 0 || id >= len(dd.Vars) {
			return nil
		}
		relation = core.Cudd_bddAnd(dd, relation, core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, id), nextFunc))
		if relation == nil {
			return nil
		}
	}
	quant := core.Cudd_ReadOne(dd)
	piCount := len(ci) - registers
	for i := 0; i < piCount; i++ {
		id := ci[i].NodeID()
		quant = core.Cudd_bddAnd(dd, quant, core.Cudd_bddIthVar(dd, id))
	}
	for i := 0; i < registers; i++ {
		id := ci[piCount+i].NodeID()
		quant = core.Cudd_bddAnd(dd, quant, core.Cudd_bddIthVar(dd, id))
	}
	if quant == nil {
		return nil
	}
	image := core.Cudd_bddExistAbstract(dd, relation, quant)
	if image == nil {
		return nil
	}
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = i
	}
	for i := 0; i < registers; i++ {
		nextID := outputs[len(outputs)-registers+i].NodeID()
		currentID := ci[piCount+i].NodeID()
		permutation[nextID] = currentID
	}
	_ = count
	return core.Cudd_bddPermute(dd, image, permutation)
}

var _ = aig.Edge(0)
