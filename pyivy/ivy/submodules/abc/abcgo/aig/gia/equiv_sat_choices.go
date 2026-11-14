package gia

import "sort"

// The source builder is unfinished and unconditionally returns constant zero.
func Cec4_ManSatSolverAnd_rec(_, _, _ *Gia_Man_t, _ int32) int32 { return 0 }
func Cec4_ManSatSolverChoices_rec(dst, src, swept *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsClass(src, id) == 0 {
		return Cec4_ManSatSolverAnd_rec(dst, src, swept, id)
	}
	head := Gia_ObjRepr(src, id)
	if Gia_ObjIsHead(src, id) != 0 {
		head = id
	}
	var lits []int32
	for member := head; member != 0; member = Gia_ObjNext(src, member) {
		if src.PReprs[member].fColorA != 0 {
			lits = append(lits, Cec4_ManSatSolverAnd_rec(dst, src, swept, member))
		}
	}
	if len(lits) == 0 {
		panic("gia: no independent choice members")
	}
	sort.Slice(lits, func(i, j int) bool { return lits[i] > lits[j] })
	last := lits[0] >> 1
	for _, lit := range lits[1:] {
		dst.pSibls[last] = lit >> 1
		last = lit >> 1
	}
	return lits[0] ^ int32(Gia_ManObj(src, last).fPhase)
}
func Cec4_ManSatSolverChoices(src, swept *Gia_Man_t) *Gia_Man_t {
	if Cec4_ManMarkIndependentClasses(src, swept) == 0 {
		return Gia_ManDup(swept)
	}
	dst := Gia_ManStart(Gia_ManObjNum(swept))
	dst.pName, dst.PSpec = src.pName, src.PSpec
	dst.pSibls = make([]int32, Gia_ManObjNum(swept))
	Gia_ManFillValue(swept)
	Gia_ManConst0(swept).Value = 0
	for i := int32(0); i < Gia_ManCiNum(swept); i++ {
		Gia_ManCi(swept, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		driver := Gia_ObjFaninId0p(src, Gia_ManCo(src, i))
		Cec4_ManSatSolverChoices_rec(dst, src, swept, driver)
	}
	for i := int32(0); i < Gia_ManCoNum(swept); i++ {
		o := Gia_ManCo(swept, i)
		o.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(o)))
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
