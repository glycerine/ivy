package abc

// Library-derived timing setup from misc/tim/timMan.c:Tim_ManCreate.

import "slices"

type TimLibraryBox struct {
	Inputs, Outputs int
	Delays          []float32
	Black           bool
}

// Configure initializes delay tables and optional primary constraints in one
// atomic operation. A nil library gives every box its own unit-delay table;
// with a library, nonnegative box table IDs index shared concrete entries and
// -1 boxes receive appended unit tables.
func (m *TimManager) Configure(library []TimLibraryBox, arrivals, required []float32, options TimDelayOptions) error {
	if m == nil || len(m.tables) != 0 || options.MaxTables < 0 || options.MaxValues < 0 {
		return ErrDimension
	}
	primaryCIs := make([]int, 0, len(m.cis))
	primaryCOs := make([]int, 0, len(m.cos))
	for i, terminal := range m.cis {
		if terminal.Box < 0 {
			primaryCIs = append(primaryCIs, i)
		}
	}
	for i, terminal := range m.cos {
		if terminal.Box < 0 {
			primaryCOs = append(primaryCOs, i)
		}
	}
	if arrivals != nil && len(arrivals) != len(primaryCIs) || required != nil && len(required) != len(primaryCOs) {
		return ErrDimension
	}
	if options.MaxTables == 0 {
		options.MaxTables = 1 << 20
	}
	if options.MaxValues == 0 {
		options.MaxValues = 1 << 20
	}
	if library != nil && len(library) > options.MaxTables {
		return ErrLimit
	}
	work := m.Clone()
	work.tables = nil
	work.tableValues = 0
	if library != nil {
		work.tables = make([]TimDelayTable, len(library))
	}
	for boxID := range work.boxes {
		box := &work.boxes[boxID]
		if library == nil || box.DelayTable == -1 {
			tableID := len(work.tables)
			count, err := checkedSize(0, box.Inputs, box.Outputs, options.MaxValues)
			if err != nil {
				return err
			}
			values := make([]float32, count)
			for i := range values {
				values[i] = 1
			}
			if err = work.SetDelayTable(tableID, TimDelayTable{Inputs: box.Inputs, Outputs: box.Outputs, Delays: values}, options); err != nil {
				return err
			}
			box.DelayTable = tableID
			continue
		}
		if box.DelayTable < 0 || box.DelayTable >= len(library) {
			return ErrDimension
		}
		entry := library[box.DelayTable]
		count, err := checkedSize(0, entry.Inputs, entry.Outputs, options.MaxValues)
		if err != nil {
			return err
		}
		if entry.Inputs != box.Inputs || entry.Outputs != box.Outputs || count != len(entry.Delays) {
			return ErrDimension
		}
		box.Black = entry.Black
		if work.tables[box.DelayTable].Inputs != 0 {
			continue
		}
		if err = work.SetDelayTable(box.DelayTable, TimDelayTable{Inputs: entry.Inputs, Outputs: entry.Outputs, Delays: entry.Delays}, options); err != nil {
			return err
		}
	}
	if arrivals != nil {
		for i, terminalID := range primaryCIs {
			work.cis[terminalID].Arrival = arrivals[i]
		}
	}
	if required != nil {
		for i, terminalID := range primaryCOs {
			work.cos[terminalID].Required = required[i]
		}
	}
	// Ensure the caller cannot retain a delay slice through the library.
	for i := range work.tables {
		work.tables[i].Delays = slices.Clone(work.tables[i].Delays)
	}
	*m = *work
	return nil
}
