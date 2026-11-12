// Sequential canonicalization ported from ABC ivyCanon.c (Alan Mishchenko).
package ivy

func Ivy_CanonPair_rec(p *Ivy_Man_t, ghost *Ivy_Obj_t) *Ivy_Obj_t {
	a, b := Ivy_ObjFanin0(ghost), Ivy_ObjFanin1(ghost)
	typ := ghost.Type
	c0, c1 := Ivy_ObjFaninC0(ghost), Ivy_ObjFaninC1(ghost)
	if Ivy_ObjIsNode(ghost) == 0 {
		panic("binary canonicalization needs node")
	}
	if Ivy_ObjIsLatch(a) == 0 || Ivy_ObjIsLatch(b) == 0 {
		if existing := Ivy_TableLookup(p, ghost); existing != nil {
			return existing
		}
		return Ivy_ObjCreate(p, ghost)
	}
	result := Ivy_Oper(p, Ivy_NotCond(Ivy_ObjFanin0(a), c0), Ivy_NotCond(Ivy_ObjFanin0(b), c1), typ)
	init0, init1 := Ivy_InitNotCond(a.Init, c0), Ivy_InitNotCond(b.Init, c1)
	init := Ivy_InitAnd(init0, init1)
	if typ == IVY_EXOR {
		init = Ivy_InitExor(init0, init1)
	}
	return Ivy_Latch(p, result, init)
}
func Ivy_CanonAnd(p *Ivy_Man_t, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	return Ivy_CanonPair_rec(p, Ivy_ObjCreateGhost(p, a, b, IVY_AND, IVY_INIT_NONE))
}
func Ivy_CanonExor(p *Ivy_Man_t, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	compl := Ivy_IsComplement(a) ^ Ivy_IsComplement(b)
	result := Ivy_CanonPair_rec(p, Ivy_ObjCreateGhost(p, Ivy_Regular(a), Ivy_Regular(b), IVY_EXOR, IVY_INIT_NONE))
	return Ivy_NotCond(result, compl)
}
func Ivy_CanonLatch(p *Ivy_Man_t, node *Ivy_Obj_t, init Ivy_Init_t) *Ivy_Obj_t {
	compl := Ivy_IsComplement(node)
	ghost := Ivy_ObjCreateGhost(p, Ivy_Regular(node), nil, IVY_LATCH, Ivy_InitNotCond(init, compl))
	result := Ivy_TableLookup(p, ghost)
	if result == nil {
		result = Ivy_ObjCreate(p, ghost)
	}
	return Ivy_NotCond(result, compl)
}
