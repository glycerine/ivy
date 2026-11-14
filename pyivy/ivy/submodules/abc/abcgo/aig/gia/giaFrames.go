package gia

import (
	"fmt"

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

// Gia_ManUnrollDup makes an ordered, object-preserving clone. vLimit receives
// the one conservative level boundary used by clients that inspect the order.
func Gia_ManUnrollDup(src *Gia_Man_t, limits *Vec_Int_t) *Gia_Man_t {
	if src == nil || limits == nil || Vec_IntSize(limits) != 0 {
		return nil
	}
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.PSpec = src.pName, src.PSpec
	Vec_IntPush(limits, 1)
	copyMap := make([]int32, src.nObjs)
	for i := range copyMap {
		copyMap[i] = -1
	}
	copyMap[0] = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		lit := Gia_ManAppendCi(dst)
		copyMap[Gia_ObjId(src, Gia_ManCi(src, i))] = lit
		Gia_ObjSetValue(Gia_ManObj(dst, lit>>1), Gia_ObjCioId(Gia_ManCi(src, i)))
	}
	for id := int32(1); id < src.nObjs; id++ {
		obj := Gia_ManObj(src, id)
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			a := remapGiaLiteral(copyMap, Gia_ObjFaninLit0(obj, id))
			b := remapGiaLiteral(copyMap, Gia_ObjFaninLit1(obj, id))
			copyMap[id] = Gia_ManAppendAnd(dst, a, b)
		case Gia_ObjIsCo(obj) != 0:
			copyMap[id] = Gia_ManAppendCo(dst, remapGiaLiteral(copyMap, Gia_ObjFaninLit0(obj, id)))
		}
		if copyMap[id] >= 0 {
			Gia_ObjSetValue(Gia_ManObj(dst, copyMap[id]>>1), id)
		}
	}
	Vec_IntPush(limits, Gia_ManObjNum(dst))
	Gia_ManSetRegNum(dst, 0)
	return dst
}

// Gia_ManUnrollAbs returns conservative reverse-time object sets. Each entry
// packs the distance from the requested frame in the high bits and the source
// object ID in the low bits; including the full cone is safe for abstraction.
func Gia_ManUnrollAbs(src *Gia_Man_t, frames int32) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(maxInt(0, int(frames)))
	if src == nil || frames <= 0 {
		return result
	}
	bits := int32(0)
	for n := Gia_ManObjNum(src); n > 1<<bits; bits++ {
	}
	for end := int32(0); end < frames; end++ {
		entries := vec.Vec_IntAlloc(0)
		for frame := int32(0); frame <= end; frame++ {
			for id := int32(1); id < src.nObjs; id++ {
				obj := Gia_ManObj(src, id)
				if Gia_ObjIsAnd(obj) != 0 || Gia_ObjIsRo(src, obj) != 0 {
					vec.Vec_IntPush(entries, ((end-frame)<<bits)|id)
				}
			}
		}
		vec.Vec_IntSort(entries, 1)
		vec.Vec_PtrPush(result, entries)
	}
	return result
}

// Gia_ManUnrollStart prepares incremental frame construction. With
// fSaveLastLit, primary inputs are appended as each frame is requested.
func Gia_ManUnrollStart(src *Gia_Man_t, params *Gia_ParFra_t) any {
	if src == nil || params == nil || Gia_ManRegNum(src) < 0 || params.NFrames < 0 {
		return nil
	}
	capacity := int64(Gia_ManObjNum(src)) * int64(maxInt(1, int(params.NFrames)))
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
		for id := int32(1); id < src.nObjs; id++ {
			obj := Gia_ManObj(src, id)
			if Gia_ObjIsAnd(obj) != 0 {
				a := remapGiaLiteral(mapping, Gia_ObjFaninLit0(obj, id))
				b := remapGiaLiteral(mapping, Gia_ObjFaninLit1(obj, id))
				if manager.pPars.FDisableSt != 0 {
					mapping[id] = Gia_ManAppendAnd2(dst, a, b)
				} else {
					mapping[id] = Gia_ManHashAnd(dst, a, b)
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
