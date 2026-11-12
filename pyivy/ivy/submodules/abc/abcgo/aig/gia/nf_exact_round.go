package gia

import "github.com/glycerine/abcgo/misc/vec"

// exactAreaRound walks referenced gates backwards so newly tightened input
// requirements are available when their drivers are considered.
func (p *nfMatcher) exactAreaRound(r *nfMappingRefs) int {
	g := p.cuts.graph
	for i := range p.required {
		p.required[i] = [2]int32{nfInfinity, nfInfinity}
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		fanin, phase := Gia_ObjFaninId0(co, id), Gia_ObjFaninC0(co)
		required := r.delay
		if p.cuts.pars.fDoAverage != 0 {
			kind := r.selected[fanin][phase]
			if kind < 0 {
				kind = 0
			}
			required = p.objects[fanin][phase][kind].delay * (100 + p.cuts.pars.nRelaxRatio) / 100
		}
		p.updateRequired(fanin, phase, required)
	}
	changed := 0
	for index := len(g.pObjs) - 1; index >= 0; index-- {
		obj := g.pObjs[index]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			if r.refs[id][1] > 0 {
				p.updateRequired(id, 0, p.required[id][1]-p.invDelay)
			}
			if r.refs[id][0] > 0 {
				p.updateRequired(Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj), p.required[id][0])
			}
			continue
		}
		// An inverted phase depends on the other phase of the same object.
		for phase := int32(0); phase < 2; phase++ {
			if r.refs[id][phase] == 0 {
				continue
			}
			kind := r.selected[id][phase]
			if kind < 0 {
				panic("gia: missing exact-area match")
			}
			if p.objects[id][phase][kind].inverted {
				p.updateRequired(id, phase^1, p.required[id][phase]-p.invDelay)
			}
		}
		for phase := int32(0); phase < 2; phase++ {
			if r.refs[id][phase] == 0 {
				continue
			}
			kind := r.selected[id][phase]
			if p.objects[id][phase][kind].inverted {
				continue
			}
			if p.optimizeExactMatch(r, id, phase) {
				changed++
			}
			// Propagate even if the old implementation was retained.
			match := p.objects[id][phase][r.selected[id][phase]]
			cut := p.cuts.storage.at(p.cuts.sets[id])[match.cutOffset:]
			cell := &p.cuts.library.cells[match.gate]
			leaves := Nf_CutLeaves(cut)
			for pin := int32(0); pin < Nf_CutSize(cut); pin++ {
				fanin := leaves[Nf_CfgVar(match.cfg, pin)]
				polarity := Nf_CfgCompl(match.cfg, pin)
				p.updateRequired(fanin, polarity, p.required[id][phase]-cell.Delay(int(pin)))
			}
		}
	}
	for _, id := range vec.Vec_IntArray(g.vCis) {
		if r.refs[id][1] > 0 {
			p.updateRequired(id, 0, p.required[id][1]-p.invDelay)
		}
	}
	p.refreshExactStats(r)
	return changed
}
