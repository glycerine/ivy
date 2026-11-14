package gia

import "github.com/glycerine/abcgo/misc/vec"

type ofCutManager struct {
	graph                     *Gia_Man_t
	pars                      *Jf_Par_t
	storage                   *ofCutStorage
	truth                     *nfTruthStore
	sets, flows, delays, refs []int32
	counts                    [4]float64
}

func newOfCutManager(g *Gia_Man_t, pars *Jf_Par_t) *ofCutManager {
	if pars.NCutNum <= 1 || pars.NCutNum > 32 || pars.NLutSize <= 1 || pars.NLutSize > 6 {
		panic("gia: invalid Of mapper parameters")
	}
	if Gia_ManHasChoices(g) != 0 {
		Gia_ManSetPhase(g)
	}
	n := int(g.nObjs)
	p := &ofCutManager{graph: g, pars: pars, storage: newOfCutStorage(), sets: make([]int32, n), flows: make([]int32, n), delays: make([]int32, n)}
	refs := vec.Vec_IntAlloc(0)
	Mf_ManSetFlowRefs(g, refs)
	p.refs = append([]int32(nil), vec.Vec_IntArray(refs)...)
	if pars.FCutMin != 0 {
		p.truth = &nfTruthStore{memory: vec.Vec_MemAllocForTT(6, 0)}
	}
	return p
}
func (p *ofCutManager) prepare(dst []ofCut, id int32, addUnit bool) int32 {
	if p.sets[id] == 0 {
		dst[0] = ofUnitCut(id)
		return 1
	}
	data := p.storage.at(p.sets[id])
	count, pos := data[0], 1
	for i := int32(0); i < count; i++ {
		size, function := ofStoredSize(data[pos:]), ofStoredFunction(data[pos:])
		pos++
		c := &dst[i]
		*c = ofCut{size: uint32(size), function: uint32(function)}
		copy(c.leaves[:], data[pos:pos+int(size)])
		pos += int(size) + 3
		c.sign = Nf_CutGetSign(c.leaves[:size])
	}
	if addUnit && count > 0 && dst[0].size > 1 {
		dst[count] = ofUnitCut(id)
		count++
	}
	return count
}
func (p *ofCutManager) params(c *ofCut, id int32) {
	c.delay, c.flow = 0, 0
	for _, leaf := range c.leaves[:c.size] {
		c.delay = max(c.delay, p.delays[leaf])
		c.flow += p.flows[leaf]
	}
	if c.size > 1 {
		c.delay++
		c.flow += 100 * (int32(c.size) + p.pars.NAreaTuner)
	}
	c.flow /= max(1, 2*p.refs[id])
}
func (p *ofCutManager) computeTruth(out *ofCut, op byte, inputs []*ofCut, complements []int32) {
	var converted [3]Nf_Cut_t
	var ptrs [3]*Nf_Cut_t
	for i, c := range inputs {
		converted[i] = Nf_Cut_t{iFunc: c.function, nLeaves: c.size, pLeaves: c.leaves}
		ptrs[i] = &converted[i]
	}
	dest := Nf_Cut_t{iFunc: out.function, nLeaves: out.size, pLeaves: out.leaves}
	p.truth.compute(ptrs[:len(inputs)], complements, &dest, op)
	out.function, out.size, out.leaves = dest.iFunc, dest.nLeaves, dest.pLeaves
	out.sign = Nf_CutGetSign(out.leaves[:out.size])
}
func (p *ofCutManager) merge(id int32) {
	obj := Gia_ManObj(p.graph, id)
	var left, right, third, workspace [32]ofCut
	var cuts [32]*ofCut
	for i := range cuts {
		cuts[i] = &workspace[i]
	}
	n0 := p.prepare(left[:], Gia_ObjFaninId0(obj, id), true)
	n1 := p.prepare(right[:], Gia_ObjFaninId1(obj, id), true)
	c0, c1 := Gia_ObjFaninC0(obj), Gia_ObjFaninC1(obj)
	limit, cap := p.pars.NLutSize, p.pars.NCutNum
	var n int32
	if sibling := Gia_ObjSibl(p.graph, id); sibling != 0 {
		count := p.prepare(third[:], sibling, false)
		complement := Gia_ObjPhase(obj) ^ Gia_ObjPhase(Gia_ObjSiblObj(p.graph, id))
		for i := int32(0); i < count; i++ {
			*cuts[n] = third[i]
			if p.truth != nil {
				cuts[n].function ^= uint32(complement)
			}
			p.params(cuts[n], id)
			n = ofAddCut(cuts[:], n, cap)
		}
	}
	if Gia_ObjIsMuxId(p.graph, id) != 0 {
		n2 := p.prepare(third[:], Gia_ObjFaninId2(p.graph, id), true)
		c2 := Gia_ObjFaninC2(p.graph, obj)
		p.counts[0] += float64(n0 * n1 * n2)
		for i := int32(0); i < n0; i++ {
			for j := int32(0); j < n1; j++ {
				for k := int32(0); k < n2; k++ {
					a, b, c := &left[i], &right[j], &third[k]
					if Nf_CutCountBits(a.sign|b.sign|c.sign) > limit {
						continue
					}
					p.counts[1]++
					if !ofMergeCuts(cuts[n], limit, a, b, c) || ofLastContained(cuts[:], n) {
						continue
					}
					p.counts[2]++
					if p.truth != nil {
						p.computeTruth(cuts[n], 'm', []*ofCut{a, b, c}, []int32{c0, c1, c2})
					}
					p.params(cuts[n], id)
					n = ofAddCut(cuts[:], n, cap)
				}
			}
		}
	} else {
		p.counts[0] += float64(n0 * n1)
		for i := int32(0); i < n0; i++ {
			for j := int32(0); j < n1; j++ {
				a, b := &left[i], &right[j]
				if int32(a.size+b.size) > limit && Nf_CutCountBits(a.sign|b.sign) > limit {
					continue
				}
				p.counts[1]++
				if !ofMergeCuts(cuts[n], limit, a, b) || ofLastContained(cuts[:], n) {
					continue
				}
				p.counts[2]++
				if p.truth != nil {
					op := byte('&')
					if Gia_ObjIsXor(obj) != 0 {
						op = '^'
					}
					p.computeTruth(cuts[n], op, []*ofCut{a, b}, []int32{c0, c1})
				}
				p.params(cuts[n], id)
				n = ofAddCut(cuts[:], n, cap)
			}
		}
	}
	if n <= 0 || n >= cap {
		panic("gia: invalid number of Of cuts")
	}
	p.flows[id], p.delays[id] = cuts[0].flow, cuts[0].delay
	p.sets[id] = p.storage.save(cuts[:n])
	p.counts[3] += float64(n)
}
func (p *ofCutManager) computeCuts() {
	for _, obj := range p.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			f := Gia_ObjFaninId0(obj, id)
			p.flows[id], p.delays[id] = p.flows[f], p.delays[f]
		} else {
			p.merge(id)
		}
	}
	for _, obj := range p.graph.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsBuf(obj) == 0 {
			p.storage.lift(p.sets[obj.id])
		}
	}
}
