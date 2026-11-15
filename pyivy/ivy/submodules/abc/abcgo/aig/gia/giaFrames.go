package gia

import (
	"math/bits"

	"github.com/glycerine/abcgo/misc/vec"
)

// giaUnrollDupRec retains the source DFS order and publishes both directions:
// original Value is a copy literal; destination Value is the original ID.
func giaUnrollDupRec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	copyLit := func(lit int32) int32 {
		fanin := src.pObjs[lit>>1]
		giaUnrollDupRec(dst, src, fanin)
		return int32(fanin.Value) ^ (lit & 1)
	}
	a, b, c, kind := src.Gate(obj.id)
	var lit int32
	switch kind {
	case 1:
		lit = Gia_ManAppendCi(dst)
	case 2:
		lit = Gia_ManAppendAnd(dst, copyLit(a), copyLit(b))
	case 3:
		lit = Gia_ManAppendXorReal(dst, copyLit(a), copyLit(b))
	case 4:
		// Visit stored data pins before control, extending the two-pin C walk.
		otherwise, then := copyLit(c), copyLit(b)
		lit = Gia_ManAppendMuxReal(dst, copyLit(a), then, otherwise)
	case 5:
		lit = Gia_ManAppendBuf(dst, copyLit(a))
	case 6:
		lit = Gia_ManAppendCo(dst, copyLit(a))
	default:
		panic("gia: invalid object in unrolling order")
	}
	obj.Value = uint32(lit)
	dst.pObjs[lit>>1].Value = uint32(obj.id)
}

// Gia_ManUnrollDup orders PO cones, then the reached registers' RI cones in
// successive dependency tents. Unreachable objects are not copied.
func Gia_ManUnrollDup(src *Gia_Man_t, limits *Vec_Int_t) *Gia_Man_t {
	if src == nil || limits == nil || Vec_IntSize(limits) != 0 {
		return nil
	}
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	dst.fGiaSimple = src.fGiaSimple
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	Vec_IntPush(limits, dst.nObjs)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaUnrollDupRec(dst, src, Gia_ManPo(src, i))
	}
	Vec_IntPush(limits, dst.nObjs)
	for i := int32(1); i < dst.nObjs; i++ {
		if i == Vec_IntEntryLast(limits) {
			Vec_IntPush(limits, dst.nObjs)
		}
		original := src.pObjs[dst.pObjs[i].Value]
		if Gia_ObjIsRo(src, original) != 0 {
			ri := Gia_ObjRoToRi(src, original)
			if ri.Value != ^uint32(0) {
				panic("gia: register input already ordered")
			}
			giaUnrollDupRec(dst, src, ri)
		}
	}
	Gia_ManSetRegNum(dst, 0)
	return dst
}

// Gia_ManUnrollAbs packs the newly reached dependency tent of each earlier
// timeframe, retaining the source sparse reverse-time abstraction schedule.
func Gia_ManUnrollAbs(src *Gia_Man_t, frames int32) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(maxInt(0, int(frames)))
	if src == nil || frames <= 0 {
		return result
	}
	shift := uint(bits.Len(uint(src.nObjs - 1)))
	limits := Vec_IntAlloc(0)
	ordered := Gia_ManUnrollDup(src, limits)
	defer Gia_ManStop(ordered)
	for end := int32(0); end < frames; end++ {
		size := ordered.nObjs
		if end+1 < int32(Vec_IntSize(limits)) {
			size = Vec_IntEntry(limits, int(end+1))
		}
		row := Vec_IntAlloc(int(size))
		for frame := int32(0); frame <= end; frame++ {
			start, stop := int32(0), int32(0)
			if frame < int32(Vec_IntSize(limits)) {
				start = Vec_IntEntry(limits, int(frame))
			}
			if frame+1 < int32(Vec_IntSize(limits)) {
				stop = Vec_IntEntry(limits, int(frame+1))
			}
			for i := stop - 1; i >= start; i-- {
				original := src.pObjs[ordered.pObjs[i].Value]
				if Gia_ObjIsCo(original) != 0 || Gia_ObjIsPi(src, original) != 0 {
					continue
				}
				Vec_IntPush(row, ((end-frame)<<shift)|original.id)
			}
		}
		Vec_IntSort(row, 1)
		Vec_PtrPush(result, row)
	}
	return result
}

func Gia_ManFraSetDefaultParams(params *Gia_ParFra_t) {
	if params == nil {
		return
	}
	*params = Gia_ParFra_t{NFrames: 32}
}

func remapGiaLiteral(mapping []int32, literal int32) int32 {
	id := int(literal >> 1)
	if id < 0 || id >= len(mapping) || mapping[id] < 0 {
		panic("gia: unroll encountered an unmapped source object")
	}
	return mapping[id] ^ (literal & 1)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
