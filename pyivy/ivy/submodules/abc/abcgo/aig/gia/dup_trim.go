package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManPoIsToRemove(p *Gia_Man_t, obj *Gia_Obj_t, value int32) int32 {
	if Gia_ObjIsCo(obj) == 0 {
		panic("gia: expected output")
	}
	if value < -1 || value > 1 {
		panic("gia: invalid output value")
	}
	edge, _, _, _ := p.Gate(Gia_ObjId(p, obj))
	return giaBool(edge>>1 == 0 && (value == -1 || value == edge&1))
}

func Gia_ManDupTrimmedNonZero(src *Gia_Man_t) *Vec_Int_t {
	if Gia_ManPoNum(src)%2 != 0 {
		panic("gia: dual outputs require pairs")
	}
	miter := Gia_ManTransformMiter(src)
	swept := Gia_ManSeqStructSweep(miter, 1, 1, 0)
	Gia_ManStop(miter)
	defer Gia_ManStop(swept)
	result := vec.Vec_IntAlloc(int(Gia_ManPoNum(swept)))
	for i := int32(0); i < Gia_ManPoNum(swept); i++ {
		obj := Gia_ManPo(swept, i)
		if Gia_ObjFaninId0(obj, obj.id) != 0 {
			vec.Vec_IntPush(result, i)
		}
	}
	return result
}
func Gia_ManDupTrimmed(src *Gia_Man_t, trimCis, trimCos, dual, value int32) *Gia_Man_t {
	var nonzero *Vec_Int_t
	if dual != 0 && trimCos != 0 {
		nonzero = Gia_ManDupTrimmedNonZero(src)
	}
	refs := make([]int32, src.nObjs)
	for _, obj := range src.pObjs {
		for _, id := range giaObjectFanins(src, obj) {
			refs[id]++
		}
	}
	out := giaDupMappedStart(src)
	retained := int32(0)
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		if trimCis == 0 || refs[Gia_ManPi(src, i).id] > 0 {
			retained++
		}
	}
	if retained == 0 {
		Gia_ManAppendCi(out)
	}
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		obj := Gia_ManCi(src, i)
		if trimCis == 0 || refs[obj.id] > 0 || Gia_ObjIsRo(src, obj) != 0 {
			obj.Value = uint32(Gia_ManAppendCi(out))
		}
	}
	giaDupMappedNodes(out, src)
	if dual != 0 && trimCos != 0 {
		if vec.Vec_IntSize(nonzero) == 0 {
			Gia_ManAppendCo(out, 0)
			Gia_ManAppendCo(out, 0)
		} else {
			for _, index := range vec.Vec_IntArray(nonzero) {
				giaDupMappedOutput(out, src, Gia_ManPo(src, 2*index))
				giaDupMappedOutput(out, src, Gia_ManPo(src, 2*index+1))
			}
		}
	} else {
		count := int32(0)
		for i := int32(0); i < Gia_ManPoNum(src); i++ {
			obj := Gia_ManPo(src, i)
			if trimCos == 0 || Gia_ManPoIsToRemove(src, obj, value) == 0 {
				count++
			}
		}
		if count == 0 {
			Gia_ManAppendCo(out, 0)
		}
		for i := int32(0); i < Gia_ManPoNum(src); i++ {
			obj := Gia_ManPo(src, i)
			if trimCos == 0 || Gia_ManPoIsToRemove(src, obj, value) == 0 {
				giaDupMappedOutput(out, src, obj)
			}
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	if dual != 0 && trimCos != 0 {
		swept := Gia_ManSeqStructSweep(out, 1, 1, 0)
		Gia_ManStop(out)
		return swept
	}
	return out
}
func Gia_ManDupTrimmed2(src *Gia_Man_t) *Gia_Man_t {
	refs := make([]int32, src.nObjs)
	for _, obj := range src.pObjs {
		for _, id := range giaObjectFanins(src, obj) {
			refs[id]++
		}
	}
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		obj := Gia_ManPo(src, i)
		refs[Gia_ObjFaninId0(obj, obj.id)]--
	}
	out := giaDupMappedStart(src)
	retained := 0
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		if refs[Gia_ManPi(src, i).id] > 0 {
			retained++
		}
	}
	if retained == 0 {
		Gia_ManAppendCi(out)
	}
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		obj := Gia_ManCi(src, i)
		if refs[obj.id] > 0 || Gia_ObjIsRo(src, obj) != 0 {
			obj.Value = uint32(Gia_ManAppendCi(out))
		}
	}
	giaDupMappedNodes(out, src)
	keep := func(obj *Gia_Obj_t) bool {
		fanin := Gia_ObjFanin0(obj)
		return Gia_ObjIsConst0(fanin) == 0 && Gia_ObjIsPi(src, fanin) == 0
	}
	count := 0
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		if keep(Gia_ManPo(src, i)) {
			count++
		}
	}
	if count == 0 {
		Gia_ManAppendCo(out, 0)
	}
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		obj := Gia_ManPo(src, i)
		if keep(obj) {
			giaDupMappedOutput(out, src, obj)
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
