package gia

import (
	"fmt"
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManUnr_t owns incremental frame-building state. The returned pNew graph
// is owned by the caller; stopping this manager never destroys that graph.
type Gia_ManUnr_t struct {
	pAig    *Gia_Man_t
	pPars   *Gia_ParFra_t
	pNew    *Gia_Man_t
	state   []int32
	frame   int32
	LastLit int32
	stopped bool
}

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

// Gia_ManUnrollStart prepares incremental frame construction. With
// fSaveLastLit, primary inputs are appended as each frame is requested.
func Gia_ManUnrollStart(src *Gia_Man_t, params *Gia_ParFra_t) any {
	if src == nil || params == nil || Gia_ManRegNum(src) < 0 || params.NFrames < 0 {
		return nil
	}
	reserveFrames := int64(maxInt(1, int(params.NFrames)))
	if params.fSaveLastLit != 0 {
		// Incremental callers can converge long before the requested bound.
		reserveFrames = 1
	}
	capacity := int64(Gia_ManObjNum(src)) * reserveFrames
	if capacity > int64(objectNoFanin) {
		return nil
	}
	if capacity < 1 {
		capacity = 1
	}
	manager := &Gia_ManUnr_t{pAig: src, pPars: params, LastLit: 0}
	manager.pNew = Gia_ManStart(int32(capacity))
	manager.pNew.pName, manager.pNew.PSpec = src.pName, src.PSpec
	Gia_ManHashStart(manager.pNew)
	nPis, nRegs := Gia_ManPiNum(src), Gia_ManRegNum(src)
	if params.fSaveLastLit == 0 {
		for i := int32(0); i < params.NFrames*nPis; i++ {
			Gia_ManAppendCi(manager.pNew)
		}
	}
	manager.state = make([]int32, nRegs)
	for i := int32(0); i < nRegs; i++ {
		if params.FInit != 0 {
			manager.state[i] = 0
		} else {
			manager.state[i] = Gia_ManAppendCi(manager.pNew)
		}
	}
	return manager
}

// Gia_ManUnrollAdd appends frames until fMax and returns the live graph.
func Gia_ManUnrollAdd(raw any, fMax int32) any {
	manager, ok := raw.(*Gia_ManUnr_t)
	if !ok || manager == nil || manager.stopped || manager.pAig == nil || manager.pNew == nil || fMax < manager.frame {
		return nil
	}
	if manager.pPars.NFrames > 0 && fMax > manager.pPars.NFrames {
		return nil
	}
	src, dst := manager.pAig, manager.pNew
	nPis, nRegs := Gia_ManPiNum(src), Gia_ManRegNum(src)
	for manager.frame < fMax {
		frame := manager.frame
		frameInputs := make([]int32, nPis)
		for i := int32(0); i < nPis; i++ {
			if manager.pPars.fSaveLastLit != 0 {
				frameInputs[i] = Gia_ManAppendCi(dst)
			} else {
				frameInputs[i] = Gia_Obj2Lit(dst, Gia_ManPi(dst, frame*nPis+i))
			}
		}
		mapping := make([]int32, src.nObjs)
		for i := range mapping {
			mapping[i] = -1
		}
		mapping[0] = 0
		for i := int32(0); i < Gia_ManCiNum(src); i++ {
			obj := Gia_ManCi(src, i)
			id := Gia_ObjId(src, obj)
			if i < nPis {
				mapping[id] = frameInputs[i]
			} else {
				mapping[id] = manager.state[i-nPis]
			}
		}
		and, xor, mux := Gia_ManHashAnd, Gia_ManHashXor, Gia_ManHashMux
		if manager.pPars.FDisableSt != 0 {
			and, xor, mux = Gia_ManAppendAnd2, Gia_ManAppendXor2, Gia_ManAppendMux2
		}
		for id := int32(1); id < src.nObjs; id++ {
			obj := Gia_ManObj(src, id)
			if Gia_ObjIsAnd(obj) != 0 {
				a, b, c, kind := src.Gate(id)
				a = remapGiaLiteral(mapping, a)
				switch kind {
				case 2:
					mapping[id] = and(dst, a, remapGiaLiteral(mapping, b))
				case 3:
					mapping[id] = xor(dst, a, remapGiaLiteral(mapping, b))
				case 4:
					mapping[id] = mux(dst, a, remapGiaLiteral(mapping, b), remapGiaLiteral(mapping, c))
				case 5:
					mapping[id] = a
				default:
					panic("gia: unsupported unrolling gate")
				}
			}
		}
		nextState := make([]int32, nRegs)
		manager.LastLit = 0
		for i := int32(0); i < Gia_ManCoNum(src); i++ {
			obj := Gia_ManCo(src, i)
			lit := remapGiaLiteral(mapping, Gia_ObjFaninLit0(obj, Gia_ObjId(src, obj)))
			if i < Gia_ManPoNum(src) {
				if i == 0 {
					manager.LastLit = lit
				}
				if manager.pPars.fSaveLastLit == 0 {
					Gia_ManAppendCo(dst, lit)
				}
			} else {
				nextState[i-Gia_ManPoNum(src)] = lit
			}
		}
		manager.state = nextState
		manager.frame++
	}
	return manager.pNew
}

// Gia_ManUnrollLastLit reads the current frame's first property literal.
func Gia_ManUnrollLastLit(raw any) int32 {
	manager, ok := raw.(*Gia_ManUnr_t)
	if !ok || manager == nil || manager.stopped {
		return -1
	}
	return manager.LastLit
}

// Gia_ManUnrollStop releases lifecycle bookkeeping but leaves the result graph
// alive for its caller, matching the C ownership boundary.
func Gia_ManUnrollStop(raw any) {
	manager, ok := raw.(*Gia_ManUnr_t)
	if !ok || manager == nil || manager.stopped {
		return
	}
	manager.stopped = true
	manager.state = nil
	manager.pAig, manager.pPars, manager.pNew = nil, nil, nil
}

// Gia_ManUnroll builds a complete combinational unrolling.
func Gia_ManUnroll(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	raw := Gia_ManUnrollStart(src, params)
	if raw == nil {
		return nil
	}
	manager := raw.(*Gia_ManUnr_t)
	result, ok := Gia_ManUnrollAdd(raw, params.NFrames).(*Gia_Man_t)
	if !ok || result == nil {
		Gia_ManUnrollStop(raw)
		return nil
	}
	Gia_ManHashStop(result)
	result.pName, result.PSpec = src.pName, src.PSpec
	Gia_ManUnrollStop(raw)
	_ = manager
	return result
}

// Gia_ManFrames2 adapts the historical external entry point to the native AIG
// frame builder, then exports the resulting sequential graph to GIA.
func Gia_ManFrames2(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	if src == nil || params == nil || params.NFrames <= 0 {
		return nil
	}
	return buildGiaFrames(src, params)
}

func Gia_ManFraSetDefaultParams(params *Gia_ParFra_t) {
	if params == nil {
		return
	}
	*params = Gia_ParFra_t{NFrames: 32}
}

func Gia_ManFrames(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	if src == nil || params == nil || params.NFrames <= 0 {
		return nil
	}
	return buildGiaFrames(src, params)
}

func Gia_ManFramesInit(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	if src == nil || params == nil || params.NFrames <= 0 {
		return nil
	}
	copyParams := *params
	copyParams.FInit = 1
	return buildGiaFrames(src, &copyParams)
}

func buildGiaFrames(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	graph := Gia_ManToAigSimple(src)
	withRegisters := int32(0)
	if params.FInit == 0 {
		withRegisters = 1
	}
	unrolled := aig.Aig_ManFrames(graph, params.NFrames, params.FInit, 1, withRegisters, 0, nil)
	result := Gia_ManFromAigSimple(unrolled)
	result.pName, result.PSpec = src.pName, src.PSpec
	if params.FVerbose != 0 {
		fmt.Printf("Unrolled %d frames: %d PIs, %d POs, %d registers.\n", params.NFrames, Gia_ManPiNum(result), Gia_ManPoNum(result), Gia_ManRegNum(result))
	}
	return result
}

func Gia_ManFramesInitSpecial(src *Gia_Man_t, frames, verbose int32) *Gia_Man_t {
	if src == nil {
		return nil
	}
	params := Gia_ParFra_t{NFrames: frames, FInit: 1, FVerbose: verbose}
	return Gia_ManFramesInit(src, &params)
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
