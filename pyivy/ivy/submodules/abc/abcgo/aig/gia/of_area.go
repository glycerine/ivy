package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

// adjustCutRefs recursively adds or removes a cut's mapped fanins and returns
// the incremental LUT cost. Primary inputs have no selected cut.
func (p *ofMapper) adjustCutRefs(cut []int32, add bool) int32 {
	cost := int32(1)
	if p.iteration&1 == 0 {
		cost = p.cutArea(ofStoredSize(cut))
	}
	for i := range ofStoredLeaves(cut) {
		id := ofStoredVar(cut, i)
		o := &p.objects[id]
		if o.best == 0 {
			continue
		}
		if add {
			old := o.refs
			o.refs++
			if old == 0 {
				cost += p.adjustCutRefs(p.cuts.storage.at(o.best), true)
			}
		} else {
			if o.refs <= 0 {
				panic("gia: invalid Of reference decrement")
			}
			o.refs--
			if o.refs == 0 {
				cost += p.adjustCutRefs(p.cuts.storage.at(o.best), false)
			}
		}
	}
	return cost
}
func (p *ofMapper) candidateArea(cut []int32) int32 {
	area := p.adjustCutRefs(cut, true)
	if removed := p.adjustCutRefs(cut, false); removed != area {
		panic("gia: inconsistent Of reference restoration")
	}
	return area
}
func (p *ofMapper) exactBest(id, required int32) (int32, int32) {
	best, cost := int32(0), int32(math.MaxInt32)
	p.eachCut(id, func(handle int32, cut []int32) {
		if ofStoredDelay1(cut) > required {
			return
		}
		area := p.candidateArea(cut)
		if area < cost {
			best, cost = handle, area
		}
	})
	if best == 0 {
		panic("gia: Of node has no exact-area cut")
	}
	return best, cost
}

// forwardExact updates cut arrivals in topological order and recovers area
// while keeping each object's previously propagated required time.
func (p *ofMapper) forwardExact() {
	for _, obj := range p.cuts.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		o := &p.objects[id]
		if Gia_ObjIsBuf(obj) != 0 {
			o.delay = p.objects[Gia_ObjFaninId0(obj, id)].delay
			continue
		}
		before := int32(0)
		if o.refs > 0 {
			before = p.adjustCutRefs(p.cuts.storage.at(o.best), false)
		}
		p.eachCut(id, func(_ int32, cut []int32) {
			delay := int32(0)
			for i := range ofStoredLeaves(cut) {
				delay = max(delay, p.objects[ofStoredVar(cut, i)].delay+p.cuts.pars.NDelayLut1)
			}
			ofStoredSetDelay1(cut, delay)
		})
		best, area := p.exactBest(id, o.required)
		if o.refs > 0 {
			if area > before {
				panic("gia: Of forward exact area increased")
			}
			p.adjustCutRefs(p.cuts.storage.at(best), true)
		}
		o.best = best
		o.delay = ofStoredDelay1(p.cuts.storage.at(best))
	}
}

// backwardExact propagates output requirements and improves referenced cuts.
func (p *ofMapper) backwardExact() {
	for i := range p.objects {
		p.objects[i].required = math.MaxInt32
	}
	delay := int32(0)
	for _, id := range vec.Vec_IntArray(p.cuts.graph.vCos) {
		co := Gia_ManObj(p.cuts.graph, id)
		delay = max(delay, p.objects[Gia_ObjFaninId0(co, id)].delay)
	}
	for _, id := range vec.Vec_IntArray(p.cuts.graph.vCos) {
		co := Gia_ManObj(p.cuts.graph, id)
		p.updateRequired(Gia_ObjFaninId0(co, id), delay)
	}
	p.cuts.pars.Delay = word(delay)
	p.cuts.pars.Area = 0
	p.cuts.pars.Edge = 0
	for i := len(p.cuts.graph.pObjs) - 1; i >= 0; i-- {
		obj := p.cuts.graph.pObjs[i]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		o := &p.objects[id]
		if o.refs == 0 {
			continue
		}
		if Gia_ObjIsBuf(obj) != 0 {
			p.updateRequired(Gia_ObjFaninId0(obj, id), o.required)
			continue
		}
		before := p.adjustCutRefs(p.cuts.storage.at(o.best), false)
		best, area := p.exactBest(id, o.required)
		if area > before {
			panic("gia: Of backward exact area increased")
		}
		o.best = best
		cut := p.cuts.storage.at(best)
		for i := range ofStoredLeaves(cut) {
			p.updateRequired(ofStoredVar(cut, i), o.required-p.cuts.pars.NDelayLut1)
		}
		p.adjustCutRefs(cut, true)
		p.cuts.pars.Area++
		p.cuts.pars.Edge += word(ofStoredSize(cut))
	}
	p.iteration++
}

// backwardRequired propagates requirements through the current selected cuts
// without changing their reference counts.
func (p *ofMapper) backwardRequired() {
	for i := range p.objects {
		p.objects[i].required = math.MaxInt32
	}
	delay := int32(0)
	for _, id := range vec.Vec_IntArray(p.cuts.graph.vCos) {
		co := Gia_ManObj(p.cuts.graph, id)
		delay = max(delay, p.objects[Gia_ObjFaninId0(co, id)].delay)
	}
	for _, id := range vec.Vec_IntArray(p.cuts.graph.vCos) {
		co := Gia_ManObj(p.cuts.graph, id)
		p.updateRequired(Gia_ObjFaninId0(co, id), delay)
	}
	p.cuts.pars.Delay = word(delay)
	p.cuts.pars.Area = 0
	p.cuts.pars.Edge = 0
	for i := len(p.cuts.graph.pObjs) - 1; i >= 0; i-- {
		obj := p.cuts.graph.pObjs[i]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		o := &p.objects[id]
		if o.refs == 0 {
			continue
		}
		if Gia_ObjIsBuf(obj) != 0 {
			p.updateRequired(Gia_ObjFaninId0(obj, id), o.required)
			continue
		}
		cut := p.cuts.storage.at(o.best)
		for pin := range ofStoredLeaves(cut) {
			p.updateRequired(ofStoredVar(cut, pin), o.required-p.cuts.pars.NDelayLut1)
		}
		p.cuts.pars.Area++
		p.cuts.pars.Edge += word(ofStoredSize(cut))
	}
}
