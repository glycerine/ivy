package gia

import "fmt"

// Gia_ManDupZeroUndc changes 1-initialized register coordinates and replaces
// unknown initial states with fresh PIs selected only in the first frame.
func Gia_ManDupZeroUndc(src *Gia_Man_t, initial string, newPIs, simple, verbose int32) *Gia_Man_t {
	if len(initial) != int(Gia_ManRegNum(src)) || newPIs < 0 {
		panic("gia: invalid register initialization")
	}
	unknown := make([]int32, Gia_ManRegNum(src))
	count := Gia_ManPiNum(src)
	ones := 0
	for i, ch := range []byte(initial) {
		unknown[i] = -1
		switch ch {
		case 'x', 'X':
			unknown[i] = count
			count++
		case '1':
			ones++
		case '0':
		default:
			panic("gia: invalid initial state value")
		}
	}
	out := giaDupMappedStart(src)
	out.fGiaSimple = simple
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := Gia_ManPiNum(src); i < count; i++ {
		Gia_ManAppendCi(out)
	}
	for i := int32(0); i < newPIs; i++ {
		Gia_ManAppendCi(out)
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	reset := int32(-1)
	if count > Gia_ManPiNum(src) {
		reset = Gia_ManAppendCi(out)
	}
	refs := make([]int32, src.nObjs)
	for _, obj := range src.pObjs {
		for _, id := range giaObjectFanins(src, obj) {
			refs[id]++
		}
	}
	for i, ch := range []byte(initial) {
		obj := Gia_ManRo(src, int32(i))
		if ch == '1' {
			obj.Value ^= 1
		} else if unknown[i] >= 0 && refs[obj.id] > 0 {
			obj.Value = uint32(Gia_ManAppendMux(out, reset, int32(obj.Value), Gia_ObjToLit(out, Gia_ManCi(out, unknown[i]))))
		}
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	for i, ch := range []byte(initial) {
		ri := Gia_ManRi(src, int32(i))
		lit := giaMappedEdge(src, ri)
		if ch == '1' {
			lit ^= 1
		}
		ri.Value = uint32(Gia_ManAppendCo(out, lit))
	}
	if reset >= 0 {
		Gia_ManAppendCo(out, 1)
	}
	regs := Gia_ManRegNum(src)
	if reset >= 0 {
		regs++
	}
	Gia_ManSetRegNum(out, regs)
	if verbose != 0 {
		fmt.Printf("Converted %d 1-valued FFs and %d DC-valued FFs.\n", ones, count-Gia_ManPiNum(src))
	}
	return out
}
func Gia_ManMiter2(src *Gia_Man_t, initial string, verbose int32) *Gia_Man_t {
	if len(initial) != int(Gia_ManCiNum(src)) {
		panic("gia: invalid miter initialization")
	}
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		if initial[i] != 'x' && initial[i] != 'X' {
			panic("gia: invalid miter PI initialization")
		}
	}
	undc := Gia_ManDupZeroUndc(src, initial[Gia_ManPiNum(src):], 0, 0, verbose)
	defer Gia_ManStop(undc)
	piInit := []byte(initial[:Gia_ManPiNum(src)])
	for _, ch := range []byte(initial[Gia_ManPiNum(src):]) {
		if ch == 'x' || ch == 'X' {
			piInit = append(piInit, ch)
		}
	}
	out := Gia_ManStart(undc.nObjs)
	out.CopyIdentityFrom(undc)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(undc)
	undc.pObjs[0].Value = 0
	firstInputs := make([]int32, Gia_ManPiNum(undc))
	for i := range firstInputs {
		firstInputs[i] = Gia_ManAppendCi(out)
		Gia_ManPi(undc, int32(i)).Value = uint32(firstInputs[i])
	}
	secondInputs := make([]int32, len(firstInputs))
	for i, ch := range piInit {
		if ch == 'x' {
			secondInputs[i] = firstInputs[i]
		} else {
			secondInputs[i] = Gia_ManAppendCi(out)
		}
	}
	for i := int32(0); i < Gia_ManRegNum(undc); i++ {
		Gia_ManRo(undc, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaComposeNodes(out, undc, true)
	firstOutputs := make([]int32, Gia_ManCoNum(undc))
	for i := range firstOutputs {
		firstOutputs[i] = giaMappedEdge(undc, Gia_ManCo(undc, int32(i)))
	}
	for i, lit := range secondInputs {
		Gia_ManPi(undc, int32(i)).Value = uint32(lit)
	}
	for i := int32(0); i < Gia_ManRegNum(undc); i++ {
		Gia_ManRo(undc, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaComposeNodes(out, undc, true)
	secondOutputs := make([]int32, Gia_ManCoNum(undc))
	for i := range secondOutputs {
		secondOutputs[i] = giaMappedEdge(undc, Gia_ManCo(undc, int32(i)))
	}
	for i := int32(0); i < Gia_ManPoNum(undc); i++ {
		Gia_ManAppendCo(out, Gia_ManHashXor(out, firstOutputs[i], secondOutputs[i]))
	}
	for _, values := range [][]int32{firstOutputs, secondOutputs} {
		for i := int32(0); i < Gia_ManRegNum(undc); i++ {
			Gia_ManAppendCo(out, values[Gia_ManPoNum(undc)+i])
		}
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, 2*Gia_ManRegNum(undc))
	return giaCleanComposed(out)
}
