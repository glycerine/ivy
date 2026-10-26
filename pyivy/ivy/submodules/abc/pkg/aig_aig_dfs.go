package abc

// Core AIG traversal, support and cone analysis from aig/aig/aigDfs.c.

type AigTraversalOptions struct {
	MaxObjects, MaxVisits int // zero: one million objects, eight million events
	Stop                  func() bool
}

type AigDFSOptions struct {
	NodesOnly, Preorder bool
	AigTraversalOptions
}

func checkAigTraversal(g *AigManager, o AigTraversalOptions) (AigTraversalOptions, error) {
	if g == nil || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return o, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 8 << 20
	}
	if len(g.objects) > o.MaxObjects {
		return o, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return o, ErrCancelled
	}
	if e := g.Check(); e != nil {
		return o, e
	}
	return o, nil
}

func aigTraversalVisit(o AigTraversalOptions) func() error {
	visits := 0
	return func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if visits == o.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
}

func aigDFSOrder(g *AigManager, roots []AigID, seen []bool, preorder bool, visit func() error, result []AigID) ([]AigID, error) {
	type event struct {
		id   AigID
		exit bool
	}
	for _, root := range roots {
		stack := []event{{id: root}}
		for len(stack) > 0 {
			if e := visit(); e != nil {
				return nil, e
			}
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if v.exit {
				result = append(result, v.id)
				continue
			}
			if seen[v.id] {
				continue
			}
			seen[v.id] = true
			obj := g.objects[v.id]
			if preorder {
				result = append(result, v.id)
				if next := g.choiceNext(v.id); next != InvalidAigID {
					var e error
					result, e = aigDFSOrder(g, []AigID{next}, seen, false, visit, result)
					if e != nil {
						return nil, e
					}
				}
			} else {
				stack = append(stack, event{id: v.id, exit: true})
			}
			switch obj.Type {
			case AigConstant1, AigCI:
			case AigCO, AigBuffer:
				stack = append(stack, event{id: obj.Fanin0.ID()})
			case AigAND:
				stack = append(stack, event{id: obj.Fanin1.ID()}, event{id: obj.Fanin0.ID()})
			default:
				return nil, ErrDimension
			}
			if !preorder {
				if next := g.choiceNext(v.id); next != InvalidAigID {
					stack = append(stack, event{id: next})
				}
			}
		}
	}
	return result, nil
}

// DFS returns native CO-rooted preorder or postorder. NodesOnly excludes the
// constant, CIs and COs. As in native Aig_ManDfs, dangling live objects reject.
func (g *AigManager) DFS(o AigDFSOptions) ([]AigID, error) {
	checked, e := checkAigTraversal(g, o.AigTraversalOptions)
	if e != nil {
		return nil, e
	}
	seen := make([]bool, len(g.objects))
	seen[0] = true
	var roots, result []AigID
	if o.NodesOnly {
		for _, id := range g.cis {
			seen[id] = true
		}
		for _, id := range g.cos {
			roots = append(roots, g.objects[id].Fanin0.ID())
		}
	} else {
		result = append(result, 0)
		roots = append(roots, g.cos...)
	}
	result, e = aigDFSOrder(g, roots, seen, o.Preorder, aigTraversalVisit(checked), result)
	if e != nil {
		return nil, e
	}
	expected := 0
	for _, obj := range g.objects {
		if obj.Type == AigNone {
			continue
		}
		if !o.NodesOnly || obj.Type == AigAND || obj.Type == AigBuffer {
			expected++
		}
	}
	if len(result) != expected {
		return nil, ErrInvariant
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

// DFSAll returns native postorder from every CO, followed by otherwise unused
// CIs. A dangling internal object is an invariant error.
func (g *AigManager) DFSAll(o AigTraversalOptions) ([]AigID, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	seen := make([]bool, len(g.objects))
	seen[0] = true
	result, e := aigDFSOrder(g, g.cos, seen, false, aigTraversalVisit(checked), []AigID{0})
	if e != nil {
		return nil, e
	}
	for _, id := range g.cis {
		if !seen[id] {
			result = append(result, id)
		}
	}
	live := 0
	for _, obj := range g.objects {
		if obj.Type != AigNone {
			live++
		}
	}
	if len(result) != live {
		return nil, ErrInvariant
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

// DFSArray returns constant-prefixed postorder from explicit object roots,
// including CO roots themselves.
func (g *AigManager) DFSArray(roots []AigID, o AigTraversalOptions) ([]AigID, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	if len(roots) > checked.MaxObjects {
		return nil, ErrLimit
	}
	for _, id := range roots {
		if uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
	}
	seen := make([]bool, len(g.objects))
	seen[0] = true
	result, e := aigDFSOrder(g, roots, seen, false, aigTraversalVisit(checked), []AigID{0})
	if e != nil {
		return nil, e
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

// DFSNodes returns native postorder from explicit roots. CO roots are replaced
// by their drivers; the constant is omitted while encountered CIs are retained.
func (g *AigManager) DFSNodes(roots []AigID, o AigTraversalOptions) ([]AigID, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	if len(roots) > checked.MaxObjects {
		return nil, ErrLimit
	}
	resolved := make([]AigID, len(roots))
	for i, id := range roots {
		if uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
		resolved[i] = id
		if g.objects[id].Type == AigCO {
			resolved[i] = g.objects[id].Fanin0.ID()
		}
	}
	seen := make([]bool, len(g.objects))
	seen[0] = true
	result, e := aigDFSOrder(g, resolved, seen, false, aigTraversalVisit(checked), nil)
	if e != nil {
		return nil, e
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

// ReverseDFS implements Aig_ManDfsReverse. It follows fanouts in native
// object-insertion order, omits COs and returns all AND/buffer objects in
// fanout-first postorder, including dangling logic. Fanout storage is owned by
// the call and the graph is not modified.
func (g *AigManager) ReverseDFS(o AigTraversalOptions) ([]AigID, error) {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return nil, err
	}
	visit := aigTraversalVisit(checked)
	fanouts := make([][]AigID, len(g.objects))
	for id, object := range g.objects {
		if err = visit(); err != nil {
			return nil, err
		}
		appendFanout := func(literal AigLit) error {
			if err := visit(); err != nil {
				return err
			}
			if !g.validSignal(literal) {
				return ErrInvariant
			}
			fanouts[literal.ID()] = append(fanouts[literal.ID()], AigID(id))
			return nil
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI:
		case AigCO, AigBuffer:
			if err = appendFanout(object.Fanin0); err != nil {
				return nil, err
			}
		case AigAND:
			if err = appendFanout(object.Fanin0); err != nil {
				return nil, err
			}
			if err = appendFanout(object.Fanin1); err != nil {
				return nil, err
			}
		default:
			return nil, ErrDimension
		}
	}
	seen := make([]bool, len(g.objects))
	for _, id := range g.cos {
		seen[id] = true
	}
	type event struct {
		id   AigID
		exit bool
	}
	result := make([]AigID, 0, len(g.hash))
	for id, object := range g.objects {
		if object.Type != AigAND && object.Type != AigBuffer {
			continue
		}
		stack := []event{{id: AigID(id)}}
		for len(stack) > 0 {
			if err = visit(); err != nil {
				return nil, err
			}
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if current.exit {
				result = append(result, current.id)
				continue
			}
			if seen[current.id] {
				continue
			}
			kind := g.objects[current.id].Type
			if kind != AigAND && kind != AigBuffer {
				return nil, ErrInvariant
			}
			seen[current.id] = true
			stack = append(stack, event{id: current.id, exit: true})
			list := fanouts[current.id]
			for i := len(list) - 1; i >= 0; i-- {
				stack = append(stack, event{id: list[i]})
			}
		}
	}
	return result, nil
}

// Support returns unique CIs in fanin-zero-first encounter order for explicit
// roots. CO roots are replaced by their drivers and constants are ignored.
func (g *AigManager) Support(roots []AigID, o AigTraversalOptions) ([]AigID, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	if len(roots) > checked.MaxObjects {
		return nil, ErrLimit
	}
	seen := make([]bool, len(g.objects))
	visit := aigTraversalVisit(checked)
	var support []AigID
	for _, source := range roots {
		if uint64(source) >= uint64(len(g.objects)) || g.objects[source].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
		if g.objects[source].Type == AigCO {
			source = g.objects[source].Fanin0.ID()
		}
		stack := []AigID{source}
		for len(stack) > 0 {
			if e := visit(); e != nil {
				return nil, e
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			obj := g.objects[id]
			switch obj.Type {
			case AigConstant1:
			case AigCI:
				support = append(support, id)
			case AigBuffer:
				stack = append(stack, obj.Fanin0.ID())
			case AigAND:
				stack = append(stack, obj.Fanin1.ID(), obj.Fanin0.ID())
			default:
				return nil, ErrDimension
			}
		}
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return support, nil
}

func (g *AigManager) SupportSize(root AigID, o AigTraversalOptions) (int, error) {
	support, e := g.Support([]AigID{root}, o)
	if e != nil {
		return 0, e
	}
	return len(support), nil
}

// CountANDsWithSupportAtMost is the non-printing form of
// Aig_SupportSizeTest. It counts every live AND whose unique CI support is no
// larger than maximum. Work and cancellation are shared across all roots.
func (g *AigManager) CountANDsWithSupportAtMost(maximum int, o AigTraversalOptions) (int, error) {
	if maximum < 0 {
		return 0, ErrDimension
	}
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return 0, err
	}
	visit := aigTraversalVisit(checked)
	epochs := make([]uint32, len(g.objects))
	epoch := uint32(0)
	count := 0
	for root, object := range g.objects {
		if object.Type != AigAND {
			continue
		}
		if epoch == ^uint32(0) {
			for i := range epochs {
				epochs[i] = 0
			}
			epoch = 1
		} else {
			epoch++
		}
		support := 0
		stack := []AigID{AigID(root)}
		for len(stack) > 0 {
			if err = visit(); err != nil {
				return 0, err
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if epochs[id] == epoch {
				continue
			}
			epochs[id] = epoch
			current := g.objects[id]
			switch current.Type {
			case AigConstant1:
			case AigCI:
				support++
			case AigBuffer:
				stack = append(stack, current.Fanin0.ID())
			case AigAND:
				stack = append(stack, current.Fanin1.ID(), current.Fanin0.ID())
			default:
				return 0, ErrDimension
			}
		}
		if support <= maximum {
			count++
		}
	}
	return count, nil
}

// SmallSupportANDCount implements Aig_SupportSizeTest's fixed threshold of 16
// without writing diagnostics.
func (g *AigManager) SmallSupportANDCount(o AigTraversalOptions) (int, error) {
	return g.CountANDsWithSupportAtMost(16, o)
}

// ConeANDMask is the owned-state equivalent of Aig_ConeMark_rec,
// Aig_ConeCleanAndMark_rec and Aig_ConeCountAndMark_rec. True entries are the
// unique AND objects reachable from root. Non-AND roots form an empty cone;
// unlike the C helpers, no manager marks or generic data fields need cleanup.
func (g *AigManager) ConeANDMask(root AigID, o AigTraversalOptions) ([]bool, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return nil, ErrInvalidLiteral
	}
	seen := make([]bool, len(g.objects))
	stack := []AigID{root}
	visit := aigTraversalVisit(checked)
	for len(stack) > 0 {
		if e := visit(); e != nil {
			return nil, e
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] || g.objects[id].Type != AigAND {
			continue
		}
		seen[id] = true
		stack = append(stack, g.objects[id].Fanin1.ID(), g.objects[id].Fanin0.ID())
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return seen, nil
}

// DAGSize counts unique ANDs in one regular-root cone without graph marks.
func (g *AigManager) DAGSize(root AigID, o AigTraversalOptions) (int, error) {
	mask, e := g.ConeANDMask(root, o)
	if e != nil {
		return 0, e
	}
	count := 0
	for _, marked := range mask {
		if marked {
			count++
		}
	}
	return count, nil
}

// CollectCut returns unique internal ANDs in native postorder, stopping at the
// supplied regular leaf IDs and excluding those leaves from the result.
func (g *AigManager) CollectCut(root AigID, leaves []AigID, o AigTraversalOptions) ([]AigID, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return nil, ErrInvalidLiteral
	}
	if len(leaves) > checked.MaxObjects {
		return nil, ErrLimit
	}
	seen := make([]bool, len(g.objects))
	for _, id := range leaves {
		if uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
		if seen[id] {
			return nil, ErrDimension
		}
		seen[id] = true
	}
	type event struct {
		id   AigID
		exit bool
	}
	stack := []event{{id: root}}
	visit := aigTraversalVisit(checked)
	var nodes []AigID
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
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return nodes, nil
}

// MaxOutputLevel returns the largest regular CO-driver level.
func (g *AigManager) MaxOutputLevel() (uint32, error) {
	if e := g.Check(); e != nil {
		return 0, e
	}
	var level uint32
	for _, id := range g.cos {
		level = max(level, g.objects[g.objects[id].Fanin0.ID()].Level)
	}
	return level, nil
}

// LevelBuckets returns every live object in object order grouped by its stored
// level. Objects above the maximum output-driver level are invariant failures.
func (g *AigManager) LevelBuckets(o AigTraversalOptions) ([][]AigID, error) {
	checked, e := checkAigTraversal(g, o)
	if e != nil {
		return nil, e
	}
	maximum, e := g.MaxOutputLevel()
	if e != nil {
		return nil, e
	}
	levels := make([][]AigID, int(maximum)+1)
	visit := aigTraversalVisit(checked)
	for id, obj := range g.objects {
		if obj.Type == AigNone {
			continue
		}
		if e := visit(); e != nil {
			return nil, e
		}
		if obj.Level > maximum {
			return nil, ErrInvariant
		}
		levels[obj.Level] = append(levels[obj.Level], AigID(id))
	}
	if checked.Stop != nil && checked.Stop() {
		return nil, ErrCancelled
	}
	return levels, nil
}

// VerifyTopologicalOrder performs the native topological check together with
// the stronger manager metadata/hash/reference validation used by Go.
func (g *AigManager) VerifyTopologicalOrder() error {
	if g == nil {
		return ErrDimension
	}
	return g.Check()
}
