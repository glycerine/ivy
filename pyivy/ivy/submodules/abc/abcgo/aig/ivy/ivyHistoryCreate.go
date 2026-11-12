// History creation helpers ported from ABC ivyHaig.c (Alan Mishchenko).
package ivy

func Ivy_HaigObjRepr(node *Ivy_Obj_t) *Ivy_Obj_t {
	if node == nil || Ivy_IsComplement(node) != 0 {
		panic("regular history node required")
	}
	if node.pEquiv == nil || node.nRefs > 0 {
		return node
	}
	seen := map[*Ivy_Obj_t]bool{node: true}
	for next := Ivy_Regular(node.pEquiv); next != nil && !seen[next]; next = Ivy_Regular(next.pEquiv) {
		if next.nRefs > 0 {
			return Ivy_NotCond(next, Ivy_IsComplement(node.pEquiv))
		}
		seen[next] = true
	}
	panic("history equivalence ring has no representative")
}
func Ivy_ManHaigCreateObj(p *Ivy_Man_t, node *Ivy_Obj_t) {
	if p.pHaig == nil || Ivy_IsComplement(node) != 0 {
		panic("history manager and regular node required")
	}
	mapped := func(edge *Ivy_Obj_t) *Ivy_Obj_t {
		if edge == nil {
			panic("missing history mapping")
		}
		return Ivy_NotCond(Ivy_HaigObjRepr(Ivy_Regular(edge)), Ivy_IsComplement(edge))
	}
	switch node.Type {
	case IVY_BUF:
		node.pEquiv = Ivy_ObjChild0Equiv(node)
	case IVY_LATCH:
		node.pEquiv = Ivy_Latch(p.pHaig, mapped(Ivy_ObjChild0Equiv(node)), node.Init)
	case IVY_AND, IVY_EXOR:
		node.pEquiv = Ivy_Oper(p.pHaig, mapped(Ivy_ObjChild0Equiv(node)), mapped(Ivy_ObjChild1Equiv(node)), node.Type)
	case IVY_PI:
		node.pEquiv = Ivy_ObjCreatePi(p.pHaig)
	case IVY_PO:
		index := len(p.vPos) - 1
		if index < len(p.pHaig.vPos) {
			node.pEquiv = p.pHaig.vPos[index]
		} else {
			node.pEquiv = Ivy_ObjCreatePo(p.pHaig, mapped(Ivy_ObjChild0Equiv(node)))
		}
	case IVY_ASSERT:
		node.pEquiv = mapped(Ivy_ObjChild0Equiv(node))
	default:
		panic("unsupported history object")
	}
}

func Ivy_ObjIsInTfi_rec(next, old *Ivy_Obj_t, levels int32) int32 {
	seen := map[*Ivy_Obj_t]int32{}
	var visit func(*Ivy_Obj_t, int32) bool
	visit = func(n *Ivy_Obj_t, left int32) bool {
		n = Ivy_Regular(n)
		if n == old {
			return true
		}
		if n == nil || left == 0 || Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
			return false
		}
		if previous, known := seen[n]; known && (previous < 0 || left >= 0 && previous >= left) {
			return false
		}
		seen[n] = left
		remaining := left - 1
		if left < 0 {
			remaining = -1
		}
		return visit(n.pFanin0, remaining) || visit(n.pFanin1, remaining)
	}
	return ivyBool(visit(next, levels))
}
func Ivy_ManHaigCreateChoice(p *Ivy_Man_t, old, next *Ivy_Obj_t) {
	if p.pHaig == nil || Ivy_IsComplement(old) != 0 || old.pEquiv == nil || Ivy_Regular(next).pEquiv == nil {
		panic("ivy: missing history choice mapping")
	}
	a := old.pEquiv
	b := Ivy_NotCond(Ivy_Regular(next).pEquiv, Ivy_IsComplement(next))
	a = Ivy_NotCond(Ivy_HaigObjRepr(Ivy_Regular(a)), Ivy_IsComplement(a))
	b = Ivy_NotCond(Ivy_HaigObjRepr(Ivy_Regular(b)), Ivy_IsComplement(b))
	ar, br := Ivy_Regular(a), Ivy_Regular(b)
	if ar == br {
		return
	}
	if ar.nRefs == 0 || br.pEquiv != nil || br.nRefs > 0 {
		p.pHaig.nClassesSkip++
		return
	}
	compl := Ivy_IsComplement(a) ^ Ivy_IsComplement(b)
	if ar.pEquiv == nil {
		br.pEquiv = Ivy_NotCond(ar, compl)
	} else {
		br.pEquiv = Ivy_NotCond(ar.pEquiv, compl)
	}
	ar.pEquiv = br
}
