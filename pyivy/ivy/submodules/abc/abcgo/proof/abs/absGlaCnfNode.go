package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Ga2_ManAddToAbsOneStatic(p *Ga2_Man_t, obj *gia.Gia_Obj_t, frame, useID int32) {
	out := Ga2_ObjFindOrAddLit(p, obj, frame)
	if out <= 1 {
		panic("abs: invalid output literal")
	}
	isConst := gia.Gia_ObjIsConst0(obj) != 0
	isRo := gia.Gia_ObjIsRo(p.pGia, obj) != 0
	if !isConst && !isRo && gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: invalid abstracted object")
	}
	proofID := int32(-1)
	if useID != 0 {
		proofID = gia.Gia_ObjId(p.pGia, obj)
	}
	if isConst || frame == 0 && isRo {
		p.pSat.AddClause(proofID, bsat.Literal(out^1))
		return
	}
	vec.Vec_IntClear(p.vLits)
	leaves := Ga2_ObjLeaves(p.pGia, obj)
	sourceFrame := frame
	if isRo {
		sourceFrame--
	}
	static := true
	for i := 0; i < vec.Vec_IntSize(leaves); i++ {
		leaf := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(leaves, i))
		lit := Ga2_ObjFindOrAddLit(p, leaf, sourceFrame)
		vec.Vec_IntPush(p.vLits, lit)
		if lit < 2 {
			static = false
		}
	}
	if static || isRo {
		Ga2_ManCnfAddStatic(p.pSat, Ga2_ObjCnf0(p, obj), Ga2_ObjCnf1(p, obj), vec.Vec_IntArray(p.vLits), out, proofID)
	} else {
		truth := Ga2_ObjComputeTruthSpecial(p.pGia, obj, leaves, p.vLits)
		Ga2_ManCnfAddDynamic(p, int32(truth&0xffff), vec.Vec_IntArray(p.vLits), out, gia.Gia_ObjId(p.pGia, obj))
	}
}
