package gia

func Gia_ObjIsMuxType(obj *Gia_Obj_t) int32 {
	if Gia_IsComplement(obj) != 0 {
		panic("gia: MUX recognition requires a regular node")
	}
	if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjFaninC0(obj) == 0 || Gia_ObjFaninC1(obj) == 0 {
		return 0
	}
	a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
	if Gia_ObjIsAnd(a) == 0 || Gia_ObjIsAnd(b) == 0 {
		return 0
	}
	left, right := [2]*Gia_Obj_t{Gia_ObjChild0(a), Gia_ObjChild1(a)}, [2]*Gia_Obj_t{Gia_ObjChild0(b), Gia_ObjChild1(b)}
	for _, x := range left {
		for _, y := range right {
			if Gia_Regular(x) == Gia_Regular(y) && Gia_IsComplement(x) != Gia_IsComplement(y) {
				return 1
			}
		}
	}
	return 0
}
func Gia_ObjRecognizeExor(obj *Gia_Obj_t, fanin0, fanin1 **Gia_Obj_t) int32 {
	if Gia_IsComplement(obj) != 0 {
		panic("gia: XOR recognition requires a regular node")
	}
	if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjFaninC0(obj) == 0 || Gia_ObjFaninC1(obj) == 0 {
		return 0
	}
	a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
	if Gia_ObjIsAnd(a) == 0 || Gia_ObjIsAnd(b) == 0 {
		return 0
	}
	if Gia_ObjFanin0(a) != Gia_ObjFanin0(b) || Gia_ObjFanin1(a) != Gia_ObjFanin1(b) || Gia_ObjFaninC0(a) == Gia_ObjFaninC0(b) || Gia_ObjFaninC1(a) == Gia_ObjFaninC1(b) {
		return 0
	}
	if fanin0 != nil {
		*fanin0 = Gia_ObjChild0(a)
	}
	if fanin1 != nil {
		*fanin1 = Gia_ObjChild1(a)
	}
	return 1
}
func Gia_ObjRecognizeMux(obj *Gia_Obj_t, then, otherwise **Gia_Obj_t) *Gia_Obj_t {
	if Gia_ObjIsMuxType(obj) == 0 {
		panic("gia: expected structural MUX")
	}
	a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
	left, right := [2]*Gia_Obj_t{Gia_ObjChild0(a), Gia_ObjChild1(a)}, [2]*Gia_Obj_t{Gia_ObjChild0(b), Gia_ObjChild1(b)}
	// Preserve the source's branch precedence for structures matching two pairs.
	for _, pair := range [][2]int{{1, 1}, {0, 0}, {0, 1}, {1, 0}} {
		i, j := pair[0], pair[1]
		if Gia_Regular(left[i]) != Gia_Regular(right[j]) || Gia_IsComplement(left[i]) == Gia_IsComplement(right[j]) {
			continue
		}
		if Gia_IsComplement(left[i]) != 0 {
			*then = Gia_Not(right[1-j])
			*otherwise = Gia_Not(left[1-i])
			return right[j]
		}
		*then = Gia_Not(left[1-i])
		*otherwise = Gia_Not(right[1-j])
		return left[i]
	}
	panic("gia: inconsistent structural MUX")
}
func Gia_ObjRecognizeMuxLits(p *Gia_Man_t, obj *Gia_Obj_t, then, otherwise *int32) int32 {
	var a, b *Gia_Obj_t
	control := Gia_ObjRecognizeMux(obj, &a, &b)
	*then, *otherwise = Gia_Obj2Lit(p, a), Gia_Obj2Lit(p, b)
	return Gia_Obj2Lit(p, control)
}
func Gia_ManCreateMuxRefs(p *Gia_Man_t) []int32 {
	result := make([]int32, len(p.pObjs))
	for _, obj := range p.pObjs {
		if Gia_ObjIsMux(p, obj) != 0 {
			result[Gia_ObjFaninId2(p, obj.id)]++
			continue
		}
		if Gia_ObjRecognizeExor(obj, nil, nil) != 0 || Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var a, b *Gia_Obj_t
		control := Gia_ObjRecognizeMux(obj, &a, &b)
		result[Gia_ObjId(p, Gia_Regular(control))]++
	}
	return result
}
