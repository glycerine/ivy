package gia

import "github.com/glycerine/abcgo/misc/vec"

// fixOutputDrivers shares a gate and inverter when both output polarities use
// separate gates and the inverter still meets the mapped delay.
func (p *nfMatcher) fixOutputDrivers(r *nfMappingRefs) int {
	g := p.cuts.graph
	changed := 0
	for _, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		driver, phase := Gia_ObjFaninId0(co, id), Gia_ObjFaninC0(co)
		obj := Gia_ManObj(g, driver)
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || r.refs[driver][0] == 0 || r.refs[driver][1] == 0 {
			continue
		}
		kind, otherKind := r.selected[driver][phase], r.selected[driver][phase^1]
		if kind < 0 || otherKind < 0 {
			panic("gia: missing Nf output driver match")
		}
		current, other := p.objects[driver][phase][kind], p.objects[driver][phase^1][otherKind]
		if current.inverted || other.inverted || other.delay+p.invDelay > r.delay {
			continue
		}
		p.adjustMatchRefs(r, driver, phase, current, false)
		r.refs[driver][phase^1]++
		other.delay += p.invDelay
		other.flow += p.invArea
		other.inverted = true
		p.objects[driver][phase][kind] = other
		changed++
	}
	if changed != 0 {
		p.refreshExactStats(r)
	}
	return changed
}
