package gia

import "github.com/glycerine/abcgo/misc/vec"

// deriveMapping records the selected gates in GIA's literal-indexed cell
// mapping. Negative entries denote an inverter or a buffer, respectively.
func (p *nfMatcher) deriveMapping(refs *nfMappingRefs) *Gia_Man_t {
	g := p.cuts.graph
	if g.VCellMapping != nil {
		panic("gia: cell mapping already exists")
	}
	mapping := vec.Vec_IntAlloc(int(2*g.nObjs + refs.edges + 2*refs.gates))
	vec.Vec_IntFill(mapping, int(2*g.nObjs), 0)
	for _, id := range vec.Vec_IntArray(g.vCis) {
		if refs.refs[id][1] != 0 {
			vec.Vec_IntWriteEntry(mapping, int(2*id+1), -1)
		}
	}
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			if refs.refs[id][1] != 0 {
				vec.Vec_IntWriteEntry(mapping, int(2*id+1), -1)
			}
			vec.Vec_IntWriteEntry(mapping, int(2*id), -2)
			continue
		}
		for phase := int32(0); phase < 2; phase++ {
			if refs.refs[id][phase] == 0 {
				continue
			}
			kind := refs.selected[id][phase]
			if kind < 0 {
				panic("gia: missing Nf selected match")
			}
			match := p.objects[id][phase][kind]
			if match.inverted {
				vec.Vec_IntWriteEntry(mapping, int(2*id+phase), -1)
				continue
			}
			if match.cutOffset <= 0 {
				panic("gia: missing Nf cut")
			}
			cut := p.cuts.storage.at(p.cuts.sets[id])[match.cutOffset:]
			leaves := Nf_CutLeaves(cut)
			vec.Vec_IntWriteEntry(mapping, int(2*id+phase), int32(vec.Vec_IntSize(mapping)))
			vec.Vec_IntPush(mapping, Nf_CutSize(cut))
			for pin := int32(0); pin < Nf_CutSize(cut); pin++ {
				vec.Vec_IntPush(mapping, 2*leaves[Nf_CfgVar(match.cfg, pin)]+Nf_CfgCompl(match.cfg, pin))
			}
			vec.Vec_IntPush(mapping, match.gate)
		}
	}
	g.VCellMapping = mapping
	return g
}
