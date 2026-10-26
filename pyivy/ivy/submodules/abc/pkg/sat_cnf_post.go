package abc

// Mapping recovery from sat/cnf/cnfPost.c:Cnf_ManPostprocess.

import (
	"errors"
	"slices"
)

type CNFPostprocessOptions struct {
	Composition                      CNFCompositionOptions
	MaxObjects, MaxCuts, MaxAttempts int // zero: one million each
	Stop                             func() bool
}
type CNFPostprocessedMapping struct {
	Cuts          []AigCNFCut // surviving reachable cuts in preorder
	References    []uint32    // owned mapping references, including terminals
	Area, Changes int
}

// PostprocessCNFMapping4 performs one native-order recovery pass on supplied
// four-input cuts, using corrected Boolean composition. Visit roots in object
// order; try fanins by stable ascending cost; absorb at most one singly used
// fanin per root when total clause cost does not increase. Composition limit
// failures skip that proposal, while global limits/cancellation abort atomically.
// Source graph/cuts remain unchanged. Supergate marks are not supported.
func PostprocessCNFMapping4(g *AigManager, cuts []AigCNFCut4, o CNFPostprocessOptions) (*CNFPostprocessedMapping, error) {
	if o.MaxObjects < 0 || o.MaxCuts < 0 || o.MaxAttempts < 0 {
		return nil, ErrDimension
	}
	if o.MaxAttempts == 0 {
		o.MaxAttempts = 1 << 20
	}
	a, b, c := o.Stop, o.Composition.Stop, o.Composition.ISOP.Stop
	stop := func() bool { return a != nil && a() || b != nil && b() || c != nil && c() }
	o.Composition.Stop = stop
	o.Composition.ISOP.Stop = nil
	if o.Composition.MaxInputs < 0 || o.Composition.MaxInputs > 16 || o.Composition.MaxVisits < 0 || o.Composition.ISOP.MaxElements < 0 || o.Composition.ISOP.MaxVisits < 0 {
		return nil, ErrDimension
	}
	scan, e := ScanCNFMapping4(g, cuts, CNFMappingScanOptions{Preorder: true, MaxObjects: o.MaxObjects, MaxCuts: o.MaxCuts, Stop: stop})
	if e != nil {
		return nil, e
	}
	byRoot := make([]*AigCNFCut, len(g.objects))
	costs := make([]int, len(g.objects))
	for _, cut := range cuts {
		if stop() {
			return nil, ErrCancelled
		}
		for i := 1; i < len(cut.Leaves); i++ {
			if cut.Leaves[i] <= cut.Leaves[i-1] {
				return nil, ErrDimension
			}
		}
		truth, e := TruthTableFromWords(len(cut.Leaves), []uint64{uint64(cut.Truth) * 0x1000100010001}, 0)
		if e != nil {
			return nil, e
		}
		v := &AigCNFCut{Root: cut.Root, Leaves: slices.Clone(cut.Leaves), Truth: truth}
		byRoot[cut.Root] = v
		pos, e := CNFSOP4Cover(cut.Truth)
		if e != nil {
			return nil, e
		}
		neg, e := CNFSOP4Cover(^cut.Truth)
		if e != nil {
			return nil, e
		}
		costs[cut.Root] = len(pos) + len(neg)
	}
	refs := scan.References
	attempts, changes := 0, 0
	for id, obj := range g.objects {
		if stop() {
			return nil, ErrCancelled
		}
		if obj.Type != AigAND || refs[id] == 0 {
			continue
		}
		top := byRoot[id]
		if top == nil {
			return nil, ErrInvariant
		}
		order := slices.Clone(top.Leaves)
		slices.SortStableFunc(order, func(a, b AigID) int { return costs[a] - costs[b] })
		for _, fanID := range order {
			if stop() {
				return nil, ErrCancelled
			}
			if g.objects[fanID].Type != AigAND {
				continue
			}
			if refs[fanID] == 0 {
				return nil, ErrInvariant
			}
			if refs[fanID] != 1 {
				continue
			}
			fan := byRoot[fanID]
			if fan == nil {
				return nil, ErrInvariant
			}
			if attempts == o.MaxAttempts {
				return nil, ErrLimit
			}
			attempts++
			composed, e := ComposeCNFCuts(*top, *fan, o.Composition)
			if e != nil {
				if errors.Is(e, ErrLimit) {
					continue
				}
				return nil, e
			}
			if composed.Cost == 127 || composed.Cost > costs[id]+costs[fanID] {
				continue
			}
			for _, old := range []*AigCNFCut{top, fan} {
				for _, leaf := range old.Leaves {
					if refs[leaf] == 0 {
						return nil, ErrInvariant
					}
					refs[leaf]--
				}
			}
			for _, leaf := range composed.Cut.Leaves {
				refs[leaf]++
			}
			if refs[fanID] != 0 {
				return nil, ErrInvariant
			}
			byRoot[id] = &composed.Cut
			byRoot[fanID] = nil
			costs[id] = composed.Cost
			costs[fanID] = 0
			changes++
			break
		}
	}
	r := &CNFPostprocessedMapping{References: refs, Changes: changes}
	seen := make([]bool, len(g.objects))
	var stack []AigID
	for _, co := range g.cos {
		stack = append(stack, g.objects[co].Fanin0.ID())
		for len(stack) > 0 {
			if stop() {
				return nil, ErrCancelled
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			if g.objects[id].Type != AigAND {
				continue
			}
			cut := byRoot[id]
			if cut == nil {
				return nil, ErrInvariant
			}
			r.Cuts = append(r.Cuts, *cut)
			r.Area += costs[id]
			for i := len(cut.Leaves) - 1; i >= 0; i-- {
				stack = append(stack, cut.Leaves[i])
			}
		}
	}
	if stop() {
		return nil, ErrCancelled
	}
	return r, nil
}

// DeriveAigPostprocessedCNF4 explicitly enables one recovery pass after native
// DAR/area-flow mapping. Default DeriveAigCNF4 retains its native initial path.
func DeriveAigPostprocessedCNF4(g *AigManager, exposedOutputs int, o AigCNFDeriveOptions, p CNFPostprocessOptions) (*AigMappedCNF, error) {
	a, b, c, d, e := o.Cuts.Stop, o.CNF.Stop, p.Stop, p.Composition.Stop, p.Composition.ISOP.Stop
	stop := func() bool {
		return a != nil && a() || b != nil && b() || c != nil && c() || d != nil && d() || e != nil && e()
	}
	o.Cuts.Stop = stop
	o.CNF.Stop = stop
	p.Stop = stop
	p.Composition.Stop = nil
	p.Composition.ISOP.Stop = nil
	generated, err := ComputeDARCuts4(g, o.Cuts)
	if err != nil {
		return nil, err
	}
	selected, err := SelectCNFAreaFlow4(generated.Graph, generated.Candidates, CNFAreaFlowOptions{MaxObjects: o.CNF.MaxObjects, Stop: stop})
	if err != nil {
		return nil, err
	}
	recovered, err := PostprocessCNFMapping4(generated.Graph, selected.Cuts, p)
	if err != nil {
		return nil, err
	}
	return WriteAigMappedCNF(generated.Graph, recovered.Cuts, exposedOutputs, o.CNF)
}
