package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManQuantDupConeSupp_rec(q, p *Gia_Man_t, obj *Gia_Obj_t, cis, objs *Vec_Int_t, keep func(any, int32) int32, data any) {
	if Gia_ObjCopyArray(p, obj.id) >= 0 {
		return
	}
	if Gia_ObjIsCi(obj) != 0 {
		lit := Gia_ManAppendCi(q)
		Gia_ManQuantSetSuppZero(q)
		if keep(data, Gia_ObjCioId(obj)) == 0 {
			Gia_ManQuantSetSuppCi(q, Gia_ManObj(q, lit>>1))
		}
		Gia_ObjSetCopyArray(p, obj.id, lit)
		vec.Vec_IntPush(cis, obj.id)
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unsupported quantification cone")
	}
	Gia_ManQuantDupConeSupp_rec(q, p, Gia_ObjFanin0(obj), cis, objs, keep, data)
	Gia_ManQuantDupConeSupp_rec(q, p, Gia_ObjFanin1(obj), cis, objs, keep, data)
	a := Gia_ObjCopyArray(p, Gia_ObjFaninId0(obj, obj.id)) ^ Gia_ObjFaninC0(obj)
	b := Gia_ObjCopyArray(p, Gia_ObjFaninId1(obj, obj.id)) ^ Gia_ObjFaninC1(obj)
	Gia_ObjSetCopyArray(p, obj.id, Gia_ManHashAnd(q, a, b))
	vec.Vec_IntPush(objs, obj.id)
}
func Gia_ManQuantDupConeSupp(p *Gia_Man_t, lit int32, keep func(any, int32) int32, data any, outCis **Vec_Int_t, outLit *int32) *Gia_Man_t {
	root := Gia_ManObj(p, lit>>1)
	if Gia_ObjIsAnd(root) == 0 {
		panic("gia: quantification root must be AND")
	}
	vec.Vec_IntFillExtra(&p.vCopies, int(p.nObjs), -1)
	q := Gia_ManStart(1000)
	Gia_ManHashStart(q)
	Gia_ManQuantSetSuppStart(q)
	cis, objs := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	// Constant zero may occur in the cone and has a fixed copy.
	Gia_ObjSetCopyArray(p, 0, 0)
	Gia_ManQuantDupConeSupp_rec(q, p, root, cis, objs, keep, data)
	if outLit != nil {
		*outLit = Gia_ObjCopyArray(p, lit>>1) ^ (lit & 1)
	}
	Gia_ObjSetCopyArray(p, 0, -1)
	for _, id := range vec.Vec_IntArray(cis) {
		Gia_ObjSetCopyArray(p, id, -1)
	}
	for _, id := range vec.Vec_IntArray(objs) {
		Gia_ObjSetCopyArray(p, id, -1)
	}
	for i, id := range vec.Vec_IntArray(cis) {
		vec.Vec_IntWriteEntry(cis, i, Gia_ObjCioId(Gia_ManObj(p, id)))
	}
	if outCis != nil {
		*outCis = cis
	}
	return q
}
func Gia_ManQuantExist_rec(p *Gia_Man_t, id int32, result []int32) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		Gia_ObjCopyGetTwoArray(p, id, result)
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		result[0], result[1] = 0, 1
		Gia_ObjCopySetTwoArray(p, id, result)
		return
	}
	fans := [2]int32{Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)}
	complements := [2]int32{Gia_ObjFaninC0(obj), Gia_ObjFaninC1(obj)}
	var values [2][2]int32
	for i, fan := range fans {
		if Gia_ManQuantCheckSupp(p, fan, p.iSuppPi) != 0 {
			Gia_ManQuantExist_rec(p, fan, values[i][:])
		} else {
			values[i] = [2]int32{fan << 1, fan << 1}
		}
	}
	result[0] = Gia_ManHashAnd(p, values[0][0]^complements[0], values[1][0]^complements[1])
	result[1] = Gia_ManHashAnd(p, values[0][1]^complements[0], values[1][1]^complements[1])
	Gia_ObjCopySetTwoArray(p, id, result)
}
func Gia_ManQuantCollect_rec(p *Gia_Man_t, id int32, cis, side, ands *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	if Gia_ManQuantCheckOverlap(p, id) == 0 {
		vec.Vec_IntPush(side, id)
		return
	}
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		vec.Vec_IntPush(cis, id)
		return
	}
	Gia_ManQuantCollect_rec(p, Gia_ObjFaninId0(obj, id), cis, side, ands)
	Gia_ManQuantCollect_rec(p, Gia_ObjFaninId1(obj, id), cis, side, ands)
	vec.Vec_IntPush(ands, id)
}
func Gia_ManQuantCollect(p *Gia_Man_t, id int32, keep func(any, int32) int32, data any, cis, side, ands *Vec_Int_t) {
	Gia_ManQuantMarkUsedCis(p, keep, data)
	Gia_ManIncrementTravId(p)
	Gia_ManQuantCollect_rec(p, id, cis, side, ands)
}
func Gia_ManQuantExist2Dup(p *Gia_Man_t, lit int32, cis, side, ands *Vec_Int_t, outLit *int32) *Gia_Man_t {
	q := Gia_ManStart(int32(vec.Vec_IntSize(side) + vec.Vec_IntSize(cis) + 10*vec.Vec_IntSize(ands) + 1))
	Gia_ManQuantSetSuppStart(q)
	Gia_ManHashStart(q)
	vec.Vec_IntFillExtra(&p.vCopies, int(p.nObjs), -1)
	for _, id := range vec.Vec_IntArray(side) {
		Gia_ObjSetCopyArray(p, id, Gia_ManAppendCi(q))
		Gia_ManQuantSetSuppZero(q)
	}
	for _, id := range vec.Vec_IntArray(cis) {
		x := Gia_ManAppendCi(q)
		Gia_ObjSetCopyArray(p, id, x)
		Gia_ManQuantSetSuppZero(q)
		Gia_ManQuantSetSuppCi(q, Gia_ManObj(q, x>>1))
	}
	for _, id := range vec.Vec_IntArray(ands) {
		obj := Gia_ManObj(p, id)
		a := Gia_ObjCopyArray(p, Gia_ObjFaninId0(obj, id)) ^ Gia_ObjFaninC0(obj)
		b := Gia_ObjCopyArray(p, Gia_ObjFaninId1(obj, id)) ^ Gia_ObjFaninC1(obj)
		Gia_ObjSetCopyArray(p, id, Gia_ManHashAnd(q, a, b))
	}
	if outLit != nil {
		*outLit = Gia_ObjCopyArray(p, lit>>1) ^ (lit & 1)
	}
	for _, group := range []*Vec_Int_t{side, cis, ands} {
		for _, id := range vec.Vec_IntArray(group) {
			Gia_ObjSetCopyArray(p, id, -1)
		}
	}
	return q
}

func Gia_ManQuantExist2(p *Gia_Man_t, lit int32, keep func(any, int32) int32, data any) int32 {
	if lit < 2 {
		return lit
	}
	root := Gia_ManObj(p, lit>>1)
	if Gia_ObjIsCi(root) != 0 {
		if keep(data, Gia_ObjCioId(root)) != 0 {
			return lit
		}
		return 1
	}
	var cis *Vec_Int_t
	var out int32
	q := Gia_ManQuantDupConeSupp(p, lit, keep, data, &cis, &out)
	if q.iSuppPi == 0 {
		Gia_ManStop(q)
		return lit
	}
	current, next := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	vec.Vec_IntPush(current, out)
	for q.iSuppPi--; q.iSuppPi >= 0; q.iSuppPi-- {
		vec.Vec_IntFillExtra(&q.vCopiesTwo, int(2*q.nObjs), -1)
		vec.Vec_IntClear(next)
		Gia_ManIncrementTravId(q)
		for _, entry := range vec.Vec_IntArray(current) {
			var pair [2]int32
			Gia_ManQuantExist_rec(q, entry>>1, pair[:])
			for _, x := range pair {
				x ^= entry & 1
				if x == 0 {
					continue
				}
				if x == 1 {
					Gia_ManStop(q)
					return 1
				}
				vec.Vec_IntPushUnique(next, x)
			}
		}
		current, next = next, current
		if vec.Vec_IntSize(current) == 0 {
			Gia_ManStop(q)
			return 0
		}
	}
	for i, x := range vec.Vec_IntArray(current) {
		vec.Vec_IntWriteEntry(current, i, x^1)
	}
	out = Gia_ManHashAndMulti(q, current) ^ 1
	Gia_ManAppendCo(q, out)
	result := Gia_ManDupConeBack(p, q, cis)
	Gia_ManStop(q)
	return result
}
func Gia_ManQuantExistInt(p *Gia_Man_t, lit int32, cis, side, ands *Vec_Int_t) int32 {
	if lit < 2 {
		return lit
	}
	if vec.Vec_IntSize(cis) == 0 {
		return lit
	}
	if vec.Vec_IntSize(ands) == 0 {
		for _, id := range vec.Vec_IntArray(cis) {
			if id == lit>>1 {
				return 1
			}
		}
		return lit
	}
	var out int32
	q := Gia_ManQuantExist2Dup(p, lit, cis, side, ands, &out)
	vec.Vec_IntFillExtra(&q.vCopiesTwo, int(2*q.nObjs), -1)
	Gia_ObjCopySetTwoArray(q, 0, []int32{0, 0})
	for i := int32(0); i < Gia_ManCiNum(q); i++ {
		Gia_ObjCopySetTwoArray(q, i+1, []int32{(i + 1) << 1, (i + 1) << 1})
	}
	for i := Gia_ManCiNum(q) - 1; i >= int32(vec.Vec_IntSize(side)); i-- {
		Gia_ObjCopySetTwoArray(q, i+1, []int32{0, 1})
		q.iSuppPi--
		vec.Vec_IntFillExtra(&q.vCopiesTwo, int(2*q.nObjs), -1)
		Gia_ManIncrementTravId(q)
		var pair [2]int32
		Gia_ManQuantExist_rec(q, out>>1, pair[:])
		out = Gia_ManHashOr(q, pair[0]^(out&1), pair[1]^(out&1))
		Gia_ObjCopySetTwoArray(q, i+1, []int32{(i + 1) << 1, (i + 1) << 1})
	}
	originalSize := vec.Vec_IntSize(side)
	vec.Vec_IntAppend(side, cis)
	Gia_ManAppendCo(q, out)
	result := Gia_ManDupConeBackObjs(p, q, side)
	vec.Vec_IntShrink(side, originalSize)
	Gia_ManStop(q)
	return result
}
func Gia_ManQuantExist(p *Gia_Man_t, lit int32, keep func(any, int32) int32, data any) int32 {
	if lit < 2 {
		return lit
	}
	root := Gia_ManObj(p, lit>>1)
	if Gia_ObjIsCi(root) != 0 {
		if keep(data, Gia_ObjCioId(root)) == 0 {
			return 1
		}
		return lit
	}
	cis, side, ands := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	Gia_ManQuantCollect(p, lit>>1, keep, data, cis, side, ands)
	return Gia_ManQuantExistInt(p, lit, cis, side, ands)
}
