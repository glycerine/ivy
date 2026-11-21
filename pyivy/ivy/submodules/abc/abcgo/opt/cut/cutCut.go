package cut

import "fmt"

func (p *Cut_Man_t) countAllocation() {
	p.check()
	p.stats.Allocated++
	p.stats.Current++
	p.stats.Peak = max(p.stats.Peak, p.stats.Current)
}

func Cut_CutAlloc(p *Cut_Man_t) *Cut_Cut_t {
	p.countAllocation()
	return p.newCut()
}

func (p *Cut_Man_t) newCut() *Cut_Cut_t {
	c := &Cut_Cut_t{NVarsMax: p.params.NVarsMax, FSimul: p.simulation, owner: p.lifetime}
	c.self = c
	if p.params.FTruth != 0 {
		if words := Cut_TruthWords(c.NVarsMax); words == 1 {
			c.truth = c.shortTruth[:]
		} else {
			c.truth = make([]uint32, words)
		}
	}
	return c
}
func Cut_CutRecycle(p *Cut_Man_t, c *Cut_Cut_t) {
	if p == nil || c == nil || c.self != c || c.owner == nil || c.owner != p.lifetime || c.owner.stopped {
		panic("cut: foreign or recycled cut")
	}
	c.owner = nil
	c.self = nil
	if c.temporary {
		return
	}
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

// copyPayload gives the destination its own mutable boundary and truth words.
func (c *Cut_Cut_t) copyPayload(source *Cut_Cut_t) {
	c.FSimul, c.FCompl, c.USign = source.FSimul, source.FCompl, source.USign
	if source.leaves != nil {
		c.setLeaves(source.leaves)
	}
	if len(source.truth) == 1 {
		c.shortTruth[0] = source.truth[0]
		c.truth = c.shortTruth[:]
	} else {
		c.truth = append(c.truth[:0], source.truth...)
	}
}

// Buffer and oracle outputs are regular cuts. A sequential input snapshot is
// temporary and ABC excludes it from the manager's cut statistics.
func (p *Cut_Man_t) copyBufferCut(source *Cut_Cut_t) *Cut_Cut_t {
	c := Cut_CutAlloc(p)
	c.copyPayload(source)
	if len(c.leaves) == 1 {
		p.stats.Trivial++
	}
	return c
}

func Cut_CutDupList(p *Cut_Man_t, list *Cut_Cut_t) *Cut_Cut_t {
	if list == nil {
		return nil
	}
	p.check()
	var head *Cut_Cut_t
	tail := &head
	for list != nil {
		count := 0
		for next := list; next != nil && count < 32; next = next.Next {
			count++
		}
		block := make([]Cut_Cut_t, count)
		for i := range block {
			c := &block[i]
			c.self = c
			c.owner, c.temporary, c.NVarsMax = p.lifetime, true, list.NVarsMax
			c.Num0, c.Num1, c.proof = list.Num0, list.Num1, list.proof
			c.UCanon0, c.UCanon1 = list.UCanon0, list.UCanon1
			c.copyPayload(list)
			*tail, tail = c, &c.Next
			list = list.Next
		}
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
	c.shortLeaves[0] = node
	c.leaves = c.shortLeaves[:1]
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
