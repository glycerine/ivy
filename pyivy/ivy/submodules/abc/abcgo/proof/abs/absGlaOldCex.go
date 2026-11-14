package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_GlaPrepareCexAndMap maps the abstraction boundary to frame assignments.
func Gia_GlaPrepareCexAndMap(p *Gla_Man_t, cex **util.Abc_Cex_t, mapping **vec.Vec_Int_t) {
	boundary := vec.Vec_IntAlloc(1000)
	for _, id := range vec.Vec_IntArray(p.vAbs) {
		object := Gla_ManObj(p, id)
		if object.fConst == 0 && object.fRo == 0 && object.fAnd == 0 {
			panic("invalid GLA abstraction object")
		}
		for k := uint32(0); k < object.nFanins; k++ {
			fanin := Gla_ManObj(p, object.Fanins[k])
			if fanin.fAbs == 0 {
				vec.Vec_IntPush(boundary, fanin.iGiaObj)
			}
		}
	}
	vec.Vec_IntUniqify(boundary)
	frames := p.pPars.IFrame
	result := util.Abc_CexAlloc(0, int32(vec.Vec_IntSize(boundary)), frames+1)
	result.IFrame = frames
	for f := int32(0); f <= frames; f++ {
		for k, id := range vec.Vec_IntArray(boundary) {
			if Gla_ObjSatValue(p, id, f) != 0 {
				util.Abc_InfoSetBit(result.PData, f*result.NPis+int32(k))
			}
		}
	}
	*mapping, *cex = boundary, result
}

// Gla_ManDeriveCex copies assignments for real inputs into an ABC witness.
func Gla_ManDeriveCex(p *Gla_Man_t, inputs *vec.Vec_Int_t) *util.Abc_Cex_t {
	frame := p.pPars.IFrame
	result := util.Abc_CexAlloc(gia.Gia_ManRegNum(p.pGia), gia.Gia_ManPiNum(p.pGia), frame+1)
	result.IPo = 0
	result.IFrame = frame
	for _, id := range vec.Vec_IntArray(inputs) {
		object := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsPi(p.pGia, object) == 0 {
			continue
		}
		for f := int32(0); f <= frame; f++ {
			if Gla_ObjSatValue(p, id, f) != 0 {
				util.Abc_InfoSetBit(result.PData, result.NRegs+f*result.NPis+gia.Gia_ObjCioId(object))
			}
		}
	}
	return result
}
