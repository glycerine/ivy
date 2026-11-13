package gia

// Gia_ManExtendCopy2Array adds unassigned entries for newly appended objects
// without changing SAT IDs or other copies already stored in the array.
func Gia_ManExtendCopy2Array(p *Gia_Man_t) {
	Vec_IntFillExtra(&p.vCopies2, int(p.nObjs), -1)
}
