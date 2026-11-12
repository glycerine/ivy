package cut

import "fmt"

func Cut_CutAlloc(p *Cut_Man_t) *Cut_Cut_t {
	p.check()
	c := &Cut_Cut_t{NVarsMax: p.params.NVarsMax, FSimul: p.simulation, owner: p, live: true}
	if p.params.FTruth != 0 {
		c.truth = make([]uint32, Cut_TruthWords(c.NVarsMax))
	}
	p.owned[c] = struct{}{}
	p.stats.Allocated++
	p.stats.Current++
	p.stats.Peak = max(p.stats.Peak, p.stats.Current)
	return c
}
func Cut_CutRecycle(p *Cut_Man_t, c *Cut_Cut_t) {
	if c == nil || !c.live || c.owner != p {
		panic("cut: foreign or recycled cut")
	}
	delete(p.owned, c)
	c.live = false
	c.owner = nil
	p.stats.Recycled++
	p.stats.Current--
	if len(c.leaves) == 1 {
		p.stats.Trivial--
	}
}
func Cut_CutCompare(a, b *Cut_Cut_t) int32 {
	if len(a.leaves) < len(b.leaves) {
		return -1
	}
	if len(a.leaves) > len(b.leaves) {
		return 1
	}
	for i, x := range a.leaves {
		if x < b.leaves[i] {
			return -1
		}
		if x > b.leaves[i] {
			return 1
		}
	}
	return 0
}
func Cut_CutDupList(p *Cut_Man_t, list *Cut_Cut_t) *Cut_Cut_t {
	var head *Cut_Cut_t
	tail := &head
	for c := list; c != nil; c = c.Next {
		copy := Cut_CutAlloc(p)
		copy.Num0, copy.Num1 = c.Num0, c.Num1
		copy.proof = c.proof
		copy.FSimul, copy.FCompl = c.FSimul, c.FCompl
		copy.USign, copy.UCanon0, copy.UCanon1 = c.USign, c.UCanon0, c.UCanon1
		copy.leaves = append([]int32(nil), c.leaves...)
		copy.truth = append(copy.truth[:0], c.truth...)
		if len(copy.leaves) == 1 {
			p.stats.Trivial++
		}
		*tail = copy
		tail = &copy.Next
	}
	return head
}
func Cut_CutRecycleList(p *Cut_Man_t, list *Cut_Cut_t) {
	for list != nil {
		next := list.Next
		Cut_CutRecycle(p, list)
		list = next
	}
}
func Cut_CutMergeLists(a, b *Cut_Cut_t) *Cut_Cut_t {
	var head *Cut_Cut_t
	tail := &head
	for a != nil && b != nil {
		var c *Cut_Cut_t
		if Cut_CutCompare(a, b) < 0 {
			c, a = a, a.Next
		} else {
			c, b = b, b.Next
		}
		*tail = c
		tail = &c.Next
	}
	if a != nil {
		*tail = a
	} else {
		*tail = b
	}
	return head
}
func Cut_CutNumberList(list *Cut_Cut_t) {
	var number uint16
	for c := list; c != nil; c = c.Next {
		c.Num0 = number & 2047
		number++
	}
}
func Cut_CutCreateTriv(p *Cut_Man_t, node int32) *Cut_Cut_t {
	if p.params.FSeq != 0 {
		node <<= CUT_SHIFT
	}
	c := Cut_CutAlloc(p)
	c.leaves = []int32{node}
	c.USign = Cut_NodeSign(node)
	for i := range c.truth {
		c.truth[i] = 0xaaaaaaaa
	}
	p.stats.Trivial++
	return c
}
func Cut_CutCountList(list *Cut_Cut_t) int32 {
	var n int32
	for ; list != nil; list = list.Next {
		n++
	}
	return n
}
func Cut_CutPrint(c *Cut_Cut_t, seq int32) {
	fmt.Printf("%d : {", len(c.leaves))
	for _, leaf := range c.leaves {
		if seq != 0 {
			fmt.Printf(" %d(%d)", leaf>>CUT_SHIFT, leaf&CUT_MASK)
		} else {
			fmt.Printf(" %d", leaf)
		}
	}
	fmt.Print(" }")
}
func Cut_CutPrintList(list *Cut_Cut_t, seq int32) {
	for c := list; c != nil; c = c.Next {
		Cut_CutPrint(c, seq)
		fmt.Println()
	}
}
func Cut_CutPrintMerge(c, a, b *Cut_Cut_t) {
	Cut_CutPrint(a, 0)
	fmt.Print(" + ")
	Cut_CutPrint(b, 0)
	fmt.Print(" = ")
	Cut_CutPrint(c, 0)
	fmt.Println()
}
