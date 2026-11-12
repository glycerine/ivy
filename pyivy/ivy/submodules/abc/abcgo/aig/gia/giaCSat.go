package gia

import (
	"fmt"
	"time"
)

type Cbs_Par_t struct {
	nBTLimit, nJustLimit, nBTThis, nBTThisNc, nJustThis, nBTTotal, nJustTotal int32
	fUseHighest, fUseLowest, fUseMaxFF, fVerbose                              int32
}

// Queue tails are logical lengths; suspended justification frontiers remain in
// the backing slice until a branch restores its saved head and tail.
type Cbs_Que_t struct {
	iHead, iTail int32
	pData        []*Gia_Obj_t
}
type Cbs_Man_t = Cbs_Man_t_
type Cbs_Man_t_ struct {
	Pars                                                                       Cbs_Par_t
	pAig                                                                       *Gia_Man_t
	pProp, pJust, pClauses                                                     Cbs_Que_t
	vLevReas, vModel                                                           *Vec_Int_t
	nSatUnsat, nSatSat, nSatUndec, nSatTotal, nConfUnsat, nConfSat, nConfUndec int32
	timeSatUnsat, timeSatSat, timeSatUndec, timeTotal                          time.Duration
}

func Cbs_VarIsAssigned(obj *Gia_Obj_t) int32 { return int32(obj.fMark0) }
func Cbs_VarAssign(obj *Gia_Obj_t) {
	if obj.fMark0 != 0 {
		panic("gia: CBS double assignment")
	}
	obj.fMark0 = 1
}
func Cbs_VarUnassign(obj *Gia_Obj_t) {
	if obj.fMark0 == 0 {
		panic("gia: CBS unassigns unknown value")
	}
	obj.fMark0, obj.fMark1 = 0, 0
	obj.Value = ^uint32(0)
}
func Cbs_VarValue(obj *Gia_Obj_t) int32 {
	if obj.fMark0 == 0 {
		panic("gia: CBS reads unknown value")
	}
	return int32(obj.fMark1)
}
func Cbs_VarSetValue(obj *Gia_Obj_t, value int32) {
	if obj.fMark0 == 0 || value < 0 || value > 1 {
		panic("gia: invalid CBS assignment")
	}
	obj.fMark1 = uint32(value)
}
func Cbs_VarIsJust(obj *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsAnd(obj) != 0 && Cbs_VarIsAssigned(Gia_ObjFanin0(obj)) == 0 && Cbs_VarIsAssigned(Gia_ObjFanin1(obj)) == 0)
}
func Cbs_VarFanin0Value(obj *Gia_Obj_t) int32 {
	if Cbs_VarIsAssigned(Gia_ObjFanin0(obj)) == 0 {
		return 2
	}
	return Cbs_VarValue(Gia_ObjFanin0(obj)) ^ Gia_ObjFaninC0(obj)
}
func Cbs_VarFanin1Value(obj *Gia_Obj_t) int32 {
	if Cbs_VarIsAssigned(Gia_ObjFanin1(obj)) == 0 {
		return 2
	}
	return Cbs_VarValue(Gia_ObjFanin1(obj)) ^ Gia_ObjFaninC1(obj)
}
func Cbs_VarDecLevel(p *Cbs_Man_t, obj *Gia_Obj_t) int32 {
	return Vec_IntEntry(p.vLevReas, 3*int(obj.Value))
}
func Cbs_VarReason0(p *Cbs_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	return p.pAig.pObjs[obj.id+Vec_IntEntry(p.vLevReas, 3*int(obj.Value)+1)]
}
func Cbs_VarReason1(p *Cbs_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	return p.pAig.pObjs[obj.id+Vec_IntEntry(p.vLevReas, 3*int(obj.Value)+2)]
}
func Cbs_ClauseDecLevel(p *Cbs_Man_t, clause int32) int32 {
	return Cbs_VarDecLevel(p, p.pClauses.pData[clause])
}
func Cbs_SetDefaultParams(parameters *Cbs_Par_t) {
	*parameters = Cbs_Par_t{nBTLimit: 1000, nJustLimit: 100, fUseHighest: 1, fVerbose: 1}
}
func Cbs_ManSetConflictNum(p *Cbs_Man_t, limit int32) { p.Pars.nBTLimit = limit }
func Cbs_ManAlloc(graph *Gia_Man_t) *Cbs_Man_t {
	for _, obj := range graph.pObjs {
		if Gia_ObjIsXor(obj) != 0 || Gia_ObjIsMux(graph, obj) != 0 {
			panic("gia: CBS expects ordinary AND graph")
		}
	}
	if graph.Refs == nil {
		Gia_ManCreateRefs(graph)
	}
	Gia_ManCleanMark0(graph)
	Gia_ManCleanMark1(graph)
	Gia_ManFillValue(graph)
	p := &Cbs_Man_t{pAig: graph, vLevReas: Vec_IntAlloc(1000), vModel: Vec_IntAlloc(1000)}
	p.pClauses = Cbs_Que_t{iHead: 1, iTail: 1, pData: make([]*Gia_Obj_t, 1, 10000)}
	Cbs_SetDefaultParams(&p.Pars)
	return p
}
func Cbs_ManStop(p *Cbs_Man_t) {
	if p != nil {
		if p.pProp.iTail != 0 {
			Cbs_ManCancelUntil(p, 0)
		}
		*p = Cbs_Man_t{}
	}
}
func Cbs_ReadModel(p *Cbs_Man_t) *Vec_Int_t { return p.vModel }
func Cbs_ManCheckLimits(p *Cbs_Man_t) int32 {
	return giaBool(p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit)
}
func Cbs_ManSaveModel(p *Cbs_Man_t, model *Vec_Int_t) {
	Vec_IntClear(model)
	p.pProp.iHead = 0
	for _, obj := range p.pProp.pData[:p.pProp.iTail] {
		if Gia_ObjIsCi(obj) != 0 {
			Vec_IntPush(model, Gia_ObjCioId(obj)<<1|(Cbs_VarValue(obj)^1))
		}
	}
}
func Cbs_ManSaveModelAll(p *Cbs_Man_t, model *Vec_Int_t) {
	Vec_IntClear(model)
	p.pProp.iHead = 0
	for _, obj := range p.pProp.pData[:p.pProp.iTail] {
		Vec_IntPush(model, obj.id<<1|(Cbs_VarValue(obj)^1))
	}
}
func Cbs_QueIsEmpty(queue *Cbs_Que_t) int32 { return giaBool(queue.iHead == queue.iTail) }
func Cbs_QuePush(queue *Cbs_Que_t, obj *Gia_Obj_t) {
	if Gia_IsComplement(obj) != 0 {
		panic("gia: CBS queues hold regular objects")
	}
	if int(queue.iTail) == len(queue.pData) {
		queue.pData = append(queue.pData, obj)
	} else {
		queue.pData[queue.iTail] = obj
	}
	queue.iTail++
}
func Cbs_QueHasNode(queue *Cbs_Que_t, obj *Gia_Obj_t) int32 {
	for _, candidate := range queue.pData[queue.iHead:queue.iTail] {
		if candidate == obj {
			return 1
		}
	}
	return 0
}
func Cbs_QueStore(queue *Cbs_Que_t, head, tail *int32) {
	*head, *tail = queue.iHead, queue.iTail
	for i := *head; i < *tail; i++ {
		Cbs_QuePush(queue, queue.pData[i])
	}
	queue.iHead = *tail
}
func Cbs_QueRestore(queue *Cbs_Que_t, head, tail int32) { queue.iHead, queue.iTail = head, tail }
func Cbs_QueFinish(queue *Cbs_Que_t) int32 {
	head := queue.iHead
	if head >= queue.iTail {
		panic("gia: empty CBS clause")
	}
	Cbs_QuePush(queue, nil)
	queue.iHead = queue.iTail
	return head
}
func Cbs_VarFaninFanoutMax(p *Cbs_Man_t, obj *Gia_Obj_t) int32 {
	return max(Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(obj)), Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(obj)))
}
func Cbs_ManDecideHighest(p *Cbs_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, obj := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if best == nil || obj.id > best.id {
			best = obj
		}
	}
	return best
}
func Cbs_ManDecideLowest(p *Cbs_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, obj := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if best == nil || obj.id < best.id {
			best = obj
		}
	}
	return best
}
func Cbs_ManDecideMaxFF(p *Cbs_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	var maximum int32
	for _, obj := range p.pJust.pData[p.pJust.iHead:p.pJust.iTail] {
		if count := Cbs_VarFaninFanoutMax(p, obj); count > maximum {
			maximum = count
			best = obj
		}
	}
	return best
}
func Cbs_ManCancelUntil(p *Cbs_Man_t, bound int32) {
	if bound > p.pProp.iTail {
		panic("gia: invalid CBS trail bound")
	}
	for _, obj := range p.pProp.pData[bound:p.pProp.iTail] {
		Cbs_VarUnassign(obj)
	}
	p.pProp.iHead, p.pProp.iTail = bound, bound
	Vec_IntShrink(p.vLevReas, int(3*bound))
}
func Cbs_ManAssign(p *Cbs_Man_t, literal *Gia_Obj_t, level int32, reason0, reason1 *Gia_Obj_t) {
	obj := Gia_Regular(literal)
	Cbs_VarAssign(obj)
	Cbs_VarSetValue(obj, Gia_IsComplement(literal)^1)
	obj.Value = uint32(p.pProp.iTail)
	Cbs_QuePush(&p.pProp, obj)
	Vec_IntPush(p.vLevReas, level)
	var a, b int32
	if reason0 != nil {
		a = reason0.id - obj.id
	}
	if reason1 != nil {
		b = reason1.id - obj.id
	}
	Vec_IntPush(p.vLevReas, a)
	Vec_IntPush(p.vLevReas, b)
}
func Cbs_ManClauseSize(p *Cbs_Man_t, clause int32) int32 {
	var count int32
	for p.pClauses.pData[clause+count] != nil {
		count++
	}
	return count
}
func Cbs_ManPrintClause(p *Cbs_Man_t, level, clause int32) {
	fmt.Printf("Level %2d: ", level)
	for i := clause; p.pClauses.pData[i] != nil; i++ {
		obj := p.pClauses.pData[i]
		fmt.Printf("%d=%d(%d) ", obj.id, Cbs_VarValue(obj), Cbs_VarDecLevel(p, obj))
	}
	fmt.Println()
}
func Cbs_ManPrintClauseNew(p *Cbs_Man_t, level, clause int32) { Cbs_ManPrintClause(p, level, clause) }
func Cbs_ManDeriveReason(p *Cbs_Man_t, level int32) {
	queue := &p.pClauses
	write := queue.iHead + 1
	var visited []*Gia_Obj_t
	for read := write; read < queue.iTail; read++ {
		obj := queue.pData[read]
		if obj.fMark0 == 0 {
			continue
		}
		obj.fMark0 = 0
		visited = append(visited, obj)
		nodeLevel := Cbs_VarDecLevel(p, obj)
		if nodeLevel < level {
			queue.pData[write] = obj
			write++
			continue
		}
		if nodeLevel != level {
			panic("gia: CBS reason is above decision level")
		}
		reason := Cbs_VarReason0(p, obj)
		if reason == obj {
			queue.pData[queue.iHead] = obj
			continue
		}
		Cbs_QuePush(queue, reason)
		if reason = Cbs_VarReason1(p, obj); reason != obj {
			Cbs_QuePush(queue, reason)
		}
	}
	if queue.pData[queue.iHead] == nil {
		panic("gia: CBS conflict has no decision")
	}
	queue.iTail = write
	for _, obj := range visited {
		obj.fMark0 = 1
	}
}
func Cbs_ManAnalyze(p *Cbs_Man_t, level int32, obj, a, b *Gia_Obj_t) int32 {
	queue := &p.pClauses
	Cbs_QuePush(queue, nil)
	Cbs_QuePush(queue, obj)
	Cbs_QuePush(queue, a)
	if b != nil {
		Cbs_QuePush(queue, b)
	}
	Cbs_ManDeriveReason(p, level)
	return Cbs_QueFinish(queue)
}
func Cbs_ManResolve(p *Cbs_Man_t, level, first, second int32) int32 {
	queue := &p.pClauses
	if queue.pData[first] != queue.pData[second] {
		panic("gia: CBS clauses have different pivots")
	}
	Cbs_QuePush(queue, nil)
	maximum := int32(-1)
	for _, clause := range []int32{first, second} {
		for index := clause + 1; queue.pData[index] != nil; index++ {
			obj := queue.pData[index]
			if obj.fMark0 == 0 {
				continue
			}
			obj.fMark0 = 0
			Cbs_QuePush(queue, obj)
			maximum = max(maximum, Cbs_VarDecLevel(p, obj))
		}
	}
	for i := queue.iHead + 1; i < queue.iTail; i++ {
		queue.pData[i].fMark0 = 1
	}
	Cbs_ManDeriveReason(p, maximum)
	return Cbs_QueFinish(queue)
}
func Cbs_ManPropagateOne(p *Cbs_Man_t, obj *Gia_Obj_t, level int32) int32 {
	if Gia_ObjIsCi(obj) != 0 || obj.id == 0 {
		return 0
	}
	a, b := Cbs_VarFanin0Value(obj), Cbs_VarFanin1Value(obj)
	if Cbs_VarValue(obj) != 0 {
		if a == 0 || b == 0 {
			if a == 0 && b != 0 {
				return Cbs_ManAnalyze(p, level, obj, Gia_ObjFanin0(obj), nil)
			}
			if a != 0 && b == 0 {
				return Cbs_ManAnalyze(p, level, obj, Gia_ObjFanin1(obj), nil)
			}
			return Cbs_ManAnalyze(p, level, obj, Gia_ObjFanin0(obj), Gia_ObjFanin1(obj))
		}
		if a == 2 {
			Cbs_ManAssign(p, Gia_ObjChild0(obj), level, obj, nil)
		}
		if b == 2 {
			if Cbs_VarIsAssigned(Gia_ObjFanin1(obj)) == 0 {
				Cbs_ManAssign(p, Gia_ObjChild1(obj), level, obj, nil)
			} else if Cbs_VarFanin1Value(obj) == 0 {
				return Cbs_ManAnalyze(p, level, obj, Gia_ObjFanin1(obj), nil)
			}
		}
		return 0
	}
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 && b == 1 {
		return Cbs_ManAnalyze(p, level, obj, Gia_ObjFanin0(obj), Gia_ObjFanin1(obj))
	}
	if a == 1 || b == 1 {
		return Cbs_ManPropagateTwo(p, obj, level)
	}
	if Cbs_QueHasNode(&p.pJust, obj) == 0 {
		Cbs_QuePush(&p.pJust, obj)
	}
	return 0
}
func Cbs_ManPropagateTwo(p *Cbs_Man_t, obj *Gia_Obj_t, level int32) int32 {
	a, b := Cbs_VarFanin0Value(obj), Cbs_VarFanin1Value(obj)
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 && b == 1 {
		return Cbs_ManAnalyze(p, level, obj, Gia_ObjFanin0(obj), Gia_ObjFanin1(obj))
	}
	if a == 2 {
		Cbs_ManAssign(p, Gia_Not(Gia_ObjChild0(obj)), level, obj, Gia_ObjFanin1(obj))
	}
	if b == 2 {
		Cbs_ManAssign(p, Gia_Not(Gia_ObjChild1(obj)), level, obj, Gia_ObjFanin0(obj))
	}
	return 0
}
func Cbs_ManPropagate(p *Cbs_Man_t, level int32) int32 {
	for {
		for index := p.pProp.iHead; index < p.pProp.iTail; index++ {
			if clause := Cbs_ManPropagateOne(p, p.pProp.pData[index], level); clause != 0 {
				return clause
			}
		}
		p.pProp.iHead = p.pProp.iTail
		write := p.pJust.iHead
		for index := p.pJust.iHead; index < p.pJust.iTail; index++ {
			obj := p.pJust.pData[index]
			if Cbs_VarIsJust(obj) != 0 {
				p.pJust.pData[write] = obj
				write++
			} else if clause := Cbs_ManPropagateTwo(p, obj, level); clause != 0 {
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
func Cbs_ManSolve_rec(p *Cbs_Man_t, level int32) int32 {
	if clause := Cbs_ManPropagate(p, level); clause != 0 {
		return clause
	}
	if Cbs_QueIsEmpty(&p.pJust) != 0 {
		return 0
	}
	p.Pars.nJustThis = max(p.Pars.nJustThis, p.pJust.iTail-p.pJust.iHead)
	if Cbs_ManCheckLimits(p) != 0 {
		return 0
	}
	bound := p.pProp.iHead
	var head, tail int32
	Cbs_QueStore(&p.pJust, &head, &tail)
	var variable *Gia_Obj_t
	switch {
	case p.Pars.fUseHighest != 0:
		variable = Cbs_ManDecideHighest(p)
	case p.Pars.fUseLowest != 0:
		variable = Cbs_ManDecideLowest(p)
	case p.Pars.fUseMaxFF != 0:
		variable = Cbs_ManDecideMaxFF(p)
	default:
		panic("gia: CBS has no decision heuristic")
	}
	var decision *Gia_Obj_t
	if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(variable)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(variable)) {
		decision = Gia_Not(Gia_ObjChild0(variable))
	} else {
		decision = Gia_Not(Gia_ObjChild1(variable))
	}
	Cbs_ManAssign(p, decision, level+1, nil, nil)
	first := Cbs_ManSolve_rec(p, level+1)
	if first == 0 {
		return 0
	}
	if p.pClauses.pData[first] != Gia_Regular(decision) {
		return first
	}
	Cbs_ManCancelUntil(p, bound)
	Cbs_QueRestore(&p.pJust, head, tail)
	Cbs_ManAssign(p, Gia_Not(decision), level+1, nil, nil)
	second := Cbs_ManSolve_rec(p, level+1)
	if second == 0 {
		return 0
	}
	if p.pClauses.pData[second] != Gia_Regular(decision) {
		return second
	}
	clause := Cbs_ManResolve(p, level, first, second)
	p.Pars.nBTThis++
	return clause
}
func cbsSolve(p *Cbs_Man_t, first, second *Gia_Obj_t, all bool) int32 {
	if p.pProp.iTail != 0 || p.pJust.iTail != 0 {
		panic("gia: CBS solver is already active")
	}
	Vec_IntClear(p.vModel)
	p.Pars.nBTThis, p.Pars.nJustThis, p.Pars.nBTThisNc = 0, 0, 0
	status := int32(0)
	roots := []*Gia_Obj_t{first}
	if second != nil {
		roots = append(roots, second)
	}
	for _, root := range roots {
		obj := Gia_Regular(root)
		if obj.id == 0 {
			if Gia_IsComplement(root) == 0 {
				status = 1
				break
			}
			continue
		}
		if Cbs_VarIsAssigned(obj) != 0 {
			if Cbs_VarValue(obj) != (Gia_IsComplement(root) ^ 1) {
				status = 1
				break
			}
			continue
		}
		Cbs_ManAssign(p, root, 0, nil, nil)
	}
	if status == 0 && p.pProp.iTail != 0 {
		if Cbs_ManSolve_rec(p, 0) != 0 || Cbs_ManCheckLimits(p) != 0 {
			status = 1
		} else if all {
			Cbs_ManSaveModelAll(p, p.vModel)
		} else {
			Cbs_ManSaveModel(p, p.vModel)
		}
	}
	Cbs_ManCancelUntil(p, 0)
	p.pJust.iHead, p.pJust.iTail = 0, 0
	p.pClauses.iHead, p.pClauses.iTail = 1, 1
	p.Pars.nBTTotal += p.Pars.nBTThis
	p.Pars.nJustTotal = max(p.Pars.nJustTotal, p.Pars.nJustThis)
	if Cbs_ManCheckLimits(p) != 0 {
		return -1
	}
	return status
}
func Cbs_ManSolve(p *Cbs_Man_t, root *Gia_Obj_t) int32 { return cbsSolve(p, root, nil, false) }
func Cbs_ManSolve2(p *Cbs_Man_t, first, second *Gia_Obj_t) int32 {
	return cbsSolve(p, first, second, true)
}
func Cbs_ManSatPrintStats(p *Cbs_Man_t) {
	fmt.Printf("CO=%d AND=%d Conf=%d JustMax=%d Unsat=%d Sat=%d Undef=%d TotalTime=%s\n", Gia_ManCoNum(p.pAig), Gia_ManAndNum(p.pAig), p.Pars.nBTLimit, p.Pars.nJustLimit, p.nSatUnsat, p.nSatSat, p.nSatUndec, p.timeTotal)
}
func Cbs_ManSolveMiterNc(graph *Gia_Man_t, conflicts int32, statusOut **Vec_Str_t, proveZero, verbose int32) *Vec_Int_t {
	if graph.nRegs != 0 {
		panic("gia: CBS miter must be combinational")
	}
	started := time.Now()
	p := Cbs_ManAlloc(graph)
	p.Pars.nBTLimit = conflicts
	defer Cbs_ManStop(p)
	statusVector := Vec_StrAlloc(int(Gia_ManPoNum(graph)))
	store := Vec_IntAlloc(10000)
	for index, id := range Vec_IntArray(graph.vCos) {
		start := time.Now()
		status := Cbs_ManSolve(p, Gia_ObjChild0(graph.pObjs[id]))
		Vec_StrPush(statusVector, byte(status))
		if status == 1 {
			if proveZero != 0 {
				Gia_ManPatchCoDriver(graph, int32(index), 0)
			}
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
			for _, literal := range Vec_IntArray(p.vModel) {
				Vec_IntPush(store, literal)
			}
			p.nSatSat++
			p.nConfSat += p.Pars.nBTThis
			p.timeSatSat += time.Since(start)
		}
	}
	p.nSatTotal = Gia_ManPoNum(graph)
	p.timeTotal = time.Since(started)
	if verbose != 0 {
		Cbs_ManSatPrintStats(p)
	}
	if statusOut != nil {
		*statusOut = statusVector
	}
	return store
}
