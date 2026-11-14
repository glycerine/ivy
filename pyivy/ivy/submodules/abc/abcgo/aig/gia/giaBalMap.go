package gia

import "fmt"

// Window membership uses the manager traversal epoch. Mark0 is temporary and
// cleared before returning; Value retains the usual source-to-copy literals.
func Gia_ManHighlight_rec(p *Gia_Man_t, id int32) {
	stack := []int32{id}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, id)
		obj := p.pObjs[id]
		if Gia_ObjIsAnd(obj) != 0 {
			stack = append(stack, giaObjectFanins(p, obj)...)
		}
	}
}
func Gia_ManPrepareWin(p *Gia_Man_t, outputs *Vec_Int_t, inputsOut, outputsOut, gatesOut **Vec_Int_t, poOnly int32) {
	Gia_ManIncrementTravId(p)
	for _, index := range Vec_IntArray(outputs) {
		if index < 0 || index >= Gia_ManCoNum(p) {
			panic("gia: invalid window output index")
		}
		Gia_ManHighlight_rec(p, Gia_ObjFaninId0p(p, Gia_ManCo(p, index)))
	}
	Gia_ManCleanMark0(p)
	if poOnly != 0 {
		for _, index := range Vec_IntArray(outputs) {
			Gia_ObjFanin0(Gia_ManCo(p, index)).fMark0 = 1
		}
	} else {
		for _, obj := range p.pObjs[1:] {
			if Gia_ObjIsCi(obj) != 0 {
				continue
			}
			if Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsTravIdCurrentId(p, obj.id) == 0 {
				continue
			}
			for _, id := range giaObjectFanins(p, obj) {
				p.pObjs[id].fMark0 = 1
			}
		}
	}
	inputs, result, gates := Vec_IntAlloc(1000), Vec_IntAlloc(1000), Vec_IntAlloc(1000)
	for _, obj := range p.pObjs[1:] {
		if Gia_ObjIsTravIdCurrentId(p, obj.id) == 0 {
			continue
		}
		if Gia_ObjIsCi(obj) != 0 {
			Vec_IntPush(inputs, obj.id)
		} else if obj.fMark0 != 0 {
			Vec_IntPush(result, obj.id)
		}
		if Gia_ObjIsAnd(obj) != 0 {
			Vec_IntPush(gates, obj.id)
		}
	}
	// Standalone extraction must also expose a selected CI/constant driver. The C
	// implementation's else-if silently omitted these zero-depth outputs.
	if poOnly != 0 {
		for id, obj := range p.pObjs {
			if obj.fMark0 != 0 && (id == 0 || Gia_ObjIsCi(obj) != 0) {
				Vec_IntPush(result, int32(id))
			}
		}
	}
	Gia_ManCleanMark0(p)
	*inputsOut, *outputsOut, *gatesOut = inputs, result, gates
}
func giaWindowCopyGate(dst, src *Gia_Man_t, obj *Gia_Obj_t, hash bool) int32 {
	a, b, c, kind := src.Gate(obj.id)
	switch kind {
	case 2:
		if hash {
			return Gia_ManHashAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		}
		return Gia_ManAppendAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
	case 3:
		if hash {
			return Gia_ManHashXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		}
		return Gia_ManAppendXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
	case 4:
		if hash {
			return Gia_ManHashMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
		}
		return Gia_ManAppendMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
	case 5:
		return Gia_ManAppendBuf(dst, Gia_ObjLitCopy(src, a))
	default:
		panic("gia: unsupported window gate")
	}
}
func Gia_ManExtractWin(src *Gia_Man_t, outputs *Vec_Int_t, poOnly int32) *Gia_Man_t {
	var inputs, boundary, gates *Vec_Int_t
	Gia_ManPrepareWin(src, outputs, &inputs, &boundary, &gates, poOnly)
	dst := Gia_ManStart(int32(Vec_IntSize(inputs) + Vec_IntSize(boundary) + Vec_IntSize(gates) + 1))
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for _, id := range Vec_IntArray(inputs) {
		src.pObjs[id].Value = uint32(Gia_ManAppendCi(dst))
	}
	for _, id := range Vec_IntArray(gates) {
		src.pObjs[id].Value = uint32(giaWindowCopyGate(dst, src, src.pObjs[id], false))
	}
	for _, id := range Vec_IntArray(boundary) {
		Gia_ManAppendCo(dst, int32(src.pObjs[id].Value))
	}
	return dst
}
func Gia_ManInsertWin(src *Gia_Man_t, outputs *Vec_Int_t, window *Gia_Man_t) *Gia_Man_t {
	var inputs, boundary, gates *Vec_Int_t
	Gia_ManPrepareWin(src, outputs, &inputs, &boundary, &gates, 0)
	if Gia_ManCiNum(window) != int32(Vec_IntSize(inputs)) || Gia_ManCoNum(window) != int32(Vec_IntSize(boundary)) {
		panic("gia: replacement window interface mismatch")
	}
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil || window.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for _, id := range Vec_IntArray(src.vCis) {
		src.pObjs[id].Value = uint32(Gia_ManAppendCi(dst))
	}
	window.pObjs[0].Value = 0
	for i, id := range Vec_IntArray(window.vCis) {
		window.pObjs[id].Value = src.pObjs[Vec_IntEntry(inputs, i)].Value
	}
	Gia_ManHashAlloc(dst)
	for _, obj := range window.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(giaWindowCopyGate(dst, window, obj, true))
		}
	}
	for i, id := range Vec_IntArray(window.vCos) {
		src.pObjs[Vec_IntEntry(boundary, i)].Value = uint32(Gia_ObjFanin0Copy(window.pObjs[id]))
	}
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsTravIdCurrentId(src, obj.id) == 0 {
			obj.Value = uint32(giaWindowCopyGate(dst, src, obj, true))
		}
	}
	for _, id := range Vec_IntArray(src.vCos) {
		Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(src.pObjs[id]))
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyInterfaceMetadata(dst, src)
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
func Gia_ManFindLatest(p *Gia_Man_t, target, window int32) *Vec_Int_t {
	outputs := Vec_IntAlloc(1000)
	if Gia_ManHasMapping(p) != 0 {
		levels := make([]int32, p.nObjs)
		var maximum int32
		for id := int32(1); id < p.nObjs; id++ {
			if Gia_ObjIsLut(p, id) == 0 {
				continue
			}
			for _, fanin := range Gia_ObjLutFanins(p, id) {
				levels[id] = max(levels[id], levels[fanin])
			}
			levels[id]++
			maximum = max(maximum, levels[id])
		}
		if window != 0 {
			target = int32((1 - 0.01*float64(window)) * float64(maximum))
		}
		if maximum < target {
			fmt.Printf("The maximum mapped level (%d) is less than the target level (%d).\n", maximum, target)
		}
		for i, id := range Vec_IntArray(p.vCos) {
			if levels[Gia_ObjFaninId0(p.pObjs[id], id)] >= target {
				Vec_IntPush(outputs, int32(i))
			}
		}
	} else {
		maximum := Gia_ManLevelNum(p)
		if window != 0 {
			target = int32((1 - 0.01*float64(window)) * float64(maximum))
		}
		if maximum < target {
			fmt.Printf("The maximum AIG level (%d) is less than the target level (%d).\n", maximum, target)
		}
		for i, id := range Vec_IntArray(p.vCos) {
			if Gia_ObjLevel(p, p.pObjs[id]) >= target {
				Vec_IntPush(outputs, int32(i))
			}
		}
	}
	return outputs
}
func Gia_ManExtractWindow(p *Gia_Man_t, target, window, verbose int32) *Gia_Man_t {
	if (target == 0) == (window == 0) {
		panic("gia: choose target level or percentage window")
	}
	outputs := Gia_ManFindLatest(p, target, window)
	if verbose != 0 {
		fmt.Printf("Collected %d outputs to extract.\n", Vec_IntSize(outputs))
	}
	if Vec_IntSize(outputs) == 0 {
		return Gia_ManDup(p)
	}
	return Gia_ManExtractWin(p, outputs, 1)
}
