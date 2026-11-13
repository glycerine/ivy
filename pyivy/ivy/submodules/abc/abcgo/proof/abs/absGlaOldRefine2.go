package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gla_ManRefinement2 simulates the abstraction, traces relevant pseudo-inputs,
// and checks their sufficiency with ternary simulation before remapping them.
func Gla_ManRefinement2(p *Gla_Man_t) *vec.Vec_Int_t {
	p.refSign++
	if p.refSign == 0 || p.refSign > 0x7fffffff {
		for i := range p.pvRefis {
			clear(p.pvRefis[i])
		}
		p.refSign = 1
	}
	sign := p.refSign
	cos, pis, ppis, nodes := vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000)
	Gla_ManCollect(p, pis, ppis, cos, nodes)
	g := p.pGia
	for frame := int32(0); frame <= p.pPars.iFrame; frame++ {
		zero := Gla_ObjRef(p, gia.Gia_ManConst0(g), frame)
		*zero = Rfn_Obj_t{Sign: sign}
		for _, id := range vec.Vec_IntArray(pis) {
			r := Gla_ObjRef(p, gia.Gia_ManObj(g, id), frame)
			*r = Rfn_Obj_t{Value: uint32(Gla_ObjSatValue(p, id, frame)), Sign: sign}
		}
		for i, id := range vec.Vec_IntArray(ppis) {
			o := gia.Gia_ManObj(g, id)
			if gia.Gia_ObjIsAnd(o) == 0 && gia.Gia_ObjIsRo(g, o) == 0 {
				panic("invalid GLA pseudo-input")
			}
			r := Gla_ObjRef(p, o, frame)
			*r = Rfn_Obj_t{Value: uint32(Gla_ObjSatValue(p, id, frame)), Prio: uint32(i + 1), fPPi: 1, Sign: sign}
		}
		for _, id := range vec.Vec_IntArray(nodes) {
			o := gia.Gia_ManObj(g, id)
			r := Gla_ObjRef(p, o, frame)
			*r = Rfn_Obj_t{Sign: sign}
			if gia.Gia_ObjIsRo(g, o) != 0 {
				if frame > 0 {
					previous := Gla_ObjRef(p, gia.Gia_ObjRoToRi(g, o), frame-1)
					r.Value, r.Prio = previous.Value, previous.Prio
				}
				continue
			}
			if gia.Gia_ObjIsAnd(o) == 0 {
				panic("invalid GLA internal node")
			}
			a, b := Gla_ObjRef(p, gia.Gia_ObjFanin0(o), frame), Gla_ObjRef(p, gia.Gia_ObjFanin1(o), frame)
			av, bv := a.Value^uint32(gia.Gia_ObjFaninC0(o)), b.Value^uint32(gia.Gia_ObjFaninC1(o))
			r.Value = av & bv
			if int(id) < len(p.pObj2Obj) && p.pObj2Obj[id] != ^uint32(0) && Gla_ManCheckVar(p, int32(p.pObj2Obj[id]), frame) != 0 && int32(r.Value) != Gla_ObjSatValue(p, id, frame) {
				fmt.Printf("Object %d has value mismatch\n", id)
			}
			if r.Value == 1 {
				r.Prio = max(a.Prio, b.Prio)
			} else if av == 0 && bv == 0 {
				r.Prio = min(a.Prio, b.Prio)
			} else if av == 0 {
				r.Prio = a.Prio
			} else {
				r.Prio = b.Prio
			}
		}
		for _, id := range vec.Vec_IntArray(cos) {
			o := gia.Gia_ManObj(g, id)
			a := Gla_ObjRef(p, gia.Gia_ObjFanin0(o), frame)
			r := Gla_ObjRef(p, o, frame)
			*r = Rfn_Obj_t{Value: a.Value ^ uint32(gia.Gia_ObjFaninC0(o)), Prio: a.Prio, Sign: sign}
		}
	}
	output := gia.Gia_ManPo(g, 0)
	r := Gla_ObjRef(p, output, p.pPars.iFrame)
	if r.Value != 1 {
		fmt.Print("\nCounter-example verification has failed!!!\n")
	}
	if r.Prio == 0 {
		witness := Gla_ManDeriveCex(p, pis)
		if err := g.SetSequentialCounterexample(witness); err != nil {
			panic(err)
		}
		return nil
	}
	selected := vec.Vec_IntAlloc(100)
	vec.Vec_IntFill(p.vObjCounts, int(p.pPars.iFrame)+1, 0)
	Gla_ManRefSelect_rec(p, gia.Gia_ObjFanin0(output), p.pPars.iFrame, selected, int32(sign))
	vec.Vec_IntUniqify(selected)
	Gla_ManVerifyUsingTerSim(p, pis, ppis, nodes, cos, selected)
	for i, id := range vec.Vec_IntArray(selected) {
		if id < 0 || int(id) >= len(p.pObj2Obj) {
			panic("invalid GLA selected node")
		}
		vec.Vec_IntWriteEntry(selected, i, int32(p.pObj2Obj[id]))
	}
	p.nObjAdded += int32(vec.Vec_IntSize(selected))
	return selected
}
