package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ga2_ObjSatValue(p *Ga2_Man_t, obj *gia.Gia_Obj_t, frame int32) int32 {
	if gia.Gia_ObjIsConst0(obj) != 0 {
		panic("abs: constant has no SAT model value")
	}
	lit := Ga2_ObjFindLit(p, obj, frame)
	if lit < 0 || lit>>1 >= p.pSat.NumVars() {
		return 0
	}
	value, ok := p.pSat.Value(lit >> 1)
	if !ok {
		panic("abs: SAT model is unavailable")
	}
	result := lit & 1
	if value {
		result ^= 1
	}
	return result
}
func Ga2_ManDeriveCex(p *Ga2_Man_t, pis *vec.Vec_Int_t) *util.Abc_Cex_t {
	c := util.Abc_CexAlloc(gia.Gia_ManRegNum(p.pGia), gia.Gia_ManPiNum(p.pGia), p.pPars.IFrame+1)
	c.IPo = 0
	c.IFrame = p.pPars.IFrame
	for i := 0; i < vec.Vec_IntSize(pis); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(pis, i))
		if gia.Gia_ObjIsPi(p.pGia, obj) == 0 {
			continue
		}
		for frame := int32(0); frame <= c.IFrame; frame++ {
			if Ga2_ObjSatValue(p, obj, frame) != 0 {
				util.Abc_InfoSetBit(c.PData, c.NRegs+frame*c.NPis+gia.Gia_ObjCioId(obj))
			}
		}
	}
	return c
}
