package abc

// Fast marking from sat/cnf/cnfFast.c:Cnf_DeriveFastMark.

type FastCNFMarking struct {
	Marked                                    []bool // indexed by original object ID
	OutputDriverViolations, AndGateViolations int
}

// MarkFastCNF preserves native mux groups, sharing/complement boundaries and
// single-leaf support recovery. Graph state remains unchanged; diagnostics are
// returned rather than printed. MaxVisits bounds all passes and cone events.
func MarkFastCNF(g *AigManager, o FastCNFConeOptions) (*FastCNFMarking, error) {
	if e := checkFastCNFCone(g, 0, o); e != nil {
		return nil, e
	}
	visit := fastCNFConeVisit(o)
	marked := make([]bool, len(g.objects))
	group := make([]bool, len(g.objects))
	support := make([]int, len(g.objects))
	for _, obj := range g.objects {
		if obj.Type == AigBuffer || obj.Type == AigXOR {
			return nil, ErrDimension
		}
	}
	for _, id := range g.cis {
		if e := visit(); e != nil {
			return nil, e
		}
		marked[id] = true
	}
	for _, id := range g.cos {
		if e := visit(); e != nil {
			return nil, e
		}
		marked[g.objects[id].Fanin0.ID()] = true
	}
	for id, obj := range g.objects {
		if obj.Type != AigAND {
			continue
		}
		if e := visit(); e != nil {
			return nil, e
		}
		if _, ok := aigRecognizeMux(g, AigID(id)); !ok {
			continue
		}
		a, b := obj.Fanin0.ID(), obj.Fanin1.ID()
		if group[a] || group[b] || g.objects[a].RefCount > 1 || g.objects[b].RefCount > 1 {
			continue
		}
		group[id], group[a], group[b] = true, true, true
		marked[id] = true
		for _, child := range []AigID{a, b} {
			n := g.objects[child]
			marked[n.Fanin0.ID()] = true
			marked[n.Fanin1.ID()] = true
		}
	}
	for id, obj := range g.objects {
		if obj.Type != AigAND {
			continue
		}
		if e := visit(); e != nil {
			return nil, e
		}
		if obj.RefCount > 1 {
			marked[id] = true
		}
		for _, l := range []AigLit{obj.Fanin0, obj.Fanin1} {
			if l.Complemented() && !group[l.ID()] {
				marked[l.ID()] = true
			}
		}
	}
	for id, obj := range g.objects {
		if obj.Type != AigAND || !marked[id] {
			continue
		}
		if e := visit(); e != nil {
			return nil, e
		}
		if group[id] {
			mux, ok := aigRecognizeMux(g, AigID(id))
			if !ok {
				continue
			}
			a, b := mux.WhenFalse.ID(), mux.WhenTrue.ID()
			if !marked[a] || !marked[b] {
				return nil, ErrInvariant
			}
			refs := uint32(1)
			if a == b {
				refs = 2
			}
			for _, leaf := range []AigID{a, b} {
				if !group[leaf] && g.objects[leaf].Type != AigCI && g.objects[leaf].RefCount == refs && support[leaf] < 3 {
					marked[leaf] = false
					break
				}
			}
			continue
		}
		leaves, e := collectFastCNFLeaves(g, AigID(id), marked, true, visit)
		if e != nil {
			return nil, e
		}
		support[id] = len(leaves)
		if len(leaves) >= 6 {
			continue
		}
		for _, l := range leaves {
			leaf := l.ID()
			if !marked[leaf] {
				return nil, ErrInvariant
			}
			if group[leaf] || g.objects[leaf].Type == AigCI || g.objects[leaf].RefCount > 1 {
				continue
			}
			if support[leaf] <= 0 {
				return nil, ErrInvariant
			}
			if len(leaves)-1+support[leaf] > 6 {
				continue
			}
			marked[leaf] = false
			support[id] = 6
			break
		}
	}
	result := &FastCNFMarking{Marked: marked}
	for _, id := range g.cos {
		if e := visit(); e != nil {
			return nil, e
		}
		if !marked[g.objects[id].Fanin0.ID()] {
			result.OutputDriverViolations++
		}
	}
	stamps := make([]uint32, len(g.objects))
	var epoch uint32
	for id, obj := range g.objects {
		if obj.Type != AigAND || !marked[id] {
			continue
		}
		leaves, e := collectFastCNFLeaves(g, AigID(id), marked, false, visit)
		if e != nil {
			return nil, e
		}
		if len(leaves) <= 6 {
			continue
		}
		epoch++
		stack := []AigID{AigID(id)}
		for len(stack) > 0 {
			if e := visit(); e != nil {
				return nil, e
			}
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if node != AigID(id) && marked[node] || stamps[node] == epoch {
				continue
			}
			stamps[node] = epoch
			n := g.objects[node]
			if n.Type != AigAND {
				return nil, ErrInvariant
			}
			for _, l := range []AigLit{n.Fanin0, n.Fanin1} {
				if l.Complemented() && !marked[l.ID()] {
					result.AndGateViolations++
				}
				stack = append(stack, l.ID())
			}
		}
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}
