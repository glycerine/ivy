package abc

// Insertion-based initial-state discovery from src/sat/bmc/bmcInse.c.

import "slices"

type GiaInsertionValue uint8

const (
	GiaInsertionZero GiaInsertionValue = iota
	GiaInsertionOne
	GiaInsertionUnknown
)

func (value GiaInsertionValue) valid() bool { return value <= GiaInsertionUnknown }

type GiaInsertionOptions struct {
	Frames, Words                       int
	MaxFrames, MaxWords, MaxEntries     int
	MaxInputValues, MaxSelectedPatterns int
	Stop                                func() bool
}

type GiaInsertionResult struct {
	Classes          []GiaInitialClass
	Inputs           []GiaInsertionValue
	Target           []GiaInsertionValue
	SelectedPatterns []int
	UnknownCounts    []int
}

func (result GiaInsertionResult) Clone() GiaInsertionResult {
	result.Classes = slices.Clone(result.Classes)
	result.Inputs = slices.Clone(result.Inputs)
	result.Target = slices.Clone(result.Target)
	result.SelectedPatterns = slices.Clone(result.SelectedPatterns)
	result.UnknownCounts = slices.Clone(result.UnknownCounts)
	return result
}

func normalizeGiaInsertionOptions(options GiaInsertionOptions) (GiaInsertionOptions, error) {
	if options.Frames < 0 || options.Words < 1 || options.MaxFrames < 0 || options.MaxWords < 0 || options.MaxEntries < 0 || options.MaxInputValues < 0 || options.MaxSelectedPatterns < 0 {
		return options, ErrDimension
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 16
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	if options.MaxInputValues == 0 {
		options.MaxInputValues = 16 << 20
	}
	if options.MaxSelectedPatterns == 0 {
		options.MaxSelectedPatterns = 1 << 20
	}
	if options.Frames > options.MaxFrames || options.Words > options.MaxWords || options.Frames > options.MaxSelectedPatterns {
		return options, ErrLimit
	}
	return options, nil
}

type giaInsertionSimulation struct {
	manager   *GiaManager
	words     int
	knownZero []uint64
	knownOne  []uint64
	stop      func() bool
	random    func() uint32
}

func newGiaInsertionSimulation(manager *GiaManager, options GiaInsertionOptions) (*giaInsertionSimulation, error) {
	entries, err := checkedSize(0, len(manager.objects), options.Words, options.MaxEntries/2)
	if err != nil {
		return nil, err
	}
	z, w := uint32(3716960521), uint32(2174103536)
	random := func() uint32 {
		z = 36969*(z&65535) + (z >> 16)
		w = 18000*(w&65535) + (w >> 16)
		return (z << 16) + w
	}
	random() // Gia_ManRandomW(1) resets before both identical 32-bit calls.
	return &giaInsertionSimulation{
		manager: manager, words: options.Words,
		knownZero: make([]uint64, entries), knownOne: make([]uint64, entries),
		stop: options.Stop, random: random,
	}, nil
}

func (simulation *giaInsertionSimulation) stopped() bool {
	return simulation.stop != nil && simulation.stop()
}

func (simulation *giaInsertionSimulation) span(id GiaID) ([]uint64, []uint64) {
	begin := int(id) * simulation.words
	return simulation.knownZero[begin : begin+simulation.words], simulation.knownOne[begin : begin+simulation.words]
}

func (simulation *giaInsertionSimulation) initialize(target []GiaInsertionValue) error {
	manager := simulation.manager
	if len(target) != manager.registers {
		return ErrDimension
	}
	for register, id := range manager.cos[manager.PrimaryOutputCount():] {
		value := target[register]
		if !value.valid() {
			return ErrDimension
		}
		zero, one := simulation.span(id)
		for word := 0; word < simulation.words; word++ {
			if simulation.stopped() {
				return ErrCancelled
			}
			switch value {
			case GiaInsertionZero:
				zero[word], one[word] = ^uint64(0), 0
			case GiaInsertionOne:
				zero[word], one[word] = 0, ^uint64(0)
			case GiaInsertionUnknown:
				zero[word], one[word] = 0, 0
			}
		}
	}
	return nil
}

func (simulation *giaInsertionSimulation) read(literal GiaLit) ([]uint64, []uint64) {
	zero, one := simulation.span(literal.ID())
	if literal.Complemented() {
		return one, zero
	}
	return zero, one
}

func (simulation *giaInsertionSimulation) simulateObject(id GiaID) error {
	manager, object := simulation.manager, simulation.manager.objects[id]
	zero, one := simulation.span(id)
	switch {
	case id == 0:
		for word := range zero {
			zero[word], one[word] = ^uint64(0), 0
		}
	case object.IsCI():
		ordinal := int(object.Diff1())
		if ordinal < manager.PrimaryInputCount() {
			for word := range zero {
				if simulation.stopped() {
					return ErrCancelled
				}
				value := uint64(simulation.random())<<32 | uint64(simulation.random())
				zero[word], one[word] = value, ^value
			}
			return nil
		}
		register := ordinal - manager.PrimaryInputCount()
		sourceZero, sourceOne := simulation.span(manager.cos[manager.PrimaryOutputCount()+register])
		copy(zero, sourceZero)
		copy(one, sourceOne)
	case object.IsAnd():
		left, err := object.Fanin0(id)
		if err != nil {
			return err
		}
		right, err := object.Fanin1(id)
		if err != nil {
			return err
		}
		leftZero, leftOne := simulation.read(left)
		rightZero, rightOne := simulation.read(right)
		for word := range zero {
			zero[word] = leftZero[word] | rightZero[word]
			one[word] = leftOne[word] & rightOne[word]
		}
	case object.IsCO():
		driver, err := object.Fanin0(id)
		if err != nil {
			return err
		}
		sourceZero, sourceOne := simulation.read(driver)
		copy(zero, sourceZero)
		copy(one, sourceOne)
	default:
		return ErrInvariant
	}
	return nil
}

func (simulation *giaInsertionSimulation) simulateRound() error {
	for id := range simulation.manager.objects {
		if simulation.stopped() {
			return ErrCancelled
		}
		if err := simulation.simulateObject(GiaID(id)); err != nil {
			return err
		}
	}
	return nil
}

func (simulation *giaInsertionSimulation) highestScore() (pattern, unknown int, err error) {
	patterns := 64 * simulation.words
	bestKnown := -1
	manager := simulation.manager
	for sample := 0; sample < patterns; sample++ {
		if simulation.stopped() {
			return 0, 0, ErrCancelled
		}
		known := 0
		for _, id := range manager.cos[manager.PrimaryOutputCount():] {
			zero, one := simulation.span(id)
			known += int((zero[sample/64] | one[sample/64]) >> uint(sample&63) & 1)
		}
		if known > bestKnown {
			bestKnown, pattern = known, sample
		}
	}
	return pattern, manager.registers - bestKnown, nil
}

func (simulation *giaInsertionSimulation) value(id GiaID, pattern int) (GiaInsertionValue, error) {
	zero, one := simulation.span(id)
	word, bit := pattern/64, uint(pattern&63)
	z, o := zero[word]>>bit&1 != 0, one[word]>>bit&1 != 0
	if z && o {
		return 0, ErrInvariant
	}
	if z {
		return GiaInsertionZero, nil
	}
	if o {
		return GiaInsertionOne, nil
	}
	return GiaInsertionUnknown, nil
}

func (simulation *giaInsertionSimulation) findStarting(pattern int) (target, inputs []GiaInsertionValue, err error) {
	manager := simulation.manager
	target = make([]GiaInsertionValue, manager.registers)
	inputs = make([]GiaInsertionValue, manager.PrimaryInputCount())
	for index, id := range manager.cos[manager.PrimaryOutputCount():] {
		target[index], err = simulation.value(id, pattern)
		if err != nil {
			return nil, nil, err
		}
	}
	for index, id := range manager.cis[:manager.PrimaryInputCount()] {
		inputs[index], err = simulation.value(id, pattern)
		if err != nil {
			return nil, nil, err
		}
	}
	return target, inputs, nil
}

// DiscoverInsertionInitialState implements Gia_ManInseInit,
// Gia_ManInseSimulateObj, Gia_ManInseHighestScore,
// Gia_ManInseFindStarting, Gia_ManInseSimulate, Gia_ManInsePerform and
// Gia_ManInseTest. A nil concrete initial state selects the all-zero state used
// by the source wrapper. Returned inputs are frame-major; Classes use 0/1 when
// the selected final ternary target remains unknown and 4/5 when it is fixed.
func (manager *GiaManager) DiscoverInsertionInitialState(initial []bool, options GiaInsertionOptions) (GiaInsertionResult, error) {
	fail := func(err error) (GiaInsertionResult, error) { return GiaInsertionResult{}, err }
	if manager == nil {
		return fail(ErrDimension)
	}
	if err := manager.Check(); err != nil {
		return fail(err)
	}
	options, err := normalizeGiaInsertionOptions(options)
	if err != nil {
		return fail(err)
	}
	if manager.registers == 0 {
		return fail(ErrDimension)
	}
	if initial == nil {
		initial = make([]bool, manager.registers)
	} else if len(initial) != manager.registers {
		return fail(ErrDimension)
	} else {
		initial = slices.Clone(initial)
	}
	for id, object := range manager.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return fail(ErrUnsupported)
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return fail(ErrUnsupported)
		}
	}
	inputValues, err := checkedSize(0, options.Frames, manager.PrimaryInputCount(), options.MaxInputValues)
	if err != nil {
		return fail(err)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	simulation, err := newGiaInsertionSimulation(manager, options)
	if err != nil {
		return fail(err)
	}
	target := make([]GiaInsertionValue, manager.registers)
	for index := range target {
		target[index] = GiaInsertionUnknown
	}
	if err = simulation.initialize(target); err != nil {
		return fail(err)
	}
	result := GiaInsertionResult{
		Inputs:           make([]GiaInsertionValue, 0, inputValues),
		SelectedPatterns: make([]int, 0, options.Frames),
		UnknownCounts:    make([]int, 0, options.Frames),
	}
	for frame := 0; frame < options.Frames; frame++ {
		if err = simulation.simulateRound(); err != nil {
			return fail(err)
		}
		pattern, unknown, scoreErr := simulation.highestScore()
		if scoreErr != nil {
			return fail(scoreErr)
		}
		var frameInputs []GiaInsertionValue
		target, frameInputs, err = simulation.findStarting(pattern)
		if err != nil {
			return fail(err)
		}
		result.SelectedPatterns = append(result.SelectedPatterns, pattern)
		result.UnknownCounts = append(result.UnknownCounts, unknown)
		result.Inputs = append(result.Inputs, frameInputs...)
		if err = simulation.initialize(target); err != nil {
			return fail(err)
		}
	}
	state := slices.Clone(initial)
	primary := make([]bool, manager.PrimaryInputCount())
	for frame := 0; frame < options.Frames; frame++ {
		for input := range primary {
			value := result.Inputs[frame*len(primary)+input]
			if value == GiaInsertionUnknown {
				return fail(ErrInvariant)
			}
			primary[input] = value == GiaInsertionOne
		}
		_, state, err = manager.Step(primary, state)
		if err != nil {
			return fail(err)
		}
	}
	result.Target = slices.Clone(target)
	result.Classes = make([]GiaInitialClass, manager.registers)
	for register, value := range state {
		if target[register] != GiaInsertionUnknown && (target[register] == GiaInsertionOne) != value {
			return fail(ErrInvariant)
		}
		if value {
			result.Classes[register] = GiaInitialKnownOne
		} else {
			result.Classes[register] = GiaInitialKnownZero
		}
		if target[register] != GiaInsertionUnknown {
			result.Classes[register] += 4
		}
	}
	return result, nil
}
