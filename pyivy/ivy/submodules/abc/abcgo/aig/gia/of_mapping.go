package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

type ofMappingObject struct{ best, best2, delay, delay2, required, refs, flow int32 }
type ofMapper struct {
	cuts      *ofCutManager
	objects   []ofMappingObject
	iteration int32
}

func newOfMapper(cuts *ofCutManager) *ofMapper {
	p := &ofMapper{cuts: cuts, objects: make([]ofMappingObject, len(cuts.sets))}
	for index, id := range vec.Vec_IntArray(cuts.graph.vCis) {
		if cuts.graph.vInArrs != nil {
			p.objects[id].delay = int32(10 * max(0, vec.Vec_FltEntry(cuts.graph.vInArrs, index)))
			p.objects[id].delay2 = p.objects[id].delay
		}
	}
	// Initial flow estimates are distributed over graph fanouts.
	refs := make([]int32, len(p.objects))
	for _, obj := range cuts.graph.pObjs {
		for _, id := range giaObjectFanins(cuts.graph, obj) {
			refs[id]++
		}
	}
	for _, obj := range cuts.graph.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			id := obj.id
			p.objects[id].flow = (p.objects[Gia_ObjFaninId0(obj, id)].flow + p.objects[Gia_ObjFaninId1(obj, id)].flow + 1000) / max(1, refs[id])
		}
	}
	return p
}
func (p *ofMapper) cutArea(size int32) int32 {
	if size < 2 {
		return 0
	}
	return size + p.cuts.pars.NAreaTuner
}
func (p *ofMapper) eachCut(id int32, visit func(handle int32, cut []int32)) {
	data := p.cuts.storage.at(p.cuts.sets[id])
	offset := int32(1)
	for i := int32(0); i < data[0]; i++ {
		cut := data[offset:]
		visit(p.cuts.sets[id]+offset, cut)
		offset += ofStoredSize(cut) + 4
	}
}
func (p *ofMapper) forward() {
	for _, obj := range p.cuts.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			p.objects[id].delay = p.objects[Gia_ObjFaninId0(obj, id)].delay
			continue
		}
		best, bestDelay := int32(0), int32(math.MaxInt32)
		p.eachCut(id, func(handle int32, cut []int32) {
			delay := int32(0)
			for i := range ofStoredLeaves(cut) {
				delay = max(delay, p.objects[ofStoredVar(cut, i)].delay+p.cuts.pars.NDelayLut1)
			}
			ofStoredSetDelay1(cut, delay)
			if delay < bestDelay {
				best, bestDelay = handle, delay
			}
		})
		o := &p.objects[id]
		if o.refs > 0 {
			best = o.best
		}
		if best == 0 {
			panic("gia: Of node has no cut")
		}
		o.best = best
		cut := p.cuts.storage.at(best)
		o.delay = ofStoredDelay1(cut)
		if p.iteration > 0 {
			area := 100 * p.cutArea(ofStoredSize(cut))
			for i := range ofStoredLeaves(cut) {
				area += p.objects[ofStoredVar(cut, i)].flow
			}
			o.flow = area / max(1, o.refs)
		}
	}
}
func (p *ofMapper) updateRequired(id, value int32) {
	p.objects[id].required = min(p.objects[id].required, value)
}
func (p *ofMapper) backward() {
	for i := range p.objects {
		p.objects[i].required = math.MaxInt32
		p.objects[i].refs = 0
	}
	delay := int32(0)
	for _, id := range vec.Vec_IntArray(p.cuts.graph.vCos) {
		co := Gia_ManObj(p.cuts.graph, id)
		delay = max(delay, p.objects[Gia_ObjFaninId0(co, id)].delay)
	}
	for _, id := range vec.Vec_IntArray(p.cuts.graph.vCos) {
		co := Gia_ManObj(p.cuts.graph, id)
		f := Gia_ObjFaninId0(co, id)
		p.updateRequired(f, delay)
		p.objects[f].refs++
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
			f := Gia_ObjFaninId0(obj, id)
			p.updateRequired(f, o.required)
			p.objects[f].refs++
			continue
		}
		best, bestCost := int32(0), int32(math.MaxInt32)
		p.eachCut(id, func(handle int32, cut []int32) {
			if ofStoredDelay1(cut) > o.required {
				return
			}
			cost := int32(0)
			for i := range ofStoredLeaves(cut) {
				f := ofStoredVar(cut, i)
				if p.objects[f].refs == 0 {
					cost += p.objects[f].flow
				}
			}
			if cost < bestCost {
				best, bestCost = handle, cost
			}
		})
		if best == 0 {
			panic("gia: Of node has no timing-feasible cut")
		}
		o.best = best
		cut := p.cuts.storage.at(best)
		for i := range ofStoredLeaves(cut) {
			f := ofStoredVar(cut, i)
			p.updateRequired(f, o.required-p.cuts.pars.NDelayLut1)
			p.objects[f].refs++
		}
		p.cuts.pars.Edge += word(ofStoredSize(cut))
		p.cuts.pars.Area++
	}
	p.iteration++
}
