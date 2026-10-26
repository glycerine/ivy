package abc

// Owned delay tables from misc/tim/timBox.c:Tim_ManBoxDelayTable and the
// table layout in timMan.c. Arrival/required propagation from timTime.c.

import (
	"math"
	"slices"
)

const TimNoArc float32 = -1000000000

// Delays is output-major: Delays[output*Inputs+input]. TimNoArc disables an arc.
type TimDelayTable struct {
	Inputs, Outputs int
	Delays          []float32
}
type TimDelayOptions struct{ MaxTables, MaxValues int } // zero: one million; table limit includes sparse indices

func (m *TimManager) DelayTableCount() int {
	if m == nil {
		return 0
	}
	return len(m.tables)
}

// SetDelayTable owns the supplied values and preserves native epoch caching.
// Advance traversal to invalidate previously computed box outputs after changes.
func (m *TimManager) SetDelayTable(index int, table TimDelayTable, options TimDelayOptions) error {
	if m == nil || index < 0 || table.Inputs < 1 || table.Outputs < 1 || options.MaxTables < 0 || options.MaxValues < 0 {
		return ErrDimension
	}
	if options.MaxTables == 0 {
		options.MaxTables = 1 << 20
	}
	if options.MaxValues == 0 {
		options.MaxValues = 1 << 20
	}
	if index >= options.MaxTables || len(m.tables) > options.MaxTables {
		return ErrLimit
	}
	count, err := checkedSize(0, table.Inputs, table.Outputs, options.MaxValues)
	if err != nil {
		return err
	}
	if count != len(table.Delays) {
		return ErrDimension
	}
	old := 0
	if index < len(m.tables) {
		old = len(m.tables[index].Delays)
	}
	total, err := checkedSize(m.tableValues-old, 1, count, options.MaxValues)
	if err != nil {
		return err
	}
	table.Delays = slices.Clone(table.Delays)
	if index >= len(m.tables) {
		m.tables = append(m.tables, make([]TimDelayTable, index+1-len(m.tables))...)
	}
	m.tables[index] = table
	m.tableValues = total
	return nil
}
func (m *TimManager) DelayTable(index int) (TimDelayTable, error) {
	if m == nil || index < 0 || index >= len(m.tables) || m.tables[index].Inputs == 0 {
		return TimDelayTable{}, ErrDimension
	}
	table := m.tables[index]
	table.Delays = slices.Clone(table.Delays)
	return table, nil
}
func (m *TimManager) boxTable(index int) (*TimDelayTable, error) {
	if m == nil || index < 0 || index >= len(m.boxes) {
		return nil, ErrDimension
	}
	box := m.boxes[index]
	if box.DelayTable < 0 || box.DelayTable >= len(m.tables) || m.tables[box.DelayTable].Inputs == 0 {
		return nil, ErrUnsupported
	}
	table := &m.tables[box.DelayTable]
	if table.Inputs != box.Inputs || table.Outputs != box.Outputs {
		return nil, ErrDimension
	}
	return table, nil
}
func (m *TimManager) BoxDelayTable(index int) (TimDelayTable, error) {
	table, err := m.boxTable(index)
	if err != nil {
		return TimDelayTable{}, err
	}
	result := *table
	result.Delays = slices.Clone(table.Delays)
	return result, nil
}
func timArrivalInteger(value float32) (int32, error) {
	v := float64(value)
	if math.IsNaN(v) || v < -2147483648 || v >= 2147483648 {
		return 0, ErrDimension
	}
	return int32(value), nil
}

// boxArrival preserves C's Abc_MaxInt conversion of float sums (truncation
// toward zero). Undefined nonfinite/out-of-int32 conversions reject atomically.
// Stored inputs are used even if their traversal marks are stale, as in C,
// without printing diagnostics. Advance the epoch to invalidate cached outputs.
func (m *TimManager) boxArrival(index int) (float32, error) {
	boxID := m.cis[index].Box
	table, err := m.boxTable(boxID)
	if err != nil {
		return 0, err
	}
	box := m.boxes[boxID]
	values := make([]float32, box.Outputs)
	for output := range values {
		best := int32(-1000000000)
		for input := 0; input < box.Inputs; input++ {
			delay := table.Delays[output*box.Inputs+input]
			if delay == TimNoArc {
				continue
			}
			candidate := float32(m.cos[box.FirstInput+input].Arrival + delay)
			value, err := timArrivalInteger(candidate)
			if err != nil {
				return 0, err
			}
			if value > best {
				best = value
			}
		}
		values[output] = float32(best)
	}
	m.boxes[boxID].TravID = m.epoch
	for output, value := range values {
		p := &m.cis[box.FirstOutput+output]
		p.Arrival = value
		p.TravID = m.epoch
	}
	return m.cis[index].Arrival, nil
}

// boxRequired uses the actual [output,input] disabled-arc guard. Native C
// accidentally tests [output,output], potentially reading past a row. All
// other defined float32 minimum/caching/mark behavior is preserved. Stored
// output requirements are consumed even when their traversal labels are stale.
func (m *TimManager) boxRequired(index int) (float32, error) {
	boxID := m.cos[index].Box
	table, err := m.boxTable(boxID)
	if err != nil {
		return 0, err
	}
	box := m.boxes[boxID]
	values := make([]float32, box.Inputs)
	for input := range values {
		best := TimEternity
		for output := 0; output < box.Outputs; output++ {
			delay := table.Delays[output*box.Inputs+input]
			if delay == TimNoArc {
				continue
			}
			candidate := float32(m.cis[box.FirstOutput+output].Required - delay)
			// Preserve Abc_MinFloat's comparison/selection, including IEEE ties/NaNs.
			if !(best < candidate) {
				best = candidate
			}
		}
		values[input] = best
	}
	m.boxes[boxID].TravID = m.epoch
	for input, value := range values {
		p := &m.cos[box.FirstInput+input]
		p.Required = value
		p.TravID = m.epoch
	}
	return m.cos[index].Required, nil
}
