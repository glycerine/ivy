// Boolean constructors ported from ABC ivyOper.c (Alan Mishchenko).
package ivy

func Ivy_ObjIsExorType(a, b *Ivy_Obj_t, out0, out1 **Ivy_Obj_t) int32 {
	a, b = Ivy_Regular(a), Ivy_Regular(b)
	if Ivy_ObjIsAnd(a) == 0 || Ivy_ObjIsAnd(b) == 0 || Ivy_ObjFanin0(a) != Ivy_ObjFanin0(b) || Ivy_ObjFanin1(a) != Ivy_ObjFanin1(b) || Ivy_ObjFaninC0(a) == Ivy_ObjFaninC0(b) || Ivy_ObjFaninC1(a) == Ivy_ObjFaninC1(b) {
		return 0
	}
	*out0, *out1 = a.pFanin0, a.pFanin1
	return 1
}
func Ivy_Oper(p *Ivy_Man_t, a, b *Ivy_Obj_t, typ Ivy_Type_t) *Ivy_Obj_t {
	switch typ {
	case IVY_AND:
		return Ivy_And(p, a, b)
	case IVY_EXOR:
		return Ivy_Exor(p, a, b)
	}
	panic("unsupported binary operation")
}
func ivyCheckOperands(p *Ivy_Man_t, edges ...*Ivy_Obj_t) {
	if p == nil || p.memoryStopped {
		panic("manager is stopped")
	}
	for _, edge := range edges {
		node := Ivy_Regular(edge)
		if node == nil || node.owner != p || node.recycled {
			panic("foreign, absent or recycled operand")
		}
	}
}
func Ivy_And(p *Ivy_Man_t, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	ivyCheckOperands(p, a, b)
	if a == b {
		return a
	}
	if a == Ivy_Not(b) {
		return Ivy_ManConst0(p)
	}
	if Ivy_Regular(a) == p.pConst1 {
		if a == p.pConst1 {
			return b
		}
		return a
	}
	if Ivy_Regular(b) == p.pConst1 {
		if b == p.pConst1 {
			return a
		}
		return b
	}
	return Ivy_CanonAnd(p, a, b)
}
func Ivy_Exor(p *Ivy_Man_t, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	return Ivy_Or(p, Ivy_And(p, a, Ivy_Not(b)), Ivy_And(p, Ivy_Not(a), b))
}
func Ivy_Or(p *Ivy_Man_t, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	return Ivy_Not(Ivy_And(p, Ivy_Not(a), Ivy_Not(b)))
}
func ivyLookupAnd(p *Ivy_Man_t, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	if a == b {
		return a
	}
	if a == Ivy_Not(b) {
		return Ivy_ManConst0(p)
	}
	if Ivy_Regular(a) == p.pConst1 {
		if a == p.pConst1 {
			return b
		}
		return a
	}
	if Ivy_Regular(b) == p.pConst1 {
		if b == p.pConst1 {
			return a
		}
		return b
	}
	return Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, a, b, IVY_AND, IVY_INIT_NONE))
}
func Ivy_Mux(p *Ivy_Man_t, c, a, b *Ivy_Obj_t) *Ivy_Obj_t {
	ivyCheckOperands(p, c, a, b)
	if b == Ivy_Not(a) {
		return Ivy_Exor(p, c, b)
	}
	a0, a1 := ivyLookupAnd(p, c, a), ivyLookupAnd(p, Ivy_Not(c), b)
	if a0 != nil && a1 != nil {
		if result := ivyLookupAnd(p, Ivy_Not(a0), Ivy_Not(a1)); result != nil {
			return Ivy_Not(result)
		}
	}
	b0, b1 := ivyLookupAnd(p, c, Ivy_Not(a)), ivyLookupAnd(p, Ivy_Not(c), Ivy_Not(b))
	if b0 != nil && b1 != nil {
		if result := ivyLookupAnd(p, Ivy_Not(b0), Ivy_Not(b1)); result != nil {
			return result
		}
	}
	if ivyBool(a0 != nil)+ivyBool(a1 != nil) >= ivyBool(b0 != nil)+ivyBool(b1 != nil) {
		if a0 == nil {
			a0 = Ivy_And(p, c, a)
		}
		if a1 == nil {
			a1 = Ivy_And(p, Ivy_Not(c), b)
		}
		return Ivy_Or(p, a0, a1)
	}
	if b0 == nil {
		b0 = Ivy_And(p, c, Ivy_Not(a))
	}
	if b1 == nil {
		b1 = Ivy_And(p, Ivy_Not(c), Ivy_Not(b))
	}
	return Ivy_Not(Ivy_Or(p, b0, b1))
}
func Ivy_Maj(p *Ivy_Man_t, a, b, c *Ivy_Obj_t) *Ivy_Obj_t {
	return Ivy_Or(p, Ivy_Or(p, Ivy_And(p, a, b), Ivy_And(p, a, c)), Ivy_And(p, b, c))
}
func Ivy_Multi_rec(p *Ivy_Man_t, args []*Ivy_Obj_t, n int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if n <= 0 || int64(n) > int64(len(args)) {
		panic("invalid multi-input count")
	}
	if n == 1 {
		return args[0]
	}
	half := n / 2
	return Ivy_Oper(p, Ivy_Multi_rec(p, args, half, typ), Ivy_Multi_rec(p, args[half:], n-half, typ), typ)
}
func Ivy_Multi(p *Ivy_Man_t, args []*Ivy_Obj_t, n int32, typ Ivy_Type_t) *Ivy_Obj_t {
	if typ != IVY_AND && typ != IVY_EXOR {
		panic("invalid multi-input type")
	}
	return Ivy_Multi_rec(p, args, n, typ)
}
func Ivy_Miter(p *Ivy_Man_t, pairs *[]*Ivy_Obj_t) *Ivy_Obj_t {
	if pairs == nil || len(*pairs) == 0 || len(*pairs)%2 != 0 {
		panic("miter requires nonempty pairs")
	}
	args := *pairs
	for i := 0; i < len(args); i += 2 {
		args[i/2] = Ivy_Not(Ivy_Exor(p, args[i], args[i+1]))
	}
	*pairs = args[:len(args)/2]
	return Ivy_Not(Ivy_Multi_rec(p, *pairs, int32(len(*pairs)), IVY_AND))
}
func Ivy_Latch(p *Ivy_Man_t, node *Ivy_Obj_t, init Ivy_Init_t) *Ivy_Obj_t {
	ivyCheckOperands(p, node)
	return Ivy_CanonLatch(p, node, init)
}
