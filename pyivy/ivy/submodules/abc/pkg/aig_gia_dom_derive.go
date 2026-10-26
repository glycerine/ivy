package abc

// Cone derivation from src/aig/gia/giaMffc.c:Gia_ManDomDerive.

// DeriveDominatedCone constructs one output for each assignment to the first
// cofactorVariables support objects (0..5). Remaining support objects become
// free CIs. All support CIs are retained, including fixed/unused ones, as in C.
// Support must uniquely cut every path to a CI. Constants remain constants.
// The source is immutable; maxObjects bounds source mapping and output graph
// allocation, with zero selecting one million objects.
func (m *GiaManager) DeriveDominatedCone(root GiaID, support []GiaID, cofactorVariables, maxObjects int) (*GiaManager, error) {
	if cofactorVariables < 0 || cofactorVariables > 5 || cofactorVariables > len(support) || maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects || len(support) > maxObjects {
		return nil, ErrLimit
	}
	if uint64(root) >= uint64(len(m.objects)) || !m.objects[root].IsAnd() {
		return nil, ErrInvalidLiteral
	}
	used := make([]bool, len(m.objects))
	for _, id := range support {
		if id == 0 || uint64(id) >= uint64(len(m.objects)) || m.objects[id].IsCO() || used[id] {
			return nil, ErrInvalidLiteral
		}
		used[id] = true
	}
	nodes, err := m.CollectConeAnds([]GiaID{root}, support, maxObjects)
	if err != nil {
		return nil, err
	}
	g, err := NewGiaManagerWithLimit(maxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	inputs := make([]GiaLit, len(support))
	for i := range inputs {
		inputs[i], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, len(m.objects))
	for assignment := 0; assignment < 1<<cofactorVariables; assignment++ {
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		for i, id := range support {
			mapping[id] = inputs[i]
			if i < cofactorVariables {
				mapping[id] = GiaLit((assignment >> i) & 1)
			}
		}
		translate := func(l GiaLit) (GiaLit, error) {
			if !mapping[l.ID()].Valid() {
				return InvalidGiaLit, ErrDimension
			}
			return mapping[l.ID()].NotIf(l.Complemented()), nil
		}
		for _, id := range nodes {
			obj := m.objects[id]
			first, err := obj.Fanin0(id)
			if err != nil {
				return nil, err
			}
			a, err := translate(first)
			if err != nil {
				return nil, err
			}
			if obj.IsBuffer() {
				mapping[id] = a
				continue
			}
			second, err := obj.Fanin1(id)
			if err != nil {
				return nil, err
			}
			b, err := translate(second)
			if err != nil {
				return nil, err
			}
			var value GiaLit
			if control, yes := m.MuxControl(id); yes {
				c, err := translate(control)
				if err != nil {
					return nil, err
				}
				value, err = g.HashMux(c, b, a)
				if err != nil {
					return nil, err
				}
			} else if obj.IsXor() {
				value, err = g.HashXor(a, b)
			} else {
				value, err = g.HashAnd(a, b)
			}
			if err != nil {
				return nil, err
			}
			mapping[id] = value
		}
		if !mapping[root].Valid() {
			return nil, ErrDimension
		}
		if _, err = g.AppendCO(mapping[root]); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	cleaned, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return cleaned, nil
}
