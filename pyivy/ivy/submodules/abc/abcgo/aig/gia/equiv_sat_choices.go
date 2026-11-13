package gia

import "sort"

// Cec4_ManSatSolverAnd_rec cannot reconstruct choices: the upstream routine
// is unfinished and returns constant zero, which can corrupt the result graph.
func Cec4_ManSatSolverAnd_rec(_, _, _ *Gia_Man_t, _ int32) int32 {
	panic("gia: Cec4 choice reconstruction is unavailable")
}
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
		if src.pReprs[member].fColorA != 0 {
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
	if Cec4_ManMarkIndependentClasses(src, swept) != 0 {
		panic("gia: Cec4 choice reconstruction is unavailable")
	}
	return Gia_ManDup(swept)
}
