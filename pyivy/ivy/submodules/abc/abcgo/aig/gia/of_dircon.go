package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"sort"
)

func (p *ofMapper) directCut(cut []int32) (int32, int32) {
	n := int(ofStoredSize(cut))
	fast := int(p.cuts.pars.NFastEdges)
	if fast < 0 {
		panic("gia: invalid Of fast edge count")
	}
	type entry struct {
		pin   int
		delay int32
	}
	order := make([]entry, n)
	for i := range order {
		id := ofStoredVar(cut, i)
		order[i] = entry{i, p.objects[id].delay + p.cuts.pars.NDelayLut1}
	}
	sort.SliceStable(order, func(i, j int) bool { return order[i].delay > order[j].delay })
	slow := int32(0)
	if fast < n {
		slow = order[fast].delay
	}
	delay1, delay2 := int32(0), int32(0)
	works := true
	for k, e := range order {
		id := ofStoredVar(cut, e.pin)
		after := p.objects[id].delay2 + p.cuts.pars.NDelayLut2
		delay := e.delay
		if k < fast && e.delay > slow && after < e.delay && works && Gia_ObjIsAnd(Gia_ManObj(p.cuts.graph, id)) != 0 && Gia_ObjIsBuf(Gia_ManObj(p.cuts.graph, id)) == 0 {
			delay = after
			ofStoredSetFlag(cut, e.pin, 1)
		} else {
			ofStoredSetFlag(cut, e.pin, 0)
			works = false
		}
		delay1 = max(delay1, delay)
		delay2 = max(delay2, e.delay)
	}
	if delay1 > delay2 {
		for i := 0; i < n; i++ {
			ofStoredSetFlag(cut, i, 0)
		}
		delay1 = delay2
	}
	ofStoredSetDelay1(cut, delay1)
	ofStoredSetDelay2(cut, delay2)
	return delay1, delay2
}
func (p *ofMapper) forwardDirect() {
	for _, obj := range p.cuts.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		o := &p.objects[id]
		if Gia_ObjIsBuf(obj) != 0 {
			f := Gia_ObjFaninId0(obj, id)
			o.delay, o.delay2 = p.objects[f].delay, p.objects[f].delay2
			continue
		}
		d1, d2 := int32(math.MaxInt32), int32(math.MaxInt32)
		var best1, best2 int32
		p.eachCut(id, func(handle int32, cut []int32) {
			a, b := p.directCut(cut)
			if a < d1 {
				d1, best1 = a, handle
			}
			if b < d2 {
				d2, best2 = b, handle
			}
		})
		if best1 == 0 || best2 == 0 {
			panic("gia: Of direct mapping has no cut")
		}
		o.delay, o.delay2, o.best, o.best2 = d1, d2, best1, best2
	}
}
func (p *ofMapper) backwardDirect() {
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
	pointed := make([]bool, len(p.objects))
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
		if pointed[id] {
			o.best = o.best2
		}
		if o.best == 0 {
			panic("gia: missing Of direct cut")
		}
		cut := p.cuts.storage.at(o.best)
		for pin := range ofStoredLeaves(cut) {
			f, flag := ofStoredVar(cut, pin), ofStoredFlag(cut, pin)
			direct := flag != 0 && !pointed[id]
			arc := p.cuts.pars.NDelayLut1
			if direct {
				arc = p.cuts.pars.NDelayLut2
				pointed[f] = true
			}
			p.updateRequired(f, o.required-arc)
			p.objects[f].refs++
		}
		p.cuts.pars.Area++
		p.cuts.pars.Edge += word(ofStoredSize(cut))
	}
	p.iteration++
}
