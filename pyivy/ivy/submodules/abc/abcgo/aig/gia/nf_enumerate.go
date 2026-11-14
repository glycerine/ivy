package gia

import "github.com/glycerine/abcgo/misc/vec"

// nfCutManager enumerates and stores cuts in topological order. The library
// truth table is shared so newly encountered functions retain stable IDs.
type nfCutManager struct {
	graph        *Gia_Man_t
	pars         *Jf_Par_t
	library      *nfLibrary
	storage      *nfCutStorage
	sets, delays []int32
	flows, refs  []float32
	counts       [6]float64
	exactFlow    bool
}

func newNfCutManager(g *Gia_Man_t, pars *Jf_Par_t, library *nfLibrary) *nfCutManager {
	if pars.NCutNum <= 1 || pars.NCutNum > NF_CUT_MAX || pars.NLutSize <= 1 || pars.NLutSize > NF_LEAF_MAX || library == nil || library.truth == nil {
		panic("gia: invalid Nf mapper parameters")
	}
	if Gia_ManHasChoices(g) != 0 {
		Gia_ManSetPhase(g)
	}
	n := int(g.nObjs)
	p := &nfCutManager{graph: g, pars: pars, library: library, storage: newNfCutStorage(), sets: make([]int32, n), delays: make([]int32, n), flows: make([]float32, n), refs: make([]float32, n)}
	refs := vec.Vec_IntAlloc(0)
	Mf_ManSetFlowRefs(g, refs)
	for i, v := range vec.Vec_IntArray(refs) {
		p.refs[i] = 2 * float32(v)
	}
	return p
}

func (p *nfCutManager) prepare(dst []Nf_Cut_t, id int32, addUnit bool) int32 {
	if p.sets[id] == 0 {
		return Nf_CutCreateUnit(&dst[0], id)
	}
	data := p.storage.at(p.sets[id])
	count, pos := data[0], 1
	for i := int32(0); i < count; i++ {
		size, function := Nf_CutSize(data[pos:]), Nf_CutFunc(data[pos:])
		pos++
		c := &dst[i]
		*c = Nf_Cut_t{iFunc: uint32(function), nLeaves: uint32(size), Useless: uint32(giaBool(int(function>>1) >= p.library.truth.matchCount))}
		copy(c.pLeaves[:], data[pos:pos+int(size)])
		pos += int(size)
		c.Sign = Nf_CutGetSign(c.pLeaves[:size])
	}
	if addUnit && count > 0 && dst[0].nLeaves > 1 {
		count += Nf_CutCreateUnit(&dst[count], id)
	}
	return count
}

func (p *nfCutManager) params(c *Nf_Cut_t, refs float32) {
	c.Delay, c.Flow = 0, 0
	for _, id := range c.pLeaves[:c.nLeaves] {
		c.Delay = max(c.Delay, p.delays[id])
		c.Flow += p.flows[id]
	}
	if c.nLeaves > 1 {
		c.Delay++
		c.Flow += float32(int32(c.nLeaves) + p.pars.NAreaTuner)
	}
	c.Flow /= refs
}

func (p *nfCutManager) merge(id int32) {
	obj := Gia_ManObj(p.graph, id)
	var left, right, third, workspace [NF_CUT_MAX]Nf_Cut_t
	var cuts [NF_CUT_MAX]*Nf_Cut_t
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
			cuts[n].iFunc ^= uint32(complement)
			p.params(cuts[n], p.refs[id])
			n = p.addCut(cuts[:], n, cap)
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
					if Nf_CutCountBits(a.Sign|b.Sign|c.Sign) > limit {
						continue
					}
					p.counts[1]++
					if Nf_CutMergeOrderMux(a, b, c, cuts[n], limit) == 0 || Nf_SetLastCutIsContained(cuts[:], n) != 0 {
						continue
					}
					p.counts[2]++
					if Nf_CutComputeTruthMux6(p.library.truth, a, b, c, c0, c1, c2, cuts[n]) != 0 {
						cuts[n].Sign = Nf_CutGetSign(cuts[n].pLeaves[:cuts[n].nLeaves])
					}
					p.params(cuts[n], p.refs[id])
					n = p.addCut(cuts[:], n, cap)
				}
			}
		}
	} else {
		p.counts[0] += float64(n0 * n1)
		for i := int32(0); i < n0; i++ {
			for j := int32(0); j < n1; j++ {
				a, b := &left[i], &right[j]
				if int32(a.nLeaves+b.nLeaves) > limit && Nf_CutCountBits(a.Sign|b.Sign) > limit {
					continue
				}
				p.counts[1]++
				if Nf_CutMergeOrder(a, b, cuts[n], limit) == 0 || Nf_SetLastCutIsContained(cuts[:], n) != 0 {
					continue
				}
				p.counts[2]++
				if Nf_CutComputeTruth6(p.library.truth, a, b, c0, c1, cuts[n], Gia_ObjIsXor(obj)) != 0 {
					cuts[n].Sign = Nf_CutGetSign(cuts[n].pLeaves[:cuts[n].nLeaves])
				}
				p.params(cuts[n], p.refs[id])
				n = p.addCut(cuts[:], n, cap)
			}
		}
	}
	if n <= 0 || n >= cap {
		panic("gia: invalid number of Nf cuts")
	}
	p.flows[id], p.delays[id] = cuts[0].Flow, cuts[0].Delay
	p.sets[id] = p.storage.save(cuts[:n], false)
	p.counts[3] += float64(n)
	for _, c := range cuts[:n] {
		if c.Useless == 0 {
			p.counts[4]++
			p.counts[5] += float64(len(p.library.matches[c.iFunc>>1]))
		}
	}
}

func (p *nfCutManager) computeCuts() {
	for _, obj := range p.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjIsBuf(obj) != 0 {
			fanin := Gia_ObjFaninId0(obj, id)
			p.flows[id], p.delays[id] = p.flows[fanin], p.delays[fanin]
		} else {
			p.merge(id)
		}
	}
}

// addCut uses Pf's exact flow comparison when requested; Nf uses its
// epsilon-aware comparison.
func (p *nfCutManager) addCut(cuts []*Nf_Cut_t, n, limit int32) int32 {
	if !p.exactFlow {
		return Nf_SetAddCut(cuts, n, limit)
	}
	if n == 0 {
		return 1
	}
	n = Nf_SetLastCutContainsArea(cuts, n)
	for i := n; i > 0; i-- {
		a, b := cuts[i-1], cuts[i]
		less := a.Useless < b.Useless || a.Useless == b.Useless && (a.Flow < b.Flow || a.Flow == b.Flow && (a.Delay < b.Delay || a.Delay == b.Delay && a.nLeaves < b.nLeaves))
		if less {
			break
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
	return min(n+1, limit-1)
}
