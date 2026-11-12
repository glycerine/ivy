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
	case IVY_AND:
		node.pEquiv = Ivy_And(p.pHaig, mapped(Ivy_ObjChild0Equiv(node)), mapped(Ivy_ObjChild1Equiv(node)))
	default:
		panic("history creation supports buffers, latches and ANDs")
	}
}
