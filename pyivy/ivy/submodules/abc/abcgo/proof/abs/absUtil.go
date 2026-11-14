// Go port of absUtil.c by Alan Mishchenko, UC Berkeley (2005).
package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Abs_ParSetDefaults(p *Abs_Par_t) {
	*p = Abs_Par_t{NFramesPast: 4, nLearnedMax: 1000, NLearnedStart: 1000,
		NLearnedDelta: 200, NLearnedPerce: 70, NRatioMax: 30, FPropFanout: 1,
		IFrame: -1, iFrameProved: -1, NFramesNoChangeLim: 2}
}

// Gia_VtaConvertToGla counts frame occurrences of each abstracted object.
func Gia_VtaConvertToGla(p *gia.Gia_Man_t, vVta *vec.Vec_Int_t) *vec.Vec_Int_t {
	nObjs := gia.Gia_ManObjNum(p)
	nFrames := vec.Vec_IntEntry(vVta, 0)
	if nFrames < 0 || int(nFrames+1) >= vec.Vec_IntSize(vVta) || vec.Vec_IntEntry(vVta, int(nFrames+1)) != int32(vec.Vec_IntSize(vVta)) {
		panic("abs: invalid VTA frame offsets")
	}
	bits := util.Abc_Base2Log(uint32(nObjs))
	mask := int32((uint32(1) << uint(bits)) - 1)
	assert(nObjs <= mask, "VTA object count fits the source mask")
	out := vec.Vec_IntStart(int(nObjs))
	for _, entry := range vec.Vec_IntArray(vVta)[nFrames+2:] {
		id := entry & mask
		if id >= nObjs {
			panic("abs: invalid VTA object")
		}
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsRo(p, obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 && gia.Gia_ObjIsConst0(obj) == 0 {
			panic("abs: invalid abstraction object")
		}
		vec.Vec_IntAddToEntry(out, int(id), 1)
	}
	vec.Vec_IntWriteEntry(out, 0, nFrames)
	return out
}

// Gia_VtaConvertFromGla expands each GLA member into its frame records.
func Gia_VtaConvertFromGla(p *gia.Gia_Man_t, vGla *vec.Vec_Int_t, nFrames int32) *vec.Vec_Int_t {
	if nFrames < 0 {
		panic("abs: negative frame count")
	}
	nObjs := gia.Gia_ManObjNum(p)
	bits := util.Abc_Base2Log(uint32(nObjs))
	mask := int32((uint32(1) << uint(bits)) - 1)
	assert(nObjs <= mask, "VTA object count fits the source mask")
	nGlaSize := vec.Vec_IntSum(vGla)
	out := vec.Vec_IntAlloc(1000)
	vec.Vec_IntPush(out, nFrames)
	counter := nFrames + 2
	for i := int32(0); i <= nFrames; i++ {
		vec.Vec_IntPush(out, counter)
		counter += (i + 1) * nGlaSize
	}
	for i := int32(0); i < nFrames; i++ {
		for k := int32(0); k <= i; k++ {
			for id, entry := range vec.Vec_IntArray(vGla) {
				if entry != 0 {
					vec.Vec_IntPush(out, (k<<uint(bits))|int32(id))
				}
			}
		}
	}
	if vec.Vec_IntEntry(out, int(nFrames+1)) != int32(vec.Vec_IntSize(out)) {
		panic("abs: GLA counts do not match frame records")
	}
	return out
}

func Gia_FlaConvertToGla_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, out *vec.Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	vec.Vec_IntWriteEntry(out, int(gia.Gia_ObjId(p, obj)), 1)
	if gia.Gia_ObjIsRo(p, obj) != 0 {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND or register output")
	}
	id := gia.Gia_ObjId(p, obj)
	a, b, c, kind := p.Gate(id)
	if kind != 2 && kind != 3 && kind != 4 && kind != 5 {
		panic("abs: unsupported gate in flop abstraction")
	}
	Gia_FlaConvertToGla_rec(p, gia.Gia_ManObj(p, a>>1), out)
	if kind == 2 || kind == 3 || kind == 4 {
		Gia_FlaConvertToGla_rec(p, gia.Gia_ManObj(p, b>>1), out)
	}
	if kind == 4 {
		Gia_FlaConvertToGla_rec(p, gia.Gia_ManObj(p, c>>1), out)
	}
}

func Gia_FlaConvertToGla(p *gia.Gia_Man_t, fla *vec.Vec_Int_t) *vec.Vec_Int_t {
	gia.Gia_ManIncrementTravId(p)
	gia.Gia_ObjSetTravIdCurrent(p, gia.Gia_ManConst0(p))
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ObjSetTravIdCurrent(p, gia.Gia_ManPi(p, i))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(fla, int(i)) == 0 {
			gia.Gia_ObjSetTravIdCurrent(p, gia.Gia_ManRo(p, i))
		}
	}
	out := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	vec.Vec_IntWriteEntry(out, 0, 1)
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		Gia_FlaConvertToGla_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManPo(p, i)), out)
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(fla, int(i)) != 0 {
			Gia_FlaConvertToGla_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManRi(p, i)), out)
		}
	}
	return out
}

func Gia_GlaConvertToFla(p *gia.Gia_Man_t, gla *vec.Vec_Int_t) *vec.Vec_Int_t {
	out := vec.Vec_IntStart(int(gia.Gia_ManRegNum(p)))
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(gla, int(gia.Gia_ObjId(p, gia.Gia_ManRo(p, i)))) != 0 {
			vec.Vec_IntWriteEntry(out, int(i), 1)
		}
	}
	return out
}
func Gia_GlaCountFlops(p *gia.Gia_Man_t, gla *vec.Vec_Int_t) int32 {
	var count int32
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(gla, int(gia.Gia_ObjId(p, gia.Gia_ManRo(p, i)))) != 0 {
			count++
		}
	}
	return count
}
func Gia_GlaCountNodes(p *gia.Gia_Man_t, gla *vec.Vec_Int_t) int32 {
	var count int32
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, i)) != 0 && vec.Vec_IntEntry(gla, int(i)) != 0 {
			count++
		}
	}
	return count
}
