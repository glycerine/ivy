package abc

// Ordered CI substitution from src/aig/gia/giaDup.c:Gia_ManDupDfsCiMap.

// RemapInputs keeps every CI but substitutes its uses according to replacements.
// InvalidGiaLit means identity. Other entries must be constants or phased source
// CIs at the same/earlier terminal index. Substitutions compose in CI order;
// a self-reference may invert that CI. Later CIs and gate references are errors.
// Nil outputs keeps original CO order/roles. Non-nil outputs selects phased
// source signals, retaining the source register count but assigning no constraint
// roles; its final RegisterCount entries explicitly become RIs. Such selection
// must contain enough outputs for that bank. Hashing lowers extended nodes.
// This is substitution, not an equivalence/state-encoding claim.
func (m *GiaManager) RemapInputs(replacements, outputs []GiaLit, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	if len(replacements) != len(m.cis) {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries || len(outputs) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	if outputs != nil && len(outputs) < m.registers {
		return nil, ErrDimension
	}
	for index, l := range replacements {
		if l == InvalidGiaLit {
			continue
		}
		if !m.validSignal(l) {
			return nil, ErrInvalidLiteral
		}
		if l.ID() == 0 {
			continue
		}
		obj := m.objects[l.ID()]
		if !obj.IsCI() {
			return nil, ErrDimension
		}
		ordinal, e := obj.TerminalIndex()
		if e != nil {
			return nil, e
		}
		if ordinal > index {
			return nil, ErrDimension
		}
	}
	for _, l := range outputs {
		if !m.validSignal(l) {
			return nil, ErrInvalidLiteral
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
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
	for index, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
		if replacements[index] != InvalidGiaLit {
			mapping[id], err = mappedGiaSignal(mapping, replacements[index])
			if err != nil {
				return nil, err
			}
		}
	}
	selected := outputs
	if outputs == nil {
		selected = make([]GiaLit, len(m.cos))
		for i, id := range m.cos {
			selected[i], err = m.objects[id].Fanin0(id)
			if err != nil {
				return nil, err
			}
		}
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for _, root := range selected {
		if stopped() {
			return nil, ErrCancelled
		}
		nodes, e := m.CollectTransitiveFanins([]GiaID{root.ID()}, options.MaxMappingEntries)
		if e != nil {
			return nil, e
		}
		for _, id := range nodes {
			if mapping[id].Valid() {
				continue
			}
			if stopped() {
				return nil, ErrCancelled
			}
			mapping[id], err = hashGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return nil, err
			}
		}
		copied, e := mappedGiaSignal(mapping, root)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(copied); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.registers = m.registers
	if outputs == nil {
		g.constraints = m.constraints
	}
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
