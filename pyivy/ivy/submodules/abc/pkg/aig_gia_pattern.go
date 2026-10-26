package abc

// Partial-pattern verification from src/aig/gia/giaPat.c.

type GiaPatternAssignment struct {
	Input int
	Value bool
}

type GiaPatternVerification struct {
	Cone     []GiaID
	Value    GiaTernary
	Asserted bool
}

// CollectPatternCone ports Gia_SatCollectCone[_rec]. It returns CIs and gates
// once in fanin postorder. The root is a regular CI or gate object, never a CO
// wrapper or constant. Owned traversal state replaces object marks.
func (m *GiaManager) CollectPatternCone(root GiaID, maxObjects int) ([]GiaID, error) {
	if m == nil || maxObjects < 0 {
		return nil, ErrDimension
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if root == 0 || uint64(root) >= uint64(len(m.objects)) || (!m.objects[root].IsCI() && !m.objects[root].IsAnd()) {
		return nil, ErrInvalidLiteral
	}
	return m.collectPatternCone(root, maxObjects, nil)
}

func (m *GiaManager) collectPatternCone(root GiaID, maxObjects int, stopped func() bool) ([]GiaID, error) {
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	seen := make([]bool, len(m.objects))
	seen[0] = true
	type visit struct {
		id   GiaID
		exit bool
	}
	stack := []visit{{id: root}}
	result := make([]GiaID, 0)
	for len(stack) > 0 {
		if stopped != nil && stopped() {
			return nil, ErrCancelled
		}
		last := len(stack) - 1
		item := stack[last]
		stack = stack[:last]
		if item.exit {
			result = append(result, item.id)
			continue
		}
		if seen[item.id] {
			continue
		}
		seen[item.id] = true
		children, count, err := m.crossCutFanins(item.id)
		if err != nil {
			return nil, err
		}
		stack = append(stack, visit{id: item.id, exit: true})
		for i := count - 1; i >= 0; i-- {
			stack = append(stack, visit{id: children[i]})
		}
	}
	return result, nil
}

// VerifyOutputPattern ports Gia_SatVerifyPattern without manager marks or
// stdout. Unassigned CIs are unknown; assignments use CI ordinals and later
// duplicate assignments win. Value is the ternary output and Asserted reports
// whether it is definitely one. Cone is an owned simulation postorder.
func (m *GiaManager) VerifyOutputPattern(output int, assignments []GiaPatternAssignment, options GiaCofactorOptions) (GiaPatternVerification, error) {
	var result GiaPatternVerification
	if m == nil || output < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return result, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	if output >= len(m.cos) || len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxMappingEntries || len(assignments) > options.MaxMappingEntries {
		if output >= len(m.cos) {
			return result, ErrDimension
		}
		return result, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return result, ErrCancelled
	}
	co := m.cos[output]
	driver, err := m.objects[co].Fanin0(co)
	if err != nil {
		return result, err
	}
	if driver.ID() == 0 {
		return result, ErrInvalidLiteral
	}
	result.Cone, err = m.collectPatternCone(driver.ID(), options.MaxMappingEntries, stopped)
	if err != nil {
		return GiaPatternVerification{}, err
	}
	values := make([]GiaTernary, len(m.objects))
	values[0] = GiaTernaryZero
	for _, id := range result.Cone {
		values[id] = GiaTernaryUnknown
	}
	for _, assignment := range assignments {
		if assignment.Input < 0 || assignment.Input >= len(m.cis) {
			return GiaPatternVerification{}, ErrDimension
		}
		if assignment.Value {
			values[m.cis[assignment.Input]] = GiaTernaryOne
		} else {
			values[m.cis[assignment.Input]] = GiaTernaryZero
		}
	}
	read := func(l GiaLit) GiaTernary {
		value := values[l.ID()]
		if l.Complemented() {
			value = giaTernaryNot(value)
		}
		return value
	}
	for _, id := range result.Cone {
		if stopped() {
			return GiaPatternVerification{}, ErrCancelled
		}
		obj := m.objects[id]
		if obj.IsCI() {
			continue
		}
		first, faninErr := obj.Fanin0(id)
		if faninErr != nil {
			return GiaPatternVerification{}, faninErr
		}
		a := read(first)
		if obj.IsBuffer() {
			values[id] = a
			continue
		}
		second, faninErr := obj.Fanin1(id)
		if faninErr != nil {
			return GiaPatternVerification{}, faninErr
		}
		b := read(second)
		if control, yes := m.MuxControl(id); yes {
			c := read(control)
			switch {
			case c == GiaTernaryZero:
				values[id] = a
			case c == GiaTernaryOne:
				values[id] = b
			case a == b:
				values[id] = a
			default:
				values[id] = GiaTernaryUnknown
			}
		} else if obj.IsXor() {
			values[id] = giaTernaryXor(a, b)
		} else {
			values[id] = giaTernaryAnd(a, b)
		}
	}
	result.Value = read(driver)
	result.Asserted = result.Value == GiaTernaryOne
	return result, nil
}
