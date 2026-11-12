package fpga

import (
	"encoding/binary"
	"fmt"
	"sort"
)

func Fpga_CutRegular(c *Fpga_Cut_t) *Fpga_Cut_t {
	if c != nil && c.regular != nil {
		return c.regular
	}
	return c
}
func Fpga_CutIsComplement(c *Fpga_Cut_t) int32 { return boolInt(c != nil && c.regular != nil) }
func Fpga_CutNot(c *Fpga_Cut_t) *Fpga_Cut_t {
	if c == nil {
		return nil
	}
	if c.regular != nil {
		return c.regular
	}
	if c.neg == nil {
		c.neg = &Fpga_Cut_t{regular: c}
	}
	return c.neg
}
func Fpga_CutNotCond(c *Fpga_Cut_t, b int32) *Fpga_Cut_t {
	if b != 0 {
		return Fpga_CutNot(c)
	}
	return c
}
func leafKey(leaves []*Fpga_Node_t) string {
	b := make([]byte, 4*len(leaves))
	for i, n := range leaves {
		binary.LittleEndian.PutUint32(b[4*i:], uint32(n.Num))
	}
	return string(b)
}
func cutSignature(leaves []*Fpga_Node_t) uint32 {
	var s uint32
	for _, n := range leaves {
		s |= 1 << uint(uint32(n.Num)%31)
	}
	return s
}
func unionLeaves(a, b []*Fpga_Node_t, max int) []*Fpga_Node_t {
	r := make([]*Fpga_Node_t, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		var n *Fpga_Node_t
		if j == len(b) || i < len(a) && a[i].Num < b[j].Num {
			n = a[i]
			i++
		} else if i == len(a) || b[j].Num < a[i].Num {
			n = b[j]
			j++
		} else {
			n = a[i]
			i++
			j++
		}
		r = append(r, n)
		if len(r) > max {
			return nil
		}
	}
	return r
}
func expandTruth(c *Fpga_Cut_t, leaves []*Fpga_Node_t, complement bool) uint64 {
	positions := make([]int, len(c.Leaves))
	for i, n := range c.Leaves {
		positions[i] = sort.Search(len(leaves), func(j int) bool { return leaves[j].Num >= n.Num })
	}
	var result uint64
	for assignment := 0; assignment < 1<<len(leaves); assignment++ {
		index := 0
		for i, pos := range positions {
			index |= ((assignment >> pos) & 1) << i
		}
		if (c.Truth>>index&1 != 0) != complement {
			result |= 1 << assignment
		}
	}
	return result
}
func Fpga_MappingCreatePiCuts(p *Fpga_Man_t) {
	for i, n := range p.Inputs {
		n.Cuts = Fpga_CutCreateSimple(p, n)
		n.Best = n.Cuts
		n.Best.Arrival = p.InputArrivals[i]
	}
}
func Fpga_MappingCuts(p *Fpga_Man_t) {
	Fpga_MappingCreatePiCuts(p)
	t := Fpga_CutTableStart(p)
	for _, n := range p.Ands.Nodes {
		if Fpga_NodeIsAnd(n) == 0 {
			continue
		}
		n.Cuts = Fpga_CutCompute(p, t, n)
	}
	p.Stats.Cuts = int(Fpga_CutCountAll(p))
}
func Fpga_CutCompute(p *Fpga_Man_t, t *Fpga_CutTable_t, n *Fpga_Node_t) *Fpga_Cut_t {
	Fpga_CutTableRestart(t)
	list := Fpga_CutMergeLists(p, t, Fpga_Regular(n.One).Cuts, Fpga_Regular(n.Two).Cuts, Fpga_IsComplement(n.One), Fpga_IsComplement(n.Two), 0, 0)
	for alt := n.NextE; alt != nil; alt = alt.NextE {
		for c := alt.Cuts; c != nil; c = c.Next {
			if c == alt.Cuts {
				continue
			}
			if t.Cuts[leafKey(c.Leaves)] != nil {
				continue
			}
			copy := Fpga_CutDup(p, c)
			copy.Root = n
			copy.Next = nil
			if alt.Phase != n.Phase {
				copy.Truth ^= truthMask(len(copy.Leaves))
			}
			t.Cuts[leafKey(copy.Leaves)] = copy
			t.Array = append(t.Array, copy)
		}
	}
	_ = list
	cuts := append([]*Fpga_Cut_t(nil), t.Array...)
	sort.SliceStable(cuts, func(i, j int) bool { return len(cuts[i].Leaves) < len(cuts[j].Leaves) })
	limit := p.CutLimit
	if limit <= 0 {
		limit = FPGA_CUTS_MAX_USE
	}
	if len(cuts) > limit-1 {
		cuts = cuts[:limit-1]
	}
	simple := Fpga_CutCreateSimple(p, n)
	simple.Next = Fpga_CutArray2List(cuts, int32(len(cuts)))
	for c := simple.Next; c != nil; c = c.Next {
		c.Root = n
	}
	return simple
}

// The source's subset-dominance filter is disabled: a larger LUT can have a better library cost.
func Fpga_CutFilter(p *Fpga_Man_t, n *Fpga_Node_t) {
	t := Fpga_CutTableStart(p)
	n.Cuts = Fpga_CutSortCuts(p, t, n.Cuts)
}
func Fpga_CutMergeLists(p *Fpga_Man_t, t *Fpga_CutTable_t, a, b *Fpga_Cut_t, ca, cb, pa, pb int32) *Fpga_Cut_t {
	for x := a; x != nil; x = x.Next {
		for y := b; y != nil; y = y.Next {
			leaves := unionLeaves(x.Leaves, y.Leaves, int(p.Library.LutMax))
			if leaves == nil {
				continue
			}
			key := leafKey(leaves)
			if t.Cuts[key] != nil {
				continue
			}
			c := &Fpga_Cut_t{Leaves: leaves, One: x, Two: y, Signature: cutSignature(leaves)}
			c.Truth = expandTruth(x, leaves, ca != 0) & expandTruth(y, leaves, cb != 0)
			t.Cuts[key] = c
			t.Array = append(t.Array, c)
			if len(t.Array) >= FPGA_CUTS_MAX_COMPUTE {
				return Fpga_CutArray2List(t.Array, int32(len(t.Array)))
			}
		}
	}
	return Fpga_CutArray2List(t.Array, int32(len(t.Array)))
}
func Fpga_CutMergeLists2(p *Fpga_Man_t, t *Fpga_CutTable_t, a, b *Fpga_Cut_t, ca, cb, pa, pb int32) *Fpga_Cut_t {
	return Fpga_CutMergeLists(p, t, a, b, ca, cb, pa, pb)
}
func Fpga_CutMergeTwo(a, b *Fpga_Cut_t, nodes []*Fpga_Node_t, max int32) int32 {
	r := unionLeaves(a.Leaves, b.Leaves, int(max))
	if r == nil {
		return 0
	}
	copy(nodes, r)
	return int32(len(r))
}
func Fpga_CutUnionLists(a, b *Fpga_Cut_t) *Fpga_Cut_t {
	if a == nil {
		return b
	}
	last := a
	for last.Next != nil {
		last = last.Next
	}
	last.Next = b
	return a
}
func Fpga_CutBelongsToList(list *Fpga_Cut_t, nodes []*Fpga_Node_t, n int32) int32 {
	key := leafKey(nodes[:n])
	for c := list; c != nil; c = c.Next {
		if leafKey(c.Leaves) == key {
			return 1
		}
	}
	return 0
}
func Fpga_CutCountAll(p *Fpga_Man_t) int32 {
	var total int32
	for _, n := range p.Nodes {
		total += Fpga_CutListCount(n.Cuts)
	}
	return total
}
func Fpga_CutsCleanSign(p *Fpga_Man_t) {
	for _, n := range p.Nodes {
		for c := n.Cuts; c != nil; c = c.Next {
			c.Signature = 0
		}
	}
}
func Fpga_CutsCleanRoot(p *Fpga_Man_t) {
	for _, n := range p.Nodes {
		for c := n.Cuts; c != nil; c = c.Next {
			c.Mark = false
		}
	}
}
func Fpga_CutListPrint(p *Fpga_Man_t, n *Fpga_Node_t) {
	for c := n.Cuts; c != nil; c = c.Next {
		Fpga_CutPrint(p, n, c)
	}
}
func Fpga_CutListPrint2(p *Fpga_Man_t, n *Fpga_Node_t)            { Fpga_CutListPrint(p, n) }
func Fpga_CutPrint_(p *Fpga_Man_t, c *Fpga_Cut_t, n *Fpga_Node_t) { Fpga_CutPrint(p, n, c) }
func Fpga_CutTableStart(p *Fpga_Man_t) *Fpga_CutTable_t {
	return &Fpga_CutTable_t{Manager: p, Cuts: make(map[string]*Fpga_Cut_t)}
}
func Fpga_CutTableStop(t *Fpga_CutTable_t) { t.Cuts = nil; t.Array = nil }
func Fpga_CutTableHash(nodes []*Fpga_Node_t, n int32) uint32 {
	var h uint32
	for _, v := range nodes[:n] {
		h = h*12582917 + uint32(v.Num)
	}
	return h
}
func Fpga_CutTableLookup(t *Fpga_CutTable_t, nodes []*Fpga_Node_t, n int32) int32 {
	return boolInt(t.Cuts[leafKey(nodes[:n])] != nil)
}
func Fpga_CutTableConsider(p *Fpga_Man_t, t *Fpga_CutTable_t, nodes []*Fpga_Node_t, n int32) *Fpga_Cut_t {
	key := leafKey(nodes[:n])
	if t.Cuts[key] != nil {
		return nil
	}
	c := &Fpga_Cut_t{Leaves: append([]*Fpga_Node_t(nil), nodes[:n]...)}
	c.Signature = cutSignature(c.Leaves)
	t.Cuts[key] = c
	t.Array = append(t.Array, c)
	return c
}
func Fpga_CutTableRestart(t *Fpga_CutTable_t) { clear(t.Cuts); t.Array = t.Array[:0] }
func Fpga_CutSortCutsCompare(a, b **Fpga_Cut_t) int32 {
	return int32(len((*a).Leaves) - len((*b).Leaves))
}
func Fpga_CutSortCuts(p *Fpga_Man_t, t *Fpga_CutTable_t, list *Fpga_Cut_t) *Fpga_Cut_t {
	a := []*Fpga_Cut_t{}
	for c := list; c != nil; c = c.Next {
		a = append(a, c)
	}
	sort.SliceStable(a, func(i, j int) bool { return len(a[i].Leaves) < len(a[j].Leaves) })
	return Fpga_CutArray2List(a, int32(len(a)))
}
func Fpga_CutList2Array(a []*Fpga_Cut_t, list *Fpga_Cut_t) int32 {
	var n int32
	for c := list; c != nil; c = c.Next {
		a[n] = c
		n++
	}
	return n
}
func Fpga_CutArray2List(a []*Fpga_Cut_t, n int32) *Fpga_Cut_t {
	if n == 0 {
		return nil
	}
	for i := 0; i < int(n)-1; i++ {
		a[i].Next = a[i+1]
	}
	a[n-1].Next = nil
	return a[0]
}
func Fpga_CutPrint(p *Fpga_Man_t, n *Fpga_Node_t, c *Fpga_Cut_t) {
	fmt.Fprintf(p.Output, "%d:", n.Num)
	for _, leaf := range c.Leaves {
		fmt.Fprintf(p.Output, " %d", leaf.Num)
	}
	fmt.Fprintf(p.Output, " arrival=%g flow=%g\n", c.Arrival, c.Flow)
}
