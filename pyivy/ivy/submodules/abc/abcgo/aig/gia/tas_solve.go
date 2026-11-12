package gia

import "github.com/glycerine/abcgo/misc/vec"

func Tas_ManSolve_rec(p *Tas_Man_t, level int32) int32 {
	if Tas_QueIsEmpty(&p.pProp) != 0 {
		panic("gia: SAT search has no pending propagation")
	}
	if conflict := Tas_ManPropagate(p, level); conflict != 0 {
		Tas_ManCreateCls(p, conflict)
		return conflict
	}
	if Tas_QueIsEmpty(&p.pJust) != 0 {
		return 0
	}
	p.Pars.nJustThis = max(p.Pars.nJustThis, p.pJust.iTail-p.pJust.iHead)
	if Tas_ManCheckLimits(p) != 0 {
		return 0
	}
	propHead := p.pProp.iHead
	var justHead, justTail int32
	Tas_QueStore(&p.pJust, &justHead, &justTail)
	var node, decision *Gia_Obj_t
	switch {
	case p.Pars.fUseActive != 0:
		decision = Tas_ManFindActive(p)
	case p.Pars.fUseHighest != 0:
		node = Tas_ManDecideHighest(p)
	case p.Pars.fUseLowest != 0:
		node = Tas_ManDecideLowest(p)
	case p.Pars.fUseMaxFF != 0:
		node = Tas_ManDecideMaxFF(p)
	default:
		panic("gia: no SAT decision strategy")
	}
	if node != nil {
		if Tas_VarIsJust(node) == 0 {
			panic("gia: decision node is already justified")
		}
		if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(node)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(node)) {
			decision = Gia_Not(Gia_ObjChild0(node))
		} else {
			decision = Gia_Not(Gia_ObjChild1(node))
		}
	}
	if decision == nil {
		panic("gia: SAT decision variable missing")
	}
	Tas_ManAssign(p, decision, level+1, nil, nil)
	first := Tas_ManSolve_rec(p, level+1)
	if first == 0 {
		return 0
	}
	if p.pClauses.pData[first] != Gia_Regular(decision) {
		return first
	}
	Tas_ManCancelUntil(p, propHead)
	Tas_QueRestore(&p.pJust, justHead, justTail)
	Tas_ManAssign(p, Gia_Not(decision), level+1, nil, nil)
	second := Tas_ManSolve_rec(p, level+1)
	if second == 0 {
		return 0
	}
	if p.pClauses.pData[second] != Gia_Regular(decision) {
		return second
	}
	conflict := Tas_ManResolve(p, level, first, second)
	Tas_ManCreateCls(p, conflict)
	p.Pars.nBTThis++
	return conflict
}

// Tas_ManSolveArray solves the conjunction of requested object edges.
// It returns 0 for SAT, 1 for UNSAT and -1 if resource limits were reached.
func Tas_ManSolveArray(p *Tas_Man_t, objects []*Gia_Obj_t) int32 {
	vec.Vec_IntClear(p.vModel)
	if Tas_QueIsEmpty(&p.pProp) == 0 || p.pProp.iHead != 0 || Tas_QueIsEmpty(&p.pJust) == 0 || p.pJust.iHead != 0 || p.pClauses.iHead != 1 || p.pClauses.iTail != 1 {
		panic("gia: SAT manager is not reset")
	}
	zero := Gia_ManConst0(p.pAig)
	one := Gia_Not(zero)
	seen := make(map[*Gia_Obj_t]bool)
	for _, edge := range objects {
		if edge == nil {
			panic("gia: nil SAT assumption")
		}
		if edge == zero || seen[Gia_Not(edge)] {
			return 1
		}
		Gia_ObjId(p.pAig, Gia_Regular(edge))
		seen[edge] = true
	}
	if p.pAig.Refs == nil {
		Gia_ManCreateRefs(p.pAig)
	}
	p.Pars.nBTThis, p.Pars.nJustThis, p.Pars.nBTThisNc = 0, 0, 0
	for _, edge := range objects {
		if edge != one && Tas_VarIsAssigned(Gia_Regular(edge)) == 0 {
			Tas_ManAssign(p, edge, 0, nil, nil)
		}
	}
	result := int32(0)
	if p.pProp.iTail != 0 {
		if Tas_ManSolve_rec(p, 0) == 0 && Tas_ManCheckLimits(p) == 0 {
			Tas_ManSaveModel(p, p.vModel)
		} else {
			result = 1
		}
	}
	Tas_ManCancelUntil(p, 0)
	p.pJust.iHead, p.pJust.iTail = 0, 0
	p.pClauses.iHead, p.pClauses.iTail = 1, 1
	if p.nClauses > 0 {
		p.pStore.reset()
		for _, literal := range vec.Vec_IntArray(p.vWatchLits) {
			p.pWatches[literal] = 0
		}
		vec.Vec_IntClear(p.vWatchLits)
		p.nClauses = 0
	}
	for _, id := range vec.Vec_IntArray(p.vActiveVars) {
		p.pActivity[id] = 0
	}
	vec.Vec_IntClear(p.vActiveVars)
	p.Pars.nBTTotal += p.Pars.nBTThis
	p.Pars.nJustTotal = max(p.Pars.nJustTotal, p.Pars.nJustThis)
	if Tas_ManCheckLimits(p) != 0 {
		result = -1
	}
	return result
}

func Tas_ManSolve(p *Tas_Man_t, first, second *Gia_Obj_t) int32 {
	if first == nil {
		panic("gia: nil SAT root")
	}
	objects := []*Gia_Obj_t{first}
	if second != nil {
		objects = append(objects, second)
	}
	return Tas_ManSolveArray(p, objects)
}
