package gia

import "github.com/glycerine/abcgo/misc/vec"

// nfMappingRefs holds the selected matches and phase-specific references from
// a reverse traversal. A selection of -1 denotes an unused phase.
type nfMappingRefs struct {
	refs                    [][2]int32
	selected                [][2]int8
	area                    float32
	gates, edges, inverters int32
	delay                   int32
}

func (p *nfMatcher) updateRequired(id, phase, value int32) {
	p.required[id][phase] = min(p.required[id][phase], value)
}

// mappingRefs propagates the output delay requirement through selected gates
// and blends the resulting phase references for the next area-flow round.
func (p *nfMatcher) mappingRefs(iteration int32) *nfMappingRefs {
	g := p.cuts.graph
	r := &nfMappingRefs{refs: make([][2]int32, len(p.objects)), selected: make([][2]int8, len(p.objects))}
	for i := range r.selected {
		r.selected[i] = [2]int8{-1, -1}
		p.required[i] = [2]int32{nfInfinity, nfInfinity}
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		fanin := Gia_ObjFaninId0(co, id)
		phase := Gia_ObjFaninC0(co)
		r.delay = max(r.delay, p.objects[fanin][phase][0].delay)
	}
	r.delay = p.delayTarget(max(r.delay, p.cuts.pars.MapDelay))
	for outputIndex, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		fanin := Gia_ObjFaninId0(co, id)
		phase := Gia_ObjFaninC0(co)
		req := p.outputRequired(outputIndex, p.objects[fanin][phase][0].delay, r.delay)
		r.refs[fanin][phase]++
		p.updateRequired(fanin, phase, req)
	}
	pick := func(id, phase int32) int8 {
		if p.objects[id][phase][1].delay <= p.required[id][phase] {
			return 1
		}
		return 0
	}
	gate := func(id, phase int32, kind int8) {
		m := p.objects[id][phase][kind]
		if m.inverted || m.cutOffset <= 0 {
			panic("gia: invalid Nf gate selection")
		}
		r.selected[id][phase] = kind
		cut := p.cuts.storage.at(p.cuts.sets[id])[m.cutOffset:]
		cell := &p.cuts.library.cells[m.gate]
		if Nf_CutSize(cut) != cell.Fanins() {
			panic("gia: Nf match fanin mismatch")
		}
		leaves := Nf_CutLeaves(cut)
		for pin := int32(0); pin < Nf_CutSize(cut); pin++ {
			fanin := leaves[Nf_CfgVar(m.cfg, pin)]
			polarity := Nf_CfgCompl(m.cfg, pin)
			r.refs[fanin][polarity]++
			p.updateRequired(fanin, polarity, p.required[id][phase]-cell.Delay(int(pin)))
		}
		r.area += cell.AreaF
		r.gates++
		r.edges += Nf_CutSize(cut)
	}
	for i := len(g.pObjs) - 1; i >= 0; i-- {
		obj := g.pObjs[i]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			if r.refs[id][0] == 0 && r.refs[id][1] == 0 {
				continue
			}
			if r.refs[id][1] != 0 {
				r.refs[id][0]++
				p.updateRequired(id, 0, p.required[id][1]-p.invDelay)
				r.area += p.invArea
				r.gates++
				r.edges++
				r.inverters++
				r.selected[id][1] = 0
			}
			fanin, phase := Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj)
			r.refs[fanin][phase]++
			p.updateRequired(fanin, phase, p.required[id][0])
			r.selected[id][0] = 0
			continue
		}
		if r.refs[id][0] == 0 && r.refs[id][1] == 0 {
			continue
		}
		var kinds [2]int8
		for phase := int32(0); phase < 2; phase++ {
			if r.refs[id][phase] != 0 {
				kinds[phase] = pick(id, phase)
			} else {
				kinds[phase] = -1
			}
		}
		if kinds[0] >= 0 && kinds[1] >= 0 && p.objects[id][0][kinds[0]].inverted && p.objects[id][1][kinds[1]].inverted {
			kinds = [2]int8{0, 0}
		}
		for phase := int32(0); phase < 2; phase++ {
			kind := kinds[phase]
			if kind < 0 || !p.objects[id][phase][kind].inverted {
				continue
			}
			opposite := phase ^ 1
			r.selected[id][phase] = kind
			r.refs[id][opposite]++
			p.updateRequired(id, opposite, p.required[id][phase]-p.invDelay)
			kinds[opposite] = pick(id, opposite)
			if p.objects[id][opposite][kinds[opposite]].inverted {
				kinds[opposite] = 0
			}
			if p.objects[id][opposite][kinds[opposite]].inverted {
				panic("gia: circular Nf inverter selection")
			}
			r.area += p.invArea
			r.gates++
			r.edges++
			r.inverters++
		}
		for phase := int32(0); phase < 2; phase++ {
			if kinds[phase] >= 0 && !p.objects[id][phase][kinds[phase]].inverted {
				gate(id, phase, kinds[phase])
			}
		}
	}
	for _, id := range vec.Vec_IntArray(g.vCis) {
		if r.refs[id][1] != 0 {
			r.refs[id][0]++
			p.updateRequired(id, 0, p.required[id][1]-p.invDelay)
			r.area += p.invArea
			r.gates++
			r.edges++
			r.inverters++
		}
	}
	coefficient := float32(1) / float32(1+(iteration+1)*(iteration+1))
	for id := range p.flowRefs {
		for phase := 0; phase < 2; phase++ {
			p.flowRefs[id][phase] = max(1, coefficient*p.flowRefs[id][phase]+(1-coefficient)*max(1, float32(r.refs[id][phase])))
		}
	}
	p.cuts.pars.MapDelay = r.delay
	p.cuts.pars.MapAreaF = r.area
	p.cuts.pars.Area = word(r.gates)
	p.cuts.pars.Edge = word(r.edges)
	return r
}
