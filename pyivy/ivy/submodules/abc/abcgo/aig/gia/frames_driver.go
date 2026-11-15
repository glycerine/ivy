package gia

import "fmt"

// Gia_ManFra_t owns the per-frame support sets of initialized unrolling.
type Gia_ManFra_t struct {
	pAig               *Gia_Man_t
	pPars              *Gia_ParFra_t
	vIns, vAnds, vOuts [][]int32
}

func Gia_ManFraStart(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_ManFra_t {
	return &Gia_ManFra_t{pAig: src, pPars: params}
}
func Gia_ManFraStop(p *Gia_ManFra_t) {
	if p != nil {
		*p = Gia_ManFra_t{}
	}
}
func Gia_ManFraSupports(p *Gia_ManFra_t) {
	n := int(p.pPars.NFrames)
	p.vIns, p.vAnds, p.vOuts = make([][]int32, n), make([][]int32, n), make([][]int32, n)
	var nextInputs []int32
	Gia_ManIncrementTravId(p.pAig)
	for frame := n - 1; frame >= 0; frame-- {
		roots := Gia_ManCollectPoIds(p.pAig)
		for _, id := range nextInputs {
			obj := p.pAig.pObjs[id]
			if Gia_ObjIsRo(p.pAig, obj) != 0 {
				Vec_IntPush(roots, Gia_ObjRoToRi(p.pAig, obj).id)
			}
		}
		inputs, nodes := Vec_IntAlloc(100), Vec_IntAlloc(100)
		Gia_ManCollectCis(p.pAig, Vec_IntArray(roots), int32(Vec_IntSize(roots)), inputs)
		// CollectAnds shares its caller's epoch. CollectCis just marked this
		// cone, so start a fresh epoch here rather than collecting no nodes.
		Gia_ManIncrementTravId(p.pAig)
		Gia_ManCollectAnds(p.pAig, Vec_IntArray(roots), int32(Vec_IntSize(roots)), nodes, nil)
		p.vIns[frame], p.vAnds[frame], p.vOuts[frame] = Vec_IntArray(inputs), Vec_IntArray(nodes), Vec_IntArray(roots)
		nextInputs = p.vIns[frame]
	}
}

func giaFramesCopyGate(dst, src *Gia_Man_t, obj *Gia_Obj_t, hash bool) int32 {
	read := func(lit int32) int32 { return Gia_ObjLitCopy(src, lit) }
	if hash {
		return giaHashMappedGate(dst, src, obj, read)
	}
	a, b, c, kind := src.Gate(obj.id)
	switch kind {
	case 2:
		return Gia_ManAppendAnd2(dst, read(a), read(b))
	case 3:
		return Gia_ManAppendXor2(dst, read(a), read(b))
	case 4:
		return Gia_ManAppendMux2(dst, read(a), read(b), read(c))
	case 5:
		return read(a)
	default:
		panic("gia: invalid frame gate")
	}
}

func giaFramesCleanup(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	before := Gia_ManAndNum(p)
	if Gia_ManCombMarkUsed(p) < before {
		q := Gia_ManDupMarked(p)
		Gia_ManStop(p)
		p = q
	}
	if verbose != 0 {
		fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", before, Gia_ManAndNum(p))
	}
	return p
}

func Gia_ManFramesInit(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	if src == nil || params == nil || params.NFrames <= 0 {
		return nil
	}
	p := Gia_ManFraStart(src, params)
	Gia_ManFraSupports(p)
	capacity := 0
	for i := range p.vIns {
		capacity += len(p.vIns[i]) + len(p.vAnds[i]) + len(p.vOuts[i])
	}
	q := Gia_ManStart(int32(max(1, capacity)))
	q.CopyIdentityFrom(src)
	Gia_ManHashAlloc(q)
	src.pObjs[0].Value = 0
	for frame := range p.vIns {
		if params.FVerbose != 0 {
			fmt.Printf("Frame %3d : CI = %6d. AND = %6d. CO = %6d.\n", frame, len(p.vIns[frame]), len(p.vAnds[frame]), len(p.vOuts[frame]))
		}
		for i := int32(0); i < Gia_ManPiNum(src); i++ {
			Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(q))
		}
		for _, id := range p.vIns[frame] {
			obj := src.pObjs[id]
			if Gia_ObjIsRo(src, obj) != 0 {
				obj.Value = 0
				if frame != 0 {
					obj.Value = Gia_ObjRoToRi(src, obj).Value
				}
			}
		}
		for _, id := range p.vAnds[frame] {
			src.pObjs[id].Value = uint32(giaFramesCopyGate(q, src, src.pObjs[id], true))
		}
		for _, id := range p.vOuts[frame] {
			obj := src.pObjs[id]
			lit := Gia_ObjFanin0Copy(obj)
			if Gia_ObjIsPo(src, obj) != 0 {
				lit = Gia_ManAppendCo(q, lit)
			}
			obj.Value = uint32(lit)
		}
	}
	Gia_ManFraStop(p)
	Gia_ManHashStop(q)
	return giaFramesCleanup(q, params.FVerbose)
}

func Gia_ManFrames(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	if src == nil || params == nil || params.NFrames <= 0 {
		return nil
	}
	if params.FInit != 0 {
		return Gia_ManFramesInit(src, params)
	}
	capacity := int64(params.NFrames) * int64(src.nObjs)
	if capacity > int64(objectNoFanin) {
		return nil
	}
	q := Gia_ManStart(int32(max(int64(1), capacity)))
	q.CopyIdentityFrom(src)
	hash := params.FDisableSt == 0
	if hash {
		Gia_ManHashAlloc(q)
	}
	var poLits []int32
	if params.FOrPos != 0 {
		poLits = make([]int32, Gia_ManPoNum(src))
	}
	src.pObjs[0].Value = 0
	for f := int32(0); f < params.NFrames; f++ {
		for i := int32(0); i < Gia_ManPiNum(src); i++ {
			Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(q))
		}
	}
	for f := int32(0); f < params.NFrames; f++ {
		for i := int32(0); i < src.nRegs; i++ {
			obj := Gia_ManRo(src, i)
			if f == 0 {
				obj.Value = uint32(Gia_ManAppendCi(q))
			} else {
				obj.Value = Gia_ObjRoToRi(src, obj).Value
			}
		}
		for i := int32(0); i < Gia_ManPiNum(src); i++ {
			Gia_ManPi(src, i).Value = uint32(Gia_ManCiLit(q, f*Gia_ManPiNum(src)+i))
		}
		for _, obj := range src.pObjs[1:] {
			if Gia_ObjIsAnd(obj) != 0 {
				obj.Value = uint32(giaFramesCopyGate(q, src, obj, hash))
			}
		}
		for i := int32(0); i < Gia_ManPoNum(src); i++ {
			obj := Gia_ManPo(src, i)
			lit := Gia_ObjFanin0Copy(obj)
			if poLits != nil {
				if hash {
					poLits[i] = Gia_ManHashOr(q, poLits[i], lit)
				} else {
					poLits[i] = Gia_ManAppendAnd2(q, poLits[i]^1, lit^1) ^ 1
				}
			} else {
				obj.Value = uint32(Gia_ManAppendCo(q, lit))
			}
		}
		if f == params.NFrames-1 && poLits != nil {
			for i, lit := range poLits {
				Gia_ManPo(src, int32(i)).Value = uint32(Gia_ManAppendCo(q, lit))
			}
		}
		for i := int32(0); i < src.nRegs; i++ {
			obj := Gia_ManRi(src, i)
			lit := Gia_ObjFanin0Copy(obj)
			if f == params.NFrames-1 {
				lit = Gia_ManAppendCo(q, lit)
			}
			obj.Value = uint32(lit)
		}
	}
	if hash {
		Gia_ManHashStop(q)
	}
	Gia_ManSetRegNum(q, src.nRegs)
	out := giaFramesCleanup(q, params.FVerbose)
	// These are the same logical registers, with a multi-frame transition.
	// Preserve Go's explicit reset cube after any structural cleanup.
	out.aigerLatchInits = append([]int32(nil), src.aigerLatchInits...)
	return out
}

// Gia_ManFramesInitSpecial returns state after the longest completed prefix
// whose hashed property literals are zero, stopping before the first other PO.
func Gia_ManFramesInitSpecial(src *Gia_Man_t, frames, verbose int32) *Gia_Man_t {
	if src == nil || frames < 0 {
		return nil
	}
	if frames > 0 {
		fmt.Printf("Computing specialized unrolling with %d frames...\n", frames)
	}
	q := Gia_ManStart(src.nObjs)
	q.CopyIdentityFrom(src)
	Gia_ManHashAlloc(q)
	src.pObjs[0].Value = 0
	// If the first property is nonzero, the preceding state is still zero;
	// do not return stale RI scratch left by an earlier graph operation.
	for i := int32(0); i < src.nRegs; i++ {
		Gia_ManRi(src, i).Value = 0
	}
	f := int32(0)
	for ; frames == 0 || f < frames; f++ {
		if verbose != 0 && f%100 == 0 {
			fmt.Printf("%6d : ", f)
			Gia_ManPrintStats(q, nil)
		}
		for i := int32(0); i < src.nRegs; i++ {
			obj := Gia_ManRo(src, i)
			obj.Value = 0
			if f != 0 {
				obj.Value = Gia_ObjRoToRi(src, obj).Value
			}
		}
		for i := int32(0); i < Gia_ManPiNum(src); i++ {
			Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(q))
		}
		for _, obj := range src.pObjs[1:] {
			if Gia_ObjIsAnd(obj) != 0 {
				obj.Value = uint32(giaFramesCopyGate(q, src, obj, true))
			}
		}
		nonzero := false
		for i := int32(0); i < Gia_ManPoNum(src); i++ {
			if Gia_ObjFanin0Copy(Gia_ManPo(src, i)) != 0 {
				nonzero = true
				break
			}
		}
		if nonzero {
			break
		}
		for i := int32(0); i < src.nRegs; i++ {
			obj := Gia_ManRi(src, i)
			obj.Value = uint32(Gia_ObjFanin0Copy(obj))
		}
	}
	if verbose != 0 {
		fmt.Printf("Computed prefix of %d frames.\n", f)
	}
	for i := int32(0); i < src.nRegs; i++ {
		Gia_ManAppendCo(q, int32(Gia_ManRi(src, i).Value))
	}
	Gia_ManHashStop(q)
	out := Gia_ManCleanup(q)
	if verbose != 0 {
		fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(q), Gia_ManAndNum(out))
	}
	Gia_ManStop(q)
	return out
}
