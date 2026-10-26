package abc

// Enable removal from src/aig/gia/giaEnable.c:
// Gia_ManRemoveEnables2 and Gia_ManRemoveEnables.

type GiaEnableRemoval struct {
	Graph      *GiaManager
	Recognized []bool
	Controls   []GiaLit
	Data       []GiaLit
	Classes    []int
}

// RemoveMuxEnables removes recognized register self-loop muxes. For each
// recognized register, its next-state function becomes the mux's non-feedback
// branch with the original RI phase applied. Nonmatches retain their original
// driver. Controls are normalized so true selects the update branch. Results
// are indexed by register and refer to source literals.
func (m *GiaManager) RemoveMuxEnables(options GiaCofactorOptions) (GiaEnableRemoval, error) {
	result, err := m.prepareEnableRemoval(options)
	if err != nil {
		return GiaEnableRemoval{}, err
	}
	for register, ri := range m.cos[m.PrimaryOutputCount():] {
		if options.Stop != nil && options.Stop() {
			return GiaEnableRemoval{}, ErrCancelled
		}
		driver, err := m.objects[ri].Fanin0(ri)
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		control, whenTrue, whenFalse, ok, err := m.RecognizeMux(driver.ID())
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		if !ok {
			continue
		}
		ro := m.cis[m.PrimaryInputCount()+register]
		switch {
		case whenFalse.ID() == ro:
			result.Recognized[register] = true
			result.Controls[register] = control
			result.Data[register] = whenTrue.NotIf(driver.Complemented())
		case whenTrue.ID() == ro:
			result.Recognized[register] = true
			result.Controls[register] = control.Not()
			result.Data[register] = whenFalse.NotIf(driver.Complemented())
		}
	}
	result.Graph, err = m.rebuildWithoutEnables(result.Data, result.Recognized, true, options)
	if err != nil {
		return GiaEnableRemoval{}, err
	}
	return result, nil
}

// RemoveStructuralEnables ports the older Gia_ManRemoveEnables recognizer.
// It accepts the explicit OR-of-two-AND representation, removes recognized
// feedback branches and assigns one-based classes to equal control literals.
// Class zero denotes an unrecognized register.
func (m *GiaManager) RemoveStructuralEnables(options GiaCofactorOptions) (GiaEnableRemoval, error) {
	result, err := m.prepareEnableRemoval(options)
	if err != nil {
		return GiaEnableRemoval{}, err
	}
	classes := make(map[GiaLit]int)
	for register, ri := range m.cos[m.PrimaryOutputCount():] {
		if options.Stop != nil && options.Stop() {
			return GiaEnableRemoval{}, ErrCancelled
		}
		driver, err := m.objects[ri].Fanin0(ri)
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		root := m.objects[driver.ID()]
		if !root.IsAnd() || root.IsBuffer() || root.IsXor() || !root.Complement0() || !root.Complement1() {
			continue
		}
		leftRoot, _ := root.Fanin0(driver.ID())
		rightRoot, _ := root.Fanin1(driver.ID())
		left, right := m.objects[leftRoot.ID()], m.objects[rightRoot.ID()]
		if !left.IsAnd() || !right.IsAnd() {
			continue
		}
		children := [2][2]GiaLit{}
		children[0][0], err = left.Fanin0(leftRoot.ID())
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		children[0][1], err = left.Fanin1(leftRoot.ID())
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		children[1][0], err = right.Fanin0(rightRoot.ID())
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		children[1][1], err = right.Fanin1(rightRoot.ID())
		if err != nil {
			return GiaEnableRemoval{}, err
		}
		ro := GiaLit(m.cis[m.PrimaryInputCount()+register]) << 1
		feedback := ro.NotIf(!driver.Complemented())
		branch, slot := -1, -1
		for side := 0; side < 2 && branch < 0; side++ {
			for input := 0; input < 2; input++ {
				if children[side][input] == feedback {
					branch, slot = side, input
					break
				}
			}
		}
		if branch < 0 {
			continue
		}
		control := children[branch][1-slot].Not()
		other := 1 - branch
		dataSlot := 0
		if children[other][0].ID() == control.ID() {
			dataSlot = 1
		}
		data := children[other][dataSlot].NotIf(!driver.Complemented())
		class, exists := classes[control]
		if !exists {
			class = len(classes) + 1
			classes[control] = class
		}
		result.Recognized[register] = true
		result.Controls[register] = control
		result.Data[register] = data
		result.Classes[register] = class
	}
	result.Graph, err = m.rebuildWithoutEnables(result.Data, result.Recognized, false, options)
	if err != nil {
		return GiaEnableRemoval{}, err
	}
	return result, nil
}

func (m *GiaManager) prepareEnableRemoval(options GiaCofactorOptions) (GiaEnableRemoval, error) {
	var result GiaEnableRemoval
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return result, ErrDimension
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxMappingEntries {
		return result, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return result, ErrCancelled
	}
	result.Recognized = make([]bool, m.registers)
	result.Controls = make([]GiaLit, m.registers)
	result.Data = make([]GiaLit, m.registers)
	result.Classes = make([]int, m.registers)
	for i := range result.Controls {
		result.Controls[i] = InvalidGiaLit
		result.Data[i] = InvalidGiaLit
	}
	return result, nil
}

func (m *GiaManager) rebuildWithoutEnables(data []GiaLit, recognized []bool, hashed bool, options GiaCofactorOptions) (*GiaManager, error) {
	maxObjects := options.MaxObjects
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	target, err := NewGiaManagerWithLimit(maxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	if hashed {
		if err = target.HashStart(); err != nil {
			return nil, err
		}
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	translate := func(l GiaLit) (GiaLit, error) {
		if !l.Valid() || uint64(l.ID()) >= uint64(len(mapping)) || !mapping[l.ID()].Valid() {
			return InvalidGiaLit, ErrInvariant
		}
		return mapping[l.ID()].NotIf(l.Complemented()), nil
	}
	for _, id := range m.cis {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		mapping[id], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		id := GiaID(i)
		a, err := obj.Fanin0(id)
		if err != nil {
			return nil, err
		}
		left, err := translate(a)
		if err != nil {
			return nil, err
		}
		if obj.IsBuffer() {
			if hashed {
				mapping[id] = left
			} else {
				mapping[id], err = target.AppendBuffer(left)
			}
		} else {
			b, faninErr := obj.Fanin1(id)
			if faninErr != nil {
				return nil, faninErr
			}
			right, translateErr := translate(b)
			if translateErr != nil {
				return nil, translateErr
			}
			if control, yes := m.MuxControl(id); yes {
				condition, controlErr := translate(control)
				if controlErr != nil {
					return nil, controlErr
				}
				if hashed {
					mapping[id], err = target.HashMuxReal(condition, right, left)
				} else {
					mapping[id], err = target.AppendMuxReal(condition, right, left)
				}
			} else if obj.IsXor() {
				if hashed {
					mapping[id], err = target.HashXorReal(left, right)
				} else {
					mapping[id], err = target.AppendXorReal(left, right)
				}
			} else if hashed {
				mapping[id], err = target.HashAnd(left, right)
			} else {
				mapping[id], err = target.AppendAnd(left, right)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	for output, id := range m.cos {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		if output >= m.PrimaryOutputCount() {
			register := output - m.PrimaryOutputCount()
			if recognized[register] {
				driver = data[register]
			}
		}
		copy, err := translate(driver)
		if err != nil {
			return nil, err
		}
		if _, err = target.AppendCO(copy); err != nil {
			return nil, err
		}
	}
	if hashed {
		target.HashStop()
	}
	if err = target.SetRegisterCount(m.registers); err != nil {
		return nil, err
	}
	if err = target.SetConstraintCount(m.constraints); err != nil {
		return nil, err
	}
	cleaned, _, err := target.Cleanup()
	if err != nil {
		return nil, err
	}
	return cleaned, nil
}
