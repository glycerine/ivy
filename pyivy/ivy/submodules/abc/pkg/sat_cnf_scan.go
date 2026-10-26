package abc

// Mapping DFS from sat/cnf/cnfUtil.c:Cnf_ManScanMapping and recursion.

import "slices"

type CNFMappingScanOptions struct {
	Preorder                       bool
	MaxObjects, MaxCuts, MaxVisits int // zero: 1M, 1M, 8M
	Supergates                     []AigID
	Stop                           func() bool
}
type CNFMappingScan struct {
	Roots      []AigID      // every mapped node in requested DFS order
	Cuts       []AigCNFCut4 // reachable cuts in requested DFS order
	References []uint32     // mapping-edge references, including terminal and repeated edges
	Area       int          // total positive/negative SOP clause cost of distinct reachable cuts
}

// ScanCNFMapping4 visits CO drivers in output order and cut leaves in supplied
// order. Preorder matches native optimized CNF emission; postorder matches its
// mapping-array traversal. Graph structural references remain unchanged. Cuts
// need only exist for reachable ANDs; supplied truth tables are trusted. Owned
// results replace C's destructive reference-counter updates. Supergates names
// roots carrying the source fMarkB bit; they appear in Roots but not Cuts.
func ScanCNFMapping4(g *AigManager, cuts []AigCNFCut4, o CNFMappingScanOptions) (*CNFMappingScan, error) {
	if g == nil || o.MaxObjects < 0 || o.MaxCuts < 0 || o.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxCuts == 0 {
		o.MaxCuts = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 8 << 20
	}
	if len(g.objects) > o.MaxObjects || len(cuts) > o.MaxCuts || len(o.Supergates) > o.MaxCuts {
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
	byRoot := make([]int, len(g.objects))
	for i := range byRoot {
		byRoot[i] = -1
	}
	costs := make([]int, len(cuts))
	for i, c := range cuts {
		if e := check(); e != nil {
			return nil, e
		}
		if uint64(c.Root) >= uint64(len(g.objects)) || g.objects[c.Root].Type != AigAND || byRoot[c.Root] >= 0 || len(c.Leaves) > 4 {
			return nil, ErrDimension
		}
		for _, leaf := range c.Leaves {
			if leaf >= c.Root {
				return nil, ErrDimension
			}
			typ := g.objects[leaf].Type
			if typ != AigAND && typ != AigCI && typ != AigConstant1 {
				return nil, ErrDimension
			}
		}
		mask := (1 << uint(len(c.Leaves))) - 1
		for a := 0; a < 16; a++ {
			if c.Truth>>uint(a)&1 != c.Truth>>uint(a&mask)&1 {
				return nil, ErrDimension
			}
		}
		p, e := CNFSOP4Cover(c.Truth)
		if e != nil {
			return nil, e
		}
		n, e := CNFSOP4Cover(^c.Truth)
		if e != nil {
			return nil, e
		}
		costs[i] = len(p) + len(n)
		byRoot[c.Root] = i
	}
	super := make([]bool, len(g.objects))
	for _, id := range o.Supergates {
		if uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type != AigAND || super[id] || byRoot[id] < 0 {
			return nil, ErrDimension
		}
		super[id] = true
	}
	r := &CNFMappingScan{References: make([]uint32, len(g.objects))}
	type event struct {
		id       AigID
		exit     bool
		cutIndex int
	}
	stack := make([]event, 0)
	visits := 0
	visitSuper := func() error {
		if e := check(); e != nil {
			return e
		}
		if visits == o.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
	appendNode := func(id AigID, cutIndex int) {
		r.Roots = append(r.Roots, id)
		if cutIndex >= 0 {
			c := cuts[cutIndex]
			c.Leaves = slices.Clone(c.Leaves)
			r.Cuts = append(r.Cuts, c)
		}
	}
	for _, co := range g.cos {
		stack = append(stack, event{id: g.objects[co].Fanin0.ID()})
		for len(stack) > 0 {
			if e := check(); e != nil {
				return nil, e
			}
			if visits == o.MaxVisits {
				return nil, ErrLimit
			}
			visits++
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if v.exit {
				appendNode(v.id, v.cutIndex)
				continue
			}
			old := r.References[v.id]
			if old == (1<<26)-1 {
				return nil, ErrLimit
			}
			r.References[v.id]++
			obj := g.objects[v.id]
			if old != 0 || obj.Type == AigCI || obj.Type == AigConstant1 {
				continue
			}
			if obj.Type != AigAND || byRoot[v.id] < 0 {
				return nil, ErrDimension
			}
			index := byRoot[v.id]
			if index < 0 {
				return nil, ErrDimension
			}
			if super[v.id] {
				leaves, e := collectAigSupergate(g, v.id, r.References, visitSuper)
				if e != nil {
					return nil, e
				}
				if r.Area > int(^uint32(0)>>1)-(len(leaves)+1) {
					return nil, ErrLimit
				}
				r.Area += len(leaves) + 1
				if o.Preorder {
					appendNode(v.id, -1)
				} else {
					stack = append(stack, event{id: v.id, exit: true, cutIndex: -1})
				}
				for i := len(leaves) - 1; i >= 0; i-- {
					stack = append(stack, event{id: leaves[i].ID()})
				}
				continue
			}
			c := cuts[index]
			if r.Area > int(^uint32(0)>>1)-costs[index] {
				return nil, ErrLimit
			}
			r.Area += costs[index]
			if o.Preorder {
				appendNode(v.id, index)
			} else {
				stack = append(stack, event{id: v.id, exit: true, cutIndex: index})
			}
			for i := len(c.Leaves) - 1; i >= 0; i-- {
				stack = append(stack, event{id: c.Leaves[i]})
			}
		}
	}
	if e := check(); e != nil {
		return nil, e
	}
	return r, nil
}

// WriteAigAreaFlowCNF4 connects supplied candidate selection, reachable mapping
// scanning and native-order CNF emission. Automatic DAR cut enumeration remains
// separate. All stages preserve the caller's graph and candidate storage.
func WriteAigAreaFlowCNF4(g *AigManager, candidates []AigCNFCut4, exposedOutputs int, o MappedCNFOptions) (*AigMappedCNF, error) {
	m, e := SelectCNFAreaFlow4(g, candidates, CNFAreaFlowOptions{MaxObjects: o.MaxObjects, Stop: o.Stop})
	if e != nil {
		return nil, e
	}
	scan, e := ScanCNFMapping4(g, m.Cuts, CNFMappingScanOptions{Preorder: true, MaxObjects: o.MaxObjects, Stop: o.Stop})
	if e != nil {
		return nil, e
	}
	return WriteAigMappedCNF4(g, scan.Cuts, exposedOutputs, o)
}
