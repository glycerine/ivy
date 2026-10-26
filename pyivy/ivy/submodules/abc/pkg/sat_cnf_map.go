package abc

// Area-flow cut selection from sat/cnf/cnfMap.c:Cnf_DeriveMapping.

import (
	"math"
	"slices"
)

type CNFAreaFlowOptions struct {
	MaxObjects, MaxCuts int // zero: one million each
	Stop                func() bool
}
type CNFAreaFlowMapping struct {
	Cuts   []AigCNFCut4 // one selected cut per AND, in object order; not yet output-cone filtered
	Flows  []uint32     // native unsigned signatures, indexed by object ID
	Values []uint16     // native eleven-bit reference sums for selected cuts
}

// CNFSuperAreaFlow implements the compiled but currently uncalled
// Cnf_CutSuperAreaFlow helper. Every supergate entry contributes to the base
// cost; AND entries additionally contribute their selected flow divided by
// their structural reference count. Literal phase is ignored. Checked overflow
// replaces C int wraparound.
func CNFSuperAreaFlow(g *AigManager, super []AigLit, flows []uint32) (uint32, error) {
	if g == nil {
		return 0, ErrDimension
	}
	if err := g.Check(); err != nil {
		return 0, err
	}
	if len(flows) != len(g.objects) {
		return 0, ErrDimension
	}
	area := uint64(100) * uint64(len(super)+1)
	if area > math.MaxUint32 {
		return 0, ErrLimit
	}
	for _, literal := range super {
		if !g.validSignal(literal) {
			return 0, ErrInvalidLiteral
		}
		object := g.objects[literal.ID()]
		if object.Type != AigAND {
			continue
		}
		if object.RefCount == 0 {
			return 0, ErrInvariant
		}
		area += uint64(flows[literal.ID()] / object.RefCount)
		if area > math.MaxUint32 {
			return 0, ErrLimit
		}
	}
	return uint32(area), nil
}

// SelectCNFAreaFlow4 selects among supplied nontrivial cuts using native SOP
// cost, integer division, unsigned flow arithmetic and eleven-bit reference
// sums. Equal flow favors larger reference sum; exact ties retain the first
// candidate. Every AND requires candidates. Cut truth is trusted. Graph and
// input cuts are unchanged; outputs own their storage. This is the initial
// mapping pass, before cut transfer, output-cone scanning and area recovery.
func SelectCNFAreaFlow4(g *AigManager, candidates []AigCNFCut4, o CNFAreaFlowOptions) (*CNFAreaFlowMapping, error) {
	if g == nil || o.MaxObjects < 0 || o.MaxCuts < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxCuts == 0 {
		o.MaxCuts = 1 << 20
	}
	if len(g.objects) > o.MaxObjects || len(candidates) > o.MaxCuts {
		return nil, ErrLimit
	}
	stop := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		return nil
	}
	if e := stop(); e != nil {
		return nil, e
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	groups := make([][]int, len(g.objects))
	for i, c := range candidates {
		if e := stop(); e != nil {
			return nil, e
		}
		if uint64(c.Root) >= uint64(len(g.objects)) || g.objects[c.Root].Type != AigAND || len(c.Leaves) > 4 {
			return nil, ErrDimension
		}
		for _, leaf := range c.Leaves {
			if leaf >= c.Root {
				return nil, ErrDimension
			}
			typ := g.objects[leaf].Type
			if typ != AigCI && typ != AigConstant1 && typ != AigAND {
				return nil, ErrDimension
			}
		}
		// Small truth tables use the native four-variable stretched representation.
		n := len(c.Leaves)
		mask := uint16((uint32(1) << uint(1<<uint(n))) - 1)
		base := c.Truth & mask
		stretched := base
		for shift := 1 << uint(n); shift < 16; shift *= 2 {
			stretched |= stretched << uint(shift)
		}
		if stretched != c.Truth {
			return nil, ErrDimension
		}
		groups[c.Root] = append(groups[c.Root], i)
	}
	r := &CNFAreaFlowMapping{Flows: make([]uint32, len(g.objects))}
	for id, obj := range g.objects {
		if e := stop(); e != nil {
			return nil, e
		}
		if obj.Type == AigBuffer || obj.Type == AigXOR {
			return nil, ErrDimension
		}
		if obj.Type != AigAND {
			continue
		}
		if len(groups[id]) == 0 {
			return nil, ErrDimension
		}
		best := -1
		var bestFlow uint32
		var bestValue uint16
		for _, index := range groups[id] {
			if e := stop(); e != nil {
				return nil, e
			}
			c := candidates[index]
			pos, e := CNFSOP4Cover(c.Truth)
			if e != nil {
				return nil, e
			}
			neg, e := CNFSOP4Cover(^c.Truth)
			if e != nil {
				return nil, e
			}
			flow := uint32(10 * (len(pos) + len(neg)))
			var value uint16
			for _, leaf := range c.Leaves {
				node := g.objects[leaf]
				value = uint16((uint32(value) + node.RefCount) & 2047)
				if node.Type == AigAND {
					if node.RefCount == 0 {
						return nil, ErrInvariant
					}
					flow += uint32(int32(r.Flows[leaf]) / int32(node.RefCount))
				}
			}
			if best < 0 || flow < bestFlow || flow == bestFlow && value > bestValue {
				best = index
				bestFlow = flow
				bestValue = value
			}
		}
		c := candidates[best]
		c.Leaves = slices.Clone(c.Leaves)
		r.Cuts = append(r.Cuts, c)
		r.Values = append(r.Values, bestValue)
		r.Flows[id] = bestFlow
	}
	if e := stop(); e != nil {
		return nil, e
	}
	return r, nil
}
