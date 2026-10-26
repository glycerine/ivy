package abc

// Mux-pattern recognition from src/aig/gia/giaUtil.c:
// Gia_ObjIsMuxType and Gia_ObjRecognizeMux.

// RecognizeMux returns the mux driving the regular object: control ? whenTrue
// : whenFalse. It recognizes native muxes and complemented-AND mux patterns,
// including XOR-shaped patterns. Input polarity and ABC's pattern-selection
// precedence are preserved. Nonmatches return ok=false; invalid IDs error.
func (m *GiaManager) RecognizeMux(id GiaID) (control, whenTrue, whenFalse GiaLit, ok bool, err error) {
	control, whenTrue, whenFalse = InvalidGiaLit, InvalidGiaLit, InvalidGiaLit
	obj, err := m.Object(id)
	if err != nil {
		return control, whenTrue, whenFalse, false, err
	}
	if c, yes := m.MuxControl(id); yes {
		t, err := obj.Fanin1(id)
		if err != nil {
			return control, whenTrue, whenFalse, false, err
		}
		e, err := obj.Fanin0(id)
		return c, t, e, err == nil, err
	}
	plain := func(id GiaID) bool {
		o := m.objects[id]
		_, native := m.MuxControl(id)
		return o.IsAnd() && !o.IsBuffer() && !o.IsXor() && !native
	}
	if !plain(id) {
		return control, whenTrue, whenFalse, false, nil
	}
	left, err := obj.Fanin0(id)
	if err != nil {
		return control, whenTrue, whenFalse, false, err
	}
	right, err := obj.Fanin1(id)
	if err != nil {
		return control, whenTrue, whenFalse, false, err
	}
	if !left.Complemented() || !right.Complemented() || !plain(left.ID()) || !plain(right.ID()) {
		return control, whenTrue, whenFalse, false, nil
	}
	children := [2][2]GiaLit{}
	for i, node := range []GiaLit{left, right} {
		o := m.objects[node.ID()]
		children[i][0], err = o.Fanin0(node.ID())
		if err != nil {
			return control, whenTrue, whenFalse, false, err
		}
		children[i][1], err = o.Fanin1(node.ID())
		if err != nil {
			return control, whenTrue, whenFalse, false, err
		}
	}
	for _, pair := range [][2]int{{1, 1}, {0, 0}, {0, 1}, {1, 0}} {
		a, b := pair[0], pair[1]
		c0, c1 := children[0][a], children[1][b]
		if c0 != c1.Not() {
			continue
		}
		if c0.Complemented() {
			return c1, children[1][1-b].Not(), children[0][1-a].Not(), true, nil
		}
		return c0, children[0][1-a].Not(), children[1][1-b].Not(), true, nil
	}
	return control, whenTrue, whenFalse, false, nil
}
