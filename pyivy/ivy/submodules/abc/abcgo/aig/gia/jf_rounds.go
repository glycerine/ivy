package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"time"
)

func Jf_ObjPropagateBuf(p *Jf_Man_t, obj *Gia_Obj_t, reverse int32) {
	// The original mapper explicitly rejects buffers at this point.
	panic("JF mapping does not support buffers")
}
func Jf_ManComputeCuts(p *Jf_Man_t, edge int32) {
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Aig: CI = %d  CO = %d  AND = %d    LutSize = %d  CutMax = %d  Rounds = %d\n", Gia_ManCiNum(p.pGia), Gia_ManCoNum(p.pGia), Gia_ManAndNum(p.pGia), p.pPars.NLutSize, p.pPars.NCutNum, p.pPars.NRounds)
	}
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsCi(obj) != 0 || Gia_ObjIsBuf(obj) != 0 {
			Jf_ObjAssignCut(p, obj)
		}
		if Gia_ObjIsBuf(obj) != 0 {
			Jf_ObjPropagateBuf(p, obj, 0)
		} else if Gia_ObjIsAnd(obj) != 0 {
			Jf_ObjComputeCuts(p, obj, edge)
		}
	}
	if p.pPars.FVerbose != 0 {
		fmt.Printf("CutPair = %d  Merge = %d  Eval = %d  Cut = %d  Time = %s  Cuts = %.2f MB\n", p.CutCount[0], p.CutCount[1], p.CutCount[2], p.CutCount[3], time.Since(p.clkStart), p.pMem.memory()/(1<<20))
	}
}
func Jf_ManComputeDelay(p *Jf_Man_t, evaluate int32) int32 {
	if evaluate != 0 {
		for _, obj := range p.pGia.pObjs {
			if Gia_ObjIsBuf(obj) != 0 {
				Jf_ObjPropagateBuf(p, obj, 0)
			} else if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(p.pGia, obj) > 0 {
				vec.Vec_IntWriteEntry(&p.vArr, int(obj.id), Jf_CutArr(p, Jf_ObjCutBest(p, obj.id)))
			}
		}
	}
	var delay int32
	for _, id := range vec.Vec_IntArray(p.pGia.vCos) {
		driver := Gia_ObjFanin0(Gia_ManObj(p.pGia, id))
		if Gia_ObjRefNum(p.pGia, driver) <= 0 {
			panic("JF output driver is not referenced")
		}
		delay = max(delay, Jf_ObjArr(p, driver.id))
	}
	return delay
}
func Jf_ManComputeRefs(p *Jf_Man_t) int32 {
	g := p.pGia
	if g.Refs == nil {
		panic("JF missing references")
	}
	clear(g.Refs)
	p.pPars.Area = 0
	p.pPars.Edge = 0
	for i := len(g.pObjs) - 1; i >= 0; i-- {
		obj := g.pObjs[i]
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsBuf(obj) != 0 {
			Gia_ObjRefInc(g, Gia_ObjFanin0(obj))
		} else if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(g, obj) > 0 {
			if !Jf_ObjIsUnit(obj) {
				panic("JF mapped object is non-unit")
			}
			cut := Jf_ObjCutBest(p, obj.id)
			Jf_CutRef(p, cut)
			if p.pPars.FGenCnf != 0 {
				p.pPars.Clause += word(Jf_CutCnfSize(p, cut))
			}
			p.pPars.Edge += word(Jf_CutSize(cut))
			p.pPars.Area++
		}
	}
	refs, flow := vec.Vec_FltArray(&p.vRefs), vec.Vec_FltArray(&p.vFlow)
	for i := range refs {
		old := refs[i]
		var newRef float32
		if p.pPars.FOptEdge != 0 {
			newRef = max(1, 0.8*old+0.2*float32(g.Refs[i]))
		} else {
			newRef = max(1, 0.2*old+0.8*float32(g.Refs[i]))
		}
		flow[i] = flow[i] * old / newRef
		refs[i] = newRef
		if flow[i] < 0 {
			panic("JF negative flow")
		}
	}
	p.pPars.Delay = word(Jf_ManComputeDelay(p, 1))
	return int32(p.pPars.Area)
}
func Jf_ObjComputeBestCut(p *Jf_Man_t, obj *Gia_Obj_t, edge, ela int32) {
	id := obj.id
	packed := Jf_ObjCuts(p, id)
	var best []int32
	bestArea := float32(math.MaxInt32)
	bestTime, tm := int32(math.MaxInt32), int32(math.MaxInt32)
	for _, cut := range jfCutList(packed) {
		if Jf_CutIsTriv(cut, id) {
			continue
		}
		if edge != 0 && ela == 0 {
			Jf_CutSetCost(cut, Jf_CutSize(cut))
		}
		var area float32
		if ela != 0 {
			area = float32(Jf_CutArea(p, cut, edge))
		} else {
			area = Jf_CutFlow(p, cut) + float32(Jf_CutCost(cut))
		}
		better := best == nil || bestArea > area+JF_EPSILON
		if !better && bestArea > area-JF_EPSILON {
			tm = Jf_CutArr(p, cut)
			better = bestTime > tm
		}
		if better {
			best, bestArea, bestTime = cut, area, tm
		}
	}
	if best == nil {
		panic("JF no nontrivial best cut")
	}
	vec.Vec_IntWriteEntry(&p.vArr, int(id), Jf_CutArr(p, best))
	if ela == 0 {
		vec.Vec_FltWriteEntry(&p.vFlow, int(id), bestArea/Jf_ObjRefs(p, id))
	}
	Jf_ObjSetBestCut(packed, best, p.vTemp)
}
func Jf_ManPropagateFlow(p *Jf_Man_t, edge int32) {
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsBuf(obj) != 0 {
			Jf_ObjPropagateBuf(p, obj, 0)
		} else if Gia_ObjIsAnd(obj) != 0 && Jf_ObjIsUnit(obj) {
			Jf_ObjComputeBestCut(p, obj, edge, 0)
		}
	}
	Jf_ManComputeRefs(p)
}
func Jf_ManPropagateEla(p *Jf_Man_t, edge int32) {
	p.pPars.Area = 0
	p.pPars.Edge = 0
	p.pPars.Clause = 0
	for i := len(p.pGia.pObjs) - 1; i >= 0; i-- {
		obj := p.pGia.pObjs[i]
		if Gia_ObjIsBuf(obj) != 0 {
			Jf_ObjPropagateBuf(p, obj, 1)
		} else if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(p.pGia, obj) > 0 {
			if !Jf_ObjIsUnit(obj) {
				panic("JF mapped object is non-unit")
			}
			if Jf_CutCheckMffc(p, Jf_ObjCutBest(p, obj.id), 50) != 0 {
				before := Jf_CutDeref_rec(p, Jf_ObjCutBest(p, obj.id))
				Jf_ObjComputeBestCut(p, obj, edge, 1)
				after := Jf_CutRef_rec(p, Jf_ObjCutBest(p, obj.id))
				if before < after {
					panic("JF exact area increased")
				}
			}
			if p.pPars.FGenCnf != 0 {
				p.pPars.Clause += word(Jf_CutCnfSize(p, Jf_ObjCutBest(p, obj.id)))
			}
			p.pPars.Edge += word(Jf_CutSize(Jf_ObjCutBest(p, obj.id)))
			p.pPars.Area++
		}
	}
	p.pPars.Delay = word(Jf_ManComputeDelay(p, 1))
}
