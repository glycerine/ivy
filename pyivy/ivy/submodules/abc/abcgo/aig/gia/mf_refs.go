package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Mf_ManSetFlowRefs(p *Gia_Man_t, refs *Vec_Int_t) {
	vec.Vec_IntFill(refs, int(p.nObjs), 0)
	add := func(obj *Gia_Obj_t, delta int32) {
		if Gia_ObjIsAnd(obj) != 0 {
			vec.Vec_IntAddToEntry(refs, int(obj.id), delta)
		}
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		add(Gia_ObjFanin0(obj), 1)
		add(Gia_ObjFanin1(obj), 1)
		if p.pMuxes != nil {
			if Gia_ObjIsMuxId(p, obj.id) != 0 {
				add(Gia_ObjFanin2(p, obj), 1)
			}
		} else if Gia_ObjIsMuxType(obj) != 0 {
			var then, otherwise *Gia_Obj_t
			control := Gia_ObjRecognizeMux(obj, &then, &otherwise)
			control = Gia_Regular(control)
			then = Gia_Regular(then)
			otherwise = Gia_Regular(otherwise)
			add(control, -1)
			if then == otherwise {
				add(then, -1)
			}
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		add(Gia_ObjFanin0(Gia_ManObj(p, id)), 1)
	}
	for i, v := range vec.Vec_IntArray(refs) {
		if v < 1 {
			vec.Vec_IntWriteEntry(refs, i, 1)
		}
	}
}
func Mf_ManSetMapRefs(p *Mf_Man_t) int32 {
	coef := float32(1) / float32(1+(p.Iter+1)*(p.Iter+1))
	var delay int32
	for _, id := range vec.Vec_IntArray(p.pGia.vCos) {
		driver := Gia_ObjFanin0(Gia_ManObj(p.pGia, id))
		delay = max(delay, int32(Mf_ManObj(p, driver.id).Delay))
	}
	pars := p.pPars
	if pars.DelayTarget == -1 && pars.NRelaxRatio != 0 {
		pars.DelayTarget = int32(float32(delay) * (100 + float32(pars.NRelaxRatio)) / 100)
	}
	if pars.DelayTarget != -1 {
		if delay <= pars.DelayTarget {
			delay = pars.DelayTarget
		} else if pars.NRelaxRatio == 0 {
			fmt.Printf("Relaxing user-specified delay target from %d to %d.\n", pars.DelayTarget, delay)
		}
	}
	pars.Delay = word(delay)
	if p.fUseEla == 0 {
		for _, id := range vec.Vec_IntArray(p.pGia.vCos) {
			Mf_ObjMapRefInc(p, Gia_ObjFanin0(Gia_ManObj(p.pGia, id)).id)
		}
	}
	pars.Area, pars.Edge, pars.Clause = 0, 0, 0
	for i := len(p.pGia.pObjs) - 1; i >= 0; i-- {
		obj := p.pGia.pObjs[i]
		if Gia_ObjIsAnd(obj) == 0 || Mf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Mf_ObjCutBest(p, obj.id)
		if p.fUseEla == 0 {
			for _, id := range cut[1 : 1+Mf_CutSize(cut)] {
				Mf_ObjMapRefInc(p, id)
			}
		}
		pars.Edge += word(Mf_CutSize(cut))
		pars.Area++
		if pars.FGenCnf != 0 || pars.FGenLit != 0 {
			pars.Clause += word(Mf_CutArea(p, Mf_CutSize(cut), Mf_CutFunc(cut)))
		}
	}
	for i := range p.pLfObjs {
		o := &p.pLfObjs[i]
		o.nFlowRefs = coef*o.nFlowRefs + (1-coef)*max(1, float32(o.nMapRefs))
	}
	return int32(pars.Area)
}
