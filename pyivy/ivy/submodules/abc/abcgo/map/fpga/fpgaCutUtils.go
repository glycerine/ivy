package fpga

func Fpga_CutAlloc(p *Fpga_Man_t) *Fpga_Cut_t { return &Fpga_Cut_t{} }
func Fpga_CutDup(p *Fpga_Man_t, c *Fpga_Cut_t) *Fpga_Cut_t {
	q := *c
	q.Leaves = append([]*Fpga_Node_t(nil), c.Leaves...)
	q.PinOrder = append([]int(nil), c.PinOrder...)
	q.regular = nil
	q.neg = nil
	return &q
}
func Fpga_CutFree(p *Fpga_Man_t, c *Fpga_Cut_t) {
	if c != nil {
		c.Next = nil
	}
}
func Fpga_CutCreateSimple(p *Fpga_Man_t, n *Fpga_Node_t) *Fpga_Cut_t {
	n = Fpga_Regular(n)
	leaf := n
	truth := uint64(2)
	if n.Repr != nil {
		leaf = n.Repr
		if n.Phase != leaf.Phase {
			truth ^= 3
		}
	}
	return &Fpga_Cut_t{Root: n, Leaves: []*Fpga_Node_t{leaf}, Truth: truth, Signature: cutSignature([]*Fpga_Node_t{leaf})}
}
func Fpga_CutGetRootArea(p *Fpga_Man_t, c *Fpga_Cut_t) float32 {
	if len(c.Leaves) == 0 {
		return 0
	}
	return p.Library.Areas[len(c.Leaves)]
}
func Fpga_CutListAppend(a, b *Fpga_Cut_t) *Fpga_Cut_t { return Fpga_CutUnionLists(a, b) }
func Fpga_CutListRecycle(p *Fpga_Man_t, list, save *Fpga_Cut_t) {
	for list != nil {
		next := list.Next
		if list != save {
			list.Next = nil
		}
		list = next
	}
}
func Fpga_CutListCount(c *Fpga_Cut_t) int32 {
	var n int32
	for ; c != nil; c = c.Next {
		n++
	}
	return n
}
func Fpga_CutGetParameters(p *Fpga_Man_t, c *Fpga_Cut_t) {
	Fpga_TimeCutComputeArrival(p, c)
	c.Flow = cutAreaFlow(p, c, true)
}
func Fpga_CutGetAreaFlow(p *Fpga_Man_t, c *Fpga_Cut_t) float32 {
	c.Flow = cutAreaFlow(p, c, false)
	return c.Flow
}
func cutAreaFlow(p *Fpga_Man_t, c *Fpga_Cut_t, estimated bool) float32 {
	area := Fpga_CutGetRootArea(p, c)
	for _, leaf := range c.Leaves {
		if Fpga_NodeIsAnd(leaf) == 0 {
			continue
		}
		refs := float32(leaf.Refs)
		if estimated {
			refs = leaf.EstimatedFanouts
		}
		if refs <= 0 {
			refs = 1
		}
		if leaf.Refs == 0 {
			refs = 1
		}
		area += leaf.Best.Flow / refs
	}
	return area
}

// Ref/Deref pair operations leave all counters unchanged when used for cost queries.
func Fpga_CutGetAreaRefed(p *Fpga_Man_t, c *Fpga_Cut_t) float32 {
	if len(c.Leaves) == 1 && c.One == nil {
		return 0
	}
	a := Fpga_CutDeref(p, c.Root, c, 0)
	Fpga_CutRef(p, c.Root, c, 0)
	return a
}
func Fpga_CutGetAreaDerefed(p *Fpga_Man_t, c *Fpga_Cut_t) float32 {
	if len(c.Leaves) == 1 && c.One == nil {
		return 0
	}
	a := Fpga_CutRef(p, c.Root, c, 0)
	Fpga_CutDeref(p, c.Root, c, 0)
	return a
}
func Fpga_CutRef(p *Fpga_Man_t, n *Fpga_Node_t, c *Fpga_Cut_t, fanouts int32) float32 {
	area := Fpga_CutGetRootArea(p, c)
	for _, leaf := range c.Leaves {
		leaf.Refs++
		if leaf.Refs == 1 && Fpga_NodeIsAnd(leaf) != 0 {
			area += Fpga_CutRef(p, leaf, leaf.Best, fanouts)
		}
	}
	return area
}
func Fpga_CutDeref(p *Fpga_Man_t, n *Fpga_Node_t, c *Fpga_Cut_t, fanouts int32) float32 {
	area := Fpga_CutGetRootArea(p, c)
	for _, leaf := range c.Leaves {
		if leaf.Refs <= 0 {
			panic("fpga: reference underflow")
		}
		leaf.Refs--
		if leaf.Refs == 0 && Fpga_NodeIsAnd(leaf) != 0 {
			area += Fpga_CutDeref(p, leaf, leaf.Best, fanouts)
		}
	}
	return area
}
func Fpga_MappingSetUsedCuts(p *Fpga_Man_t) {
	for _, n := range p.Nodes {
		for c := n.Cuts; c != nil; c = c.Next {
			c.Mark = false
		}
	}
	for _, n := range p.Mapping.Nodes {
		n.Best.Mark = true
	}
}
