package abc

// Port of the live interpolation behavior in src/sat/bsat/satSolver2i.c.

type GiaOutputRelationInterpolation struct {
	Graph          *GiaManager
	OriginalInputs int
	ClaimInput     int
	SourceObjects  int
	ResultObjects  int
}

// InterpolateOutputRelation implements the observable contract of
// Gia_ManInterTest. For a combinational one-output graph, it returns a graph
// whose final input is a claimed output value and whose sole output is true
// exactly when that claim equals the source output.
func (m *GiaManager) InterpolateOutputRelation(options GiaCofactorOptions) (GiaOutputRelationInterpolation, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return GiaOutputRelationInterpolation{}, ErrDimension
	}
	if err := m.Check(); err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	if m.registers != 0 || len(m.cos) != 1 {
		return GiaOutputRelationInterpolation{}, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if len(m.objects) > options.MaxMappingEntries {
		return GiaOutputRelationInterpolation{}, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return GiaOutputRelationInterpolation{}, ErrCancelled
	}

	output := m.objects[m.cos[0]]
	driver, err := output.Fanin0(m.cos[0])
	if err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	used := make([]bool, len(m.objects))
	stack := []GiaID{driver.ID()}
	for len(stack) > 0 {
		if stopped() {
			return GiaOutputRelationInterpolation{}, ErrCancelled
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if used[id] {
			continue
		}
		used[id] = true
		object := m.objects[id]
		if id == 0 || object.IsCI() {
			continue
		}
		if !object.IsAnd() {
			return GiaOutputRelationInterpolation{}, ErrInvariant
		}
		first, firstErr := object.Fanin0(id)
		second, secondErr := object.Fanin1(id)
		if firstErr != nil || secondErr != nil {
			return GiaOutputRelationInterpolation{}, ErrInvariant
		}
		stack = append(stack, first.ID(), second.ID())
		if control, yes := m.MuxControl(id); yes {
			stack = append(stack, control.ID())
		}
	}

	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	target.Name = "interpolant"
	if err = target.HashStart(); err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis {
		if stopped() {
			return GiaOutputRelationInterpolation{}, ErrCancelled
		}
		mapping[id], err = target.AppendCI()
		if err != nil {
			return GiaOutputRelationInterpolation{}, err
		}
	}
	claim, err := target.AppendCI()
	if err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	translate := func(literal GiaLit) (GiaLit, error) {
		if uint64(literal.ID()) >= uint64(len(mapping)) || !mapping[literal.ID()].Valid() {
			return InvalidGiaLit, ErrInvariant
		}
		return mapping[literal.ID()].NotIf(literal.Complemented()), nil
	}
	for id := GiaID(1); uint64(id) < uint64(len(m.objects)); id++ {
		if !used[id] || m.objects[id].IsCI() {
			continue
		}
		if stopped() {
			return GiaOutputRelationInterpolation{}, ErrCancelled
		}
		object := m.objects[id]
		first, _ := object.Fanin0(id)
		first, err = translate(first)
		if err != nil {
			return GiaOutputRelationInterpolation{}, err
		}
		if object.IsBuffer() {
			mapping[id] = first
			continue
		}
		second, _ := object.Fanin1(id)
		second, err = translate(second)
		if err != nil {
			return GiaOutputRelationInterpolation{}, err
		}
		if control, yes := m.MuxControl(id); yes {
			control, err = translate(control)
			if err == nil {
				mapping[id], err = target.HashMux(control, second, first)
			}
		} else if object.IsXor() {
			mapping[id], err = target.HashXor(first, second)
		} else {
			mapping[id], err = target.HashAnd(first, second)
		}
		if err != nil {
			return GiaOutputRelationInterpolation{}, err
		}
	}
	result, err := translate(driver)
	if err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	result, err = target.HashXor(result, claim)
	if err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	if _, err = target.AppendCO(result.Not()); err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	target.HashStop()
	if err = target.Check(); err != nil {
		return GiaOutputRelationInterpolation{}, err
	}
	return GiaOutputRelationInterpolation{
		Graph:          target,
		OriginalInputs: len(m.cis),
		ClaimInput:     len(m.cis),
		SourceObjects:  len(m.objects),
		ResultObjects:  len(target.objects),
	}, nil
}
