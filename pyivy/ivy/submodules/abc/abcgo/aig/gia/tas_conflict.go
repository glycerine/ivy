package gia

import "github.com/glycerine/abcgo/misc/vec"

func Tas_ManDeriveReason(p *Tas_Man_t, level int32) {
	q := &p.pClauses
	if q.iHead+1 >= q.iTail || q.pData[q.iHead] != nil {
		panic("gia: invalid conflict reason queue")
	}
	var marked []*Gia_Obj_t
	defer func() {
		for _, node := range marked {
			node.fPhase = 0
		}
	}()
	k := q.iHead + 1
	for i := k; i < q.iTail; i++ {
		node := q.pData[i]
		if node.fPhase != 0 {
			continue
		}
		node.fPhase = 1
		marked = append(marked, node)
		litLevel := Tas_VarDecLevel(p, node)
		if litLevel < level {
			q.pData[k] = node
			k++
			continue
		}
		if litLevel != level {
			panic("gia: conflict reason above decision level")
		}
		if Tas_VarHasReasonCls(p, node) != 0 {
			lits := Tas_VarReasonCls(p, node).literals()
			if len(lits) == 0 || lits[0]>>1 != node.id {
				panic("gia: invalid clause reason")
			}
			for _, lit := range lits[1:] {
				reason := Gia_ManObj(p.pAig, lit>>1)
				if Tas_VarIsAssigned(reason) == 0 || Tas_LitIsTrue(reason, lit) != 0 {
					panic("gia: clause reason is not false")
				}
				Tas_QuePush(q, reason)
			}
		} else {
			reason := Tas_VarReason0(p, node)
			if reason == node {
				if q.pData[q.iHead] == nil {
					q.pData[q.iHead] = node
				} else if level == 0 {
					Tas_QuePush(q, node)
				} else {
					panic("gia: multiple decision variables at one level")
				}
				continue
			}
			Tas_QuePush(q, reason)
			if reason = Tas_VarReason1(p, node); reason != node {
				Tas_QuePush(q, reason)
			}
		}
	}
	if q.pData[q.iHead] == nil {
		panic("gia: failed to derive conflict reason")
	}
	q.iTail = k
}

func Tas_ManAnalyze(p *Tas_Man_t, level int32, variable, fanin0, fanin1 *Gia_Obj_t) int32 {
	q := &p.pClauses
	if Tas_QueIsEmpty(q) == 0 || Tas_VarIsAssigned(variable) == 0 || Tas_VarIsAssigned(fanin0) == 0 || fanin1 != nil && Tas_VarIsAssigned(fanin1) == 0 {
		panic("gia: invalid SAT conflict")
	}
	Tas_QuePush(q, nil)
	Tas_QuePush(q, variable)
	Tas_QuePush(q, fanin0)
	if fanin1 != nil {
		Tas_QuePush(q, fanin1)
	}
	Tas_ManDeriveReason(p, level)
	return Tas_QueFinish(q)
}

func Tas_ManResolve(p *Tas_Man_t, _ int32, first, second int32) int32 {
	q := &p.pClauses
	if Tas_QueIsEmpty(q) == 0 || q.pData[first] == nil || q.pData[first] != q.pData[second] {
		panic("gia: invalid conflict resolution")
	}
	Tas_QuePush(q, nil)
	seen := make(map[*Gia_Obj_t]bool)
	maxLevel := int32(-1)
	for _, handle := range []int32{first, second} {
		for i := handle + 1; q.pData[i] != nil; i++ {
			node := q.pData[i]
			if seen[node] {
				continue
			}
			seen[node] = true
			Tas_QuePush(q, node)
			maxLevel = max(maxLevel, Tas_VarDecLevel(p, node))
		}
	}
	Tas_ManDeriveReason(p, maxLevel)
	return Tas_QueFinish(q)
}

func Tas_ManAllocCls(p *Tas_Man_t, words int32) *Tas_Cls_t {
	if words < 3 {
		panic("gia: invalid SAT clause size")
	}
	p.nClauses++
	return p.pStore.allocate(int(words - 3))
}

func Tas_ManWatchClause(p *Tas_Man_t, clause *Tas_Cls_t, literal int32) {
	lits := clause.literals()
	if len(lits) < 2 || literal < 0 || int(literal>>1) >= int(p.pAig.nObjs) {
		panic("gia: invalid watched clause")
	}
	index := 0
	if lits[0] != literal {
		index = 1
		if lits[1] != literal {
			panic("gia: literal not watched")
		}
	}
	watch := literal ^ 1
	clause.setNext(index, p.pWatches[watch])
	if p.pWatches[watch] == 0 {
		vec.Vec_IntPush(p.vWatchLits, watch)
	}
	p.pWatches[watch] = Tas_ClsHandle(p, clause)
}

func Tas_ManCreateCls(p *Tas_Man_t, handle int32) *Tas_Cls_t {
	if Tas_QueIsEmpty(&p.pClauses) == 0 {
		panic("gia: clause queue is not empty")
	}
	count := Tas_ManClauseSize(p, handle)
	if count == 1 {
		return nil
	}
	if count < 1 {
		panic("gia: empty conflict clause")
	}
	clause := Tas_ManAllocCls(p, count+3)
	for i := int32(0); i < count; i++ {
		clause.literals()[i] = Tas_VarToLit(p, p.pClauses.pData[handle+i]) ^ 1
	}
	Tas_ManWatchClause(p, clause, clause.literals()[0])
	Tas_ManWatchClause(p, clause, clause.literals()[1])
	return clause
}

func Tas_ManCreateFromCls(p *Tas_Man_t, clause *Tas_Cls_t, level int32) int32 {
	q := &p.pClauses
	if Tas_QueIsEmpty(q) == 0 {
		panic("gia: clause queue is not empty")
	}
	Tas_QuePush(q, nil)
	for _, lit := range clause.literals() {
		node := Gia_ManObj(p.pAig, lit>>1)
		if Tas_VarIsAssigned(node) == 0 || Tas_LitIsTrue(node, lit) != 0 {
			panic("gia: conflict clause is not false")
		}
		Tas_QuePush(q, node)
	}
	Tas_ManDeriveReason(p, level)
	return Tas_QueFinish(q)
}

func Tas_ManPropagateWatch(p *Tas_Man_t, level, literal int32) int32 {
	if literal < 0 || int(literal) >= len(p.pWatches) {
		panic("gia: invalid propagated literal")
	}
	falseLit := literal ^ 1
	previous := int32(0)
	current := p.pWatches[literal]
	for current != 0 {
		clause := Tas_ClsFromHandle(p, current)
		lits := clause.literals()
		if lits[0] == falseLit {
			lits[0], lits[1] = lits[1], lits[0]
			a, b := clause.next(0), clause.next(1)
			clause.setNext(0, b)
			clause.setNext(1, a)
		}
		if lits[1] != falseLit {
			panic("gia: broken SAT watch list")
		}
		node := Gia_ManObj(p.pAig, lits[0]>>1)
		if Tas_VarIsAssigned(node) != 0 && Tas_LitIsTrue(node, lits[0]) != 0 {
			previous, current = current, clause.next(1)
			continue
		}
		replacement := -1
		for i := 2; i < len(lits); i++ {
			candidate := Gia_ManObj(p.pAig, lits[i]>>1)
			if Tas_VarIsAssigned(candidate) == 0 || Tas_LitIsTrue(candidate, lits[i]) != 0 {
				replacement = i
				break
			}
		}
		if replacement >= 0 {
			lits[1], lits[replacement] = lits[replacement], falseLit
			next := clause.next(1)
			if previous == 0 {
				p.pWatches[literal] = next
			} else {
				Tas_ClsFromHandle(p, previous).setNext(1, next)
			}
			Tas_ManWatchClause(p, clause, lits[1])
			current = next
			continue
		}
		if Tas_VarIsAssigned(node) == 0 {
			Tas_ManAssign(p, Gia_ObjFromLit(p.pAig, lits[0]), level, nil, clause)
			previous, current = current, clause.next(1)
			continue
		}
		return Tas_ManCreateFromCls(p, clause, level)
	}
	return 0
}
