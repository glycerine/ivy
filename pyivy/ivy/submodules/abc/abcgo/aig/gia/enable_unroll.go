package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManTransferFrames(aig, frames *Gia_Man_t, nFrames int32, newGraph *Gia_Man_t, signals *Vec_Int_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(100)
	for _, id := range vec.Vec_IntArray(signals) {
		obj := Gia_ManObj(aig, id)
		if Gia_ObjIsCand(obj) == 0 {
			panic("gia: expected candidate signal")
		}
		for frame := int32(0); frame < nFrames; frame++ {
			lit := Gia_ObjCopyF(aig, frame, obj)
			if lit < 0 {
				continue
			}
			mapped := Gia_ObjValue(Gia_ManObj(frames, lit>>1))
			if mapped != 0 && mapped != -1 {
				vec.Vec_IntPushUnique(result, mapped>>1)
			}
		}
	}
	return result
}
func Gia_ManUnrollInit(p *Gia_Man_t, nFrames int32) *Gia_Man_t {
	if nFrames <= 0 {
		panic("gia: frame count must be positive")
	}
	vec.Vec_IntFill(&p.vCopies, int(nFrames*p.nObjs), -1)
	q := Gia_ManStart(nFrames * p.nObjs)
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ManHashAlloc(q)
	for i := int32(0); i < p.nRegs; i++ {
		Gia_ObjSetCopyF(p, 0, Gia_ManRo(p, i), 0)
	}
	for frame := int32(0); frame < nFrames; frame++ {
		Gia_ObjSetCopyF(p, frame, Gia_ManConst0(p), 0)
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			Gia_ObjSetCopyF(p, frame, Gia_ManPi(p, i), Gia_ManAppendCi(q))
		}
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				Gia_ObjSetCopyF(p, frame, obj, Gia_ManHashAnd(q, Gia_ObjFanin0CopyF(p, frame, obj), Gia_ObjFanin1CopyF(p, frame, obj)))
			}
		}
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			obj := Gia_ManCo(p, i)
			Gia_ObjSetCopyF(p, frame, obj, Gia_ObjFanin0CopyF(p, frame, obj))
		}
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			Gia_ManAppendCo(q, Gia_ObjCopyF(p, frame, Gia_ManPo(p, i)))
		}
		if frame+1 < nFrames {
			for i := int32(0); i < p.nRegs; i++ {
				Gia_ObjSetCopyF(p, frame+1, Gia_ManRo(p, i), Gia_ObjCopyF(p, frame, Gia_ManRi(p, i)))
			}
		}
	}
	Gia_ManHashStop(q)
	return q
}
func Gia_ManUnrollAndCofactor(p *Gia_Man_t, nFrames, limit, verbose int32) *Gia_Man_t {
	frames := Gia_ManUnrollInit(p, nFrames)
	aig := Gia_ManCleanup(frames)
	signals := Gia_ManDetectSeqSignalsWithFanout(p, limit, verbose)
	mapped := Gia_ManTransferFrames(p, frames, nFrames, aig, signals)
	Gia_ManStop(frames)
	vec.Vec_IntErase(&p.vCopies)
	result := Gia_ManDupCofAllInt(aig, mapped, verbose)
	Gia_ManStop(aig)
	return result
}

func Gia_ManRemoveEnables2(p *Gia_Man_t) *Gia_Man_t {
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ManHashAlloc(q)
	Gia_ManFillValue(p)
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ObjSetValue(Gia_ManCi(p, i), Gia_ManAppendCi(q))
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		obj := Gia_ManPo(p, i)
		Gia_ObjSetValue(obj, Gia_ObjFanin0Copy(obj))
	}
	for i := int32(0); i < p.nRegs; i++ {
		ri := Gia_ManRi(p, i)
		node := Gia_ObjFanin0(ri)
		if Gia_ObjIsMuxType(node) == 0 {
			fmt.Printf("Cannot recognize enable of flop %d.\n", i)
			Gia_ObjSetValue(ri, Gia_ObjFanin0Copy(ri))
			continue
		}
		var data1, data0 *Gia_Obj_t
		control := Gia_ObjRecognizeMux(node, &data1, &data0)
		ro := Gia_ObjRiToRo(p, ri)
		if Gia_Regular(data0) != ro && Gia_Regular(data1) != ro {
			fmt.Printf("Cannot recognize self-loop of enable flop %d.\n", i)
			Gia_ObjSetValue(ri, Gia_ObjFanin0Copy(ri))
			continue
		}
		if Gia_ObjFaninC0(ri) == 0 {
			data0 = Gia_Not(data0)
			data1 = Gia_Not(data1)
		}
		if Gia_IsComplement(control) != 0 {
			data0, data1 = data1, data0
		}
		data := data0
		if Gia_Regular(data0) == ro {
			data = data1
		}
		Gia_ObjSetValue(ri, Gia_ObjValue(Gia_Regular(data))^(1-Gia_IsComplement(data)))
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManAppendCo(q, Gia_ObjValue(Gia_ManCo(p, i)))
	}
	Gia_ManHashStop(q)
	Gia_ManSetRegNum(q, p.nRegs)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}

func Gia_ManRemoveEnables(p *Gia_Man_t) *Gia_Man_t {
	controls := []*Gia_Obj_t{nil}
	data := make([]*Gia_Obj_t, p.nRegs)
	classes := vec.Vec_IntAlloc(int(p.nRegs))
	failures := 0
	for i := int32(0); i < p.nRegs; i++ {
		ri := Gia_ManRi(p, i)
		driver := Gia_ObjFanin0(ri)
		complement := Gia_ObjFaninC0(ri)
		failure := ""
		if Gia_ObjIsAnd(driver) == 0 {
			failure = "driver is not a node"
		} else if Gia_ObjFaninC0(driver) == 0 || Gia_ObjFaninC1(driver) == 0 {
			failure = "driver is not an OR gate"
		} else if Gia_ObjIsAnd(Gia_ObjFanin0(driver)) == 0 || Gia_ObjIsAnd(Gia_ObjFanin1(driver)) == 0 {
			failure = "driver fanin is not a node"
		}
		var control, value *Gia_Obj_t
		if failure == "" {
			a, b := Gia_ObjFanin0(driver), Gia_ObjFanin1(driver)
			ro := Gia_NotCond(Gia_ObjRiToRo(p, ri), 1-complement)
			for _, pair := range [][2]*Gia_Obj_t{{a, b}, {b, a}} {
				branch, other := pair[0], pair[1]
				if Gia_ObjChild0(branch) == ro {
					control = Gia_Not(Gia_ObjChild1(branch))
				} else if Gia_ObjChild1(branch) == ro {
					control = Gia_Not(Gia_ObjChild0(branch))
				} else {
					continue
				}
				if Gia_ObjFanin0(other) == Gia_Regular(control) {
					value = Gia_ObjChild1(other)
				} else {
					value = Gia_ObjChild0(other)
				}
				break
			}
			if control == nil {
				failure = "does not have a self-loop"
			}
		}
		if failure != "" {
			fmt.Printf("The flop %d %s.\n", i, failure)
			vec.Vec_IntPush(classes, 0)
			failures++
			continue
		}
		class := 0
		for j := 1; j < len(controls); j++ {
			if controls[j] == control {
				class = j
				break
			}
		}
		if class == 0 {
			controls = append(controls, control)
			class = len(controls) - 1
		}
		data[i] = Gia_NotCond(value, 1-complement)
		vec.Vec_IntPush(classes, int32(class))
	}
	count := len(controls)
	if failures == 0 {
		count--
	}
	fmt.Printf("Detected %d classes.\n", count)
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	for _, obj := range p.pObjs[1:] {
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		case Gia_ObjIsCi(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendCi(q))
		case Gia_ObjIsPo(p, obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj)))
		}
	}
	for i := int32(0); i < p.nRegs; i++ {
		ri := Gia_ManRi(p, i)
		lit := Gia_ObjFanin0Copy(ri)
		if data[i] != nil {
			lit = Gia_ObjValue(Gia_Regular(data[i])) ^ Gia_IsComplement(data[i])
		}
		Gia_ObjSetValue(ri, Gia_ManAppendCo(q, lit))
	}
	Gia_ManSetRegNum(q, p.nRegs)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	result.VFlopClasses = classes
	return result
}
