package abc

// XOR recognition from src/aig/gia/giaUtil.c:Gia_ObjRecognizeExor.

// RecognizeXor returns inputs whose XOR equals the regular object. It handles
// native XORs and complemented-AND XOR/XNOR patterns without mutating scratch.
func (m *GiaManager) RecognizeXor(id GiaID) (a, b GiaLit, ok bool, err error) {
	a, b = InvalidGiaLit, InvalidGiaLit
	obj, err := m.Object(id)
	if err != nil {
		return a, b, false, err
	}
	if _, native := m.MuxControl(id); native || !obj.IsAnd() || obj.IsBuffer() {
		return a, b, false, nil
	}
	l, err := obj.Fanin0(id)
	if err != nil {
		return a, b, false, err
	}
	r, err := obj.Fanin1(id)
	if err != nil {
		return a, b, false, err
	}
	if obj.IsXor() {
		return l, r, true, nil
	}
	if !l.Complemented() || !r.Complemented() {
		return a, b, false, nil
	}
	for _, child := range []GiaLit{l, r} {
		node := m.objects[child.ID()]
		_, native := m.MuxControl(child.ID())
		if !node.IsAnd() || node.IsXor() || native {
			return a, b, false, nil
		}
	}
	x, y := m.objects[l.ID()], m.objects[r.ID()]
	x0, err := x.Fanin0(l.ID())
	if err != nil {
		return a, b, false, err
	}
	x1, err := x.Fanin1(l.ID())
	if err != nil {
		return a, b, false, err
	}
	y0, err := y.Fanin0(r.ID())
	if err != nil {
		return a, b, false, err
	}
	y1, err := y.Fanin1(r.ID())
	if err != nil {
		return a, b, false, err
	}
	if x0 != y0.Not() || x1 != y1.Not() {
		return a, b, false, nil
	}
	return x0, x1, true, nil
}

// MuxControlReferenceCounts counts mux-control uses, excluding XOR patterns
// even though they also admit mux decompositions. All objects, including
// dangling cones, participate. maxObjects zero selects one million objects.
func (m *GiaManager) MuxControlReferenceCounts(maxObjects int) ([]uint64, error) {
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	refs := make([]uint64, len(m.objects))
	for i := range m.objects {
		if _, _, yes, err := m.RecognizeXor(GiaID(i)); err != nil {
			return nil, err
		} else if yes {
			continue
		}
		control, _, _, yes, err := m.RecognizeMux(GiaID(i))
		if err != nil {
			return nil, err
		}
		if yes {
			refs[control.ID()]++
		}
	}
	return refs, nil
}
