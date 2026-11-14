package gia

import "github.com/glycerine/abcgo/misc/vec"

func (p *ofMapper) deriveMapping() *Gia_Man_t {
	g := p.cuts.graph
	if g.VMapping != nil || p.cuts.pars.FCutMin != 0 {
		panic("gia: invalid Of mapping derivation")
	}
	mapping := vec.Vec_IntAlloc(int(g.nObjs) + int(p.cuts.pars.Edge) + 2*int(p.cuts.pars.Area))
	vec.Vec_IntFill(mapping, int(g.nObjs), 0)
	var packing *Vec_Int_t
	if p.cuts.pars.NFastEdges != 0 {
		packing = vec.Vec_IntAlloc(16)
		vec.Vec_IntPush(packing, 0)
	}
	pointed := make([]bool, int(g.nObjs))
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || p.objects[obj.id].refs == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			panic("gia: Of LUT mapping contains a buffer")
		}
		cut := p.cuts.storage.at(p.objects[id].best)
		vec.Vec_IntWriteEntry(mapping, int(id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, ofStoredSize(cut))
		for pin := range ofStoredLeaves(cut) {
			vec.Vec_IntPush(mapping, ofStoredVar(cut, pin))
		}
		vec.Vec_IntPush(mapping, id)
		if packing == nil || pointed[id] {
			continue
		}
		start := vec.Vec_IntSize(packing)
		vec.Vec_IntPush(packing, 0)
		vec.Vec_IntPush(packing, id)
		for pin := range ofStoredLeaves(cut) {
			if ofStoredFlag(cut, pin) != 0 {
				f := ofStoredVar(cut, pin)
				vec.Vec_IntPush(packing, f)
				pointed[f] = true
			}
		}
		vec.Vec_IntWriteEntry(packing, start, int32(vec.Vec_IntSize(packing)-start-1))
		vec.Vec_IntAddToEntry(packing, 0, 1)
	}
	g.VMapping = mapping
	g.VPacking = packing
	return g
}
