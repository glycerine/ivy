package abc

// Conjunction decomposition from src/aig/gia/giaDup.c:
// Gia_ManDupWithConstrCollectAnd_rec and Gia_ManDupWithConstr.

// ExtractConjunctiveConstraint splits a single combinational, uncomplemented
// ordinary-AND output into a main output and one positive constraint output.
// Their conjunction equals the original output. The highest-level boundary
// literal becomes main (first DFS occurrence breaks ties); other boundary
// literals form the assumption. C emits the complement of that assumption;
// Go deliberately uses positive constraint polarity for its verification APIs.
// Existing constraints and nondecomposable roots are rejected. Extended nodes
// are opaque boundaries; source scratch is untouched. All CIs are retained.
func (m *GiaManager) ExtractConjunctiveConstraint(options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if m.registers != 0 || len(m.cos) != 1 || m.constraints != 0 {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	root, err := m.objects[m.cos[0]].Fanin0(m.cos[0])
	if err != nil {
		return nil, err
	}
	ordinary := func(l GiaLit) bool {
		obj := m.objects[l.ID()]
		_, mux := m.MuxControl(l.ID())
		return !l.Complemented() && obj.IsAnd() && !obj.IsBuffer() && !obj.IsXor() && !mux
	}
	if !ordinary(root) {
		return nil, ErrUnsupported
	}
	stack := []GiaLit{root}
	expanded := make([]bool, len(m.objects))
	seen := make([]bool, 2*len(m.objects))
	leaves := make([]GiaLit, 0)
	for len(stack) > 0 {
		if stopped() {
			return nil, ErrCancelled
		}
		l := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !ordinary(l) {
			if !seen[l] {
				seen[l] = true
				leaves = append(leaves, l)
			}
			continue
		}
		if expanded[l.ID()] {
			continue
		}
		expanded[l.ID()] = true
		a, e := m.objects[l.ID()].Fanin0(l.ID())
		if e != nil {
			return nil, e
		}
		b, e := m.objects[l.ID()].Fanin1(l.ID())
		if e != nil {
			return nil, e
		}
		stack = append(stack, b, a)
	}
	if len(leaves) < 2 {
		return nil, ErrUnsupported
	}
	levels, _, err := m.Levels(options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	best := 0
	for i := 1; i < len(leaves); i++ {
		if levels[leaves[i].ID()] > levels[leaves[best].ID()] {
			best = i
		}
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for i, obj := range m.objects {
		if stopped() {
			return nil, ErrCancelled
		}
		if obj.IsCI() {
			mapping[i], err = g.AppendCI()
		} else if obj.IsAnd() {
			mapping[i], err = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		}
		if err != nil {
			return nil, err
		}
	}
	assumption := GiaConst1
	for i, l := range leaves {
		if stopped() {
			return nil, ErrCancelled
		}
		if i == best {
			continue
		}
		copied, e := mappedGiaSignal(mapping, l)
		if e != nil {
			return nil, e
		}
		assumption, err = g.HashAnd(assumption, copied)
		if err != nil {
			return nil, err
		}
	}
	main, err := mappedGiaSignal(mapping, leaves[best])
	if err != nil {
		return nil, err
	}
	if _, err = g.AppendCO(main); err != nil {
		return nil, err
	}
	if _, err = g.AppendCO(assumption); err != nil {
		return nil, err
	}
	g.HashStop()
	g.constraints = 1
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
