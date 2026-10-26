package abc

// Native checks and evaluation of the representation in src/aig/gia/gia.h.

import "fmt"

func (m *GiaManager) Check() error {
	fail := func(message string) error { return fmt.Errorf("%w: GIA %s", ErrInvariant, message) }
	if len(m.objects) == 0 || uint64(len(m.objects)) > uint64(MaxGiaID)+1 || len(m.objects) > m.objectLimit {
		return fail("object count")
	}
	if m.objects[0].word0 != GiaNone || m.objects[0].word1&^giaTerm != GiaNone {
		return fail("constant zero")
	}
	if m.registers < 0 || m.registers > len(m.cis) || m.registers > len(m.cos) || m.constraints < 0 || m.constraints > m.PrimaryOutputCount() {
		return fail("terminal partition")
	}
	if m.muxes != nil && len(m.muxes) != len(m.objects) {
		return fail("mux array dimensions")
	}
	if m.muxes != nil && m.muxes[0] != 0 {
		return fail("constant mux control")
	}
	ci, co, xors, buffers, muxes := 0, 0, 0, 0, 0
	unique := make(map[giaHashKey]GiaID)
	for i := 1; i < len(m.objects); i++ {
		id := GiaID(i)
		obj := m.objects[i]
		switch {
		case obj.IsCI():
			if ci >= len(m.cis) || m.cis[ci] != id || obj.Diff1() != uint32(ci) || obj.Complement0() || obj.Complement1() {
				return fail("CI metadata")
			}
			ci++
		case obj.IsCO():
			if co >= len(m.cos) || m.cos[co] != id || obj.Diff1() != uint32(co) || obj.Complement1() {
				return fail("CO metadata")
			}
			co++
		case obj.IsAnd():
			if obj.IsBuffer() {
				buffers++
			} else if obj.IsXor() {
				xors++
			}
		default:
			return fail("object type")
		}
		if obj.IsAnd() || obj.IsCO() {
			a, err := obj.Fanin0(id)
			if err != nil || !m.validSignal(a) || a.ID() >= id {
				return fail("first fanin")
			}
			if obj.IsAnd() {
				b, err := obj.Fanin1(id)
				if err != nil || !m.validSignal(b) || b.ID() >= id {
					return fail("second fanin")
				}
				if obj.IsBuffer() && a != b {
					return fail("buffer polarity")
				}
				if control, yes := m.MuxControl(id); yes {
					if obj.IsBuffer() || obj.IsXor() || !m.validSignal(control) || control.ID() == 0 || control.ID() >= id || control.ID() == a.ID() || control.ID() == b.ID() {
						return fail("native mux")
					}
					muxes++
				}
				if m.hash != nil {
					key := m.objectKey(id)
					if _, exists := unique[key]; !exists {
						unique[key] = id
					}
					if m.hash[key] != unique[key] {
						return fail("hash entry")
					}
				}
			}
		}
		if !obj.IsAnd() {
			if _, yes := m.MuxControl(id); yes {
				return fail("terminal mux control")
			}
		}
	}
	if ci != len(m.cis) || co != len(m.cos) || xors != m.xors || buffers != m.buffers || muxes != m.muxCount {
		return fail("object categories")
	}
	if m.hash != nil && len(unique) != len(m.hash) {
		return fail("hash count")
	}
	return nil
}
func (m *GiaManager) evaluateValues(inputs []bool) ([]bool, error) {
	if len(m.objects) == 0 {
		return nil, ErrInvariant
	}
	if len(inputs) != len(m.cis) {
		return nil, ErrDimension
	}
	values := make([]bool, len(m.objects))
	for i, id := range m.cis {
		values[id] = inputs[i]
	}
	read := func(l GiaLit) bool { return values[l.ID()] != l.Complemented() }
	for i, obj := range m.objects {
		if !obj.IsAnd() && !obj.IsCO() {
			continue
		}
		a, _ := obj.Fanin0(GiaID(i))
		if obj.IsCO() || obj.IsBuffer() {
			values[i] = read(a)
			continue
		}
		b, _ := obj.Fanin1(GiaID(i))
		if c, yes := m.MuxControl(GiaID(i)); yes {
			if read(c) {
				values[i] = read(b)
			} else {
				values[i] = read(a)
			}
		} else if obj.IsXor() {
			values[i] = read(a) != read(b)
		} else {
			values[i] = read(a) && read(b)
		}
	}
	return values, nil
}
func (m *GiaManager) Evaluate(inputs []bool) ([]bool, error) {
	values, err := m.evaluateValues(inputs)
	if err != nil {
		return nil, err
	}
	outputs := make([]bool, len(m.cos))
	for i, id := range m.cos {
		outputs[i] = values[id]
	}
	return outputs, nil
}
func (m *GiaManager) Step(primary, state []bool) (outputs, next []bool, err error) {
	if len(primary) != m.PrimaryInputCount() || len(state) != m.registers {
		return nil, nil, ErrDimension
	}
	inputs := make([]bool, len(m.cis))
	copy(inputs, primary)
	copy(inputs[len(primary):], state)
	all, err := m.Evaluate(inputs)
	if err != nil {
		return nil, nil, err
	}
	n := m.PrimaryOutputCount()
	return all[:n:n], all[n:], nil
}

// SetPhases calculates Gia_ObjPhase under all-zero CIs without borrowing
// user marks or Value fields. Later graph mutations leave existing phases
// as source-compatible scratch data; call again when a fresh phase is needed.
func (m *GiaManager) SetPhases() error {
	if err := m.Check(); err != nil {
		return err
	}
	values, err := m.evaluateValues(make([]bool, len(m.cis)))
	if err != nil {
		return err
	}
	for i, v := range values {
		m.objects[i].word1 &^= giaTerm
		if v {
			m.objects[i].word1 |= giaTerm
		}
	}
	return nil
}
