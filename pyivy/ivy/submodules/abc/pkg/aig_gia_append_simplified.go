package abc

// Non-strashing trivial simplification from src/aig/gia/gia.h:
// Gia_ManAppendAnd2/Or2/Mux2/Maj2/Xor2/XorReal2 with fGiaSimple=0.

// AppendAndSimplified folds constants and equal/complemented literals, then
// appends a raw AND. It does not merge duplicate existing gates. Like C's
// AppendAnd2, simplifications may succeed even when no object capacity remains.
func (m *GiaManager) AppendAndSimplified(a, b GiaLit) (GiaLit, error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	if a == GiaConst0 || b == GiaConst0 || a == b.Not() {
		return GiaConst0, nil
	}
	if a == GiaConst1 {
		return b, nil
	}
	if b == GiaConst1 || a == b {
		return a, nil
	}
	return m.AppendAnd(a, b)
}

// AppendOrSimplified uses De Morgan's law and AppendAndSimplified.
func (m *GiaManager) AppendOrSimplified(a, b GiaLit) (GiaLit, error) {
	l, e := m.AppendAndSimplified(a.Not(), b.Not())
	if e != nil {
		return InvalidGiaLit, e
	}
	return l.Not(), nil
}

// AppendMuxSimplified builds C's three-AND mux with local simplification.
// A capacity failure may leave intermediate gates, as with other low-level
// multi-gate constructors; source signal validation precedes all mutation.
func (m *GiaManager) AppendMuxSimplified(control, whenTrue, whenFalse GiaLit) (GiaLit, error) {
	if m == nil || !m.validSignal(control) || !m.validSignal(whenTrue) || !m.validSignal(whenFalse) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	low, e := m.AppendAndSimplified(control.Not(), whenFalse)
	if e != nil {
		return InvalidGiaLit, e
	}
	high, e := m.AppendAndSimplified(control, whenTrue)
	if e != nil {
		return InvalidGiaLit, e
	}
	joined, e := m.AppendAndSimplified(low.Not(), high.Not())
	if e != nil {
		return InvalidGiaLit, e
	}
	return joined.Not(), nil
}

// AppendMajoritySimplified follows C's OR/AND decomposition with local folds.
func (m *GiaManager) AppendMajoritySimplified(a, b, c GiaLit) (GiaLit, error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) || !m.validSignal(c) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	either, e := m.AppendOrSimplified(b, c)
	if e != nil {
		return InvalidGiaLit, e
	}
	first, e := m.AppendAndSimplified(a, either)
	if e != nil {
		return InvalidGiaLit, e
	}
	second, e := m.AppendAndSimplified(b, c)
	if e != nil {
		return InvalidGiaLit, e
	}
	return m.AppendOrSimplified(first, second)
}

// AppendXorSimplified lowers XOR using the source's simplifying mux pattern.
func (m *GiaManager) AppendXorSimplified(a, b GiaLit) (GiaLit, error) {
	return m.AppendMuxSimplified(a, b.Not(), b)
}

// AppendXorRealSimplified folds constants/equal/complemented literals before
// appending a native XOR; duplicate nontrivial XORs remain distinct.
func (m *GiaManager) AppendXorRealSimplified(a, b GiaLit) (GiaLit, error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	if a == GiaConst0 {
		return b, nil
	}
	if a == GiaConst1 {
		return b.Not(), nil
	}
	if b == GiaConst0 {
		return a, nil
	}
	if b == GiaConst1 {
		return a.Not(), nil
	}
	if a == b {
		return GiaConst0, nil
	}
	if a == b.Not() {
		return GiaConst1, nil
	}
	return m.AppendXorReal(a, b)
}
