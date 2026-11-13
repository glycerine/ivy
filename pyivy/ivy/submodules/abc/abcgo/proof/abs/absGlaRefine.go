package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ga2_GlaPrepareCexAndMap(p *Ga2_Man_t, cex **util.Abc_Cex_t, mapping **vec.Vec_Int_t) {
	m := vec.Vec_IntAlloc(1000)
	for i := 1; i < vec.Vec_IntSize(p.vValues); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vValues, i))
		if Ga2_ObjIsAbs(p, obj) {
			continue
		}
		if gia.Gia_ObjPhase(obj) == 0 || !Ga2_ObjIsLeaf(p, obj) || gia.Gia_ObjIsAnd(obj) == 0 && gia.Gia_ObjIsCi(obj) == 0 {
			panic("abs: invalid GLA pseudo-input")
		}
		vec.Vec_IntPush(m, gia.Gia_ObjId(p.pGia, obj))
	}
	c := util.Abc_CexAlloc(0, int32(vec.Vec_IntSize(m)), p.pPars.iFrame+1)
	c.IFrame = p.pPars.iFrame
	for f := int32(0); f <= c.IFrame; f++ {
		for i := 0; i < vec.Vec_IntSize(m); i++ {
			obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(m, i))
			if Ga2_ObjSatValue(p, obj, f) != 0 {
				util.Abc_InfoSetBit(c.PData, f*c.NPis+int32(i))
			}
		}
	}
	*cex = c
	*mapping = m
}
func Ga2_ManRefine(p *Ga2_Man_t) *vec.Vec_Int_t {
	if p.pPars.fAddLayer != 0 {
		result := vec.Vec_IntAlloc(100)
		for i := 1; i < vec.Vec_IntSize(p.vValues); i++ {
			obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vValues, i))
			if Ga2_ObjIsAbs(p, obj) {
				continue
			}
			if gia.Gia_ObjPhase(obj) == 0 || !Ga2_ObjIsLeaf(p, obj) || gia.Gia_ObjIsAnd(obj) == 0 && gia.Gia_ObjIsCi(obj) == 0 {
				panic("abs: invalid GLA pseudo-input")
			}
			if gia.Gia_ObjIsPi(p.pGia, obj) == 0 {
				vec.Vec_IntPush(result, gia.Gia_ObjId(p.pGia, obj))
			}
		}
		p.nObjAdded += int32(vec.Vec_IntSize(result))
		return result
	}
	var cex *util.Abc_Cex_t
	var mapping *vec.Vec_Int_t
	Ga2_GlaPrepareCexAndMap(p, &cex, &mapping)
	result := Rnm_ManRefine(p.pRnm, cex, mapping, p.pPars.fPropFanout, p.pPars.fNewRefine, 1)
	if vec.Vec_IntSize(result) == 0 {
		p.pGia.ClearSequentialCounterexample()
		actual := Ga2_ManDeriveCex(p, mapping)
		if err := p.pGia.SetSequentialCounterexample(actual); err != nil {
			panic(err)
		}
		return nil
	}
	k := 0
	for i := 0; i < vec.Vec_IntSize(result); i++ {
		id := vec.Vec_IntEntry(result, i)
		if !Ga2_ObjIsAbs(p, gia.Gia_ManObj(p.pGia, id)) {
			vec.Vec_IntWriteEntry(result, k, id)
			k++
		}
	}
	vec.Vec_IntShrink(result, k)
	for i := 0; i < k; i++ {
		if gia.Gia_ObjPhase(gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(result, i))) == 0 {
			panic("abs: unmarked refinement object")
		}
	}
	p.nObjAdded += int32(k)
	return result
}
