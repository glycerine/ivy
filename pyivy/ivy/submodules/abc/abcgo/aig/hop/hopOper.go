// Boolean constructors, ported from ABC hopOper.c (Alan Mishchenko).
package hop

func Hop_IthVar(p *Hop_Man_t, i int32) *Hop_Obj_t {
	if i < 0 {
		panic("negative variable")
	}
	for Hop_ManPiNum(p) <= i {
		Hop_ObjCreatePi(p)
	}
	return Hop_ManPi(p, i)
}
func Hop_Oper(p *Hop_Man_t, a, b *Hop_Obj_t, t Hop_Type_t) *Hop_Obj_t {
	switch t {
	case AIG_AND:
		return Hop_And(p, a, b)
	case AIG_EXOR:
		return Hop_Exor(p, a, b)
	}
	panic("unsupported Hop operation")
}
func Hop_And(p *Hop_Man_t, a, b *Hop_Obj_t) *Hop_Obj_t {
	if a == b {
		return a
	}
	if a == Hop_Not(b) {
		return Hop_ManConst0(p)
	}
	if Hop_Regular(a) == p.pConst1 {
		if a == p.pConst1 {
			return b
		}
		return a
	}
	if Hop_Regular(b) == p.pConst1 {
		if b == p.pConst1 {
			return a
		}
		return b
	}
	ghost := Hop_ObjCreateGhost(p, a, b, AIG_AND)
	if node := Hop_TableLookup(p, ghost); node != nil {
		return node
	}
	return Hop_ObjCreate(p, ghost)
}
func Hop_Exor(p *Hop_Man_t, a, b *Hop_Obj_t) *Hop_Obj_t {
	return Hop_Or(p, Hop_And(p, a, Hop_Not(b)), Hop_And(p, Hop_Not(a), b))
}
func Hop_Or(p *Hop_Man_t, a, b *Hop_Obj_t) *Hop_Obj_t {
	return Hop_Not(Hop_And(p, Hop_Not(a), Hop_Not(b)))
}
func Hop_Mux(p *Hop_Man_t, c, a, b *Hop_Obj_t) *Hop_Obj_t {
	return Hop_Or(p, Hop_And(p, c, a), Hop_And(p, Hop_Not(c), b))
}
func Hop_Maj(p *Hop_Man_t, a, b, c *Hop_Obj_t) *Hop_Obj_t {
	return Hop_Or(p, Hop_And(p, a, b), Hop_And(p, c, Hop_Or(p, a, b)))
}
func Hop_Multi_rec(p *Hop_Man_t, args []*Hop_Obj_t, n int32, t Hop_Type_t) *Hop_Obj_t {
	if n <= 0 || int(n) > len(args) {
		panic("invalid multi-input count")
	}
	if n == 1 {
		return args[0]
	}
	half := n / 2
	return Hop_Oper(p, Hop_Multi_rec(p, args[:half], half, t), Hop_Multi_rec(p, args[half:n], n-half, t), t)
}
func Hop_Multi(p *Hop_Man_t, args []*Hop_Obj_t, n int32, t Hop_Type_t) *Hop_Obj_t {
	return Hop_Multi_rec(p, args, n, t)
}
func Hop_Miter(p *Hop_Man_t, v *Vec_Ptr_t) *Hop_Obj_t {
	if Vec_PtrSize(v) == 0 || Vec_PtrSize(v)%2 != 0 {
		panic("miter requires pairs")
	}
	args := make([]*Hop_Obj_t, Vec_PtrSize(v)/2)
	for i := int32(0); i < Vec_PtrSize(v); i += 2 {
		args[i/2] = Hop_Not(Hop_Exor(p, Vec_PtrArray(v)[i].(*Hop_Obj_t), Vec_PtrArray(v)[i+1].(*Hop_Obj_t)))
	}
	Vec_PtrShrink(v, Vec_PtrSize(v)/2)
	for i, a := range args {
		Vec_PtrArray(v)[i] = a
	}
	return Hop_Not(Hop_Multi_rec(p, args, int32(len(args)), AIG_AND))
}
func Hop_CreateAnd(p *Hop_Man_t, n int32) *Hop_Obj_t {
	r := Hop_ManConst1(p)
	for i := int32(0); i < n; i++ {
		r = Hop_And(p, r, Hop_IthVar(p, i))
	}
	return r
}
func Hop_CreateOr(p *Hop_Man_t, n int32) *Hop_Obj_t {
	r := Hop_ManConst0(p)
	for i := int32(0); i < n; i++ {
		r = Hop_Or(p, r, Hop_IthVar(p, i))
	}
	return r
}
func Hop_CreateExor(p *Hop_Man_t, n int32) *Hop_Obj_t {
	r := Hop_ManConst0(p)
	for i := int32(0); i < n; i++ {
		r = Hop_Exor(p, r, Hop_IthVar(p, i))
	}
	return r
}

func Hop_ObjIsExorType(a, b *Hop_Obj_t, x, y **Hop_Obj_t) int32 {
	ghost := &Hop_Obj_t{Type: AIG_AND, pFanin0: a, pFanin1: b}
	return Hop_ObjRecognizeExor(ghost, x, y)
}
