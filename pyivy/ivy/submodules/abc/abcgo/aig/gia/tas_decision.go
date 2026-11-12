package gia

import "github.com/glycerine/abcgo/misc/vec"

func Tas_VarFaninFanoutMax(p *Tas_Man_t, node *Gia_Obj_t) int32 {
	if Gia_ObjIsAnd(node) == 0 {
		panic("gia: expected SAT AND node")
	}
	return max(Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(node)), Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(node)))
}

func tasJustNodes(p *Tas_Man_t) []*Gia_Obj_t {
	q := &p.pJust
	end := q.iHead
	for end < q.iTail && q.pData[end] != nil {
		end++
	}
	return q.pData[q.iHead:end]
}

func Tas_ManFindActive(p *Tas_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	var score float32
	for _, node := range tasJustNodes(p) {
		if Gia_ObjIsAnd(node) == 0 {
			panic("gia: expected SAT AND node")
		}
		for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(node), Gia_ObjFanin1(node)} {
			activity := p.pActivity[Gia_ObjId(p.pAig, fanin)]
			if best == nil || activity > score || activity == score && fanin.id > best.id {
				best, score = fanin, activity
			}
		}
	}
	return best
}
func Tas_ManDecideHighestFanin(p *Tas_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, node := range tasJustNodes(p) {
		if Gia_ObjIsAnd(node) == 0 {
			panic("gia: expected SAT AND node")
		}
		for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(node), Gia_ObjFanin1(node)} {
			if best == nil || fanin.id > best.id {
				best = fanin
			}
		}
	}
	return best
}
func Tas_ManDecideHighest(p *Tas_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, node := range tasJustNodes(p) {
		if best == nil || node.id > best.id {
			best = node
		}
	}
	return best
}
func Tas_ManDecideHighestA(p *Tas_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, node := range tasJustNodes(p) {
		if best == nil || p.pActivity[best.id] < p.pActivity[node.id] {
			best = node
		}
	}
	return best
}
func Tas_ManDecideLowest(p *Tas_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	for _, node := range tasJustNodes(p) {
		if best == nil || node.id < best.id {
			best = node
		}
	}
	return best
}
func Tas_ManDecideMaxFF(p *Tas_Man_t) *Gia_Obj_t {
	var best *Gia_Obj_t
	var maximum int32
	for _, node := range tasJustNodes(p) {
		if score := Tas_VarFaninFanoutMax(p, node); score > maximum {
			best, maximum = node, score
		}
	}
	return best
}
func Tas_ManCancelUntil(p *Tas_Man_t, bound int32) {
	if bound < 0 || bound > p.pProp.iTail {
		panic("gia: invalid SAT backtrack bound")
	}
	for i := bound; i < p.pProp.iTail && p.pProp.pData[i] != nil; i++ {
		Tas_VarUnassign(p.pProp.pData[i])
	}
	p.pProp.iHead, p.pProp.iTail = bound, bound
	vec.Vec_IntShrink(p.vLevReas, int(3*bound))
}

// reason1 is either a second reason node or a clause when reason0 is nil.
func Tas_ManAssign(p *Tas_Man_t, edge *Gia_Obj_t, level int32, reason0 *Gia_Obj_t, reason1 any) {
	node := Gia_Regular(edge)
	if node == nil || Gia_ObjIsCand(node) == 0 || node.Value != ^uint32(0) {
		panic("gia: invalid SAT assignment")
	}
	var offset0, offset1 int32
	if reason0 != nil {
		offset0 = Gia_ObjId(p.pAig, reason0) - Gia_ObjId(p.pAig, node)
	}
	switch reason := reason1.(type) {
	case nil:
	case *Gia_Obj_t:
		if reason != nil {
			offset1 = Gia_ObjId(p.pAig, reason) - Gia_ObjId(p.pAig, node)
		}
	case *Tas_Cls_t:
		if reason0 != nil {
			panic("gia: clause reason with first node reason")
		}
		offset1 = Tas_ClsHandle(p, reason)
	default:
		panic("gia: invalid SAT reason")
	}
	Tas_VarAssign(node)
	Tas_VarSetValue(node, 1-Gia_IsComplement(edge))
	node.Value = uint32(p.pProp.iTail)
	Tas_QuePush(&p.pProp, node)
	vec.Vec_IntPush(p.vLevReas, level)
	vec.Vec_IntPush(p.vLevReas, offset0)
	vec.Vec_IntPush(p.vLevReas, offset1)
	if vec.Vec_IntSize(p.vLevReas) != 3*int(p.pProp.iTail) {
		panic("gia: unbalanced SAT trail")
	}
}

func Tas_ManClauseSize(p *Tas_Man_t, handle int32) int32 {
	if handle < 0 || int(handle) >= len(p.pClauses.pData) {
		panic("gia: invalid SAT clause queue handle")
	}
	end := handle
	for int(end) < len(p.pClauses.pData) && p.pClauses.pData[end] != nil {
		end++
	}
	if int(end) == len(p.pClauses.pData) {
		panic("gia: unterminated SAT clause")
	}
	return end - handle
}
