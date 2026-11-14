package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Lf_ManSetMuxCut(p *Lf_Man_t, best *Lf_Bst_t, id, required int32) int32 {
	if Gia_ObjIsMuxId(p.pGia, id) == 0 || best.Delay[2] > required || best.Flow[2] > 1.1*best.Flow[1] {
		return 0
	}
	mux := Gia_ManObj(p.pGia, id)
	a, b := Gia_ObjFanin0(mux), Gia_ObjFanin1(mux)
	if mux.fMark0 != 0 || a.fMark0 != 0 || b.fMark0 != 0 {
		return 0
	}
	a.fMark0 = 1
	b.fMark0 = 1
	return 1
}
func Lf_ManSetMapRefsOne(p *Lf_Man_t, id int32) {
	best := Lf_ObjReadBest(p, id)
	required := Lf_ObjRequired(p, id)
	if Lf_ObjMapRefNum(p, id) <= 0 || best.Cut[0].fUsed != 0 || best.Cut[1].fUsed != 0 {
		panic("LF invalid mapping reference")
	}
	if p.pPars.FUseMux7 == 0 || Lf_ManSetMuxCut(p, best, id, required) == 0 {
		index := 0
		if Lf_BestDiffCuts(best) != 0 && best.Delay[1] <= required {
			index = 1
		}
		best.Cut[index].fUsed = 1
	}
	cut := Lf_ObjCutBest(p, id)
	if cut.fMux7 != 0 && cut.nLeaves != 3 {
		panic("LF invalid MUX cut")
	}
	for _, leaf := range cut.pLeaves[:cut.nLeaves] {
		Lf_ObjSetRequired(p, leaf, required-1)
		if Gia_ObjIsAndNotBuf(Gia_ManObj(p.pGia, leaf)) != 0 {
			Lf_ObjMapRefInc(p, leaf)
		}
	}
	if cut.fMux7 != 0 {
		p.pPars.Mux7++
		p.pPars.Edge++
		return
	}
	if vec.Vec_FltSize(&p.vSwitches) > 0 {
		p.Switches += float64(Lf_CutSwitches(p, cut))
	}
	p.pPars.Edge += word(cut.nLeaves)
	p.pPars.Area++
}
func lfRequiredDelay(p *Lf_Man_t, selected bool) int32 {
	var delay int32
	for i := int32(0); i < Gia_ManCoNum(p.pGia); i++ {
		a := Lf_ObjCoArrival(p, i)
		if selected {
			a = Lf_ObjCoArrival2(p, i)
		}
		delay = max(delay, a)
	}
	if p.pPars.DelayTarget == -1 && p.pPars.NRelaxRatio != 0 {
		p.pPars.DelayTarget = int32(float32(delay) * float32(100+p.pPars.NRelaxRatio) / 100)
	}
	if p.pPars.DelayTarget != -1 {
		if delay <= p.pPars.DelayTarget {
			delay = p.pPars.DelayTarget
		} else if p.pPars.NRelaxRatio == 0 {
			fmt.Printf("Relaxing user-specified delay target from %d to %d.\n", p.pPars.DelayTarget, delay)
		}
	}
	p.pPars.Delay = word(delay)
	return delay
}
func lfTimedRequired(p *Lf_Man_t, delay int32) *tim.Tim_Man_t {
	if p.pGia.PManTime == nil {
		return nil
	}
	manager, ok := p.pGia.PManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("LF invalid timing manager")
	}
	tim.Tim_ManIncrementTravId(manager)
	if p.pPars.FDoAverage != 0 {
		for i := int32(0); i < Gia_ManCoNum(p.pGia); i++ {
			tim.Tim_ManSetCoRequired(manager, int(i), float32(int32(float32(Lf_ObjCoArrival(p, i))*float32(100+p.pPars.NRelaxRatio)/100)))
		}
	} else {
		tim.Tim_ManInitPoRequiredAll(manager, float32(delay))
	}
	return manager
}
func lfOutputRequired(p *Lf_Man_t, i, delay int32) int32 {
	if p.pPars.FDoAverage != 0 {
		return int32(float32(Lf_ObjCoArrival(p, i)) * float32(100+p.pPars.NRelaxRatio) / 100)
	}
	return delay
}
func Lf_ManSetMapRefs(p *Lf_Man_t) int32 {
	delay := lfRequiredDelay(p, false)
	p.pPars.Mux7 = 0
	p.pPars.Area = 0
	p.pPars.Edge = 0
	p.Switches = 0
	vec.Vec_IntFill(&p.vMapRefs, int(Gia_ManAndNotBufNum(p.pGia)), 0)
	vec.Vec_IntFill(&p.vRequired, int(p.pGia.nObjs), util.ABC_INFINITY)
	if p.pPars.FUseMux7 != 0 {
		Gia_ManCleanMark0(p.pGia)
		for _, id := range vec.Vec_IntArray(p.pGia.vCis) {
			Gia_ManObj(p.pGia, id).fMark0 = 1
		}
	}
	timing := lfTimedRequired(p, delay)
	if timing != nil {
		if Gia_ManBufNum(p.pGia) != 0 {
			panic("LF timing with buffers")
		}
		for i := len(p.pGia.pObjs) - 1; i > 0; i-- {
			obj := p.pGia.pObjs[i]
			if Gia_ObjIsBuf(obj) != 0 {
				Lf_ObjSetRequired(p, Gia_ObjFanin0(obj).id, Lf_ObjRequired(p, obj.id))
			} else if Gia_ObjIsAnd(obj) != 0 {
				if Lf_ObjMapRefNum(p, obj.id) != 0 {
					Lf_ManSetMapRefsOne(p, obj.id)
				}
			} else if Gia_ObjIsCi(obj) != 0 {
				tim.Tim_ManSetCiRequired(timing, int(Gia_ObjCioId(obj)), float32(Lf_ObjRequired(p, obj.id)))
			} else if Gia_ObjIsCo(obj) != 0 {
				driver := Gia_ObjFanin0(obj)
				Lf_ObjSetRequired(p, driver.id, int32(tim.Tim_ManGetCoRequired(timing, int(Gia_ObjCioId(obj)))))
				if Gia_ObjIsAndNotBuf(driver) != 0 {
					Lf_ObjMapRefInc(p, driver.id)
				}
			} else {
				panic("LF unexpected timed object")
			}
		}
	} else {
		for i, id := range vec.Vec_IntArray(p.pGia.vCos) {
			driver := Gia_ObjFanin0(Gia_ManObj(p.pGia, id))
			Lf_ObjSetRequired(p, driver.id, lfOutputRequired(p, int32(i), delay))
			if Gia_ObjIsAndNotBuf(driver) != 0 {
				Lf_ObjMapRefInc(p, driver.id)
			}
		}
		for i := len(p.pGia.pObjs) - 1; i >= 0; i-- {
			obj := p.pGia.pObjs[i]
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			if Gia_ObjIsBuf(obj) != 0 {
				driver := Gia_ObjFanin0(obj)
				Lf_ObjSetRequired(p, driver.id, Lf_ObjRequired(p, obj.id))
				if Gia_ObjIsAndNotBuf(driver) != 0 {
					Lf_ObjMapRefInc(p, driver.id)
				}
			} else if Lf_ObjMapRefNum(p, obj.id) != 0 {
				Lf_ManSetMapRefsOne(p, obj.id)
			}
		}
	}
	if p.pPars.FUseMux7 != 0 {
		Gia_ManCleanMark0(p.pGia)
	}
	refs, flow := vec.Vec_IntArray(&p.vMapRefs), vec.Vec_FltArray(&p.vFlowRefs)
	if len(refs) != len(flow) {
		panic("LF reference size mismatch")
	}
	coef := float32(1) / float32(1+(p.Iter+1)*(p.Iter+1))
	for i := range refs {
		flow[i] = coef*flow[i] + (1-coef)*max(1, float32(refs[i]))
	}
	return int32(p.pPars.Area)
}
func Lf_ManCountMapRefsOne(p *Lf_Man_t, id int32) {
	best := Lf_ObjReadBest(p, id)
	cut := Lf_ObjCutBest(p, id)
	if Lf_ObjMapRefNum(p, id) <= 0 || Lf_BestIsMapped(best) == 0 || cut.fMux7 != 0 {
		panic("LF invalid counted cut")
	}
	required := Lf_ObjRequired(p, id)
	for _, leaf := range cut.pLeaves[:cut.nLeaves] {
		Lf_ObjSetRequired(p, leaf, required-1)
	}
	if vec.Vec_FltSize(&p.vSwitches) > 0 {
		p.Switches += float64(Lf_CutSwitches(p, cut))
	}
	p.pPars.Edge += word(cut.nLeaves)
	p.pPars.Area++
}
func Lf_ManCountMapRefs(p *Lf_Man_t) {
	delay := lfRequiredDelay(p, true)
	p.pPars.Mux7 = 0
	p.pPars.Area = 0
	p.pPars.Edge = 0
	p.Switches = 0
	vec.Vec_IntFill(&p.vRequired, int(p.pGia.nObjs), util.ABC_INFINITY)
	if p.pPars.FUseMux7 != 0 {
		Gia_ManCleanMark0(p.pGia)
	}
	timing := lfTimedRequired(p, delay)
	if timing != nil {
		for i := len(p.pGia.pObjs) - 1; i > 0; i-- {
			obj := p.pGia.pObjs[i]
			if Gia_ObjIsBuf(obj) != 0 {
				Lf_ObjSetRequired(p, Gia_ObjFanin0(obj).id, Lf_ObjRequired(p, obj.id))
			} else if Gia_ObjIsAnd(obj) != 0 {
				if Lf_ObjMapRefNum(p, obj.id) != 0 {
					Lf_ManCountMapRefsOne(p, obj.id)
				}
			} else if Gia_ObjIsCi(obj) != 0 {
				tim.Tim_ManSetCiRequired(timing, int(Gia_ObjCioId(obj)), float32(Lf_ObjRequired(p, obj.id)))
			} else if Gia_ObjIsCo(obj) != 0 {
				Lf_ObjSetRequired(p, Gia_ObjFanin0(obj).id, int32(tim.Tim_ManGetCoRequired(timing, int(Gia_ObjCioId(obj)))))
			} else {
				panic("LF unexpected timed object")
			}
		}
	} else {
		for i, id := range vec.Vec_IntArray(p.pGia.vCos) {
			Lf_ObjSetRequired(p, Gia_ObjFanin0(Gia_ManObj(p.pGia, id)).id, lfOutputRequired(p, int32(i), delay))
		}
		for i := len(p.pGia.pObjs) - 1; i >= 0; i-- {
			obj := p.pGia.pObjs[i]
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			if Gia_ObjIsBuf(obj) != 0 {
				Lf_ObjSetRequired(p, Gia_ObjFanin0(obj).id, Lf_ObjRequired(p, obj.id))
			} else if Lf_ObjMapRefNum(p, obj.id) != 0 {
				Lf_ManCountMapRefsOne(p, obj.id)
			}
		}
	}
	if p.pPars.FUseMux7 != 0 {
		Gia_ManCleanMark0(p.pGia)
	}
}
