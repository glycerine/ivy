package abc

// Ternary evaluation from src/aig/gia/gia.h:Gia_XsimNotCond/Gia_XsimAndCond
// and giaTsim.c:Gia_ManTerSimulateRound/Node/Co (owned full-object storage).

// GiaTernary uses ABC's encodings: zero=1, one=2, unknown=3. Encoding zero is
// invalid. Unknown values denote independent Boolean possibilities per gate;
// reconvergent correlations are not recovered, so evaluation is conservative.
type GiaTernary uint8

const (
	GiaTernaryZero    GiaTernary = 1
	GiaTernaryOne     GiaTernary = 2
	GiaTernaryUnknown GiaTernary = 3
)

func (v GiaTernary) Valid() bool { return v >= GiaTernaryZero && v <= GiaTernaryUnknown }
func giaTernaryNot(v GiaTernary) GiaTernary {
	if v == GiaTernaryUnknown {
		return v
	}
	return 3 - v
}
func giaTernaryAnd(a, b GiaTernary) GiaTernary {
	if a == GiaTernaryZero || b == GiaTernaryZero {
		return GiaTernaryZero
	}
	if a == GiaTernaryUnknown || b == GiaTernaryUnknown {
		return GiaTernaryUnknown
	}
	return GiaTernaryOne
}
func giaTernaryXor(a, b GiaTernary) GiaTernary {
	if a == GiaTernaryUnknown || b == GiaTernaryUnknown {
		return GiaTernaryUnknown
	}
	if a == b {
		return GiaTernaryZero
	}
	return GiaTernaryOne
}

// EvaluateTernary returns owned CO values for CI values in terminal order.
// Buffers preserve their phased input; native XORs/muxes use conservative
// three-valued truth tables, including equal mux branches with unknown control.
// Constants, PO/RI roles and source scratch are unchanged. Resource limits and
// cancellation return no partial result. MaxMappingEntries bounds value storage.
func (m *GiaManager) EvaluateTernary(inputs []GiaTernary, options GiaCofactorOptions) ([]GiaTernary, error) {
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
	if len(inputs) != len(m.cis) {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	for _, v := range inputs {
		if !v.Valid() {
			return nil, ErrDimension
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	values := make([]GiaTernary, len(m.objects))
	values[0] = GiaTernaryZero
	for i, id := range m.cis {
		values[id] = inputs[i]
	}
	read := func(l GiaLit) GiaTernary {
		v := values[l.ID()]
		if l.Complemented() {
			v = giaTernaryNot(v)
		}
		return v
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() && !obj.IsCO() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		id := GiaID(i)
		first, e := obj.Fanin0(id)
		if e != nil {
			return nil, e
		}
		a := read(first)
		if obj.IsBuffer() || obj.IsCO() {
			values[i] = a
			continue
		}
		second, e := obj.Fanin1(id)
		if e != nil {
			return nil, e
		}
		b := read(second)
		if control, yes := m.MuxControl(id); yes {
			c := read(control)
			if c == GiaTernaryZero {
				values[i] = a
			} else if c == GiaTernaryOne {
				values[i] = b
			} else if a == b {
				values[i] = a
			} else {
				values[i] = GiaTernaryUnknown
			}
		} else if obj.IsXor() {
			values[i] = giaTernaryXor(a, b)
		} else {
			values[i] = giaTernaryAnd(a, b)
		}
	}
	outputs := make([]GiaTernary, len(m.cos))
	for i, id := range m.cos {
		outputs[i] = values[id]
	}
	return outputs, nil
}
