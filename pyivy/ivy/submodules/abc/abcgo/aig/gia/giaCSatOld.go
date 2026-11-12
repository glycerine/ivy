package gia

import (
	"fmt"
	"time"
)

// Cbs0 is the original chronological circuit-SAT solver. It propagates directly
// through the AND graph and branches on un-justified zero-valued gates; it does
// not use clauses, learning, or the later CBS engines.
type Cbs0_Par_t struct {
	nBTLimit, nJustLimit, nBTThis, nJustThis, nBTTotal, nJustTotal int32
	fUseHighest, fUseLowest, fUseMaxFF, fVerbose                   int32
}
type Cbs0_Que_t struct {
	iHead, iTail int32
	pData        []*Gia_Obj_t
}
type Cbs0_Man_t struct {
	Pars                                                                       Cbs0_Par_t
	pAig                                                                       *Gia_Man_t
	pProp, pJust                                                               Cbs0_Que_t
	vModel                                                                     *Vec_Int_t
	nSatUnsat, nSatSat, nSatUndec, nSatTotal, nConfUnsat, nConfSat, nConfUndec int32
	timeSatUnsat, timeSatSat, timeSatUndec, timeTotal                          time.Duration
}

func Cbs0_VarIsAssigned(o *Gia_Obj_t) int32 { return int32(o.fMark0) }
func Cbs0_VarAssign(o *Gia_Obj_t) {
	if o.fMark0 != 0 {
		panic("gia: CBS0 duplicate assignment")
	}
	o.fMark0 = 1
}
func Cbs0_VarUnassign(o *Gia_Obj_t) {
	if o.fMark0 == 0 {
		panic("gia: CBS0 unassigns unknown variable")
	}
	o.fMark0 = 0
	o.fMark1 = 0
}
func Cbs0_VarValue(o *Gia_Obj_t) int32 {
	if o.fMark0 == 0 {
		panic("gia: CBS0 unknown variable")
	}
	return int32(o.fMark1)
}
func Cbs0_VarSetValue(o *Gia_Obj_t, value int32) {
	if o.fMark0 == 0 || value < 0 || value > 1 {
		panic("gia: CBS0 invalid variable value")
	}
	o.fMark1 = uint32(value)
}
func cbs0FaninValue(o *Gia_Obj_t, index int) int32 {
	fanin := Gia_ObjFanin0(o)
	phase := Gia_ObjFaninC0(o)
	if index == 1 {
		fanin = Gia_ObjFanin1(o)
		phase = Gia_ObjFaninC1(o)
	}
	if fanin.id == 0 {
		return phase
	}
	if Cbs0_VarIsAssigned(fanin) == 0 {
		return 2
	}
	return Cbs0_VarValue(fanin) ^ phase
}
func Cbs0_VarFanin0Value(o *Gia_Obj_t) int32 { return cbs0FaninValue(o, 0) }
func Cbs0_VarFanin1Value(o *Gia_Obj_t) int32 { return cbs0FaninValue(o, 1) }
func Cbs0_VarIsJust(o *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsAnd(o) != 0 && Cbs0_VarFanin0Value(o) == 2 && Cbs0_VarFanin1Value(o) == 2)
}
func Cbs0_SetDefaultParams(p *Cbs0_Par_t) {
	*p = Cbs0_Par_t{nBTLimit: 1000, nJustLimit: 100, fUseHighest: 1, fVerbose: 1}
}
func Cbs0_ManAlloc(graphs ...*Gia_Man_t) *Cbs0_Man_t {
	p := &Cbs0_Man_t{vModel: Vec_IntAlloc(1000)}
	Cbs0_SetDefaultParams(&p.Pars)
	if len(graphs) > 1 {
		panic("gia: CBS0 accepts one graph")
	}
	if len(graphs) == 1 {
		p.pAig = graphs[0]
		cbs0CheckGraph(p.pAig)
		if len(p.pAig.Refs) != len(p.pAig.pObjs) {
			p.pAig.Refs = nil
			Gia_ManCreateRefs(p.pAig)
		}
		Gia_ManCleanMark0(p.pAig)
		Gia_ManCleanMark1(p.pAig)
	}
	return p
}
func cbs0CheckGraph(g *Gia_Man_t) {
	if g == nil {
		panic("gia: nil CBS0 graph")
	}
	for _, o := range g.pObjs {
		if Gia_ObjIsXor(o) != 0 || Gia_ObjIsMux(g, o) != 0 {
			panic("gia: CBS0 requires an ordinary AND graph")
		}
	}
}
func Cbs0_ManStop(p *Cbs0_Man_t) {
	if p == nil {
		return
	}
	if p.pProp.iTail != 0 {
		Cbs0_ManCancelUntil(p, 0)
	}
	p.pProp = Cbs0_Que_t{}
	p.pJust = Cbs0_Que_t{}
	p.vModel = nil
	p.pAig = nil
}
func Cbs0_ReadModel(p *Cbs0_Man_t) *Vec_Int_t { return p.vModel }
func Cbs0_ManCheckLimits(p *Cbs0_Man_t) int32 {
	return giaBool(p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit)
}
func Cbs0_ManSaveModel(p *Cbs0_Man_t, v *Vec_Int_t) {
	Vec_IntClear(v)
	p.pProp.iHead = 0
	for _, o := range p.pProp.pData[:p.pProp.iTail] {
		if Gia_ObjIsCi(o) != 0 {
			Vec_IntPush(v, (Gia_ObjCioId(o)<<1)|(Cbs0_VarValue(o)^1))
		}
	}
}
func Cbs0_QueIsEmpty(p *Cbs0_Que_t) int32 { return giaBool(p.iHead == p.iTail) }
func Cbs0_QuePush(p *Cbs0_Que_t, o *Gia_Obj_t) {
	if int(p.iTail) == len(p.pData) {
		p.pData = append(p.pData, o)
	} else {
		p.pData[p.iTail] = o
	}
	p.iTail++
}
func Cbs0_QueHasNode(p *Cbs0_Que_t, o *Gia_Obj_t) int32 {
	for _, n := range p.pData[p.iHead:p.iTail] {
		if n == o {
			return 1
		}
	}
	return 0
}
func Cbs0_QueStore(p *Cbs0_Que_t, head, tail *int32) {
	*head, *tail = p.iHead, p.iTail
	for i := *head; i < *tail; i++ {
		Cbs0_QuePush(p, p.pData[i])
	}
	p.iHead = *tail
}
func Cbs0_QueRestore(p *Cbs0_Que_t, head, tail int32) {
	if head < 0 || tail < head || int(tail) > len(p.pData) {
		panic("gia: invalid CBS0 queue restoration")
	}
	p.iHead = head
	p.iTail = tail
}
func Cbs0_VarFaninFanoutMax(p *Cbs0_Man_t, o *Gia_Obj_t) int32 {
	return max(Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(o)), Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(o)))
}
func Cbs0_ManDecideHighest(p *Cbs0_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, o := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if best == nil || best.id < o.id {
			best = o
		}
	}
	return best
}
func Cbs0_ManDecideLowest(p *Cbs0_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, o := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if best == nil || best.id > o.id {
			best = o
		}
	}
	return best
}
func Cbs0_ManDecideMaxFF(p *Cbs0_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	count := int32(0)
	for _, o := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if current := Cbs0_VarFaninFanoutMax(p, o); current > count {
			count = current
			best = o
		}
	}
	return best
}
func Cbs0_ManCancelUntil(p *Cbs0_Man_t, bound int32) {
	if bound < 0 || bound > p.pProp.iTail {
		panic("gia: invalid CBS0 trail bound")
	}
	for _, o := range p.pProp.pData[bound:p.pProp.iTail] {
		Cbs0_VarUnassign(o)
	}
	p.pProp.iHead = bound
	p.pProp.iTail = bound
}

// assign returns a contradiction for shared/complemented fanins rather than
// trying to assign the same regular object twice (a source precondition).
func cbs0Assign(p *Cbs0_Man_t, edge *Gia_Obj_t) bool {
	o := Gia_Regular(edge)
	value := Gia_IsComplement(edge) ^ 1
	if o.id == 0 {
		return value != 0
	}
	if Cbs0_VarIsAssigned(o) != 0 {
		return Cbs0_VarValue(o) != value
	}
	Cbs0_VarAssign(o)
	Cbs0_VarSetValue(o, value)
	Cbs0_QuePush(&p.pProp, o)
	return false
}
func Cbs0_ManAssign(p *Cbs0_Man_t, edge *Gia_Obj_t) {
	if cbs0Assign(p, edge) {
		panic("gia: inconsistent CBS0 assignment")
	}
}
func Cbs0_ManPropagateOne(p *Cbs0_Man_t, o *Gia_Obj_t) int32 {
	if Gia_ObjIsCi(o) != 0 {
		return 0
	}
	a, b := Cbs0_VarFanin0Value(o), Cbs0_VarFanin1Value(o)
	if Cbs0_VarValue(o) != 0 {
		if a == 0 || b == 0 {
			return 1
		}
		if a == 2 && cbs0Assign(p, Gia_ObjChild0(o)) {
			return 1
		}
		if b == 2 && cbs0Assign(p, Gia_ObjChild1(o)) {
			return 1
		}
		return 0
	}
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 && b == 1 {
		return 1
	}
	if a == 1 || b == 1 {
		if a == 2 && cbs0Assign(p, Gia_Not(Gia_ObjChild0(o))) {
			return 1
		}
		if b == 2 && cbs0Assign(p, Gia_Not(Gia_ObjChild1(o))) {
			return 1
		}
		return 0
	}
	if Cbs0_QueHasNode(&p.pJust, o) == 0 {
		Cbs0_QuePush(&p.pJust, o)
	}
	return 0
}
func Cbs0_ManPropagateTwo(p *Cbs0_Man_t, o *Gia_Obj_t) int32 {
	a, b := Cbs0_VarFanin0Value(o), Cbs0_VarFanin1Value(o)
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 && b == 1 {
		return 1
	}
	if a == 2 && cbs0Assign(p, Gia_Not(Gia_ObjChild0(o))) {
		return 1
	}
	if b == 2 && cbs0Assign(p, Gia_Not(Gia_ObjChild1(o))) {
		return 1
	}
	return 0
}
func Cbs0_ManPropagate(p *Cbs0_Man_t) int32 {
	for {
		for i := p.pProp.iHead; i < p.pProp.iTail; i++ {
			if Cbs0_ManPropagateOne(p, p.pProp.pData[i]) != 0 {
				return 1
			}
		}
		p.pProp.iHead = p.pProp.iTail
		k := p.pJust.iHead
		for i := p.pJust.iHead; i < p.pJust.iTail; i++ {
			o := p.pJust.pData[i]
			if Cbs0_VarIsJust(o) != 0 {
				p.pJust.pData[k] = o
				k++
			} else if Cbs0_ManPropagateTwo(p, o) != 0 {
				return 1
			}
		}
		if k == p.pJust.iTail {
			break
		}
		p.pJust.iTail = k
	}
	return 0
}
func Cbs0_ManSolve_rec(p *Cbs0_Man_t) int32 {
	if Cbs0_ManPropagate(p) != 0 {
		return 1
	}
	if Cbs0_QueIsEmpty(&p.pJust) != 0 {
		return 0
	}
	p.Pars.nJustThis = max(p.Pars.nJustThis, p.pJust.iTail-p.pJust.iHead)
	if Cbs0_ManCheckLimits(p) != 0 {
		return 0
	}
	bound := p.pProp.iHead
	var head, tail int32
	Cbs0_QueStore(&p.pJust, &head, &tail)
	var gate *Gia_Obj_t
	switch {
	case p.Pars.fUseHighest != 0:
		gate = Cbs0_ManDecideHighest(p)
	case p.Pars.fUseLowest != 0:
		gate = Cbs0_ManDecideLowest(p)
	case p.Pars.fUseMaxFF != 0:
		gate = Cbs0_ManDecideMaxFF(p)
	default:
		panic("gia: no CBS0 decision heuristic")
	}
	decision := Gia_Not(Gia_ObjChild1(gate))
	if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(gate)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(gate)) {
		decision = Gia_Not(Gia_ObjChild0(gate))
	}
	Cbs0_ManAssign(p, decision)
	if Cbs0_ManSolve_rec(p) == 0 {
		return 0
	}
	Cbs0_ManCancelUntil(p, bound)
	Cbs0_QueRestore(&p.pJust, head, tail)
	Cbs0_ManAssign(p, Gia_Not(decision))
	if Cbs0_ManSolve_rec(p) == 0 {
		return 0
	}
	p.Pars.nBTThis++
	return 1
}
func Cbs0_ManSolve(p *Cbs0_Man_t, root *Gia_Obj_t) int32 {
	if p.pAig == nil {
		panic("gia: CBS0 graph not set")
	}
	cbs0CheckGraph(p.pAig)
	if p.pProp.iTail != 0 || p.pJust.iTail != 0 {
		panic("gia: CBS0 solver already active")
	}
	Vec_IntClear(p.vModel)
	p.Pars.nBTThis = 0
	p.Pars.nJustThis = 0
	regular := Gia_Regular(root)
	if regular == nil || regular.id < 0 || int(regular.id) >= len(p.pAig.pObjs) || p.pAig.pObjs[regular.id] != regular {
		panic("gia: foreign CBS0 root")
	}
	if regular.id != 0 && Gia_ObjIsAnd(regular) == 0 && Gia_ObjIsCi(regular) == 0 {
		panic("gia: invalid CBS0 root type")
	}
	status := int32(1)
	if regular.id == 0 {
		status = Gia_IsComplement(root) ^ 1
	} else {
		if cbs0Assign(p, root) {
			status = 1
		} else {
			status = Cbs0_ManSolve_rec(p)
		}
		if status == 0 && Cbs0_ManCheckLimits(p) == 0 {
			Cbs0_ManSaveModel(p, p.vModel)
		}
		Cbs0_ManCancelUntil(p, 0)
	}
	p.pJust.iHead = 0
	p.pJust.iTail = 0
	p.Pars.nBTTotal += p.Pars.nBTThis
	p.Pars.nJustTotal = max(p.Pars.nJustTotal, p.Pars.nJustThis)
	if Cbs0_ManCheckLimits(p) != 0 {
		return -1
	}
	return status
}
func Cbs0_ManSatPrintStats(p *Cbs0_Man_t) {
	fmt.Printf("CBS0 CO=%d AND=%d Conf=%d JustMax=%d Unsat=%d Sat=%d Undef=%d TotalTime=%s\n", Gia_ManCoNum(p.pAig), Gia_ManAndNum(p.pAig), p.Pars.nBTLimit, p.Pars.nJustLimit, p.nSatUnsat, p.nSatSat, p.nSatUndec, p.timeTotal)
}
func Cbs_ManSolveMiter(g *Gia_Man_t, conflicts int32, statusOut **Vec_Str_t, verbose int32) *Vec_Int_t {
	if g.nRegs != 0 {
		panic("gia: CBS0 miter must be combinational")
	}
	start := time.Now()
	p := Cbs0_ManAlloc(g)
	defer Cbs0_ManStop(p)
	p.Pars.nBTLimit = conflicts
	statuses := Vec_StrAlloc(int(Gia_ManPoNum(g)))
	store := Vec_IntAlloc(0)
	for i, id := range Vec_IntArray(g.vCos) {
		began := time.Now()
		status := Cbs0_ManSolve(p, Gia_ObjChild0(g.pObjs[id]))
		Vec_StrPush(statuses, byte(status))
		if status == 1 {
			p.nSatUnsat++
			p.nConfUnsat += p.Pars.nBTThis
			p.timeSatUnsat += time.Since(began)
			continue
		}
		Vec_IntPush(store, int32(i))
		if status < 0 {
			Vec_IntPush(store, -1)
			p.nSatUndec++
			p.nConfUndec += p.Pars.nBTThis
			p.timeSatUndec += time.Since(began)
		} else {
			Vec_IntPush(store, int32(Vec_IntSize(p.vModel)))
			for _, lit := range Vec_IntArray(p.vModel) {
				Vec_IntPush(store, lit)
			}
			p.nSatSat++
			p.nConfSat += p.Pars.nBTThis
			p.timeSatSat += time.Since(began)
		}
	}
	p.nSatTotal = Gia_ManPoNum(g)
	p.timeTotal = time.Since(start)
	if verbose != 0 {
		Cbs0_ManSatPrintStats(p)
	}
	if statusOut != nil {
		*statusOut = statuses
	}
	return store
}
