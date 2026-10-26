package abc

// Selected-output duplication from src/aig/gia/giaDup.c:
// Gia_ManDupSelectedOutputs and Gia_ManDupOutputVec.

// SelectOutputs copies selected combinational outputs in the requested order,
// retaining every CI. Reachable gates are copied in DFS postorder without
// hashing, preserving extended node kinds. Duplicates are allowed within C's
// selection-length bound. Constraint outputs must remain a suffix; invalid
// reorderings are rejected instead of losing their roles. Source is immutable.
func (m *GiaManager) SelectOutputs(outputs []int, options GiaCofactorOptions) (*GiaManager, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if m.registers != 0 || len(outputs) > m.OutputCount() {
		return nil, ErrDimension
	}
	if m.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	roots := make([]GiaID, len(outputs))
	constraints := 0
	inConstraints := false
	for i, output := range outputs {
		if output < 0 || output >= m.OutputCount() {
			return nil, ErrDimension
		}
		constraint := output >= m.OutputCount()-m.constraints
		if constraint {
			inConstraints = true
			constraints++
		} else if inConstraints {
			return nil, ErrDimension
		}
		roots[i] = m.cos[output]
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	nodes, err := m.CollectConeAnds(roots, nil, options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, m.ObjectCount())
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for _, id := range nodes {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = appendGiaCopiedGate(m, target, id, mapping)
		if err != nil {
			return nil, err
		}
	}
	for _, id := range roots {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		copied, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
		if _, err = target.AppendCO(copied); err != nil {
			return nil, err
		}
	}
	if err = target.SetConstraintCount(constraints); err != nil {
		return nil, err
	}
	if err = target.Check(); err != nil {
		return nil, err
	}
	return target, nil
}

// SelectOutputMask preserves source output order for entries marked true.
func (m *GiaManager) SelectOutputMask(present []bool, options GiaCofactorOptions) (*GiaManager, error) {
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(present) != m.OutputCount() {
		return nil, ErrDimension
	}
	selected := make([]int, 0)
	for i, keep := range present {
		if keep {
			selected = append(selected, i)
		}
	}
	return m.SelectOutputs(selected, options)
}
