package gia

// Gia_ManSatPartCollectSuper gathers boundary literals of an AND supergate.
func Gia_ManSatPartCollectSuper(p *Gia_Man_t, obj *Gia_Obj_t, lits []int32, n *int32) {
	if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 != 0 {
		panic("gia: invalid SAT partition node")
	}
	id := Gia_ObjId(p, obj)
	for _, lit := range [2]int32{Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)} {
		fan := Gia_ManObj(p, lit>>1)
		if fan.fMark0 != 0 || lit&1 != 0 {
			lits[*n] = lit
			*n++
		} else {
			Gia_ManSatPartCollectSuper(p, fan, lits, n)
		}
	}
}
func Gia_ManSatPartCountClauses(p *Gia_Man_t, obj *Gia_Obj_t, onset, offset *int32) {
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid SAT partition node")
	}
	id := Gia_ObjId(p, obj)
	var on, off [2]int32
	for i, lit := range [2]int32{Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)} {
		fan := Gia_ManObj(p, lit>>1)
		if fan.fMark0 != 0 {
			on[i], off[i] = 1, 1
		} else {
			Gia_ManSatPartCountClauses(p, fan, &on[i], &off[i])
			if lit&1 != 0 {
				on[i], off[i] = off[i], on[i]
			}
		}
	}
	*onset = on[0] * on[1]
	*offset = off[0] + off[1]
}
func Gia_ManSatPartCount(p *Gia_Man_t, obj *Gia_Obj_t, leaves, nodes *int32) int32 {
	if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 != 0 {
		panic("gia: invalid SAT partition node")
	}
	*nodes++
	id := Gia_ObjId(p, obj)
	var levels [2]int32
	for i, lit := range [2]int32{Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)} {
		fan := Gia_ManObj(p, lit>>1)
		if fan.fMark0 != 0 {
			*leaves++
		} else {
			levels[i] = Gia_ManSatPartCount(p, fan, leaves, nodes) + (lit & 1)
		}
	}
	return max(levels[0], levels[1])
}
func Gia_ManSatPartCountNodes(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 != 0 {
		panic("gia: invalid SAT partition node")
	}
	count := int32(1)
	id := Gia_ObjId(p, obj)
	for _, lit := range [2]int32{Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)} {
		fan := Gia_ManObj(p, lit>>1)
		if fan.fMark0 == 0 {
			count += Gia_ManSatPartCountNodes(p, fan)
		}
	}
	return count
}
