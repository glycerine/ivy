package sfm

// Port of the static inline helpers of src/opt/sfm/sfmInt.h.

func Sfm_NtkPiNum(p *Sfm_Ntk_t) int   { return p.Pis }
func Sfm_NtkPoNum(p *Sfm_Ntk_t) int   { return p.Pos }
func Sfm_NtkNodeNum(p *Sfm_Ntk_t) int { return len(p.Fanins) - p.Pis - p.Pos }

func Sfm_ObjIsPi(p *Sfm_Ntk_t, i int) bool   { return i < p.Pis }
func Sfm_ObjIsPo(p *Sfm_Ntk_t, i int) bool   { return i+p.Pos >= len(p.Fanins) }
func Sfm_ObjIsNode(p *Sfm_Ntk_t, i int) bool { return i >= p.Pis && i+p.Pos < len(p.Fanins) }
func Sfm_ObjIsFixed(p *Sfm_Ntk_t, i int) bool {
	return p.Fixed[i]
}
func Sfm_ObjAddsLevelArray(p []bool, i int) bool { return p == nil || !p[i] }
func Sfm_ObjAddsLevel(p *Sfm_Ntk_t, i int) bool  { return Sfm_ObjAddsLevelArray(p.Empty, i) }

func Sfm_ObjFaninNum(p *Sfm_Ntk_t, i int) int  { return len(p.Fanins[i]) }
func Sfm_ObjFanoutNum(p *Sfm_Ntk_t, i int) int { return len(p.Fanouts[i]) }

// Sfm_ObjRefIncrement and Sfm_ObjRefDecrement change the size of the fanout
// array in place, exactly as C's "++Sfm_ObjFoArray(p, iObj)->nSize". The
// entries beyond the size are kept in the slice capacity and are restored
// by the matching increment.
func Sfm_ObjRefIncrement(p *Sfm_Ntk_t, iObj int) int {
	a := p.Fanouts[iObj]
	p.Fanouts[iObj] = a[:len(a)+1]
	return len(a) + 1
}
func Sfm_ObjRefDecrement(p *Sfm_Ntk_t, iObj int) int {
	a := p.Fanouts[iObj]
	p.Fanouts[iObj] = a[:len(a)-1]
	return len(a) - 1
}
func Sfm_ObjFanin(p *Sfm_Ntk_t, i, k int) int  { return p.Fanins[i][k] }
func Sfm_ObjFanout(p *Sfm_Ntk_t, i, k int) int { return p.Fanouts[i][k] }

func Sfm_ObjSatVar(p *Sfm_Ntk_t, iObj int) int32 {
	assert(p.vId2Var[iObj] > 0, "Vec_IntEntry(&p->vId2Var, iObj) > 0")
	return p.vId2Var[iObj]
}
func Sfm_ObjSetSatVar(p *Sfm_Ntk_t, iObj int, Num int32) {
	assert(p.vId2Var[iObj] == -1, "Vec_IntEntry(&p->vId2Var, iObj) == -1")
	p.vId2Var[iObj] = Num
	p.vVar2Id[Num] = iObj
}
func Sfm_ObjCleanSatVar(p *Sfm_Ntk_t, Num int32) {
	iObj := p.vVar2Id[Num]
	assert(p.vId2Var[iObj] > 0, "Vec_IntEntry(&p->vId2Var, iObj) > 0")
	p.vId2Var[iObj] = -1
	p.vVar2Id[Num] = -1
}
func Sfm_NtkCleanVars(p *Sfm_Ntk_t) {
	for i := int32(1); i < p.nSatVars; i++ {
		if p.vVar2Id[i] != -1 {
			Sfm_ObjCleanSatVar(p, i)
		}
	}
}
func Sfm_ObjLevel(p *Sfm_Ntk_t, iObj int) int          { return p.Levels[iObj] }
func Sfm_ObjSetLevel(p *Sfm_Ntk_t, iObj int, Lev int)  { p.Levels[iObj] = Lev }
func Sfm_ObjLevelR(p *Sfm_Ntk_t, iObj int) int         { return p.Reverse[iObj] }
func Sfm_ObjSetLevelR(p *Sfm_Ntk_t, iObj int, Lev int) { p.Reverse[iObj] = Lev }

// Sfm_ObjUpdateFaninCount mirrors Vec_IntAddToEntry(&p->vCounts, iObj, -1),
// which returns the new value.
func Sfm_ObjUpdateFaninCount(p *Sfm_Ntk_t, iObj int) int {
	p.vCounts[iObj]--
	return p.vCounts[iObj]
}
func Sfm_ObjResetFaninCount(p *Sfm_Ntk_t, iObj int) {
	p.vCounts[iObj] = Sfm_ObjFaninNum(p, iObj) - 1
}
