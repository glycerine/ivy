package abc

// DAR-best-cut and structural-supergate scan from
// sat/cnf/cnfUtil.c:Aig_ManScanMapping and aig/aig/aigDfs.c.

import "slices"

type AigMappingScanOptions struct {
	MaxObjects, MaxCuts, MaxVisits int // zero: 1M, 1M, 8M
	Stop                           func() bool
}

type AigMappingScan struct {
	Roots      []AigID
	References []uint32
	Area       int
}

func collectAigSupergate(g *AigManager, root AigID, references []uint32, visit func() error) ([]AigLit, error) {
	seen := make(map[AigID]AigLit)
	stack := []AigLit{AigLit(root) << 1}
	var leaves []AigLit
	for len(stack) > 0 {
		if e := visit(); e != nil {
			return nil, e
		}
		l := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		id := l.ID()
		if previous, ok := seen[id]; ok {
			if previous != l {
				return []AigLit{}, nil
			}
			continue
		}
		obj := g.objects[id]
		if id != root && (l.Complemented() || obj.Type != AigAND || references[id] > 1) {
			seen[id] = l
			leaves = append(leaves, l)
			continue
		}
		if l.Complemented() || obj.Type != AigAND {
			return nil, ErrDimension
		}
		stack = append(stack, obj.Fanin1, obj.Fanin0)
	}
	if len(leaves) <= 1 {
		return nil, ErrInvariant
	}
	return leaves, nil
}

// ScanAigMapping4 follows output drivers through supplied DAR-best cuts, except
// roots selected as structural supergates. It preserves native preorder,
// dynamic mapping-edge references and clause-area cost without mutating graph
// references or marks. Cut truth is trusted as the selected root function.
func ScanAigMapping4(g *AigManager, best []AigCNFCut4, supergates []AigID, o AigMappingScanOptions) (*AigMappingScan, error) {
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
	if len(g.objects) > o.MaxObjects || len(best) > o.MaxCuts || len(supergates) > o.MaxCuts {
		return nil, ErrLimit
	}
	stopped := func() bool { return o.Stop != nil && o.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	byRoot := make([]int, len(g.objects))
	for i := range byRoot {
		byRoot[i] = -1
	}
	costs := make([]int, len(best))
	for i, cut := range best {
		if stopped() {
			return nil, ErrCancelled
		}
		if uint64(cut.Root) >= uint64(len(g.objects)) || g.objects[cut.Root].Type != AigAND || byRoot[cut.Root] >= 0 || len(cut.Leaves) > 4 {
			return nil, ErrDimension
		}
		for _, leaf := range cut.Leaves {
			if leaf >= cut.Root {
				return nil, ErrDimension
			}
			typ := g.objects[leaf].Type
			if typ != AigAND && typ != AigCI && typ != AigConstant1 {
				return nil, ErrDimension
			}
		}
		mask := (1 << uint(len(cut.Leaves))) - 1
		for assignment := 0; assignment < 16; assignment++ {
			if cut.Truth>>uint(assignment)&1 != cut.Truth>>uint(assignment&mask)&1 {
				return nil, ErrDimension
			}
		}
		positive, e := CNFSOP4Cover(cut.Truth)
		if e != nil {
			return nil, e
		}
		negative, e := CNFSOP4Cover(^cut.Truth)
		if e != nil {
			return nil, e
		}
		costs[i] = len(positive) + len(negative)
		byRoot[cut.Root] = i
	}
	super := make([]bool, len(g.objects))
	for _, id := range supergates {
		if uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type != AigAND || super[id] {
			return nil, ErrDimension
		}
		super[id] = true
	}
	r := &AigMappingScan{References: make([]uint32, len(g.objects))}
	visits := 0
	visit := func() error {
		if stopped() {
			return ErrCancelled
		}
		if visits == o.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
	addArea := func(n int) error {
		if n < 0 || r.Area > int(^uint32(0)>>1)-n {
			return ErrLimit
		}
		r.Area += n
		return nil
	}
	for _, co := range g.cos {
		stack := []AigID{g.objects[co].Fanin0.ID()}
		for len(stack) > 0 {
			if e := visit(); e != nil {
				return nil, e
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			old := r.References[id]
			if old == (1<<26)-1 {
				return nil, ErrLimit
			}
			r.References[id]++
			obj := g.objects[id]
			if old != 0 || obj.Type == AigCI || obj.Type == AigConstant1 {
				continue
			}
			if obj.Type != AigAND {
				return nil, ErrDimension
			}
			r.Roots = append(r.Roots, id)
			if super[id] {
				leaves, e := collectAigSupergate(g, id, r.References, visit)
				if e != nil {
					return nil, e
				}
				if e := addArea(len(leaves) + 1); e != nil {
					return nil, e
				}
				for i := len(leaves) - 1; i >= 0; i-- {
					stack = append(stack, leaves[i].ID())
				}
				continue
			}
			index := byRoot[id]
			if index < 0 {
				return nil, ErrDimension
			}
			if e := addArea(costs[index]); e != nil {
				return nil, e
			}
			for i := len(best[index].Leaves) - 1; i >= 0; i-- {
				stack = append(stack, best[index].Leaves[i])
			}
		}
	}
	if stopped() {
		return nil, ErrCancelled
	}
	r.Roots = slices.Clone(r.Roots)
	return r, nil
}
