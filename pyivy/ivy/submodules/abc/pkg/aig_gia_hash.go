package abc

// Port of src/aig/gia/giaHash.c: structural hashing and Boolean construction.

// HashStart indexes existing nodes in ID order. Unlike the C assertion on
// preexisting duplicate raw nodes, Go indexes their first occurrence. Raw
// duplicates remain valid objects and are preserved by cleanup/conversion.
func (m *GiaManager) HashStart() error {
	if m.hash != nil {
		return ErrInvariant
	}
	if err := m.Check(); err != nil {
		return err
	}
	table := make(map[giaHashKey]GiaID)
	for i, o := range m.objects {
		if o.IsAnd() {
			key := m.objectKey(GiaID(i))
			if _, exists := table[key]; !exists {
				table[key] = GiaID(i)
			}
		}
	}
	m.hash = table
	return nil
}
func (m *GiaManager) HashStop() { m.hash = nil }
func giaAndTrivial(a, b GiaLit) (GiaLit, bool) {
	switch {
	case a == GiaConst0 || b == GiaConst0:
		return GiaConst0, true
	case a == GiaConst1:
		return b, true
	case b == GiaConst1:
		return a, true
	case a == b:
		return a, true
	case a == b.Not():
		return GiaConst0, true
	}
	return InvalidGiaLit, false
}
func (m *GiaManager) HashAnd(a, b GiaLit) (GiaLit, error) {
	if !m.validSignal(a) || !m.validSignal(b) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	if l, yes := giaAndTrivial(a, b); yes {
		return l, nil
	}
	if m.hash == nil {
		return InvalidGiaLit, ErrInvariant
	}
	if a > b {
		a, b = b, a
	}
	if id, exists := m.hash[giaHashKey{a: a, b: b}]; exists {
		return GiaLit(id) << 1, nil
	}
	return m.AppendAnd(a, b)
}

// HashAndTry implements Gia_ManHashAndTry without allocating a new node.
func (m *GiaManager) HashAndTry(a, b GiaLit) (GiaLit, bool, error) {
	if !m.validSignal(a) || !m.validSignal(b) {
		return InvalidGiaLit, false, ErrInvalidLiteral
	}
	if l, yes := giaAndTrivial(a, b); yes {
		return l, true, nil
	}
	if m.hash == nil {
		return InvalidGiaLit, false, ErrInvariant
	}
	if a > b {
		a, b = b, a
	}
	id, yes := m.hash[giaHashKey{a: a, b: b}]
	if !yes {
		return InvalidGiaLit, false, nil
	}
	return GiaLit(id) << 1, true, nil
}
func (m *GiaManager) HashOr(a, b GiaLit) (GiaLit, error) {
	l, err := m.HashAnd(a.Not(), b.Not())
	if err != nil {
		return InvalidGiaLit, err
	}
	return l.Not(), nil
}
func (m *GiaManager) HashXor(a, b GiaLit) (GiaLit, error) {
	if !m.validSignal(a) || !m.validSignal(b) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	complement := a.Complemented() != b.Complemented()
	a = a.Regular()
	b = b.Regular()
	left, err := m.HashAnd(a, b.Not())
	if err != nil {
		return InvalidGiaLit, err
	}
	right, err := m.HashAnd(b, a.Not())
	if err != nil {
		return InvalidGiaLit, err
	}
	l, err := m.HashAnd(left.Not(), right.Not())
	if err != nil {
		return InvalidGiaLit, err
	}
	return l.NotIf(!complement), nil
}
func (m *GiaManager) HashMux(control, whenTrue, whenFalse GiaLit) (GiaLit, error) {
	if !m.validSignal(control) || !m.validSignal(whenTrue) || !m.validSignal(whenFalse) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	if whenFalse > whenTrue {
		whenFalse, whenTrue = whenTrue, whenFalse
		control = control.Not()
	}
	complement := whenTrue.Complemented()
	if complement {
		whenFalse = whenFalse.Not()
		whenTrue = whenTrue.Not()
	}
	left, err := m.HashAnd(control.Not(), whenFalse)
	if err != nil {
		return InvalidGiaLit, err
	}
	right, err := m.HashAnd(control, whenTrue)
	if err != nil {
		return InvalidGiaLit, err
	}
	l, err := m.HashAnd(left.Not(), right.Not())
	if err != nil {
		return InvalidGiaLit, err
	}
	return l.NotIf(!complement), nil
}
func (m *GiaManager) HashMajority(a, b, c GiaLit) (GiaLit, error) {
	if !m.validSignal(a) || !m.validSignal(b) || !m.validSignal(c) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	bcOr, err := m.HashOr(b, c)
	if err != nil {
		return InvalidGiaLit, err
	}
	left, err := m.HashAnd(a, bcOr)
	if err != nil {
		return InvalidGiaLit, err
	}
	right, err := m.HashAnd(b, c)
	if err != nil {
		return InvalidGiaLit, err
	}
	return m.HashOr(left, right)
}
func (m *GiaManager) HashXorReal(a, b GiaLit) (GiaLit, error) {
	if !m.validSignal(a) || !m.validSignal(b) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	switch {
	case a == GiaConst0:
		return b, nil
	case a == GiaConst1:
		return b.Not(), nil
	case b == GiaConst0:
		return a, nil
	case b == GiaConst1:
		return a.Not(), nil
	case a == b:
		return GiaConst0, nil
	case a == b.Not():
		return GiaConst1, nil
	}
	if m.hash == nil {
		return InvalidGiaLit, ErrInvariant
	}
	complement := a.Complemented() != b.Complemented()
	a = a.Regular()
	b = b.Regular()
	if a < b {
		a, b = b, a
	}
	if id, exists := m.hash[giaHashKey{a: a, b: b, kind: 1}]; exists {
		return (GiaLit(id) << 1).NotIf(complement), nil
	}
	l, err := m.AppendXorReal(a, b)
	if err != nil {
		return InvalidGiaLit, err
	}
	return l.NotIf(complement), nil
}
func (m *GiaManager) HashMuxReal(control, whenTrue, whenFalse GiaLit) (GiaLit, error) {
	if !m.validSignal(control) || !m.validSignal(whenTrue) || !m.validSignal(whenFalse) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	switch {
	case control == GiaConst0:
		return whenFalse, nil
	case control == GiaConst1:
		return whenTrue, nil
	case whenFalse == GiaConst0:
		return m.HashAnd(control, whenTrue)
	case whenFalse == GiaConst1:
		return m.HashOr(control.Not(), whenTrue)
	case whenTrue == GiaConst0:
		return m.HashAnd(control.Not(), whenFalse)
	case whenTrue == GiaConst1:
		return m.HashOr(control, whenFalse)
	case whenFalse == whenTrue:
		return whenFalse, nil
	case control == whenFalse || control == whenTrue.Not():
		return m.HashAnd(whenFalse, whenTrue)
	case control == whenTrue || control == whenFalse.Not():
		return m.HashOr(whenFalse, whenTrue)
	case whenFalse.ID() == whenTrue.ID():
		return m.HashXorReal(control, whenFalse)
	}
	if m.hash == nil {
		return InvalidGiaLit, ErrInvariant
	}
	if whenFalse > whenTrue {
		whenFalse, whenTrue = whenTrue, whenFalse
		control = control.Not()
	}
	complement := whenTrue.Complemented()
	if complement {
		whenFalse = whenFalse.Not()
		whenTrue = whenTrue.Not()
	}
	key := giaHashKey{a: whenFalse, b: whenTrue, c: control, kind: 2}
	if id, exists := m.hash[key]; exists {
		return (GiaLit(id) << 1).NotIf(complement), nil
	}
	l, err := m.AppendMuxReal(control, whenTrue, whenFalse)
	if err != nil {
		return InvalidGiaLit, err
	}
	return l.NotIf(complement), nil
}
