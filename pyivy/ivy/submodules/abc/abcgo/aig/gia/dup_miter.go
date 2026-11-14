package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManMiter_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	Gia_ManDupCofactorVar_rec(dst, src, obj)
	return int32(obj.Value)
}
func Gia_ManMiter(left, right *Gia_Man_t, duplicated, dual, seq, implic, verbose int32) *Gia_Man_t {
	var inputs, outputs int32
	if seq != 0 {
		if Gia_ManPiNum(left) != Gia_ManPiNum(right) || Gia_ManPoNum(left) != Gia_ManPoNum(right) || Gia_ManRegNum(left) == 0 || Gia_ManRegNum(right) == 0 {
			return nil
		}
		inputs = Gia_ManPiNum(left)
		outputs = Gia_ManPoNum(left)
	} else {
		if Gia_ManCiNum(left) != Gia_ManCiNum(right) || Gia_ManCoNum(left) != Gia_ManCoNum(right) {
			return nil
		}
		inputs = Gia_ManCiNum(left)
		outputs = Gia_ManCoNum(left)
	}
	if duplicated < 0 || duplicated > inputs {
		panic("gia: invalid duplicated input count")
	}
	out := Gia_ManStart(left.nObjs + right.nObjs)
	out.pName = "miter"
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(left)
	Gia_ManFillValue(right)
	left.pObjs[0].Value = 0
	right.pObjs[0].Value = 0
	for i := int32(0); i < inputs; i++ {
		Gia_ManCi(left, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < inputs; i++ {
		if i < inputs-duplicated {
			Gia_ManCi(right, i).Value = Gia_ManCi(left, i).Value
		} else {
			Gia_ManCi(right, i).Value = uint32(Gia_ManAppendCi(out))
		}
	}
	if seq != 0 {
		for _, p := range []*Gia_Man_t{left, right} {
			for i := int32(0); i < Gia_ManRegNum(p); i++ {
				Gia_ManRo(p, i).Value = uint32(Gia_ManAppendCi(out))
			}
		}
	}
	for i := int32(0); i < outputs; i++ {
		a, b := Gia_ManCo(left, i), Gia_ManCo(right, i)
		Gia_ManMiter_rec(out, left, Gia_ObjFanin0(a))
		Gia_ManMiter_rec(out, right, Gia_ObjFanin0(b))
		x, y := Gia_ObjFanin0Copy(a), Gia_ObjFanin0Copy(b)
		if dual != 0 {
			Gia_ManAppendCo(out, x)
			Gia_ManAppendCo(out, y)
		} else if implic != 0 {
			Gia_ManAppendCo(out, Gia_ManHashAnd(out, x, y^1))
		} else {
			Gia_ManAppendCo(out, Gia_ManHashXor(out, x, y))
		}
	}
	if seq != 0 {
		for _, p := range []*Gia_Man_t{left, right} {
			for i := int32(0); i < Gia_ManRegNum(p); i++ {
				obj := Gia_ManRi(p, i)
				Gia_ManMiter_rec(out, p, Gia_ObjFanin0(obj))
				giaDupMappedOutput(out, p, obj)
			}
		}
		Gia_ManSetRegNum(out, Gia_ManRegNum(left)+Gia_ManRegNum(right))
	}
	Gia_ManHashStop(out)
	// Carry each input graph's object copies through cleanup and normalization.
	clean := Gia_ManCleanup(out)
	remapCopies := func(source, intermediate *Gia_Man_t) {
		for _, obj := range source.pObjs {
			lit := Gia_ObjValue(obj)
			if lit < 0 {
				continue
			}
			mapped := Gia_ObjValue(Gia_ManObj(intermediate, lit>>1))
			if mapped < 0 {
				Gia_ObjSetValue(obj, -1)
			} else {
				Gia_ObjSetValue(obj, mapped^(lit&1))
			}
		}
	}
	remapCopies(left, out)
	remapCopies(right, out)
	Gia_ManStop(out)
	normalized := Gia_ManDupOrderAiger(clean)
	remapCopies(left, clean)
	remapCopies(right, clean)
	Gia_ManStop(clean)
	return normalized
}
func Gia_ManDupAndOr(src *Gia_Man_t, count, useOr, compl int32) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 || count < 0 {
		panic("gia: invalid output reduction")
	}
	out := giaQuantStart(src)
	giaComposeNodes(out, src, true)
	result := int32(1)
	if useOr != 0 {
		result = 0
	}
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		lit := giaMappedEdge(src, Gia_ManPo(src, i))
		if useOr != 0 {
			result = Gia_ManHashOr(out, result, lit)
		} else {
			result = Gia_ManHashAnd(out, result, lit)
		}
	}
	if compl > 0 {
		result ^= 1
	}
	for i := int32(0); i < count; i++ {
		Gia_ManAppendCo(out, result)
	}
	Gia_ManHashStop(out)
	return giaCleanComposed(out)
}
func Gia_ManMiterNames(names *Vec_Ptr_t, outputs int32) *Vec_Ptr_t {
	if outputs < 0 || outputs%2 != 0 || int(outputs) > vec.Vec_PtrSize(names) {
		panic("gia: invalid output names")
	}
	result := vec.Vec_PtrAlloc(vec.Vec_PtrSize(names) - int(outputs/2))
	for i := int32(0); i < outputs; i += 2 {
		vec.Vec_PtrPush(result, fmt.Sprintf("%s_xor_%s", vec.Vec_PtrEntry(names, int(i)).(string), vec.Vec_PtrEntry(names, int(i+1)).(string)))
	}
	for i := int(outputs); i < vec.Vec_PtrSize(names); i++ {
		vec.Vec_PtrPush(result, vec.Vec_PtrEntry(names, i))
	}
	return result
}

// mode 0 pairs adjacent outputs; mode 1 pairs the first and second halves;
// mode 2 appends false after each output; mode 3 interleaves the two halves.
func giaTransformPairs(src *Gia_Man_t, mode int) *Gia_Man_t {
	count := Gia_ManPoNum(src)
	if mode != 2 && count%2 != 0 {
		panic("gia: output pairs require an even output count")
	}
	out := giaQuantStart(src)
	giaComposeNodes(out, src, true)
	half := count / 2
	switch mode {
	case 0, 1:
		for i := int32(0); i < half; i++ {
			a, b := 2*i, 2*i+1
			if mode == 1 {
				a, b = i, half+i
			}
			Gia_ManAppendCo(out, Gia_ManHashXor(out, giaMappedEdge(src, Gia_ManPo(src, a)), giaMappedEdge(src, Gia_ManPo(src, b))))
		}
	case 2:
		for i := int32(0); i < count; i++ {
			Gia_ManAppendCo(out, giaMappedEdge(src, Gia_ManPo(src, i)))
			Gia_ManAppendCo(out, 0)
		}
	case 3:
		for i := int32(0); i < half; i++ {
			Gia_ManAppendCo(out, giaMappedEdge(src, Gia_ManPo(src, i)))
			Gia_ManAppendCo(out, giaMappedEdge(src, Gia_ManPo(src, half+i)))
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	if mode != 2 {
		out = giaCleanComposed(out)
	}
	return out
}
func Gia_ManTransformMiter(src *Gia_Man_t) *Gia_Man_t {
	out := giaTransformPairs(src, 0)
	if src.VNamesIn != nil {
		out.VNamesIn = vec.Vec_PtrDupStr(src.VNamesIn)
	}
	if src.VNamesOut != nil {
		out.VNamesOut = Gia_ManMiterNames(src.VNamesOut, Gia_ManPoNum(src))
	}
	return out
}
func Gia_ManTransformMiter2(src *Gia_Man_t) *Gia_Man_t             { return giaTransformPairs(src, 1) }
func Gia_ManTransformToDual(src *Gia_Man_t) *Gia_Man_t             { return giaTransformPairs(src, 2) }
func Gia_ManTransformTwoWord2DualOutput(src *Gia_Man_t) *Gia_Man_t { return giaTransformPairs(src, 3) }

func Gia_ManCollectOneSide_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) == 0 {
		return
	}
	for _, id := range giaObjectFanins(p, obj) {
		Gia_ManCollectOneSide_rec(p, Gia_ManObj(p, id), nodes)
	}
	vec.Vec_IntPush(nodes, obj.id)
}
func Gia_ManCollectOneSide(p *Gia_Man_t, side int32) *Vec_Int_t {
	if side != 0 && side != 1 {
		panic("gia: invalid dual-output side")
	}
	nodes := vec.Vec_IntAlloc(int(Gia_ManAndNum(p)))
	Gia_ManIncrementTravId(p)
	for i := side; i < Gia_ManPoNum(p); i += 2 {
		Gia_ManCollectOneSide_rec(p, Gia_ObjFanin0(Gia_ManPo(p, i)), nodes)
	}
	return nodes
}
func Gia_ManTransformDualOutput(src *Gia_Man_t) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 || Gia_ManPoNum(src)%2 != 0 {
		panic("gia: invalid dual-output graph")
	}
	a, b := Gia_ManCollectOneSide(src, 0), Gia_ManCollectOneSide(src, 1)
	swap := int32(0)
	if vec.Vec_IntSize(a) > vec.Vec_IntSize(b) {
		a, b = b, a
		swap = 1
	}
	out := giaQuantStart(src)
	for _, list := range []*Vec_Int_t{a, b} {
		for _, id := range vec.Vec_IntArray(list) {
			obj := src.pObjs[id]
			x, y, z, kind := src.Gate(id)
			read := func(lit int32) int32 { return Gia_ObjLitCopy(src, lit) }
			var result int32
			switch kind {
			case 2:
				result = Gia_ManHashAnd(out, read(x), read(y))
			case 3:
				result = Gia_ManHashXor(out, read(x), read(y))
			case 4:
				result = Gia_ManHashMux(out, read(x), read(y), read(z))
			case 5:
				result = read(x)
			default:
				panic("gia: invalid dual cone gate")
			}
			obj.Value = uint32(result)
		}
	}
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		Gia_ManAppendCo(out, giaMappedEdge(src, Gia_ManPo(src, i^swap)))
	}
	Gia_ManHashStop(out)
	return giaCleanComposed(out)
}
