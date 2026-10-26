package abc

// Box records/associations from misc/tim/timBox.c; box traversal labels from
// misc/tim/timTrav.c; primary/black-box counts from misc/tim/timMan.c.

type TimBox struct {
	ID, TravID                               int
	FirstInput, Inputs, FirstOutput, Outputs int // inputs are COs; outputs are CIs
	DelayTable, Copy                         int
	Black                                    bool
}

func (b TimBox) InputLast() int  { return b.FirstInput + b.Inputs - 1 }
func (b TimBox) OutputLast() int { return b.FirstOutput + b.Outputs - 1 }

type TimBoxOptions struct{ MaxBoxes, MaxPins int } // zero: one million; pins count all stored associations

// AddBox appends a box and assigns its terminal ownership/position. Overlaps
// overwrite terminal associations in insertion order, as in C. Nonpositive
// dimensions reject native range-query undefined behavior. Delay-table IDs
// are stored without evaluating them; boxed propagation remains unsupported.
func (m *TimManager) AddBox(spec TimBox, options TimBoxOptions) (int, error) {
	if m == nil || options.MaxBoxes < 0 || options.MaxPins < 0 || spec.FirstInput < 0 || spec.FirstOutput < 0 || spec.Inputs < 1 || spec.Outputs < 1 {
		return -1, ErrDimension
	}
	if spec.FirstInput > len(m.cos) || spec.Inputs > len(m.cos)-spec.FirstInput || spec.FirstOutput > len(m.cis) || spec.Outputs > len(m.cis)-spec.FirstOutput {
		return -1, ErrDimension
	}
	if int64(spec.DelayTable) < -2147483648 || int64(spec.DelayTable) > 2147483647 {
		return -1, ErrDimension
	}
	if options.MaxBoxes == 0 {
		options.MaxBoxes = 1 << 20
	}
	if options.MaxPins == 0 {
		options.MaxPins = 1 << 20
	}
	if len(m.boxes) >= options.MaxBoxes {
		return -1, ErrLimit
	}
	pins, err := checkedSize(m.boxPins, 1, spec.Inputs, options.MaxPins)
	if err != nil {
		return -1, err
	}
	pins, err = checkedSize(pins, 1, spec.Outputs, options.MaxPins)
	if err != nil {
		return -1, err
	}
	spec.ID = len(m.boxes)
	spec.TravID = 0
	spec.Copy = 0
	m.boxes = append(m.boxes, spec)
	m.boxPins = pins
	for i := 0; i < spec.Inputs; i++ {
		p := &m.cos[spec.FirstInput+i]
		p.Box = spec.ID
		p.Position = i
	}
	for i := 0; i < spec.Outputs; i++ {
		p := &m.cis[spec.FirstOutput+i]
		p.Box = spec.ID
		p.Position = i
	}
	return spec.ID, nil
}
func (m *TimManager) Box(index int) (TimBox, error) {
	if m == nil || index < 0 || index >= len(m.boxes) {
		return TimBox{}, ErrDimension
	}
	return m.boxes[index], nil
}
func (m *TimManager) BoxCount() int {
	if m == nil {
		return 0
	}
	return len(m.boxes)
}
func (m *TimManager) BoxForCI(index int) int {
	if m == nil || index < 0 || index >= len(m.cis) {
		return -1
	}
	return m.cis[index].Box
}
func (m *TimManager) BoxForCO(index int) int {
	if m == nil || index < 0 || index >= len(m.cos) {
		return -1
	}
	return m.cos[index].Box
}

// FindBoxFromCI preserves Tim_ManBoxFindFromCiNum's boundary convention. A
// primary input returns -1, an output before a later box returns the preceding
// box ID, and every output of the final box returns the native -2 sentinel.
func (m *TimManager) FindBoxFromCI(index int) (int, error) {
	if m == nil || index < 0 || index >= len(m.cis) {
		return -1, ErrDimension
	}
	if len(m.boxes) == 0 || index < m.boxes[0].FirstOutput {
		return -1, nil
	}
	for boxID, box := range m.boxes {
		if index < box.FirstOutput {
			return boxID - 1, nil
		}
	}
	return -2, nil
}
func (m *TimManager) SetBoxCopy(index, value int) error {
	if m == nil || index < 0 || index >= len(m.boxes) || int64(value) < -2147483648 || int64(value) > 2147483647 {
		return ErrDimension
	}
	m.boxes[index].Copy = value
	return nil
}
func (m *TimManager) PrimaryCounts() (int, int) {
	if m == nil {
		return 0, 0
	}
	if len(m.boxes) == 0 {
		return len(m.cis), len(m.cos)
	}
	last := m.boxes[len(m.boxes)-1]
	return m.boxes[0].FirstOutput, len(m.cos) - last.FirstInput - last.Inputs
}
func (m *TimManager) BlackBoxCounts() (int, int, int) {
	if m == nil {
		return 0, 0, 0
	}
	boxes, inputs, outputs := 0, 0, 0
	for _, b := range m.boxes {
		if b.Black {
			boxes++
			inputs += b.Inputs
			outputs += b.Outputs
		}
	}
	return boxes, inputs, outputs
}

// MarkBox labels all stored box input/output pins current or previous. At
// epoch zero a previous label is -1, matching native traversal arithmetic.
func (m *TimManager) MarkBox(index int, outputs, previous bool) error {
	b, err := m.Box(index)
	if err != nil {
		return err
	}
	epoch := m.epoch
	if previous {
		epoch--
	}
	if outputs {
		for i := 0; i < b.Outputs; i++ {
			m.cis[b.FirstOutput+i].TravID = epoch
		}
	} else {
		for i := 0; i < b.Inputs; i++ {
			m.cos[b.FirstInput+i].TravID = epoch
		}
	}
	return nil
}
