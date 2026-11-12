package gia

import (
	"fmt"
	"time"
)

type Cbs2_Par_t struct {
	nBTLimit, nJustLimit, nBTThis, nBTThisNc, nJustThis, nBTTotal, nJustTotal int32
	fUseHighest, fUseLowest, fUseMaxFF, fUseFanout, fVerbose                  int32
}
type Cbs2_Que_t struct {
	iHead, iTail int32
	pData        []int32
}
type Cbs2_Man_t struct {
	Pars                                                                                Cbs2_Par_t
	pAig                                                                                *Gia_Man_t
	pProp, pJust, pClauses                                                              Cbs2_Que_t
	vModel                                                                              *Vec_Int_t
	vAssign, vMark                                                                      []byte
	vLevReason, vWatches, vWatchUpds, vFanoutN, vFanout0, vActivity, vActStore, vJStore []int32
	nSatUnsat, nSatSat, nSatUndec, nSatTotal, nConfUnsat, nConfSat, nConfUndec          int32
	timeJFront, timeSatUnsat, timeSatSat, timeSatUndec, timeTotal                       time.Duration
	nPropCalls                                                                          [3]int32
	nFails                                                                              [2]int32
	nClauseConf                                                                         int32
	fanoutsReady                                                                        bool
}

func Cbs2_VarUnused(p *Cbs2_Man_t, id int32) int32     { return giaBool(p.vLevReason[3*id] == -1) }
func Cbs2_VarSetUnused(p *Cbs2_Man_t, id int32)        { p.vLevReason[3*id] = -1 }
func Cbs2_VarMark0(p *Cbs2_Man_t, id int32) int32      { return int32(p.vMark[id]) }
func Cbs2_VarSetMark0(p *Cbs2_Man_t, id, value int32)  { p.vMark[id] = byte(value) }
func Cbs2_VarIsAssigned(p *Cbs2_Man_t, id int32) int32 { return giaBool(p.vAssign[id] < 2) }
func Cbs2_VarUnassign(p *Cbs2_Man_t, id int32) {
	if p.vAssign[id] >= 2 {
		panic("gia: CBS2 unassigns unknown variable")
	}
	p.vAssign[id] += 2
	Cbs2_VarSetUnused(p, id)
}
func Cbs2_VarValue(p *Cbs2_Man_t, id int32) int32 { return int32(p.vAssign[id]) }
func Cbs2_VarSetValue(p *Cbs2_Man_t, id, value int32) {
	if p.vAssign[id] < 2 {
		panic("gia: CBS2 double assignment")
	}
	p.vAssign[id] = byte(value)
}
func Cbs2_VarIsJust(p *Cbs2_Man_t, obj *Gia_Obj_t, id int32) int32 {
	return giaBool(Gia_ObjIsAnd(obj) != 0 && p.vAssign[Gia_ObjFaninId0(obj, id)] >= 2 && p.vAssign[Gia_ObjFaninId1(obj, id)] >= 2)
}
func Cbs2_VarFanin0Value(p *Cbs2_Man_t, obj *Gia_Obj_t, id int32) int32 {
	return Cbs2_VarValue(p, Gia_ObjFaninId0(obj, id)) ^ Gia_ObjFaninC0(obj)
}
func Cbs2_VarFanin1Value(p *Cbs2_Man_t, obj *Gia_Obj_t, id int32) int32 {
	return Cbs2_VarValue(p, Gia_ObjFaninId1(obj, id)) ^ Gia_ObjFaninC1(obj)
}
func Cbs2_VarDecLevel(p *Cbs2_Man_t, id int32) int32      { return p.vLevReason[3*id] }
func Cbs2_VarReason0(p *Cbs2_Man_t, id int32) int32       { return p.vLevReason[3*id+1] }
func Cbs2_VarReason1(p *Cbs2_Man_t, id int32) int32       { return p.vLevReason[3*id+2] }
func Cbs2_VarReasonP(p *Cbs2_Man_t, id int32) []int32     { return p.vLevReason[3*id+1 : 3*id+3 : 3*id+3] }
func Cbs2_ClauseSize(p *Cbs2_Man_t, clause int32) int32   { return p.pClauses.pData[clause] }
func Cbs2_ClauseLits(p *Cbs2_Man_t, clause int32) []int32 { return p.pClauses.pData[clause+1:] }
func Cbs2_ClauseLit(p *Cbs2_Man_t, clause, index int32) int32 {
	return p.pClauses.pData[clause+1+index]
}
func Cbs2_ClauseNext1p(p *Cbs2_Man_t, clause int32) *int32 {
	return &p.pClauses.pData[clause+Cbs2_ClauseSize(p, clause)+2]
}
func Cbs2_ClauseSetSize(p *Cbs2_Man_t, clause, size int32) { p.pClauses.pData[clause] = size }
func Cbs2_ClauseSetLit(p *Cbs2_Man_t, clause, index, value int32) {
	p.pClauses.pData[clause+1+index] = value
}
func Cbs2_ClauseSetNext(p *Cbs2_Man_t, clause, index, value int32) {
	p.pClauses.pData[clause+Cbs2_ClauseSize(p, clause)+1+index] = value
}
func Cbs2_SetDefaultParams(pars *Cbs2_Par_t) {
	*pars = Cbs2_Par_t{nBTLimit: 1000, nJustLimit: 500, fUseHighest: 1, fUseFanout: 1, fVerbose: 1}
}
func Cbs2_ManSetConflictNum(p *Cbs2_Man_t, limit int32) { p.Pars.nBTLimit = limit }
func Cbs2_ManAlloc(graph *Gia_Man_t) *Cbs2_Man_t {
	for _, obj := range graph.pObjs {
		if Gia_ObjIsXor(obj) != 0 || Gia_ObjIsMux(graph, obj) != 0 {
			panic("gia: CBS2 expects ordinary AND graph")
		}
	}
	if graph.Refs == nil {
		Gia_ManCreateRefs(graph)
	}
	p := &Cbs2_Man_t{pAig: graph, vModel: Vec_IntAlloc(1000), vAssign: make([]byte, graph.nObjs), vMark: make([]byte, graph.nObjs), vLevReason: make([]int32, 3*graph.nObjs), vWatches: make([]int32, 2*graph.nObjs), vFanout0: make([]int32, graph.nObjs), vFanoutN: make([]int32, 2*graph.nObjs), vActivity: make([]int32, graph.nObjs)}
	for i := range p.vAssign {
		p.vAssign[i] = 2
	}
	for i := range p.vLevReason {
		p.vLevReason[i] = -1
	}
	p.pClauses = Cbs2_Que_t{iHead: 1, iTail: 1, pData: make([]int32, 1, 10000)}
	Cbs2_SetDefaultParams(&p.Pars)
	return p
}
func Cbs2_ManStop(p *Cbs2_Man_t) {
	if p != nil {
		*p = Cbs2_Man_t{}
	}
}
func Cbs2_ReadModel(p *Cbs2_Man_t) *Vec_Int_t { return p.vModel }
func Cbs2_ManCheckLimits(p *Cbs2_Man_t) int32 {
	a, b := p.Pars.nJustThis > p.Pars.nJustLimit, p.Pars.nBTThis > p.Pars.nBTLimit
	p.nFails[0] += giaBool(a)
	p.nFails[1] += giaBool(b)
	return giaBool(a || b)
}
func Cbs2_ManSaveModel(p *Cbs2_Man_t, model *Vec_Int_t) {
	Vec_IntClear(model)
	p.pProp.iHead = 0
	for _, lit := range p.pProp.pData[:p.pProp.iTail] {
		obj := p.pAig.pObjs[lit>>1]
		if Gia_ObjIsCi(obj) != 0 {
			Vec_IntPush(model, Gia_ObjCioId(obj)<<1|(lit&1))
		}
	}
}
func Cbs2_ManSaveModelAll(p *Cbs2_Man_t, model *Vec_Int_t) {
	Vec_IntClear(model)
	p.pProp.iHead = 0
	for _, lit := range p.pProp.pData[:p.pProp.iTail] {
		Vec_IntPush(model, lit)
	}
}
func Cbs2_QueIsEmpty(q *Cbs2_Que_t) int32 { return giaBool(q.iHead == q.iTail) }
func Cbs2_QuePush(q *Cbs2_Que_t, value int32) {
	if int(q.iTail) == len(q.pData) {
		q.pData = append(q.pData, value)
	} else {
		q.pData[q.iTail] = value
	}
	q.iTail++
}
func Cbs2_QueGrow(q *Cbs2_Que_t, extra int32) {
	if int(q.iTail+extra) > cap(q.pData) {
		data := make([]int32, len(q.pData), max(2*cap(q.pData), int(q.iTail+extra)))
		copy(data, q.pData)
		q.pData = data
	}
}
func Cbs2_QueHasNode(q *Cbs2_Que_t, id int32) int32 {
	for _, value := range q.pData[q.iHead:q.iTail] {
		if value == id {
			return 1
		}
	}
	return 0
}
func Cbs2_QueStore(q *Cbs2_Que_t, head, tail *int32) {
	*head, *tail = q.iHead, q.iTail
	for i := *head; i < *tail; i++ {
		Cbs2_QuePush(q, q.pData[i])
	}
	q.iHead = *tail
}
func Cbs2_QueRestore(q *Cbs2_Que_t, head, tail int32) { q.iHead, q.iTail = head, tail }
func Cbs2_VarFaninFanoutMax(p *Cbs2_Man_t, obj *Gia_Obj_t) int32 {
	return max(Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(obj)), Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(obj)))
}
func Cbs2_ManDecideHighest(p *Cbs2_Man_t) int32 {
	var best int32
	for _, id := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		best = max(best, id)
	}
	return best
}
func Cbs2_ManDecideLowest(p *Cbs2_Man_t) *Gia_Obj_t {
	var best int32
	for _, id := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if best == 0 || id < best {
			best = id
		}
	}
	return p.pAig.pObjs[best]
}
func Cbs2_ManDecideMaxFF(p *Cbs2_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	var count int32
	for _, id := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		obj := p.pAig.pObjs[id]
		if score := Cbs2_VarFaninFanoutMax(p, obj); score > count {
			best, count = obj, score
		}
	}
	return best
}
func Cbs2_ManCancelUntil(p *Cbs2_Man_t, bound int32) {
	for _, lit := range p.pProp.pData[bound:p.pProp.iTail] {
		Cbs2_VarUnassign(p, lit>>1)
	}
	p.pProp.iHead, p.pProp.iTail = bound, bound
}
func Cbs2_ManAssign(p *Cbs2_Man_t, lit, level, a, b int32) {
	id := lit >> 1
	Cbs2_VarSetValue(p, id, (lit&1)^1)
	Cbs2_QuePush(&p.pProp, lit)
	p.vLevReason[3*id], p.vLevReason[3*id+1], p.vLevReason[3*id+2] = level, a, b
}
func Cbs2_ManPrintClause(p *Cbs2_Man_t, level, clause int32) {
	fmt.Printf("Level %d clause: %v\n", level, Cbs2_ClauseLits(p, clause)[:Cbs2_ClauseSize(p, clause)])
}
func Cbs2_ManPrintCube(p *Cbs2_Man_t, level, clause int32) { Cbs2_ManPrintClause(p, level, clause) }

// The source activity bump loop is intentionally disabled; retain its API.
func Cbs2_ManBumpClause(p *Cbs2_Man_t, clause int32) {}
func Cbs2_ManBumpClean(p *Cbs2_Man_t) {
	for _, id := range p.vActStore {
		p.vActivity[id] = 0
	}
}
func Cbs2_ManCleanWatch(p *Cbs2_Man_t) {
	for _, lit := range p.vWatchUpds {
		p.vWatches[lit] = 0
	}
	p.vWatchUpds = p.vWatchUpds[:0]
}
func Cbs2_ManWatchClause(p *Cbs2_Man_t, clause, lit int32) {
	index := lit ^ 1
	if p.vWatches[index] == 0 {
		p.vWatchUpds = append(p.vWatchUpds, index)
	}
	slot := giaBool(lit == Cbs2_ClauseLit(p, clause, 1))
	Cbs2_ClauseSetNext(p, clause, slot, p.vWatches[index])
	p.vWatches[index] = clause
}
func Cbs2_QueFinish(p *Cbs2_Man_t, level int32) int32 {
	q := &p.pClauses
	clause := q.iHead
	size := q.iTail - clause - 1
	if size < 1 {
		panic("gia: empty CBS2 reason")
	}
	q.pData[clause] = size
	copyHead := q.iTail
	q.iHead = q.iTail
	if size == 1 {
		return clause
	}
	// Cube and learned clause have distinct records; watch links follow literals.
	for i := int32(0); i <= size; i++ {
		value := q.pData[clause+i]
		if i != 0 {
			value = value<<1 | Cbs2_VarValue(p, value)
		}
		Cbs2_QuePush(q, value)
	}
	Cbs2_QuePush(q, 0)
	Cbs2_QuePush(q, 0)
	q.iHead = q.iTail
	Cbs2_ManWatchClause(p, copyHead, Cbs2_ClauseLit(p, copyHead, 0))
	Cbs2_ManWatchClause(p, copyHead, Cbs2_ClauseLit(p, copyHead, 1))
	return clause
}
func Cbs2_ManDeriveReason(p *Cbs2_Man_t, level int32) int32 {
	q := &p.pClauses
	write := q.iHead + 2
	var visited []int32
	for read := write; read < q.iTail; read++ {
		id := q.pData[read]
		if p.vMark[id] != 0 {
			continue
		}
		p.vMark[id] = 1
		visited = append(visited, id)
		nodeLevel := Cbs2_VarDecLevel(p, id)
		if nodeLevel < level {
			q.pData[write] = id
			write++
			continue
		}
		if nodeLevel != level {
			panic("gia: CBS2 reason above decision level")
		}
		a, b := Cbs2_VarReason0(p, id), Cbs2_VarReason1(p, id)
		switch {
		case a == 0 && b == 0:
			q.pData[q.iHead+1] = id
		case a != 0:
			Cbs2_QuePush(q, a)
			if b != 0 {
				Cbs2_QuePush(q, b)
			}
		default:
			lits := append([]int32(nil), Cbs2_ClauseLits(p, b)[:Cbs2_ClauseSize(p, b)]...)
			for _, lit := range lits[1:] {
				Cbs2_QuePush(q, lit>>1)
			}
		}
	}
	if q.pData[q.iHead+1] == 0 {
		panic("gia: CBS2 reason has no decision")
	}
	q.iTail = write
	for _, id := range visited {
		p.vMark[id] = 0
	}
	return Cbs2_QueFinish(p, level)
}
func Cbs2_ManAnalyze(p *Cbs2_Man_t, level, id, a, b int32) int32 {
	q := &p.pClauses
	Cbs2_QuePush(q, 0)
	Cbs2_QuePush(q, 0)
	if a != 0 {
		Cbs2_QuePush(q, id)
		Cbs2_QuePush(q, a)
		if b != 0 {
			Cbs2_QuePush(q, b)
		}
	} else {
		lits := append([]int32(nil), Cbs2_ClauseLits(p, b)[:Cbs2_ClauseSize(p, b)]...)
		for _, lit := range lits {
			Cbs2_QuePush(q, lit>>1)
		}
	}
	return Cbs2_ManDeriveReason(p, level)
}
func Cbs2_ManPropagateClauses(p *Cbs2_Man_t, level, lit int32) int32 {
	falseLit := lit ^ 1
	// A link is either a literal's watch head or the second next field of a
	// prior clause. Reacquire by index after conflict analysis may grow storage.
	prior := int32(-1)
	get := func() int32 {
		if prior < 0 {
			return p.vWatches[lit]
		}
		return *Cbs2_ClauseNext1p(p, prior)
	}
	set := func(value int32) {
		if prior < 0 {
			p.vWatches[lit] = value
		} else {
			*Cbs2_ClauseNext1p(p, prior) = value
		}
	}
	for clause := get(); clause != 0; clause = get() {
		size := Cbs2_ClauseSize(p, clause)
		lits := Cbs2_ClauseLits(p, clause)
		p.nPropCalls[1]++
		if lits[0] == falseLit {
			lits[0], lits[1] = lits[1], lits[0]
			lits[size], lits[size+1] = lits[size+1], lits[size]
		}
		if lits[1] != falseLit {
			panic("gia: corrupt CBS2 watch")
		}
		if Cbs2_VarValue(p, lits[0]>>1) == ((lits[0] & 1) ^ 1) {
			prior = clause
			continue
		}
		moved := false
		for i := int32(2); i < size; i++ {
			if Cbs2_VarValue(p, lits[i]>>1) == lits[i]&1 {
				continue
			}
			lits[1], lits[i] = lits[i], falseLit
			set(*Cbs2_ClauseNext1p(p, clause))
			Cbs2_ManWatchClause(p, clause, lits[1])
			moved = true
			break
		}
		if moved {
			continue
		}
		value := Cbs2_VarValue(p, lits[0]>>1)
		if value >= 2 {
			Cbs2_ManAssign(p, lits[0], level, 0, clause)
			prior = clause
			continue
		}
		if value == lits[0]&1 {
			p.nClauseConf++
			return Cbs2_ManAnalyze(p, level, lits[0]>>1, 0, clause)
		}
	}
	return 0
}
func Cbs2_ManResolve(p *Cbs2_Man_t, level, a, b int32) int32 {
	q := &p.pClauses
	Cbs2_QuePush(q, 0)
	Cbs2_QuePush(q, 0)
	maximum := int32(-1)
	for _, clause := range []int32{a, b} {
		lits := append([]int32(nil), Cbs2_ClauseLits(p, clause)[:Cbs2_ClauseSize(p, clause)]...)
		for _, id := range lits[1:] {
			if p.vMark[id] != 0 {
				continue
			}
			p.vMark[id] = 1
			Cbs2_QuePush(q, id)
			maximum = max(maximum, Cbs2_VarDecLevel(p, id))
		}
	}
	for i := q.iHead + 2; i < q.iTail; i++ {
		p.vMark[q.pData[i]] = 0
	}
	return Cbs2_ManDeriveReason(p, maximum)
}

func Cbs2_ManPropagateOne(p *Cbs2_Man_t, id, level int32) int32 {
	obj := p.pAig.pObjs[id]
	if Gia_ObjIsCi(obj) != 0 || id == 0 {
		return 0
	}
	p.nPropCalls[0]++
	a, b := Cbs2_VarFanin0Value(p, obj, id), Cbs2_VarFanin1Value(p, obj, id)
	fa, fb := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	if Cbs2_VarValue(p, id) == 1 {
		if a == 0 || b == 0 {
			if a == 0 && b != 0 {
				return Cbs2_ManAnalyze(p, level, id, fa, 0)
			}
			if a != 0 && b == 0 {
				return Cbs2_ManAnalyze(p, level, id, fb, 0)
			}
			return Cbs2_ManAnalyze(p, level, id, fa, fb)
		}
		if a >= 2 {
			Cbs2_ManAssign(p, Gia_ObjFaninLit0(obj, id), level, id, 0)
		}
		if b >= 2 {
			Cbs2_ManAssign(p, Gia_ObjFaninLit1(obj, id), level, id, 0)
		}
		return 0
	}
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 && b == 1 {
		return Cbs2_ManAnalyze(p, level, id, fa, fb)
	}
	if a == 1 || b == 1 {
		return Cbs2_ManPropagateTwo(p, id, level)
	}
	if p.Pars.fUseFanout == 0 {
		Cbs2_QuePush(&p.pJust, id)
	}
	return 0
}
func Cbs2_ManPropagateTwo(p *Cbs2_Man_t, id, level int32) int32 {
	obj := p.pAig.pObjs[id]
	p.nPropCalls[1]++
	a, b := Cbs2_VarFanin0Value(p, obj, id), Cbs2_VarFanin1Value(p, obj, id)
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 && b == 1 {
		return Cbs2_ManAnalyze(p, level, id, Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id))
	}
	if a >= 2 {
		Cbs2_ManAssign(p, Gia_ObjFaninLit0(obj, id)^1, level, id, Gia_ObjFaninId1(obj, id))
	}
	if b >= 2 {
		Cbs2_ManAssign(p, Gia_ObjFaninLit1(obj, id)^1, level, id, Gia_ObjFaninId0(obj, id))
	}
	return 0
}
func Cbs2_ManPropagateUnassigned(p *Cbs2_Man_t, id, level int32) {
	obj := p.pAig.pObjs[id]
	a, b := Cbs2_VarFanin0Value(p, obj, id), Cbs2_VarFanin1Value(p, obj, id)
	fa, fb := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	p.nPropCalls[2]++
	switch {
	case a == 0 && b == 0:
		Cbs2_ManAssign(p, id<<1|1, level, fa, fb)
	case a == 0:
		Cbs2_ManAssign(p, id<<1|1, level, fa, 0)
	case b == 0:
		Cbs2_ManAssign(p, id<<1|1, level, fb, 0)
	case a == 1 && b == 1:
		Cbs2_ManAssign(p, id<<1, level, fa, fb)
	}
}
func Cbs2_ManPropagate(p *Cbs2_Man_t, level int32) int32 {
	for {
		for i := p.pProp.iHead; i < p.pProp.iTail; i++ {
			if clause := Cbs2_ManPropagateOne(p, p.pProp.pData[i]>>1, level); clause != 0 {
				return clause
			}
		}
		p.pProp.iHead = p.pProp.iTail
		write := p.pJust.iHead
		for i := p.pJust.iHead; i < p.pJust.iTail; i++ {
			id := p.pJust.pData[i]
			if Cbs2_VarIsJust(p, p.pAig.pObjs[id], id) != 0 {
				p.pJust.pData[write] = id
				write++
			} else if clause := Cbs2_ManPropagateTwo(p, id, level); clause != 0 {
				return clause
			}
		}
		if write == p.pJust.iTail {
			break
		}
		p.pJust.iTail = write
	}
	return 0
}
func Cbs2_ManPropagate2(p *Cbs2_Man_t, level int32) int32 {
	for i := p.pProp.iHead; i < p.pProp.iTail; i++ {
		lit := p.pProp.pData[i]
		if clause := Cbs2_ManPropagateClauses(p, level, lit); clause != 0 {
			return clause
		}
		for fan := p.vFanout0[lit>>1]; fan != 0; fan = p.vFanoutN[fan] {
			id := fan >> 1
			if Cbs2_VarIsAssigned(p, id) == 0 {
				Cbs2_ManPropagateUnassigned(p, id, level)
			} else if clause := Cbs2_ManPropagateOne(p, id, level); clause != 0 {
				return clause
			}
		}
		if clause := Cbs2_ManPropagateOne(p, lit>>1, level); clause != 0 {
			return clause
		}
	}
	p.pProp.iHead = p.pProp.iTail
	return 0
}
func cbs2UpdateDecision(p *Cbs2_Man_t, id, decision int32, score func(int32) int32) int32 {
	obj := p.pAig.pObjs[id]
	for _, lit := range []int32{Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)} {
		if decision == -1 || score(decision>>1) < score(lit>>1) {
			decision = lit ^ 1
		}
	}
	return decision
}
func Cbs2_ManUpdateDecVar2(p *Cbs2_Man_t, id, decision int32) int32 {
	return cbs2UpdateDecision(p, id, decision, func(id int32) int32 { return Gia_ObjLevelId(p.pAig, id) })
}
func Cbs2_ManUpdateDecVar3(p *Cbs2_Man_t, id, decision int32) int32 {
	return cbs2UpdateDecision(p, id, decision, func(id int32) int32 { return p.vActivity[id] })
}
func Cbs2_ManUpdateDecVar(p *Cbs2_Man_t, id, decision int32) int32 {
	return cbs2UpdateDecision(p, id, decision, func(id int32) int32 { return Gia_ObjRefNumId(p.pAig, id) })
}
func Cbs2_ManUpdateFrontier(p *Cbs2_Man_t, oldHead int32, decision *int32) int32 {
	started := time.Now()
	*decision = -1
	oldTail := p.pJust.iTail
	for i := p.pJust.iHead; i < oldTail; i++ {
		id := p.pJust.pData[i]
		if Cbs2_VarIsJust(p, p.pAig.pObjs[id], id) != 0 {
			Cbs2_QuePush(&p.pJust, id)
		}
	}
	for i := oldHead; i < p.pProp.iTail; i++ {
		id := p.pProp.pData[i] >> 1
		if Cbs2_VarIsJust(p, p.pAig.pObjs[id], id) != 0 {
			Cbs2_QuePush(&p.pJust, id)
		}
	}
	p.pProp.iHead = p.pProp.iTail
	p.pJust.iHead = oldTail
	p.timeJFront += time.Since(started)
	return Cbs2_QueIsEmpty(&p.pJust)
}
func cbs2SolveRecursive(p *Cbs2_Man_t, level int32, fanout bool) int32 {
	oldHead := p.pProp.iHead
	if fanout {
		if clause := Cbs2_ManPropagate2(p, level); clause != 0 {
			return clause
		}
		var decision int32
		if Cbs2_ManUpdateFrontier(p, oldHead, &decision) != 0 {
			return 0
		}
	} else {
		if clause := Cbs2_ManPropagate(p, level); clause != 0 {
			return clause
		}
		if Cbs2_QueIsEmpty(&p.pJust) != 0 {
			return 0
		}
	}
	p.Pars.nJustThis = max(p.Pars.nJustThis, p.pJust.iTail-p.pJust.iHead)
	if Cbs2_ManCheckLimits(p) != 0 {
		return 0
	}
	bound := p.pProp.iHead
	head, tail := p.pJust.iHead, p.pJust.iTail
	if !fanout {
		Cbs2_QueStore(&p.pJust, &head, &tail)
	}
	id := Cbs2_ManDecideHighest(p)
	obj := p.pAig.pObjs[id]
	var decision int32
	if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(obj)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(obj)) {
		decision = Gia_ObjFaninLit0(obj, id) ^ 1
	} else {
		decision = Gia_ObjFaninLit1(obj, id) ^ 1
	}
	Cbs2_ManAssign(p, decision, level+1, 0, 0)
	first := cbs2SolveRecursive(p, level+1, fanout)
	if first == 0 {
		return 0
	}
	if Cbs2_ClauseLit(p, first, 0) != decision>>1 {
		return first
	}
	Cbs2_ManCancelUntil(p, bound)
	Cbs2_QueRestore(&p.pJust, head, tail)
	Cbs2_ManAssign(p, decision^1, level+1, 0, 0)
	second := cbs2SolveRecursive(p, level+1, fanout)
	if second == 0 {
		return 0
	}
	if Cbs2_ClauseLit(p, second, 0) != decision>>1 {
		return second
	}
	clause := Cbs2_ManResolve(p, level, first, second)
	Cbs2_ManBumpClause(p, clause)
	p.Pars.nBTThis++
	return clause
}
func Cbs2_ManSolve1_rec(p *Cbs2_Man_t, level int32) int32 { return cbs2SolveRecursive(p, level, false) }
func Cbs2_ManSolve2_rec(p *Cbs2_Man_t, level int32) int32 { return cbs2SolveRecursive(p, level, true) }
func Cbs2_ManSolve_rec(p *Cbs2_Man_t, level int32) int32 {
	return cbs2SolveRecursive(p, level, p.Pars.fUseFanout != 0)
}
func cbs2Solve(p *Cbs2_Man_t, first, second int32, pair bool) int32 {
	Vec_IntClear(p.vModel)
	p.Pars.nBTThis, p.Pars.nJustThis, p.Pars.nBTThisNc = 0, 0, 0
	roots := []int32{first}
	if pair && second != 0 {
		roots = append(roots, second)
	}
	status := int32(0)
	created := p.Pars.fUseFanout != 0 && !p.fanoutsReady
	if created {
		Gia_ManIncrementTravId(p.pAig)
		for _, lit := range roots {
			Cbs2_ManCreateFanout_rec(p, lit>>1)
		}
	}
	for _, lit := range roots {
		id := lit >> 1
		if id == 0 {
			if lit == 0 {
				status = 1
				break
			}
			continue
		}
		if Cbs2_VarIsAssigned(p, id) != 0 {
			if Cbs2_VarValue(p, id) != (lit&1)^1 {
				status = 1
				break
			}
			continue
		}
		Cbs2_ManAssign(p, lit, 0, 0, 0)
	}
	if status == 0 && p.pProp.iTail != 0 {
		if Cbs2_ManSolve_rec(p, 0) != 0 || Cbs2_ManCheckLimits(p) != 0 {
			status = 1
		} else if pair {
			Cbs2_ManSaveModelAll(p, p.vModel)
		} else {
			Cbs2_ManSaveModel(p, p.vModel)
		}
	}
	Cbs2_ManCancelUntil(p, 0)
	Cbs2_ManCleanWatch(p)
	Cbs2_ManBumpClean(p)
	p.pJust.iHead, p.pJust.iTail = 0, 0
	p.pClauses.iHead, p.pClauses.iTail = 1, 1
	if created {
		for _, lit := range roots {
			Cbs2_ManDeleteFanout_rec(p, lit>>1)
		}
	}
	p.Pars.nBTTotal += p.Pars.nBTThis
	p.Pars.nJustTotal = max(p.Pars.nJustTotal, p.Pars.nJustThis)
	if Cbs2_ManCheckLimits(p) != 0 {
		return -1
	}
	return status
}
func Cbs2_ManSolve(p *Cbs2_Man_t, lit int32) int32 { return cbs2Solve(p, lit, 0, false) }
func Cbs2_ManSolve2(p *Cbs2_Man_t, first, second int32) int32 {
	return cbs2Solve(p, first, second, true)
}
func Cbs2_ManSatPrintStats(p *Cbs2_Man_t) {
	fmt.Printf("CO=%d AND=%d Conf=%d JustMax=%d Unsat=%d Sat=%d Undef=%d TotalTime=%s\n", Gia_ManCoNum(p.pAig), Gia_ManAndNum(p.pAig), p.Pars.nBTLimit, p.Pars.nJustLimit, p.nSatUnsat, p.nSatSat, p.nSatUndec, p.timeTotal)
}
func Cbs2_ObjPrintFanouts(p *Cbs2_Man_t, id int32) {
	fmt.Printf("Fanouts of node %d:", id)
	for fan := p.vFanout0[id]; fan != 0; fan = p.vFanoutN[fan] {
		fmt.Printf(" %d", fan>>1)
	}
	fmt.Println()
}
func Cbs2_ManPrintFanouts(p *Cbs2_Man_t) {
	for id, fan := range p.vFanout0 {
		if fan != 0 {
			Cbs2_ObjPrintFanouts(p, int32(id))
		}
	}
}
func Cbs2_ObjCreateFanout(p *Cbs2_Man_t, id, a, b int32) {
	p.vFanoutN[id<<1], p.vFanoutN[id<<1|1] = p.vFanout0[a], p.vFanout0[b]
	p.vFanout0[a], p.vFanout0[b] = id<<1, id<<1|1
}
func Cbs2_ObjDeleteFanout(p *Cbs2_Man_t, id int32) {
	p.vFanoutN[id<<1], p.vFanoutN[id<<1|1], p.vFanout0[id] = 0, 0, 0
}
func Cbs2_ManCreateFanout_rec(p *Cbs2_Man_t, id int32) {
	if id == 0 || Gia_ObjIsTravIdCurrentId(p.pAig, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p.pAig, id)
	obj := p.pAig.pObjs[id]
	p.fanoutsReady = true
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	Cbs2_ManCreateFanout_rec(p, a)
	Cbs2_ManCreateFanout_rec(p, b)
	Cbs2_ObjCreateFanout(p, id, a, b)
}
func Cbs2_ManDeleteFanout_rec(p *Cbs2_Man_t, id int32) {
	if id == 0 {
		return
	}
	obj := p.pAig.pObjs[id]
	Cbs2_ObjDeleteFanout(p, id)
	p.fanoutsReady = false
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	if p.vFanout0[a] != 0 {
		Cbs2_ManDeleteFanout_rec(p, a)
	}
	if p.vFanout0[b] != 0 {
		Cbs2_ManDeleteFanout_rec(p, b)
	}
}
func Cbs2_ManCheckFanouts(p *Cbs2_Man_t) {
	for id, value := range p.vFanout0 {
		if value != 0 || p.vFanoutN[2*id] != 0 || p.vFanoutN[2*id+1] != 0 {
			panic("gia: stale CBS2 fanouts")
		}
	}
}
func Cbs2_ManSolveMiterNc(graph *Gia_Man_t, conflicts int32, statusOut **Vec_Str_t, verbose int32) *Vec_Int_t {
	if graph.nRegs != 0 {
		panic("gia: CBS2 miter must be combinational")
	}
	started := time.Now()
	p := Cbs2_ManAlloc(graph)
	defer Cbs2_ManStop(p)
	p.Pars.nBTLimit = conflicts
	statuses := Vec_StrAlloc(int(Gia_ManPoNum(graph)))
	store := Vec_IntAlloc(10000)
	for index, id := range Vec_IntArray(graph.vCos) {
		start := time.Now()
		lit := Gia_ObjFaninLit0p(graph, graph.pObjs[id])
		status := Cbs2_ManSolve(p, lit)
		Vec_StrPush(statuses, byte(status))
		if status == 1 {
			p.nSatUnsat++
			p.nConfUnsat += p.Pars.nBTThis
			p.timeSatUnsat += time.Since(start)
			continue
		}
		Vec_IntPush(store, int32(index))
		if status < 0 {
			Vec_IntPush(store, -1)
			p.nSatUndec++
			p.nConfUndec += p.Pars.nBTThis
			p.timeSatUndec += time.Since(start)
		} else {
			Vec_IntPush(store, int32(Vec_IntSize(p.vModel)))
			for _, value := range Vec_IntArray(p.vModel) {
				Vec_IntPush(store, value)
			}
			p.nSatSat++
			p.nConfSat += p.Pars.nBTThis
			p.timeSatSat += time.Since(start)
		}
	}
	p.nSatTotal = Gia_ManPoNum(graph)
	p.timeTotal = time.Since(started)
	if verbose != 0 {
		Cbs2_ManSatPrintStats(p)
	}
	if statusOut != nil {
		*statusOut = statuses
	}
	return store
}
