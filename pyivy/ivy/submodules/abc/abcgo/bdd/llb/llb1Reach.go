package llb

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_ManConstructOutBdd builds the BDD for one complemented AIG edge. Caller
// variables follow the graph's CI order.
func Llb_ManConstructOutBdd(g *Aig_Man_t, edge *Aig_Obj_t, dd *DdManager) *DdNode {
	if g == nil || edge == nil || dd == nil {
		return nil
	}
	variables := make([]*core.DdNode, len(g.Inputs()))
	for i := range variables {
		variables[i] = core.Cudd_bddIthVar(dd, i)
	}
	return llbBDDForEdge(g, llbBuildBDDTable(g, dd, variables), *edge)
}

// Llb_ManConstructGroupBdd relates each group output variable to its AIG
// function, while assigning BDD variables to the group's input nodes.
func Llb_ManConstructGroupBdd(p *Llb_Man_t, group *Llb_Grp_t) *DdNode {
	if p == nil || p.dd == nil || p.pAig == nil || p.vObj2Var == nil || group == nil {
		return nil
	}
	functions := make([]*DdNode, len(p.pAig.Nodes()))
	functions[0] = core.Cudd_ReadOne(p.dd)
	if group.vIns != nil {
		for i := 0; i < vec.Vec_PtrSize(group.vIns); i++ {
			edge := llbGroupEdge(group.vIns, i)
			id := edge.NodeID()
			if id < 0 || id >= len(functions) || id >= vec.Vec_IntSize(p.vObj2Var) {
				return nil
			}
			index := vec.Vec_IntEntry(p.vObj2Var, id)
			if index < 0 || int(index) >= len(p.dd.Vars) {
				return nil
			}
			functions[id] = core.Cudd_bddIthVar(p.dd, int(index))
		}
	}
	if group.vNodes != nil {
		for i := 0; i < vec.Vec_PtrSize(group.vNodes); i++ {
			edge := llbGroupEdge(group.vNodes, i)
			node := p.pAig.Node(edge)
			left := llbBDDForEdge(p.pAig, functions, node.Fanins[0])
			right := llbBDDForEdge(p.pAig, functions, node.Fanins[1])
			if left == nil || right == nil {
				return nil
			}
			functions[edge.NodeID()] = core.Cudd_bddAnd(p.dd, left, right)
			if functions[edge.NodeID()] == nil {
				return nil
			}
		}
	}
	result := core.Cudd_ReadOne(p.dd)
	if group.vOuts == nil {
		return result
	}
	for i := 0; i < vec.Vec_PtrSize(group.vOuts); i++ {
		edge := llbGroupEdge(group.vOuts, i)
		id := edge.NodeID()
		if id < 0 || id >= len(functions) || id >= vec.Vec_IntSize(p.vObj2Var) {
			return nil
		}
		index := vec.Vec_IntEntry(p.vObj2Var, id)
		if index < 0 || int(index) >= len(p.dd.Vars) {
			return nil
		}
		output := llbBDDForEdge(p.pAig, functions, edge)
		if p.pAig.Node(edge).Kind == aig.Output {
			output = llbBDDForEdge(p.pAig, functions, p.pAig.Node(edge).Fanins[0])
		}
		if output == nil {
			return nil
		}
		result = core.Cudd_bddAnd(p.dd, result, core.Cudd_bddXnor(p.dd, output, core.Cudd_bddIthVar(p.dd, int(index))))
		if result == nil {
			return nil
		}
	}
	return result
}

func llbGroupVariables(p *Llb_Man_t, group *Llb_Grp_t, include func(int, int) bool) *DdNode {
	if p == nil || p.dd == nil || p.vObj2Var == nil || group == nil {
		return nil
	}
	cube := core.Cudd_ReadOne(p.dd)
	seen := map[int]bool{}
	for _, boundary := range []*Vec_Ptr_t{group.vIns, group.vOuts} {
		for i := 0; i < vec.Vec_PtrSize(boundary); i++ {
			edge := llbGroupEdge(boundary, i)
			id := int(aig.Aig_ObjId(edge))
			if id < 0 || id >= int(vec.Vec_IntSize(p.vObj2Var)) {
				return nil
			}
			index := int(vec.Vec_IntEntry(p.vObj2Var, id))
			if index < 0 || index >= len(p.dd.Vars) {
				return nil
			}
			if seen[index] || (include != nil && !include(id, index)) {
				continue
			}
			seen[index] = true
			cube = core.Cudd_bddAnd(p.dd, cube, core.Cudd_bddIthVar(p.dd, index))
			if cube == nil {
				return nil
			}
		}
	}
	return cube
}

func Llb_ManConstructQuantCubeIntern(p *Llb_Man_t, group *Llb_Grp_t, place int32, backward int32) *DdNode {
	if p == nil || p.pAig == nil || group == nil || p.vVarBegs == nil || p.vVarEnds == nil {
		return nil
	}
	inputs := p.pAig.Inputs()
	pis := len(inputs) - int(p.pAig.RegisterCount())
	if pis < 0 {
		return nil
	}
	primaryInput := make(map[int]bool, pis)
	for _, edge := range inputs[:pis] {
		primaryInput[int(edge.NodeID())] = true
	}
	return llbGroupVariables(p, group, func(id, _ int) bool {
		if backward != 0 && primaryInput[id] {
			return false
		}
		if id >= vec.Vec_IntSize(p.vVarBegs) || id >= vec.Vec_IntSize(p.vVarEnds) {
			return false
		}
		first, last := vec.Vec_IntEntry(p.vVarBegs, id), vec.Vec_IntEntry(p.vVarEnds, id)
		return first >= 0 && first <= last && first == last
	})
}
func Llb_ManConstructQuantCubeFwd(p *Llb_Man_t, group *Llb_Grp_t, place int32) *DdNode {
	if p == nil || group == nil || p.vVarEnds == nil {
		return nil
	}
	return llbGroupVariables(p, group, func(id, _ int) bool {
		return id < vec.Vec_IntSize(p.vVarEnds) && vec.Vec_IntEntry(p.vVarEnds, id) == place
	})
}
func Llb_ManConstructQuantCubeBwd(p *Llb_Man_t, group *Llb_Grp_t, place int32) *DdNode {
	if p == nil || group == nil || p.vVarBegs == nil || p.pAig == nil {
		return nil
	}
	inputs := p.pAig.Inputs()
	pis := len(inputs) - int(p.pAig.RegisterCount())
	if pis < 0 {
		return nil
	}
	primaryInput := make(map[int]bool, pis)
	for _, edge := range inputs[:pis] {
		primaryInput[int(edge.NodeID())] = true
	}
	return llbGroupVariables(p, group, func(id, _ int) bool {
		return !primaryInput[id] && id < vec.Vec_IntSize(p.vVarBegs) && vec.Vec_IntEntry(p.vVarBegs, id) == place
	})
}

// Llb_ManComputeInitState returns the all-zero latch state.
func Llb_ManComputeInitState(p *Llb_Man_t, dd *DdManager) *DdNode {
	if p == nil || p.pAig == nil || dd == nil {
		return nil
	}
	state := core.Cudd_ReadOne(dd)
	for i := 0; i < int(p.pAig.RegisterCount()); i++ {
		state = core.Cudd_bddAnd(dd, state, core.Cudd_Not(core.Cudd_bddIthVar(dd, i)))
		if state == nil {
			return nil
		}
	}
	return state
}

// Llb_ManComputeImage computes one exact symbolic transition image. State
// variables occupy [0,nRegs), PIs [nRegs,nRegs+nPis), and next-state variables
// follow them. It existentially removes current state and PI variables before
// renaming next-state variables back to state variables.
func Llb_ManComputeImage(p *Llb_Man_t, states *DdNode, backward int32) *DdNode {
	if p == nil || p.pAig == nil || p.dd == nil || states == nil {
		return nil
	}
	g := p.pAig
	nRegs := int(g.RegisterCount())
	inputs := g.Inputs()
	nPis := len(inputs) - nRegs
	if nRegs < 0 || nPis < 0 {
		return nil
	}
	variables := make([]*core.DdNode, len(inputs))
	for i := range inputs {
		if i < nPis {
			variables[i] = core.Cudd_bddIthVar(p.dd, nRegs+i)
		} else {
			variables[i] = core.Cudd_bddIthVar(p.dd, i-nPis)
		}
	}
	functions := llbBuildBDDTable(g, p.dd, variables)
	if functions == nil {
		return nil
	}
	relation := states
	for i := 0; i < nRegs; i++ {
		nextEdge := g.Outputs()[len(g.Outputs())-nRegs+i]
		nextValue := llbBDDForEdge(g, functions, nextEdge)
		nextVar := core.Cudd_bddIthVar(p.dd, nRegs+nPis+i)
		relation = core.Cudd_bddAnd(p.dd, relation, core.Cudd_bddXnor(p.dd, nextVar, nextValue))
		if relation == nil {
			return nil
		}
	}
	quantify := core.Cudd_ReadOne(p.dd)
	for i := 0; i < nRegs+nPis; i++ {
		quantify = core.Cudd_bddAnd(p.dd, quantify, core.Cudd_bddIthVar(p.dd, i))
		if quantify == nil {
			return nil
		}
	}
	image := core.Cudd_bddExistAbstract(p.dd, relation, quantify)
	if image == nil {
		return nil
	}
	permutation := make([]int, len(p.dd.Vars))
	for i := range permutation {
		permutation[i] = i
	}
	for i := 0; i < nRegs; i++ {
		permutation[nRegs+nPis+i] = i
	}
	return core.Cudd_bddPermute(p.dd, image, permutation)
}

// Llb_ManCreateConstraints combines the selected AIG literals into one BDD.
func Llb_ManCreateConstraints(p *Llb_Man_t, hints *Vec_Int_t, useNext int32) *DdNode {
	if p == nil || p.dd == nil {
		return nil
	}
	if hints == nil {
		return core.Cudd_ReadOne(p.dd)
	}
	inputs := make([]*core.DdNode, len(p.pAigGlo.Inputs()))
	for i := range inputs {
		inputs[i] = core.Cudd_bddIthVar(p.dd, i)
	}
	functions := llbBuildBDDTable(p.pAigGlo, p.dd, inputs)
	result := core.Cudd_ReadOne(p.dd)
	for id, polarity := range vec.Vec_IntArray(hints) {
		if polarity != 0 && polarity != 1 {
			continue
		}
		if id >= len(functions) || functions[id] == nil {
			continue
		}
		f := functions[id]
		if polarity == 0 {
			f = core.Cudd_Not(f)
		}
		result = core.Cudd_bddAnd(p.dd, result, f)
		if result == nil {
			return nil
		}
	}
	return result
}

func llbPickAssignment(dd *core.DdManager, function *core.DdNode) ([]bool, bool) {
	if dd == nil || function == nil {
		return nil, false
	}
	assignment := make([]bool, len(dd.Vars))
	var visit func(*core.DdNode) bool
	visit = func(edge *core.DdNode) bool {
		if edge == nil {
			return false
		}
		if core.Cudd_IsConstant(edge) {
			return edge == dd.One
		}
		regular, complement := core.Cudd_Regular(edge), core.Cudd_IsComplement(edge)
		index := int(regular.Index)
		low := core.Cudd_NotCond(regular.E, complement)
		before := append([]bool(nil), assignment...)
		if low != dd.LogicZero && visit(low) {
			assignment[index] = false
			return true
		}
		copy(assignment, before)
		high := core.Cudd_NotCond(regular.T, complement)
		if high != dd.LogicZero && visit(high) {
			assignment[index] = true
			return true
		}
		copy(assignment, before)
		return false
	}
	return assignment, visit(function)
}

func llbTransitionRelation(g *Aig_Man_t, dd *core.DdManager) *core.DdNode {
	if g == nil || dd == nil {
		return nil
	}
	ci := g.Inputs()
	registers := int(g.RegisterCount())
	pis := len(ci) - registers
	if pis < 0 || len(g.Outputs()) < registers {
		return nil
	}
	variables := make([]*core.DdNode, len(ci))
	for i := 0; i < pis; i++ {
		variables[i] = core.Cudd_bddIthVar(dd, registers+i)
	}
	for i := 0; i < registers; i++ {
		variables[pis+i] = core.Cudd_bddIthVar(dd, i)
	}
	functions := llbBuildBDDTable(g, dd, variables)
	if functions == nil {
		return nil
	}
	relation := core.Cudd_ReadOne(dd)
	outputs := g.Outputs()
	for i := 0; i < registers; i++ {
		next := outputs[len(outputs)-registers+i]
		index := registers + pis + i
		relation = core.Cudd_bddAnd(dd, relation, core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, index), llbBDDForEdge(g, functions, next)))
		if relation == nil {
			return nil
		}
	}
	return relation
}

// Llb_ManReachDeriveCex walks the saved reachable layers backward through the
// exact transition relation, choosing one predecessor and input per frame.
func Llb_ManReachDeriveCex(p *Llb_Man_t) *Abc_Cex_t {
	if p == nil || p.dd == nil || len(p.frameSets) == 0 || len(p.failureValues) == 0 {
		return nil
	}
	g := p.pAig
	registers := int(g.RegisterCount())
	pis := len(g.Inputs()) - registers
	frames := len(p.frameSets)
	if registers < 0 || pis < 0 {
		return nil
	}
	states := make([][]bool, frames)
	inputs := make([][]bool, frames)
	state := append([]bool(nil), p.failureValues[:registers]...)
	inputs[frames-1] = append([]bool(nil), p.failureValues[registers:registers+pis]...)
	states[frames-1] = state
	relation := llbTransitionRelation(g, p.dd)
	if relation == nil {
		return nil
	}
	for frame := frames - 2; frame >= 0; frame-- {
		candidate := p.frameSets[frame]
		candidate = core.Cudd_bddAnd(p.dd, candidate, relation)
		for i, value := range state {
			v := core.Cudd_bddIthVar(p.dd, registers+pis+i)
			if !value {
				v = core.Cudd_Not(v)
			}
			candidate = core.Cudd_bddAnd(p.dd, candidate, v)
		}
		if candidate == nil {
			return nil
		}
		assignment, ok := llbPickAssignment(p.dd, candidate)
		if !ok {
			return nil
		}
		state = append([]bool(nil), assignment[:registers]...)
		states[frame] = state
		inputs[frame] = append([]bool(nil), assignment[registers:registers+pis]...)
	}
	cex := util.Abc_CexAlloc(int32(registers), int32(pis), int32(frames))
	if cex == nil {
		return nil
	}
	cex.IFrame = int32(frames - 1)
	for i, value := range states[0] {
		if value {
			util.Abc_InfoSetBit(cex.PData, int32(i))
		}
	}
	for frame := range inputs {
		for i, value := range inputs[frame] {
			if value {
				util.Abc_InfoSetBit(cex.PData, int32(registers+frame*pis+i))
			}
		}
	}
	assignment := make([]bool, len(g.Inputs()))
	copy(assignment, inputs[frames-1])
	copy(assignment[pis:], states[frames-1])
	outputs, err := g.Evaluate(assignment)
	if err != nil {
		return nil
	}
	propertyCount := len(outputs) - registers
	for i := 0; i < propertyCount; i++ {
		if outputs[i] {
			cex.IPo = int32(i)
			return cex
		}
	}
	return nil
}

// Llb_ManReachability performs exact forward symbolic reachability on the
// native AIG. It returns 0 for a reachable bad output, 1 for a fixed point,
// and -1 when configured limits stop the search.
func Llb_ManReachability(p *Llb_Man_t, hints *Vec_Int_t, global **DdManager) int32 {
	if p == nil || p.pAig == nil || p.pPars == nil {
		return -1
	}
	g := p.pAig
	nRegs := int(g.RegisterCount())
	nPis := len(g.Inputs()) - nRegs
	if nRegs < 0 || nPis < 0 || len(g.Outputs()) < nRegs {
		return -1
	}
	p.pPars.iFrame = -1
	p.reached = nil
	p.frameSets = nil
	p.failureValues = nil
	p.dd = core.NewBDDManager(2*nRegs + nPis)
	if p.dd == nil {
		return -1
	}
	initial := Llb_ManComputeInitState(p, p.dd)
	if initial == nil {
		return -1
	}
	p.vRings = vec.Vec_PtrAlloc(16)
	current, reached := initial, initial
	limit := int(p.pPars.nIterMax)
	if limit <= 0 {
		limit = 100000
	}
	deadline := time.Time{}
	if p.pPars.TimeLimit > 0 {
		deadline = time.Now().Add(time.Duration(p.pPars.TimeLimit) * time.Second)
	}
	propertyCount := len(g.Outputs()) - nRegs
	for frame := 0; frame < limit; frame++ {
		p.pPars.iFrame = int32(frame)
		p.frameSets = append(p.frameSets, current)
		if !deadline.IsZero() && time.Now().After(deadline) {
			return -1
		}
		// Check if any state in the current frontier can assert a property.
		ciVars := make([]*core.DdNode, len(g.Inputs()))
		for i := range ciVars {
			if i < nPis {
				ciVars[i] = core.Cudd_bddIthVar(p.dd, nRegs+i)
			} else {
				ciVars[i] = core.Cudd_bddIthVar(p.dd, i-nPis)
			}
		}
		functions := llbBuildBDDTable(g, p.dd, ciVars)
		bad := core.Cudd_ReadLogicZero(p.dd)
		for i := 0; i < propertyCount; i++ {
			bad = core.Cudd_bddOr(p.dd, bad, llbBDDForEdge(g, functions, g.Outputs()[i]))
		}
		piCube := core.Cudd_ReadOne(p.dd)
		for i := 0; i < nPis; i++ {
			piCube = core.Cudd_bddAnd(p.dd, piCube, core.Cudd_bddIthVar(p.dd, nRegs+i))
		}
		badWithInputs := core.Cudd_bddAnd(p.dd, current, bad)
		badStates := core.Cudd_bddExistAbstract(p.dd, badWithInputs, piCube)
		if badStates != nil && badStates != core.Cudd_ReadLogicZero(p.dd) {
			p.reached = reached
			p.failureValues, _ = llbPickAssignment(p.dd, badWithInputs)
			if p.pPars.fSkipOutCheck == 0 && p.pAigGlo != nil {
				if cex := Llb_ManReachDeriveCex(p); cex != nil {
					_ = p.pAigGlo.SetSequentialCounterexample(cex)
				}
			}
			return 0
		}
		next := Llb_ManComputeImage(p, current, 0)
		if next == nil {
			return -1
		}
		if core.Cudd_bddLeq(p.dd, next, reached) {
			p.reached = reached
			return 1
		}
		if p.pPars.nBddMax > 0 && int32(core.Cudd_DagSize(next)) > p.pPars.nBddMax {
			p.reached = reached
			return -1
		}
		current = core.Cudd_bddAnd(p.dd, next, core.Cudd_Not(reached))
		if current == nil {
			return -1
		}
		reached = core.Cudd_bddOr(p.dd, reached, next)
		if reached == nil {
			return -1
		}
	}
	return -1
}
