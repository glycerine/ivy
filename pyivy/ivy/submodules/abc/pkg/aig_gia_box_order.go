package abc

// Normalized boxed traversal from aig/gia/giaTim.c:
// Gia_ManIsNormalized, Gia_ManOrderWithBoxes_rec and Gia_ManOrderWithBoxes.

// IsNormalized checks C's CI-prefix/CO-suffix object layout on a valid graph.
func (m *GiaManager) IsNormalized() bool {
	if m == nil {
		return false
	}
	if err := m.Check(); err != nil {
		return false
	}
	for i, id := range m.cis {
		if id != GiaID(i+1) {
			return false
		}
	}
	offset := len(m.objects) - len(m.cos)
	for i, id := range m.cos {
		if id != GiaID(offset+i) {
			return false
		}
	}
	return true
}

// OrderWithBoxes returns native staged order: constant/PIs, each box's input
// cones then COs then output CIs, followed by primary-output cones and COs.
// Dangling gates are omitted. Timing boxes must partition the CI suffix and CO
// prefix contiguously. A cone reaching a future box output rejects as a box
// dependency error. MaxVisits bounds returned object entries; MaxObjects the
// arena. Iterative private state preserves graph/timing traversal marks.
// Buffers/XOR/mux are supported; native mux controls are an explicit extension.
func (m *GiaManager) OrderWithBoxes(timing *TimManager, options GiaLUTAnalysisOptions) ([]GiaID, error) {
	if m == nil || timing == nil || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects {
		return nil, ErrLimit
	}
	if !m.IsNormalized() {
		return nil, ErrInvariant
	}
	if len(m.cis) != len(timing.cis) || len(m.cos) != len(timing.cos) {
		return nil, ErrDimension
	}
	pi, po := timing.PrimaryCounts()
	curCI, curCO := pi, 0
	for _, box := range timing.boxes {
		if box.FirstInput != curCO || box.FirstOutput != curCI {
			return nil, ErrInvariant
		}
		curCO += box.Inputs
		curCI += box.Outputs
	}
	if curCI != len(m.cis) || curCO+po != len(m.cos) {
		return nil, ErrInvariant
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	order := make([]GiaID, 0)
	states := make([]uint8, len(m.objects))
	appendID := func(id GiaID) error {
		if len(order) == options.MaxVisits {
			return ErrLimit
		}
		order = append(order, id)
		states[id] = 2
		return nil
	}
	if err := appendID(0); err != nil {
		return nil, err
	}
	for i := 0; i < pi; i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		if err := appendID(m.cis[i]); err != nil {
			return nil, err
		}
	}
	type frame struct {
		id    GiaID
		ready bool
	}
	stack := make([]frame, 0)
	collect := func(root GiaID) error {
		stack = append(stack[:0], frame{id: root})
		for len(stack) > 0 {
			if stopped() {
				return ErrCancelled
			}
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if states[item.id] == 2 {
				continue
			}
			if item.ready {
				if err := appendID(item.id); err != nil {
					return err
				}
				continue
			}
			if m.objects[item.id].IsCI() {
				return ErrInvariant
			}
			if states[item.id] == 1 {
				return ErrInvariant
			}
			states[item.id] = 1
			children, count, err := m.crossCutFanins(item.id)
			if err != nil {
				return err
			}
			stack = append(stack, frame{id: item.id, ready: true})
			for i := count - 1; i >= 0; i-- {
				if states[children[i]] != 2 {
					stack = append(stack, frame{id: children[i]})
				}
			}
		}
		return nil
	}
	collectCOs := func(first, count int) error {
		for i := 0; i < count; i++ {
			co := m.cos[first+i]
			driver, err := m.objects[co].Fanin0(co)
			if err != nil {
				return err
			}
			if err = collect(driver.ID()); err != nil {
				return err
			}
		}
		for i := 0; i < count; i++ {
			if stopped() {
				return ErrCancelled
			}
			if err := appendID(m.cos[first+i]); err != nil {
				return err
			}
		}
		return nil
	}
	for _, box := range timing.boxes {
		if err := collectCOs(box.FirstInput, box.Inputs); err != nil {
			return nil, err
		}
		for i := 0; i < box.Outputs; i++ {
			if stopped() {
				return nil, ErrCancelled
			}
			if err := appendID(m.cis[box.FirstOutput+i]); err != nil {
				return nil, err
			}
		}
	}
	if err := collectCOs(curCO, po); err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return order, nil
}
