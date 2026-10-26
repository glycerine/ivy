package abc

// Cone traversal/truth from sat/cnf/cnfFast.c.

type FastCNFConeOptions struct {
	MaxObjects, MaxVisits int // zero: one million each; visits count stack events
	Stop                  func() bool
}

func checkFastCNFCone(g *AigManager, root AigID, o FastCNFConeOptions) error {
	if g == nil || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return ErrDimension
	}
	limit := o.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if len(g.objects) > limit {
		return ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	if e := g.Check(); e != nil {
		return e
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return ErrInvalidLiteral
	}
	return nil
}
func fastCNFConeVisit(o FastCNFConeOptions) func() error {
	count := 0
	limit := o.MaxVisits
	if limit == 0 {
		limit = 1 << 20
	}
	f := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if count == limit {
			return ErrLimit
		}
		count++
		return nil
	}
	return f
}

// CollectFastCNFLeaves follows fanin-zero-first DFS and first-encounter unique
// boundary order. The root must be an AND and is expanded even when marked.
// stopComplement preserves signed boundary literals; otherwise leaves are
// regular IDs encoded as literals. Unmarked terminal encounters are errors.
func CollectFastCNFLeaves(g *AigManager, root AigID, marked []AigID, stopComplement bool, o FastCNFConeOptions) ([]AigLit, error) {
	if e := checkFastCNFCone(g, root, o); e != nil {
		return nil, e
	}
	if g.objects[root].Type != AigAND {
		return nil, ErrDimension
	}
	marks := make([]bool, len(g.objects))
	for _, id := range marked {
		if uint64(id) >= uint64(len(marks)) || g.objects[id].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
		marks[id] = true
	}
	leaves, e := collectFastCNFLeaves(g, root, marks, stopComplement, fastCNFConeVisit(o))
	if e != nil {
		return nil, e
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return leaves, nil
}

func collectFastCNFLeaves(g *AigManager, root AigID, marks []bool, stopComplement bool, visit func() error) ([]AigLit, error) {
	seen := make(map[AigLit]bool)
	stack := []AigLit{AigLit(root) << 1}
	var leaves []AigLit
	for len(stack) > 0 {
		if e := visit(); e != nil {
			return nil, e
		}
		l := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		id := l.ID()
		if id != root && (marks[id] || stopComplement && l.Complemented()) {
			if !stopComplement {
				l = AigLit(id) << 1
			}
			if !seen[l] {
				seen[l] = true
				leaves = append(leaves, l)
			}
			continue
		}
		obj := g.objects[id]
		if obj.Type != AigAND {
			return nil, ErrDimension
		}
		a, b := obj.Fanin0, obj.Fanin1
		if !stopComplement {
			a = AigLit(a.ID()) << 1
			b = AigLit(b.ID()) << 1
		}
		stack = append(stack, b, a)
	}
	return leaves, nil
}

// CollectFastCNFVolume returns regular ANDs in native postorder, stopping at
// supplied regular leaf IDs. Every terminal dependency must be a supplied leaf.
func CollectFastCNFVolume(g *AigManager, root AigID, leaves []AigID, o FastCNFConeOptions) ([]AigID, error) {
	if e := checkFastCNFCone(g, root, o); e != nil {
		return nil, e
	}
	seen := make([]bool, len(g.objects))
	for _, id := range leaves {
		if uint64(id) >= uint64(len(seen)) || g.objects[id].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
		seen[id] = true
	}
	type event struct {
		id   AigID
		exit bool
	}
	stack := []event{{id: root}}
	var nodes []AigID
	visit := fastCNFConeVisit(o)
	for len(stack) > 0 {
		if e := visit(); e != nil {
			return nil, e
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if v.exit {
			nodes = append(nodes, v.id)
			continue
		}
		if seen[v.id] {
			continue
		}
		seen[v.id] = true
		obj := g.objects[v.id]
		if obj.Type != AigAND {
			return nil, ErrDimension
		}
		stack = append(stack, event{id: v.id, exit: true}, event{id: obj.Fanin1.ID()}, event{id: obj.Fanin0.ID()})
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return nodes, nil
}

// DeriveFastCNFCutTruth assigns native variable patterns to ordered leaves and
// evaluates supplied postorder nodes. At most six leaves and 256 total entries
// are accepted. Leaves/nodes must be unique and disjoint, dependencies assigned.
// A constant leaf is an abstract cut variable, as in native Cnf_CutDeriveTruth.
func DeriveFastCNFCutTruth(g *AigManager, leaves, nodes []AigID, o FastCNFConeOptions) (uint64, error) {
	if len(leaves) > 6 || len(nodes) == 0 || len(leaves)+len(nodes) > 256 {
		return 0, ErrDimension
	}
	if e := checkFastCNFCone(g, nodes[len(nodes)-1], o); e != nil {
		return 0, e
	}
	truth, e := deriveFastCNFTruth(g, leaves, nodes, fastCNFConeVisit(o))
	if e != nil {
		return 0, e
	}
	if o.Stop != nil && o.Stop() {
		return 0, ErrCancelled
	}
	return truth, nil
}
func deriveFastCNFTruth(g *AigManager, leaves, nodes []AigID, visit func() error) (uint64, error) {
	slots := make(map[AigID]int, len(leaves)+len(nodes))
	for _, id := range leaves {
		slots[id] = -1
	}
	for _, id := range nodes {
		slots[id] = -1
	}
	var truth [256]uint64
	for i, id := range leaves {
		if e := visit(); e != nil {
			return 0, e
		}
		if uint64(id) >= uint64(len(g.objects)) || slots[id] >= 0 {
			return 0, ErrDimension
		}
		typ := g.objects[id].Type
		if typ != AigCI && typ != AigConstant1 && typ != AigAND {
			return 0, ErrDimension
		}
		slots[id] = i
		truth[i] = truthVariablePattern(i)
	}
	for i, id := range nodes {
		if e := visit(); e != nil {
			return 0, e
		}
		if uint64(id) >= uint64(len(g.objects)) || slots[id] >= 0 || g.objects[id].Type != AigAND {
			return 0, ErrDimension
		}
		obj := g.objects[id]
		a, okA := slots[obj.Fanin0.ID()]
		b, okB := slots[obj.Fanin1.ID()]
		if !okA || !okB || a < 0 || b < 0 {
			return 0, ErrDimension
		}
		x, y := truth[a], truth[b]
		if obj.Fanin0.Complemented() {
			x = ^x
		}
		if obj.Fanin1.Complemented() {
			y = ^y
		}
		slot := len(leaves) + i
		truth[slot] = x & y
		slots[id] = slot
	}
	return truth[len(leaves)+len(nodes)-1], nil
}
