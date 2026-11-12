package gia

// Gia_ManSimplifyXor reduces a sorted supergate, cancelling equal pairs and
// accumulating constant-one inputs into the first remaining literal's phase.
func Gia_ManSimplifyXor(super *Vec_Int_t) {
	count, previous, phase := 0, int32(-1), int32(0)
	for _, literal := range Vec_IntArray(super) {
		if literal == 0 {
			continue
		}
		if literal == 1 {
			phase ^= 1
			continue
		}
		if previous != literal {
			Vec_IntWriteEntry(super, count, literal)
			count++
			previous = literal
		} else {
			previous = -1
			count--
		}
	}
	Vec_IntShrink(super, count)
	if count == 0 {
		Vec_IntPush(super, phase)
	} else if phase != 0 {
		Vec_IntWriteEntry(super, 0, Vec_IntEntry(super, 0)^1)
	}
}

// Gia_ManSimplifyAnd removes sorted duplicate literals and the identity one,
// collapsing the supergate to zero when a false or complementary input occurs.
func Gia_ManSimplifyAnd(super *Vec_Int_t) {
	count, previous := 0, int32(-1)
	for _, literal := range Vec_IntArray(super) {
		if literal == 0 {
			Vec_IntFill(super, 1, 0)
			return
		}
		if literal == 1 {
			continue
		}
		if previous == -1 || previous>>1 != literal>>1 {
			Vec_IntWriteEntry(super, count, literal)
			count++
			previous = literal
		} else if previous != literal {
			Vec_IntFill(super, 1, 0)
			return
		}
	}
	Vec_IntShrink(super, count)
	if count == 0 {
		Vec_IntPush(super, 1)
	}
}

// giaSuperCollect walks in source fanin order, stopping at shared subgraphs and
// the source's 50-entry heuristic. Explicit frames avoid depth-limited stacks.
func giaSuperCollect(p *Gia_Man_t, start *Gia_Obj_t, strict int32, xor bool) {
	stack := []*Gia_Obj_t{start}
	for len(stack) > 0 {
		obj := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		regular := Gia_Regular(obj)
		stop := Gia_IsComplement(obj) != 0
		if xor && stop {
			panic("gia: complemented XOR supergate")
		}
		if xor {
			stop = Gia_ObjIsXor(regular) == 0
		} else {
			stop = stop || Gia_ObjIsAndReal(p, regular) == 0
		}
		if !stop {
			refs := Gia_ObjRefNum(p, regular)
			stop = strict != 0 && refs > 1 || refs > 2 || refs == 2 && (Gia_ObjRefNum(p, Gia_ObjFanin0(regular)) == 1 || Gia_ObjRefNum(p, Gia_ObjFanin1(regular)) == 1) || Vec_IntSize(p.vSuper) > 50
		}
		if stop {
			Vec_IntPush(p.vSuper, Gia_ObjToLit(p, obj))
			continue
		}
		if xor {
			if Gia_ObjFaninC0(regular) != 0 || Gia_ObjFaninC1(regular) != 0 {
				panic("gia: complemented XOR supergate fanin")
			}
			stack = append(stack, Gia_ObjFanin1(regular), Gia_ObjFanin0(regular))
		} else {
			stack = append(stack, Gia_ObjChild1(regular), Gia_ObjChild0(regular))
		}
	}
}

func Gia_ManSuperCollectXor_rec(p *Gia_Man_t, obj *Gia_Obj_t, strict int32) {
	giaSuperCollect(p, obj, strict, true)
}
func Gia_ManSuperCollectAnd_rec(p *Gia_Man_t, obj *Gia_Obj_t, strict int32) {
	giaSuperCollect(p, obj, strict, false)
}

func Gia_ManSuperCollect(p *Gia_Man_t, obj *Gia_Obj_t, strict int32) {
	if p.vSuper == nil {
		p.vSuper = Vec_IntAlloc(1000)
	} else {
		Vec_IntClear(p.vSuper)
	}
	switch {
	case Gia_ObjIsXor(obj) != 0:
		if Gia_ObjFaninC0(obj) != 0 || Gia_ObjFaninC1(obj) != 0 {
			panic("gia: complemented XOR supergate fanin")
		}
		Gia_ManSuperCollectXor_rec(p, Gia_ObjFanin0(obj), strict)
		Gia_ManSuperCollectXor_rec(p, Gia_ObjFanin1(obj), strict)
		Vec_IntSort(p.vSuper, 0)
		Gia_ManSimplifyXor(p.vSuper)
	case Gia_ObjIsAndReal(p, obj) != 0:
		Gia_ManSuperCollectAnd_rec(p, Gia_ObjChild0(obj), strict)
		Gia_ManSuperCollectAnd_rec(p, Gia_ObjChild1(obj), strict)
		Vec_IntSort(p.vSuper, 0)
		Gia_ManSimplifyAnd(p.vSuper)
	default:
		panic("gia: supergate root must be AND or XOR")
	}
}
