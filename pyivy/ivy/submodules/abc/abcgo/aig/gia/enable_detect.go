package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_CollectSuper_rec(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t) {
	if Gia_IsComplement(obj) != 0 || Gia_ObjIsCi(obj) != 0 {
		vec.Vec_IntPushUnique(super, Gia_ObjId(p, Gia_Regular(obj)))
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid supergate")
	}
	Gia_CollectSuper_rec(p, Gia_ObjChild0(obj), super)
	Gia_CollectSuper_rec(p, Gia_ObjChild1(obj), super)
}
func Gia_CollectSuper(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t) {
	if Gia_IsComplement(obj) != 0 {
		panic("gia: complemented supergate root")
	}
	vec.Vec_IntClear(super)
	if Gia_ObjIsAnd(obj) != 0 {
		vec.Vec_IntPushUnique(super, Gia_ObjId(p, Gia_ObjFanin0(obj)))
		vec.Vec_IntPushUnique(super, Gia_ObjId(p, Gia_ObjFanin1(obj)))
	} else {
		vec.Vec_IntPushUnique(super, Gia_ObjId(p, obj))
	}
}
func Gia_ManPrintSignals(p *Gia_Man_t, freq []int32, label string) {
	count, total := 0, int32(0)
	for _, v := range freq {
		if v > 1 {
			count++
			total += v
		}
	}
	fmt.Printf("%s (total = %d  driven = %d)\n", label, count, total)
	count = 0
	for i, v := range freq {
		if v > 10 {
			count++
			fmt.Printf("%3d :   Obj = %6d   Refs = %6d   Freq = %6d\n", count, i, p.Refs[i], v)
		}
	}
}

type giaSeqSignals struct {
	sets, resets, enables    []int32
	haveSetReset, haveEnable int32
}

func giaDetectSeqSignals(p *Gia_Man_t) giaSeqSignals {
	if p.nRegs <= 0 {
		panic("gia: sequential signal detection requires registers")
	}
	r := giaSeqSignals{sets: make([]int32, p.nObjs), resets: make([]int32, p.nObjs), enables: make([]int32, p.nObjs)}
	super := vec.Vec_IntAlloc(100)
	for i := int32(0); i < p.nRegs; i++ {
		flop := Gia_ManRi(p, i)
		node := Gia_ObjFanin0(flop)
		if Gia_ObjIsAnd(node) == 0 {
			continue
		}
		Gia_CollectSuper(p, node, super)
		target := r.resets
		if Gia_ObjFaninC0(flop) != 0 {
			target = r.sets
		}
		for _, id := range vec.Vec_IntArray(super) {
			target[id]++
		}
		if Gia_ObjIsMuxType(node) == 0 {
			continue
		}
		var a, b *Gia_Obj_t
		control := Gia_ObjRecognizeMux(node, &a, &b)
		ro := Gia_ObjRiToRo(p, flop)
		if Gia_Regular(a) != ro && Gia_Regular(b) != ro {
			continue
		}
		r.enables[Gia_ObjId(p, Gia_Regular(control))]++
		r.haveEnable++
	}
	for i := int32(0); i < p.nRegs; i++ {
		node := Gia_ObjFanin0(Gia_ManRi(p, i))
		if Gia_ObjIsAnd(node) == 0 {
			continue
		}
		Gia_CollectSuper(p, node, super)
		for _, id := range vec.Vec_IntArray(super) {
			if r.sets[id] > 1 || r.resets[id] > 1 {
				r.haveSetReset++
				break
			}
		}
	}
	return r
}
func Gia_ManDetectSeqSignals(p *Gia_Man_t, setReset, verbose int32) {
	r := giaDetectSeqSignals(p)
	p.Refs = nil
	Gia_ManCreateRefs(p)
	if verbose != 0 {
		fmt.Printf("Flops with set/reset = %6d. Flops with enable = %6d.\n", r.haveSetReset, r.haveEnable)
		if setReset != 0 {
			Gia_ManPrintSignals(p, r.sets, "Set signals")
			Gia_ManPrintSignals(p, r.resets, "Reset signals")
		}
		Gia_ManPrintSignals(p, r.enables, "Enable signals")
	}
	p.Refs = nil
}
func Gia_ManDetectSeqSignalsWithFanout(p *Gia_Man_t, limit, verbose int32) *Vec_Int_t {
	r := giaDetectSeqSignals(p)
	result := vec.Vec_IntAlloc(100)
	for _, group := range []struct {
		name string
		freq []int32
	}{{"set", r.sets}, {"reset", r.resets}, {"enable", r.enables}} {
		for i := 1; i < len(group.freq); i++ {
			if group.freq[i] > limit {
				if verbose != 0 {
					fmt.Printf("Adding %s signal %d related to %d flops.\n", group.name, i, group.freq[i])
				}
				vec.Vec_IntPushUnique(result, int32(i))
			}
		}
	}
	return result
}
