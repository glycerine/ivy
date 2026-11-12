package gia

import "github.com/glycerine/abcgo/sat/satoko"

// Gia_Min2ManSolve checks the conjunction using a Tseitin encoding. Unlike
// the C implementation's supergate optimization, encoding each AND separately
// preserves the same Boolean function without relying on pointer-tagged edges.
func Gia_Min2ManSolve(p *Gia_Man2Min_t) int32 {
	if p.iLits[0] < 2 || p.iLits[1] < 2 || p.pSat.VariableCount() != 0 {
		panic("gia: invalid support SAT state")
	}
	Vec_IntFillExtra(&p.pGia.vCopies, int(Gia_ManObjNum(p.pGia)), -1)
	Vec_IntClear(p.vSatVars)
	p.pSat.Bookmark()
	defer func() {
		p.pSat.Rollback()
		for _, id := range Vec_IntArray(p.vSatVars) {
			Gia_Min2ObjCleanSatId(p.pGia, Gia_ManObj(p.pGia, id))
		}
	}()
	var encode func(int32) int
	encode = func(id int32) int {
		obj := Gia_ManObj(p.pGia, id)
		if old := Gia_Min2ObjSatId(p.pGia, obj); old >= 0 {
			return int(old)
		}
		v := p.pSat.AddVariable(false)
		Vec_IntPush(p.vSatVars, id)
		Gia_Min2ObjSetSatId(p.pGia, obj, int32(v))
		if id == 0 {
			if !p.pSat.AddClause([]int{2*v + 1}) {
				panic("gia: inconsistent constant")
			}
		} else if Gia_ObjIsCi(obj) == 0 {
			if Gia_ObjIsAnd(obj) == 0 {
				panic("gia: invalid support SAT node")
			}
			a := encode(Gia_ObjFaninId0(obj, id))*2 + int(Gia_ObjFaninC0(obj))
			b := encode(Gia_ObjFaninId1(obj, id))*2 + int(Gia_ObjFaninC1(obj))
			if !p.pSat.AddClause([]int{a, 2*v + 1}) || !p.pSat.AddClause([]int{b, 2*v + 1}) || !p.pSat.AddClause([]int{a ^ 1, b ^ 1, 2 * v}) {
				panic("gia: inconsistent AND encoding")
			}
		}
		return v
	}
	assumptions := make([]int, 2)
	for i, lit := range p.iLits {
		assumptions[i] = 2*encode(lit>>1) + int(lit&1)
	}
	status := p.pSat.SolveAssumptions(assumptions)
	if status == satoko.SATOKO_SAT {
		if Gia_Min2ManSimulate(p) == 0 {
			panic("gia: SAT called despite a simulation witness")
		}
		for _, cis := range p.vCis {
			for _, id := range Vec_IntArray(cis) {
				value := p.pSat.ModelValue(int(Gia_Min2ObjSatId(p.pGia, Gia_ManObj(p.pGia, id))))
				Gia_Min2SimSetInputBit(p, id, giaBool(value == 1), p.iPattern)
			}
		}
		p.iPattern++
		if p.iPattern == 64 {
			p.iPattern = 1
		}
		p.nSatSat++
	}
	return giaBool(status == satoko.SATOKO_UNSAT)
}
