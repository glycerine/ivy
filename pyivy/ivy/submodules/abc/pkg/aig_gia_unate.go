package abc

// Structural unateness from src/aig/gia/giaUnate.c.

type GiaUnateClass uint8

const (
	GiaUnateBinate GiaUnateClass = iota
	GiaUnateNegative
	GiaUnatePositive
	GiaUnateIndependent
)

type GiaUnateOptions struct {
	MaxObjects int
	MaxEntries int
	Stop       func() bool
}

type GiaUnateReport struct {
	Inputs, Outputs []int
	Classes         [][]GiaUnateClass // output-major, then input order
	SupportPairs    int
	UnatePairs      int
	BinatePairs     int
}

func checkGiaUnateOptions(options GiaUnateOptions) (GiaUnateOptions, error) {
	if options.MaxObjects < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 24
	}
	return options, nil
}

func giaFlipUnate(class GiaUnateClass) GiaUnateClass {
	if class == GiaUnateNegative || class == GiaUnatePositive {
		return class ^ 3
	}
	return class
}

// AnalyzeStructuralUnateness computes the source's path-polarity analysis for
// selected CI and CO ordinals. Nil selections mean all terminals; empty but
// non-nil selections mean none. Selection order and duplicates are preserved.
// This is deliberately structural: reconvergent cancellation can make a
// function independent even when both path polarities are reported.
func (m *GiaManager) AnalyzeStructuralUnateness(inputIDs, outputIDs []int, options GiaUnateOptions) (GiaUnateReport, error) {
	options, err := checkGiaUnateOptions(options)
	if err != nil {
		return GiaUnateReport{}, err
	}
	if m == nil {
		return GiaUnateReport{}, ErrDimension
	}
	if err = m.Check(); err != nil {
		return GiaUnateReport{}, err
	}
	if len(m.objects) > options.MaxObjects {
		return GiaUnateReport{}, ErrLimit
	}
	// Packed native XOR/mux nodes satisfy Gia_ObjIsAnd but the source routine
	// would interpret them as two-input ANDs and ignore mux controls.
	if m.XorCount() != 0 || m.MuxCount() != 0 {
		return GiaUnateReport{}, ErrUnsupported
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return GiaUnateReport{}, ErrCancelled
	}
	inputs := inputIDs
	if inputIDs == nil {
		inputs = make([]int, len(m.cis))
		for i := range inputs {
			inputs[i] = i
		}
	} else {
		inputs = append([]int(nil), inputIDs...)
	}
	outputs := outputIDs
	if outputIDs == nil {
		outputs = make([]int, len(m.cos))
		for i := range outputs {
			outputs[i] = i
		}
	} else {
		outputs = append([]int(nil), outputIDs...)
	}
	for _, input := range inputs {
		if input < 0 || input >= len(m.cis) {
			return GiaUnateReport{}, ErrDimension
		}
	}
	for _, output := range outputs {
		if output < 0 || output >= len(m.cos) {
			return GiaUnateReport{}, ErrDimension
		}
	}
	entries, err := checkedSize(0, len(m.objects), len(inputs), options.MaxEntries)
	if err != nil {
		return GiaUnateReport{}, err
	}
	if _, err = checkedSize(entries, len(outputs), len(inputs), options.MaxEntries); err != nil {
		return GiaUnateReport{}, err
	}
	states := make([]GiaUnateClass, entries)
	for i := range states {
		states[i] = GiaUnateIndependent
	}
	width := len(inputs)
	for column, input := range inputs {
		states[int(m.cis[input])*width+column] = GiaUnatePositive
	}
	for i, object := range m.objects {
		if stopped() {
			return GiaUnateReport{}, ErrCancelled
		}
		if !object.IsAnd() {
			continue
		}
		id := GiaID(i)
		first, firstErr := object.Fanin0(id)
		if firstErr != nil {
			return GiaUnateReport{}, firstErr
		}
		second, secondErr := object.Fanin1(id)
		if secondErr != nil {
			return GiaUnateReport{}, secondErr
		}
		for column := 0; column < width; column++ {
			left := states[int(first.ID())*width+column]
			right := states[int(second.ID())*width+column]
			if first.Complemented() {
				left = giaFlipUnate(left)
			}
			if second.Complemented() {
				right = giaFlipUnate(right)
			}
			states[i*width+column] = left & right
		}
	}
	result := GiaUnateReport{Inputs: inputs, Outputs: outputs, Classes: make([][]GiaUnateClass, len(outputs))}
	for row, output := range outputs {
		if stopped() {
			return GiaUnateReport{}, ErrCancelled
		}
		co := m.cos[output]
		driver, driverErr := m.objects[co].Fanin0(co)
		if driverErr != nil {
			return GiaUnateReport{}, driverErr
		}
		classes := make([]GiaUnateClass, width)
		for column := range classes {
			class := states[int(driver.ID())*width+column]
			if driver.Complemented() {
				class = giaFlipUnate(class)
			}
			classes[column] = class
			switch class {
			case GiaUnateBinate:
				result.SupportPairs++
				result.BinatePairs++
			case GiaUnateNegative, GiaUnatePositive:
				result.SupportPairs++
				result.UnatePairs++
			}
		}
		result.Classes[row] = classes
	}
	if stopped() {
		return GiaUnateReport{}, ErrCancelled
	}
	return result, nil
}

// StructuralUnate is the one-CI/one-CO form of the analysis.
func (m *GiaManager) StructuralUnate(inputID, outputID int, options GiaUnateOptions) (GiaUnateClass, error) {
	report, err := m.AnalyzeStructuralUnateness([]int{inputID}, []int{outputID}, options)
	if err != nil {
		return GiaUnateBinate, err
	}
	return report.Classes[0][0], nil
}
