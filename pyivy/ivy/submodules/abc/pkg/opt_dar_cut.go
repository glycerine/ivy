package abc

// Four-input cut enumeration from opt/dar/darCut.c and darCore.c.

import "math/bits"

type DARCutOptions struct {
	CutsPerNode                       int // zero: ten; range 2..255
	SkipTruthMinimization             bool
	MaxObjects, MaxCutSlots, MaxPairs int // zero: 1M, 8M, 16M
	Stop                              func() bool
}
type DARCuts4 struct {
	Graph      *AigManager  // cleaned owned graph, preserving object IDs
	Candidates []AigCNFCut4 // used nontrivial slots in native object/slot order
}
type darCut4 struct {
	leaves [4]AigID
	n      int
	truth  uint16
	sign   uint32
	value  uint16
	used   bool
}

func darContains(a, b *darCut4) bool {
	for i := 0; i < a.n; i++ {
		found := false
		for j := 0; j < b.n; j++ {
			found = found || a.leaves[i] == b.leaves[j]
		}
		if !found {
			return false
		}
	}
	return true
}
func darFilter(set []darCut4, index int) bool {
	c := &set[index]
	for i := range set {
		p := &set[i]
		if i == index || !p.used {
			continue
		}
		if p.n > c.n {
			if p.sign&c.sign == c.sign && darContains(c, p) {
				p.used = false
			}
		} else if p.sign&c.sign == p.sign && darContains(p, c) {
			c.used = false
			return true
		}
	}
	return false
}
func darMerge(a, b darCut4) (darCut4, bool) {
	c := darCut4{sign: a.sign | b.sign, used: true}
	i, j := 0, 0
	for i < a.n || j < b.n {
		var id AigID
		if j == b.n || i < a.n && a.leaves[i] < b.leaves[j] {
			id = a.leaves[i]
			i++
		} else if i == a.n || b.leaves[j] < a.leaves[i] {
			id = b.leaves[j]
			j++
		} else {
			id = a.leaves[i]
			i++
			j++
		}
		if c.n == 4 {
			return darCut4{}, false
		}
		c.leaves[c.n] = id
		c.n++
	}
	return c, true
}
func darTruthIn(c, a *darCut4) uint16 {
	var truth uint16
	for assignment := 0; assignment < 16; assignment++ {
		index := 0
		for j := 0; j < a.n; j++ {
			for i := 0; i < c.n; i++ {
				if a.leaves[j] == c.leaves[i] {
					index |= ((assignment >> uint(i)) & 1) << uint(j)
					break
				}
			}
		}
		truth |= ((a.truth >> uint(index)) & 1) << uint(assignment)
	}
	return truth
}
func darMinimize(c *darCut4) bool {
	old := *c
	var kept [4]int
	count := 0
	for i := 0; i < old.n; i++ {
		dependent := false
		for a := 0; a < 16; a++ {
			if a&(1<<uint(i)) == 0 && old.truth>>uint(a)&1 != old.truth>>uint(a|(1<<uint(i)))&1 {
				dependent = true
				break
			}
		}
		if dependent {
			kept[count] = i
			count++
		}
	}
	if count == old.n {
		return false
	}
	c.n = count
	c.sign = 0
	c.truth = 0
	for j := 0; j < count; j++ {
		c.leaves[j] = old.leaves[kept[j]]
		c.sign |= uint32(1) << uint(c.leaves[j]&31)
	}
	for a := 0; a < 16; a++ {
		index := 0
		for j := 0; j < count; j++ {
			index |= ((a >> uint(j)) & 1) << uint(kept[j])
		}
		c.truth |= ((old.truth >> uint(index)) & 1) << uint(a)
	}
	return true
}

// ComputeDARCuts4 follows native bounded-slot eviction, ordered pair iteration,
// signature filtering, dominance and support minimization. Cleanup occurs on
// an owned clone; source IDs remain stable. Buffer and explicit XOR objects
// are unsupported. Limits/cancellation return no partial result.
func ComputeDARCuts4(g *AigManager, o DARCutOptions) (*DARCuts4, error) {
	if g == nil || o.CutsPerNode < 0 || o.MaxObjects < 0 || o.MaxCutSlots < 0 || o.MaxPairs < 0 {
		return nil, ErrDimension
	}
	if o.CutsPerNode == 0 {
		o.CutsPerNode = 10
	}
	if o.CutsPerNode < 2 || o.CutsPerNode > 255 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxCutSlots == 0 {
		o.MaxCutSlots = 8 << 20
	}
	if o.MaxPairs == 0 {
		o.MaxPairs = 16 << 20
	}
	if len(g.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	check := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		return nil
	}
	if e := check(); e != nil {
		return nil, e
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	graph := g.Clone()
	graph.Cleanup()
	sets := make([][]darCut4, len(graph.objects))
	allocated, pairs := 0, 0
	value := func(c *darCut4) uint16 {
		if c.n < 2 {
			return 1001
		}
		var refs uint32
		ones := 0
		for i := 0; i < c.n; i++ {
			r := graph.objects[c.leaves[i]].RefCount
			refs += r
			if r == 1 {
				ones++
			}
		}
		if refs > 1000 {
			refs = 1000
		}
		if ones > 3 {
			refs = uint32(5 - ones)
		}
		return uint16(refs)
	}
	for id, obj := range graph.objects {
		if e := check(); e != nil {
			return nil, e
		}
		if obj.Type == AigNone || obj.Type == AigCO {
			continue
		}
		if obj.Type != AigConstant1 && obj.Type != AigCI && obj.Type != AigAND {
			return nil, ErrDimension
		}
		if o.CutsPerNode > o.MaxCutSlots-allocated {
			return nil, ErrLimit
		}
		allocated += o.CutsPerNode
		set := make([]darCut4, o.CutsPerNode)
		sets[id] = set
		unit := darCut4{n: 1, truth: 0xaaaa, sign: uint32(1) << uint(id&31), used: true, value: 1001}
		unit.leaves[0] = AigID(id)
		if obj.Type == AigConstant1 {
			unit.n = 0
			unit.truth = 65535
			unit.sign = 0
		}
		set[0] = unit
		if obj.Type != AigAND {
			continue
		}
		early := false
		for _, a := range sets[obj.Fanin0.ID()] {
			if !a.used {
				continue
			}
			for _, b := range sets[obj.Fanin1.ID()] {
				if !b.used {
					continue
				}
				if e := check(); e != nil {
					return nil, e
				}
				if pairs == o.MaxPairs {
					return nil, ErrLimit
				}
				pairs++
				if bits.OnesCount32(a.sign|b.sign) > 4 {
					continue
				}
				free := -1
				for i := range set {
					if !set[i].used {
						free = i
						break
					}
				}
				if free < 0 {
					for _, minimum := range []int{3, 2, 0} {
						for i := range set {
							if set[i].n >= minimum && (free < 0 || set[i].value < set[free].value) {
								free = i
							}
						}
						if free >= 0 {
							break
						}
					}
					set[free].used = false
				}
				c, ok := darMerge(a, b)
				if !ok {
					continue
				}
				set[free] = c
				if darFilter(set, free) {
					continue
				}
				c = set[free]
				av, bv := darTruthIn(&c, &a), darTruthIn(&c, &b)
				if obj.Fanin0.Complemented() {
					av = ^av
				}
				if obj.Fanin1.Complemented() {
					bv = ^bv
				}
				c.truth = av & bv
				set[free] = c
				if !o.SkipTruthMinimization && darMinimize(&set[free]) {
					if darFilter(set, free) {
						return nil, ErrInvariant
					}
				}
				set[free].value = value(&set[free])
				if set[free].value == 0 {
					set[free].used = false
				} else if set[free].n < 2 {
					early = true
					break
				}
			}
			if early {
				break
			}
		}
	}
	r := &DARCuts4{Graph: graph}
	for id, set := range sets {
		if graph.objects[id].Type != AigAND {
			continue
		}
		for slot, c := range set {
			if slot == 0 || !c.used {
				continue
			}
			leaves := append([]AigID(nil), c.leaves[:c.n]...)
			r.Candidates = append(r.Candidates, AigCNFCut4{Root: AigID(id), Leaves: leaves, Truth: c.truth})
		}
	}
	if e := check(); e != nil {
		return nil, e
	}
	return r, nil
}

type AigCNFDeriveOptions struct {
	Cuts DARCutOptions
	CNF  MappedCNFOptions
}

// DeriveAigCNF4 automatically enumerates DAR cuts, selects area-flow cuts,
// scans the output cone and emits optimized clauses. Defaults match the native
// ten-slot, support-minimized path. Cleanup uses an owned graph preserving IDs;
// unneeded objects have no assigned CNF variable. Explicit buffers and XOR
// objects remain unsupported. No partial result is returned on failure.
func DeriveAigCNF4(g *AigManager, exposedOutputs int, o AigCNFDeriveOptions) (*AigMappedCNF, error) {
	cutStop, cnfStop := o.Cuts.Stop, o.CNF.Stop
	stop := func() bool { return cutStop != nil && cutStop() || cnfStop != nil && cnfStop() }
	o.Cuts.Stop = stop
	o.CNF.Stop = stop
	r, e := ComputeDARCuts4(g, o.Cuts)
	if e != nil {
		return nil, e
	}
	return WriteAigAreaFlowCNF4(r.Graph, r.Candidates, exposedOutputs, o.CNF)
}
