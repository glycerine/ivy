package aig

import (
	"fmt"
	"math/bits"
	"time"
)

/**CFile****************************************************************

  FileName    [aigCuts.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Computation of K-feasible priority cuts.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigCuts.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

// Aig_ManCutStart creates bounded linked cut slots. Graph structure must remain
// unchanged until Stop. Slot validity is separate from zero-leaf constants.
func Aig_ManCutStart(g *Graph, nCutsMax, nLeafMax, fTruth, fVerbose int32) *Aig_ManCut_t {
	if g == nil || nCutsMax < 2 || nLeafMax < 1 || nLeafMax > 16 {
		panic("aig: invalid priority-cut parameters")
	}
	words := int32(cutTruthWords(int(nLeafMax)))
	truthBytes := int32(0)
	if fTruth != 0 {
		truthBytes = 4 * words
	}
	return &Aig_ManCut_t{
		pAig: g, pCuts: make([]*Aig_Cut_t, len(g.nodes)),
		nCutsMax: nCutsMax, nLeafMax: nLeafMax, fTruth: fTruth, fVerbose: fVerbose,
		nTruthWords: words, nCutSize: 24 + 4*nLeafMax + truthBytes,
	}
}
func Aig_ManCutStop(p *Aig_ManCut_t) {
	if p == nil {
		return
	}
	if p.pAig != nil && p.pAig.cutManager == p {
		p.pAig.cutManager = nil
	}
	*p = Aig_ManCut_t{}
}
func cutTruthWords(leaves int) int {
	if leaves <= 5 {
		return 1
	}
	return 1 << uint(leaves-5)
}
func cutRepeatTruth(truth []uint32, words int) {
	for i := words; i < len(truth); i++ {
		truth[i] = truth[i%words]
	}
}
func cutLive(c *Aig_Cut_t) bool { return c != nil && (c.valid || len(c.pFanins) > 0) }
func cutSetLeaves(c *Aig_Cut_t, leaves []int32) {
	c.pFanins = append(c.pFanins[:0], leaves...)
	c.nFanins = byte(len(leaves))
	c.uSign = 0
	for _, id := range leaves {
		c.uSign |= Aig_ObjCutSign(uint32(id))
	}
}
func cutClear(c *Aig_Cut_t) {
	c.valid = false
	c.nFanins = 0
	c.pFanins = nil
	c.truth = nil
	c.uSign = 0
	c.Cost = 0
}
func cutCopy(dst, src *Aig_Cut_t) {
	next, node, size, max := dst.pNext, dst.iNode, dst.nCutSize, dst.nLeafMax
	*dst = *src
	dst.pNext = next
	dst.iNode = node
	dst.nCutSize = size
	dst.nLeafMax = max
	dst.pFanins = append([]int32(nil), src.pFanins...)
	dst.truth = append([]uint32(nil), src.truth...)
}
func Aig_CutPrint(c *Aig_Cut_t) {
	fmt.Print("{")
	for _, id := range c.pFanins {
		fmt.Printf(" %d", id)
	}
	fmt.Println(" }")
}
func Aig_ObjCutPrint(p *Aig_ManCut_t, e Edge) {
	fmt.Printf("Cuts for node %d:\n", e.NodeID())
	for c := Aig_ObjCuts(p, e); c != nil; c = c.pNext {
		if cutLive(c) {
			Aig_CutPrint(c)
		}
	}
}
func Aig_ManCutCount(p *Aig_ManCut_t, full *int32) int32 {
	var total, k int32
	for id, n := range p.pAig.nodes {
		if p.pAig.deleted[id] || n.Kind != And && n.Kind != ExclusiveOr {
			continue
		}
		for c := p.pCuts[id]; c != nil; c = c.pNext {
			if cutLive(c) {
				total++
				if len(c.pFanins) == int(p.nLeafMax) {
					k++
				}
			}
		}
	}
	if full != nil {
		*full = k
	}
	return total
}
func Aig_CutFindCost(p *Aig_ManCut_t, c *Aig_Cut_t) int32 {
	if len(c.pFanins) == 0 {
		return 0
	}
	var cost int64
	for _, id := range c.pFanins {
		cost += int64(p.pAig.nodes[id].Refs)
	}
	return int32(cost * 1000 / int64(len(c.pFanins)))
}
func Aig_CutFindCost2(p *Aig_ManCut_t, c *Aig_Cut_t) float32 {
	if len(c.pFanins) == 0 {
		return 0
	}
	var cost float32
	for _, id := range c.pFanins {
		refs := p.pAig.nodes[id].Refs
		if refs == 0 {
			return 0
		}
		cost += 1 / float32(refs)
	}
	return 1 / cost
}

// Aig_CutFindFree preserves the unit cut and clears the worst non-unit slot if
// all slots are occupied. Admission should be checked before calling it.
func Aig_CutFindFree(p *Aig_ManCut_t, e Edge) *Aig_Cut_t {
	var worst *Aig_Cut_t
	for c := Aig_ObjCuts(p, e); c != nil; c = c.pNext {
		if !cutLive(c) {
			return c
		}
		if len(c.pFanins) == 1 && c.pFanins[0] == int32(e.NodeID()) {
			continue
		}
		if worst == nil || worst.Cost < c.Cost {
			worst = c
		}
	}
	if worst == nil {
		panic("aig: no replaceable cut slot")
	}
	cutClear(worst)
	return worst
}
func Aig_CutTruthPhase(merged, child *Aig_Cut_t) uint32 {
	var phase uint32
	j := 0
	for i, id := range merged.pFanins {
		if j < len(child.pFanins) && id == child.pFanins[j] {
			phase |= 1 << uint(i)
			j++
		}
	}
	if j != len(child.pFanins) {
		panic("aig: child cut is not contained in merged cut")
	}
	return phase
}
func cutPositions(merged, child *Aig_Cut_t) []int {
	phase := Aig_CutTruthPhase(merged, child)
	positions := make([]int, 0, len(child.pFanins))
	for i := range merged.pFanins {
		if phase>>uint(i)&1 != 0 {
			positions = append(positions, i)
		}
	}
	return positions
}
func cutProjectedIndex(assignment int, positions []int) int {
	index := 0
	for i, pos := range positions {
		index |= ((assignment >> uint(pos)) & 1) << uint(i)
	}
	return index
}
func cutTruthBit(truth []uint32, index int) bool { return truth[index>>5]>>uint(index&31)&1 != 0 }
func cutCombineTruth(p *Aig_ManCut_t, out, a, b *Aig_Cut_t, complA, complB bool, kind NodeKind) []uint32 {
	if p.fTruth == 0 || len(a.truth) < int(p.nTruthWords) || len(b.truth) < int(p.nTruthWords) {
		panic("aig: truth tables unavailable")
	}
	posA, posB := cutPositions(out, a), cutPositions(out, b)
	truth := make([]uint32, p.nTruthWords)
	words := cutTruthWords(len(out.pFanins))
	for bit := 0; bit < words*32; bit++ {
		va := cutTruthBit(a.truth, cutProjectedIndex(bit, posA)) != complA
		vb := cutTruthBit(b.truth, cutProjectedIndex(bit, posB)) != complB
		value := va && vb
		if kind == ExclusiveOr {
			value = va != vb
		}
		if value {
			truth[bit>>5] |= 1 << uint(bit&31)
		}
	}
	cutRepeatTruth(truth, words)
	out.truth = truth
	return truth
}
func Aig_CutComputeTruth(p *Aig_ManCut_t, out, a, b *Aig_Cut_t, complA, complB int32) []uint32 {
	return cutCombineTruth(p, out, a, b, complA != 0, complB != 0, And)
}
func Aig_CutSupportMinimize(p *Aig_ManCut_t, c *Aig_Cut_t) int32 {
	if p.fTruth == 0 || len(c.truth) < int(p.nTruthWords) {
		panic("aig: support minimization requires truth")
	}
	var support []int
	for variable := range c.pFanins {
		essential := false
		for bit := 0; bit < 1<<uint(len(c.pFanins)); bit++ {
			if bit>>uint(variable)&1 == 0 && cutTruthBit(c.truth, bit) != cutTruthBit(c.truth, bit|(1<<uint(variable))) {
				essential = true
				break
			}
		}
		if essential {
			support = append(support, variable)
		}
	}
	if len(support) == len(c.pFanins) {
		return int32(len(support))
	}
	leaves := make([]int32, len(support))
	for i, pos := range support {
		leaves[i] = c.pFanins[pos]
	}
	truth := make([]uint32, p.nTruthWords)
	words := cutTruthWords(len(support))
	for bit := 0; bit < words*32; bit++ {
		oldIndex := 0
		for i, pos := range support {
			oldIndex |= ((bit >> uint(i)) & 1) << uint(pos)
		}
		if cutTruthBit(c.truth, oldIndex) {
			truth[bit>>5] |= 1 << uint(bit&31)
		}
	}
	cutRepeatTruth(truth, words)
	cutSetLeaves(c, leaves)
	c.truth = truth
	c.valid = true
	return int32(len(leaves))
}
func Aig_CutCheckDominance(a, b *Aig_Cut_t) bool {
	i, j := 0, 0
	for i < len(a.pFanins) && j < len(b.pFanins) {
		if a.pFanins[i] == b.pFanins[j] {
			i++
			j++
		} else if a.pFanins[i] > b.pFanins[j] {
			j++
		} else {
			return false
		}
	}
	return i == len(a.pFanins)
}
func Aig_CutFilter(p *Aig_ManCut_t, e Edge, candidate *Aig_Cut_t) bool {
	for old := Aig_ObjCuts(p, e); old != nil; old = old.pNext {
		if old == candidate || !cutLive(old) || len(old.pFanins) == 1 && old.pFanins[0] == int32(e.NodeID()) {
			continue
		}
		if len(old.pFanins) > len(candidate.pFanins) {
			if Aig_CutCheckDominance(candidate, old) {
				cutClear(old)
			}
		} else if Aig_CutCheckDominance(old, candidate) {
			cutClear(candidate)
			return true
		}
	}
	return false
}
func Aig_CutMergeOrdered(p *Aig_ManCut_t, a, b, out *Aig_Cut_t) bool {
	capacity := len(a.pFanins) + len(b.pFanins)
	if capacity > int(p.nLeafMax) {
		capacity = int(p.nLeafMax)
	}
	leaves := make([]int32, 0, capacity)
	i, j := 0, 0
	for i < len(a.pFanins) || j < len(b.pFanins) {
		var id int32
		if j == len(b.pFanins) || i < len(a.pFanins) && a.pFanins[i] < b.pFanins[j] {
			id = a.pFanins[i]
			i++
		} else if i == len(a.pFanins) || b.pFanins[j] < a.pFanins[i] {
			id = b.pFanins[j]
			j++
		} else {
			id = a.pFanins[i]
			i++
			j++
		}
		if len(leaves) == int(p.nLeafMax) {
			return false
		}
		leaves = append(leaves, id)
	}
	cutSetLeaves(out, leaves)
	out.valid = true
	return true
}
func Aig_CutMerge(p *Aig_ManCut_t, a, b, out *Aig_Cut_t) bool {
	return Aig_CutMergeOrdered(p, a, b, out)
}
func Aig_ObjPrepareCuts(p *Aig_ManCut_t, e Edge, trivial int32) *Aig_Cut_t {
	if !p.pAig.valid(e) || e.NodeID() >= len(p.pCuts) || p.pCuts[e.NodeID()] != nil {
		panic("aig: cut object already prepared or invalid")
	}
	slots := make([]Aig_Cut_t, p.nCutsMax)
	for i := range slots {
		c := &slots[i]
		c.iNode = int32(e.NodeID())
		c.nCutSize = int16(p.nCutSize)
		c.nLeafMax = byte(p.nLeafMax)
		if i+1 < len(slots) {
			c.pNext = &slots[i+1]
		}
	}
	p.pCuts[e.NodeID()] = &slots[0]
	c := &slots[0]
	if p.pAig.nodes[e.NodeID()].Kind == Constant {
		c.valid = true
		if p.fTruth != 0 {
			c.truth = make([]uint32, p.nTruthWords)
			for i := range c.truth {
				c.truth[i] = ^uint32(0)
			}
		}
	} else if trivial != 0 {
		c.valid = true
		cutSetLeaves(c, []int32{int32(e.NodeID())})
		if p.fTruth != 0 {
			c.truth = make([]uint32, p.nTruthWords)
			for i := range c.truth {
				c.truth[i] = 0xaaaaaaaa
			}
		}
	}
	return c
}
func cutCandidate(p *Aig_ManCut_t, e Edge) *Aig_Cut_t {
	return &Aig_Cut_t{iNode: int32(e.NodeID()), nLeafMax: byte(p.nLeafMax), nCutSize: int16(p.nCutSize), valid: true}
}
func cutAdmit(p *Aig_ManCut_t, e Edge, candidate *Aig_Cut_t) {
	if Aig_CutFilter(p, e, candidate) {
		return
	}
	candidate.Cost = Aig_CutFindCost(p, candidate)
	slot := Aig_CutFindFree(p, e)
	cutCopy(slot, candidate)
}
func Aig_ObjComputeCuts(p *Aig_ManCut_t, e Edge, trivial int32) {
	node := p.pAig.node(e)
	if node.Kind != And && node.Kind != ExclusiveOr && node.Kind != Buffer && node.Kind != Output {
		panic("aig: cut computation requires logic node")
	}
	Aig_ObjPrepareCuts(p, e, trivial)
	fanA := node.Fanins[0]
	a := p.pCuts[fanA.NodeID()]
	if a == nil {
		panic("aig: fanin cuts not computed")
	}
	if node.Kind == Buffer || node.Kind == Output {
		for old := a; old != nil; old = old.pNext {
			if !cutLive(old) {
				continue
			}
			candidate := cutCandidate(p, e)
			cutSetLeaves(candidate, old.pFanins)
			if p.fTruth != 0 {
				candidate.truth = append([]uint32(nil), old.truth...)
				if fanA.Complemented() {
					for i := range candidate.truth {
						candidate.truth[i] = ^candidate.truth[i]
					}
				}
			}
			cutAdmit(p, e, candidate)
		}
		return
	}
	fanB := node.Fanins[1]
	b := p.pCuts[fanB.NodeID()]
	if b == nil {
		panic("aig: fanin cuts not computed")
	}
	for x := a; x != nil; x = x.pNext {
		if !cutLive(x) {
			continue
		}
		for y := b; y != nil; y = y.pNext {
			if !cutLive(y) {
				continue
			}
			if bits.OnesCount32(x.uSign|y.uSign) > int(p.nLeafMax) {
				continue
			}
			candidate := cutCandidate(p, e)
			if !Aig_CutMerge(p, x, y, candidate) {
				continue
			}
			if p.fTruth != 0 {
				cutCombineTruth(p, candidate, x, y, fanA.Complemented(), fanB.Complemented(), node.Kind)
			}
			cutAdmit(p, e, candidate)
		}
	}
}
func Aig_ComputeCuts(g *Graph, nCutsMax, nLeafMax, fTruth, fVerbose int32) *Aig_ManCut_t {
	if g.cutManager != nil {
		panic("aig: graph already owns a cut manager")
	}
	start := time.Now()
	p := Aig_ManCutStart(g, nCutsMax, nLeafMax, fTruth, fVerbose)
	roots := make([]Edge, 0, len(g.nodes))
	for id := range g.nodes {
		if !g.deleted[id] {
			roots = append(roots, Edge(id)<<1)
		}
	}
	for _, e := range g.dfs(roots, false, false, false) {
		kind := g.nodes[e.NodeID()].Kind
		switch kind {
		case Constant, Input:
			Aig_ObjPrepareCuts(p, e, 1)
		case Output:
			Aig_ObjComputeCuts(p, e, 0)
		default:
			Aig_ObjComputeCuts(p, e, 1)
		}
	}
	if fVerbose != 0 {
		var full int32
		total := Aig_ManCutCount(p, &full)
		fmt.Printf("Nodes = %6d. Total cuts = %6d. %d-input cuts = %6d. Runtime = %s\n", len(roots), total, nLeafMax, full, time.Since(start))
	}
	g.cutManager = p
	return p
}

// Cuts returns active cut records in priority-slot order. The records belong to
// the manager; Leaves/Truth provide independent snapshots for external callers.
func (p *Aig_ManCut_t_) Cuts(e Edge) []*Aig_Cut_t {
	var cuts []*Aig_Cut_t
	for c := Aig_ObjCuts(p, e); c != nil; c = c.pNext {
		if cutLive(c) {
			cuts = append(cuts, c)
		}
	}
	return cuts
}
func (c *Aig_Cut_t) Leaves() []int32 { return append([]int32(nil), c.pFanins...) }
func (c *Aig_Cut_t) Truth() []uint32 { return append([]uint32(nil), c.truth...) }

// Valid distinguishes a computed zero-support constant cut from an unused slot.
func (c *Aig_Cut_t) Valid() bool { return cutLive(c) }
