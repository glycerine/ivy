package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

type Llb_Mnx_t struct {
	pAig                            *Aig_Man_t
	pPars                           *Gia_ParLlb_t
	dd                              *DdManager
	bBad, bReached, bCurrent, bNext *DdNode
	vRings, vRoots                  *Vec_Ptr_t
	vOrder, vVars2Q                 *Vec_Int_t
}

func Llb_Nonlin4ComputeBad(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t) *DdNode {
	if dd == nil || g == nil {
		return nil
	}
	inputs := make([]*core.DdNode, len(g.Inputs()))
	for i, e := range g.Inputs() {
		index := int(e.NodeID())
		if order != nil && int(e.NodeID()) < vec.Vec_IntSize(order) {
			index = int(vec.Vec_IntEntry(order, e.NodeID()))
		}
		if index >= 0 && index < len(dd.Vars) {
			inputs[i] = core.Cudd_bddIthVar(dd, index)
		}
	}
	fs := llbBuildBDDTable(g, dd, inputs)
	if fs == nil {
		return nil
	}
	count := len(g.Outputs()) - int(g.RegisterCount())
	bad := core.Cudd_ReadLogicZero(dd)
	for i := 0; i < count; i++ {
		bad = core.Cudd_bddOr(dd, bad, llbBDDForEdge(g, fs, g.Outputs()[i]))
	}
	return bad
}
func Llb_Nonlin4DerivePartitions(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(8)
	if dd == nil || g == nil {
		return result
	}
	inputs := make([]*core.DdNode, len(g.Inputs()))
	for i := range inputs {
		inputs[i] = core.Cudd_bddIthVar(dd, i)
	}
	fs := llbBuildBDDTable(g, dd, inputs)
	if fs != nil {
		for _, e := range g.Outputs() {
			vec.Vec_PtrPush(result, llbBDDForEdge(g, fs, e))
		}
	}
	return result
}
func Llb_Nonlin4CreateOrderSimple(g *Aig_Man_t) *Vec_Int_t { return Llb_Nonlin4FindOrder(g, nil) }
func Llb_Nonlin4CreateOrder(g *Aig_Man_t) *Vec_Int_t       { return Llb_Nonlin4FindOrder(g, nil) }
func Llb_Nonlin4CreateOrder_rec(g *Aig_Man_t, e *Aig_Obj_t, order *Vec_Int_t, counter *int32) {
	Llb_Nonlin4FindOrder_rec(g, e, order, counter)
}
func Llb_Nonlin4CollectHighRefNodes(g *Aig_Man_t, minRefs int32) *Vec_Int_t {
	r := vec.Vec_IntAlloc(16)
	if g == nil {
		return r
	}
	for _, e := range g.LiveEdges() {
		if aig.Aig_ObjRefs(g, e) >= minRefs {
			vec.Vec_IntPush(r, int32(e.NodeID()))
		}
	}
	return r
}
func Llb_Nonlin4CreateVars2Q(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, backward int32) *Vec_Int_t {
	result := vec.Vec_IntAlloc(0)
	if dd == nil || g == nil {
		return result
	}
	vec.Vec_IntFill(result, len(dd.Vars), 1)
	registers := int(g.RegisterCount())
	inputs, outputs := g.Inputs(), g.Outputs()
	if registers < 0 || registers > len(inputs) || registers > len(outputs) {
		return result
	}
	stateEdges := outputs[len(outputs)-registers:]
	if backward != 0 {
		stateEdges = inputs[len(inputs)-registers:]
	}
	for _, edge := range stateEdges {
		index := int32(edge.NodeID())
		if order != nil {
			if int(edge.NodeID()) >= int(vec.Vec_IntSize(order)) {
				return result
			}
			index = vec.Vec_IntEntry(order, int(edge.NodeID()))
		}
		if index < 0 || int(index) >= len(dd.Vars) {
			return result
		}
		vec.Vec_IntWriteEntry(result, int(index), 0)
	}
	return result
}
func Llb_Nonlin4SetupVarMap(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t) {}
func Llb_Nonlin4ComputeInitState(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, backward int32) *DdNode {
	if dd == nil || g == nil {
		return nil
	}
	registers := int(g.RegisterCount())
	inputs, outputs := g.Inputs(), g.Outputs()
	if registers < 0 || registers > len(inputs) || registers > len(outputs) {
		return nil
	}
	stateEdges := inputs[len(inputs)-registers:]
	if backward != 0 {
		stateEdges = outputs[len(outputs)-registers:]
	}
	state := core.Cudd_ReadOne(dd)
	for _, edge := range stateEdges {
		index := int32(edge.NodeID())
		if order != nil {
			if int(edge.NodeID()) >= int(vec.Vec_IntSize(order)) {
				return nil
			}
			index = vec.Vec_IntEntry(order, int(edge.NodeID()))
		}
		if index < 0 || int(index) >= len(dd.Vars) {
			return nil
		}
		state = core.Cudd_bddAnd(dd, state, core.Cudd_Not(core.Cudd_bddIthVar(dd, int(index))))
		if state == nil {
			return nil
		}
	}
	return state
}
func Llb_Nonlin4ComputeCube(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, values string, flag int32) *DdNode {
	if dd == nil || g == nil {
		return nil
	}
	cube := core.Cudd_ReadOne(dd)
	registers := int(g.RegisterCount())
	inputs, outputs := g.Inputs(), g.Outputs()
	if registers < 0 || registers > len(inputs) || registers > len(outputs) {
		return nil
	}
	valueEdges, variableEdges := inputs[len(inputs)-registers:], outputs[len(outputs)-registers:]
	if flag != 0 {
		valueEdges, variableEdges = variableEdges, valueEdges
	}
	for i := 0; i < registers; i++ {
		valueIndex := int32(valueEdges[i].NodeID())
		variableIndex := int32(variableEdges[i].NodeID())
		if order != nil {
			if int(valueEdges[i].NodeID()) >= int(vec.Vec_IntSize(order)) || int(variableEdges[i].NodeID()) >= int(vec.Vec_IntSize(order)) {
				return nil
			}
			valueIndex = vec.Vec_IntEntry(order, int(valueEdges[i].NodeID()))
			variableIndex = vec.Vec_IntEntry(order, int(variableEdges[i].NodeID()))
		}
		if valueIndex < 0 || int(valueIndex) >= len(values) || variableIndex < 0 || int(variableIndex) >= len(dd.Vars) {
			return nil
		}
		v := core.Cudd_bddIthVar(dd, int(variableIndex))
		if values[valueIndex] != 1 {
			v = core.Cudd_Not(v)
		}
		cube = core.Cudd_bddAnd(dd, cube, v)
		if cube == nil {
			return nil
		}
	}
	return cube
}
func Llb_Nonlin4RecordState(g *Aig_Man_t, order *Vec_Int_t, state []uint32, values string, backward int32) {
	if g == nil || state == nil {
		return
	}
	registers := int(g.RegisterCount())
	inputs, outputs := g.Inputs(), g.Outputs()
	if registers < 0 || registers > len(inputs) || registers > len(outputs) || len(state)*32 < registers {
		return
	}
	stateEdges := inputs[len(inputs)-registers:]
	if backward != 0 {
		stateEdges = outputs[len(outputs)-registers:]
	}
	for i, edge := range stateEdges {
		index := int32(edge.NodeID())
		if order != nil {
			if int(edge.NodeID()) >= int(vec.Vec_IntSize(order)) {
				return
			}
			index = vec.Vec_IntEntry(order, int(edge.NodeID()))
		}
		if index < 0 || int(index) >= len(values) {
			return
		}
		mask := uint32(1) << uint(i&31)
		if values[index] == 1 {
			state[i/32] |= mask
		} else {
			state[i/32] &^= mask
		}
	}
}
func Llb_Nonlin4Multiply(dd *DdManager, cube *DdNode, parts *Vec_Ptr_t) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(1)
	if dd == nil {
		return result
	}
	f := cube
	for _, x := range vec.Vec_PtrArray(parts) {
		if part, ok := x.(*core.DdNode); ok {
			f = core.Cudd_bddAnd(dd, f, part)
		}
	}
	vec.Vec_PtrPush(result, f)
	return result
}
func Llb_Nonlin4Deref(dd *DdManager, parts *Vec_Ptr_t) {
	if parts != nil {
		vec.Vec_PtrFree(parts)
	}
}
func Llb_Nonlin4DeriveCex(p *Llb_Mnx_t, backward, verbose int32) *Vec_Ptr_t { return nil }
func Llb_Nonlin4Reachability(p *Llb_Mnx_t) int32 {
	if p == nil || p.pAig == nil || p.pPars == nil {
		return -1
	}
	manager := &Llb_Man_t_{pAig: p.pAig, pAigGlo: p.pAig, pPars: p.pPars}
	status := Llb_ManReachability(manager, nil, nil)
	p.dd = manager.dd
	p.bReached = manager.reached
	return status
}
func Llb_Nonlin4Reorder(dd *DdManager, twice, verbose int32) {
	if dd != nil {
		core.Cudd_ReduceHeap(dd, core.CUDD_REORDER_SYMM_SIFT, 1)
	}
}
func Llb_MnxStart(g *Aig_Man_t, params *Gia_ParLlb_t) *Llb_Mnx_t {
	if g == nil || params == nil {
		return nil
	}
	return &Llb_Mnx_t{pAig: g, pPars: params, vOrder: Llb_Nonlin4CreateOrder(g), vRings: vec.Vec_PtrAlloc(16)}
}
func Llb_MnxStop(p *Llb_Mnx_t) {
	if p != nil && p.dd != nil {
		core.Cudd_Quit(p.dd)
		p.dd = nil
	}
}
func Llb_MnxCheckNextStateVars(p *Llb_Mnx_t) {
	if p == nil || p.pAig == nil {
		return
	}
	if int(p.pAig.RegisterCount()) > len(p.pAig.Outputs()) {
		panic("llb: invalid next-state output count")
	}
}
func Llb_Nonlin4CoreReach(g *Aig_Man_t, params *Gia_ParLlb_t) int32 {
	if g == nil || params == nil {
		return -1
	}
	if params.fSkipReach != 0 {
		return -1
	}
	return Llb_Nonlin4Reachability(Llb_MnxStart(g, params))
}

func llbBDDToAIG(dd *core.DdManager, root *core.DdNode, graph *aig.Graph, inputs []aig.Edge, memo map[*core.DdNode]aig.Edge) aig.Edge {
	if core.Cudd_IsConstant(root) {
		if root == dd.One {
			return graph.Const1()
		}
		return graph.Const0()
	}
	complement := core.Cudd_IsComplement(root)
	regular := core.Cudd_Regular(root)
	if e, ok := memo[regular]; ok {
		if complement {
			return e.Not()
		}
		return e
	}
	index := int(regular.Index)
	if index < 0 || index >= len(inputs) {
		return aig.InvalidEdge
	}
	t := llbBDDToAIG(dd, regular.T, graph, inputs, memo)
	e := llbBDDToAIG(dd, regular.E, graph, inputs, memo)
	x := inputs[index]
	value := graph.AND(x, t).Not().Not() // keep a stable positive edge
	value = graph.AND(graph.AND(x, t).Not(), graph.AND(x.Not(), e).Not()).Not()
	memo[regular] = value
	if complement {
		return value.Not()
	}
	return value
}
func Llb_ReachableStates(g *Aig_Man_t) *Aig_Man_t {
	if g == nil {
		return nil
	}
	params := &Gia_ParLlb_t{nIterMax: 100000, nBddMax: 1000000, fSkipOutCheck: 1, fSilent: 1}
	manager := &Llb_Man_t_{pAig: g, pAigGlo: g, pPars: params}
	if Llb_ManReachability(manager, nil, nil) != 1 || manager.reached == nil {
		return nil
	}
	out := aig.NewGraph()
	registers := int(g.RegisterCount())
	inputs := make([]aig.Edge, registers)
	for i := range inputs {
		inputs[i] = out.NewInput()
	}
	memo := map[*core.DdNode]aig.Edge{}
	for i := range inputs {
		_ = i
	}
	out.AddOutput(llbBDDToAIG(manager.dd, manager.reached, out, inputs, memo))
	aig.Aig_ManSetRegNum(out, 0)
	core.Cudd_Quit(manager.dd)
	return out
}
