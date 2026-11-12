package gia

// adjustMatchRefs adds or removes a match's fanin references. Area is counted
// only when a previously unused implementation becomes used, or vice versa.
// The root reference is owned by the caller and is not changed here.
func (p *nfMatcher) adjustMatchRefs(r *nfMappingRefs, id, phase int32, m nfMatch, add bool) uint64 {
	obj := Gia_ManObj(p.cuts.graph, id)
	visit := func(fanin, polarity int32) uint64 { return p.adjustFaninRef(r, fanin, polarity, add) }
	if m.inverted {
		return p.cuts.library.cells[3].AreaW + visit(id, phase^1)
	}
	if Gia_ObjIsCi(obj) != 0 || id == 0 {
		return 0
	}
	if Gia_ObjIsBuf(obj) != 0 {
		return visit(Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj))
	}
	if m.cutOffset <= 0 {
		panic("gia: missing Nf exact-area cut")
	}
	cut := p.cuts.storage.at(p.cuts.sets[id])[m.cutOffset:]
	leaves := Nf_CutLeaves(cut)
	area := p.cuts.library.cells[m.gate].AreaW
	for pin := int32(0); pin < Nf_CutSize(cut); pin++ {
		area += visit(leaves[Nf_CfgVar(m.cfg, pin)], Nf_CfgCompl(m.cfg, pin))
	}
	return area
}

func (p *nfMatcher) adjustFaninRef(r *nfMappingRefs, id, phase int32, add bool) uint64 {
	refs := &r.refs[id][phase]
	if add {
		previous := *refs
		*refs++
		if previous != 0 {
			return 0
		}
	} else {
		if *refs <= 0 {
			panic("gia: invalid Nf reference decrement")
		}
		*refs--
		if *refs != 0 {
			return 0
		}
	}
	// Unused phases retain their delay-oriented match as the candidate
	// implementation, just as the exact-area pass initializes them.
	kind := r.selected[id][phase]
	if kind < 0 {
		kind = 0
	}
	return p.adjustMatchRefs(r, id, phase, p.objects[id][phase][kind], add)
}

// incrementalArea measures adding a candidate while restoring every reference
// touched during the measurement.
func (p *nfMatcher) incrementalArea(r *nfMappingRefs, id, phase int32, m nfMatch) uint64 {
	area := p.adjustMatchRefs(r, id, phase, m, true)
	restored := p.adjustMatchRefs(r, id, phase, m, false)
	if restored != area {
		panic("gia: inconsistent Nf exact-area reference restoration")
	}
	return area
}

// bestExactMatch considers every library match for each stored cut. The
// current root's fanin references must be detached before calling it.
func (p *nfMatcher) bestExactMatch(r *nfMappingRefs, id, phase, required int32) (nfMatch, uint64, bool) {
	data := p.cuts.storage.at(p.cuts.sets[id])
	offset := int32(1)
	var best nfMatch
	var bestArea uint64
	found := false
	for i := int32(0); i < data[0]; i++ {
		cut := data[offset:]
		count, function := Nf_CutSize(cut), Nf_CutFunc(cut)
		if count == 0 {
			gate := phase ^ (function & 1)
			candidate := nfMatch{gate: gate, cutOffset: offset}
			area := p.cuts.library.cells[gate].AreaW
			if !found || area < bestArea {
				best, bestArea, found = candidate, area, true
			}
		} else if int(function>>1) < len(p.cuts.library.matches) {
			leaves := Nf_CutLeaves(cut)
			for _, match := range p.cuts.library.matches[function>>1] {
				if int32(match.cfg.fCompl)^(function&1) != phase {
					continue
				}
				cell := &p.cuts.library.cells[match.gate]
				if cell.Fanins() != count {
					panic("gia: inconsistent Nf exact-area match")
				}
				delay := int32(0)
				for pin := int32(0); pin < count; pin++ {
					fanin, polarity := leaves[Nf_CfgVar(match.cfg, pin)], Nf_CfgCompl(match.cfg, pin)
					kind := r.selected[fanin][polarity]
					if kind < 0 {
						kind = 0
					}
					delay = max(delay, p.objects[fanin][polarity][kind].delay+cell.Delay(int(pin)))
				}
				if delay > required {
					continue
				}
				cfg := match.cfg
				cfg.fCompl = 0
				candidate := nfMatch{gate: match.gate, cutOffset: offset, cfg: cfg, delay: delay}
				area := p.incrementalArea(r, id, phase, candidate)
				if !found || area < bestArea || area == bestArea && delay < best.delay {
					best, bestArea, found = candidate, area, true
				}
			}
		}
		offset += count + 1
	}
	return best, bestArea, found
}
