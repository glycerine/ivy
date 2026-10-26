package abc

// Counterexample analysis and optimization from src/sat/bmc/bmcCexTools.c.

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

type CounterexampleToolOptions struct {
	MaxObjects, MaxMappingEntries, MaxFrames, MaxBits int
	Stop                                              func() bool
}

type CounterexampleToolStatistics struct {
	Inputs, Registers, Frames                      int
	TotalInputBits, CareBits, DontCareBits         int
	EssentialBits, OptionalCareBits, MinimizedBits int
	DontCarePercent, OptionalPercent               float64
	EssentialPercent, MinimizedPercent             float64
}

type GiaAOGStatistics struct {
	Name                         string
	PrimaryInputs, Gates, Levels int
}

type CounterexampleAnalysis struct {
	States, Implications, Care, Essential, Minimized *Counterexample
	CareVerified, MinimizedVerified                  bool
	Statistics                                       CounterexampleToolStatistics
}

type GiaCareNodeStatistics struct {
	PatternCounts []int
	Average       float64
	GateCount     int
}

func normalizeCounterexampleToolOptions(options CounterexampleToolOptions) (CounterexampleToolOptions, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxFrames < 0 || options.MaxBits < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxBits == 0 {
		options.MaxBits = 64 << 20
	}
	return options, nil
}

func cexToolStopped(options CounterexampleToolOptions) bool {
	return options.Stop != nil && options.Stop()
}

func checkCounterexampleForGraph(graph *GiaManager, trace *Counterexample, options CounterexampleToolOptions) error {
	if graph == nil || trace == nil {
		return ErrDimension
	}
	if err := graph.Check(); err != nil {
		return err
	}
	if err := trace.Validate(); err != nil {
		return err
	}
	if trace.Registers() != graph.RegisterCount() || trace.Inputs() != graph.PrimaryInputCount() || trace.Property < 0 || trace.Property >= graph.PrimaryOutputCount() {
		return ErrDimension
	}
	if trace.Frames() > options.MaxFrames || trace.BitCount() > options.MaxBits || graph.ObjectCount() > options.MaxObjects || graph.ObjectCount() > options.MaxMappingEntries {
		return ErrLimit
	}
	return nil
}

// CounterexampleInputOneCount counts ones in the first realInputs input
// columns. Initial-register bits are intentionally excluded.
func CounterexampleInputOneCount(trace *Counterexample, realInputs int) (int, error) {
	if trace == nil {
		return 0, ErrDimension
	}
	if err := trace.Validate(); err != nil {
		return 0, err
	}
	if realInputs < 0 || realInputs > trace.Inputs() {
		return 0, ErrDimension
	}
	count := 0
	for frame := 0; frame < trace.Frames(); frame++ {
		for input := 0; input < realInputs; input++ {
			if value, _ := trace.Input(frame, input); value {
				count++
			}
		}
	}
	return count, nil
}

func BuildCounterexampleToolStatistics(graph *GiaManager, original, care, essential, minimized *Counterexample) (CounterexampleToolStatistics, error) {
	var result CounterexampleToolStatistics
	if graph == nil || original == nil || care == nil || essential == nil || minimized == nil {
		return result, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return result, err
	}
	for _, trace := range []*Counterexample{original, care, essential, minimized} {
		if err := trace.Validate(); err != nil {
			return result, err
		}
	}
	if original.Inputs() != graph.PrimaryInputCount() || original.Registers() != graph.RegisterCount() || care.Frames() != original.Frames() || essential.Frames() != original.Frames() || minimized.Frames() != original.Frames() {
		return result, ErrDimension
	}
	var err error
	result.Inputs, result.Registers, result.Frames = graph.PrimaryInputCount(), graph.RegisterCount(), original.Frames()
	result.TotalInputBits, err = checkedSize(0, result.Inputs, result.Frames, math.MaxInt)
	if err != nil {
		return result, err
	}
	if result.CareBits, err = CounterexampleInputOneCount(care, result.Inputs); err != nil {
		return result, err
	}
	if result.EssentialBits, err = CounterexampleInputOneCount(essential, result.Inputs); err != nil {
		return result, err
	}
	if result.MinimizedBits, err = CounterexampleInputOneCount(minimized, result.Inputs); err != nil {
		return result, err
	}
	result.DontCareBits = result.TotalInputBits - result.CareBits
	result.OptionalCareBits = result.CareBits - result.EssentialBits
	if result.TotalInputBits > 0 {
		denominator := float64(result.TotalInputBits)
		result.DontCarePercent = 100 * float64(result.DontCareBits) / denominator
		result.OptionalPercent = 100 * float64(result.OptionalCareBits) / denominator
		result.EssentialPercent = 100 * float64(result.EssentialBits) / denominator
		result.MinimizedPercent = 100 * float64(result.MinimizedBits) / denominator
	}
	return result, nil
}

func (m *GiaManager) CounterexampleAOGStatistics(maxObjects int) (GiaAOGStatistics, error) {
	var result GiaAOGStatistics
	if m == nil || maxObjects < 0 {
		return result, ErrDimension
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	levels, maximum, err := m.Levels(maxObjects)
	if err != nil {
		return result, err
	}
	_ = levels
	return GiaAOGStatistics{Name: m.Name, PrimaryInputs: m.PrimaryInputCount(), Gates: m.AndCount(), Levels: int(maximum)}, nil
}

// FormatCounterexampleInputs returns deterministic diagnostics instead of
// writing process-global stdout. Non-verbose mode returns only the summary.
func FormatCounterexampleInputs(trace *Counterexample, realInputs int, verbose bool) (string, error) {
	if trace == nil {
		return "", ErrDimension
	}
	statistics, err := trace.Statistics(realInputs)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString(FormatCounterexampleStatistics(statistics))
	builder.WriteByte('\n')
	if !verbose {
		return builder.String(), nil
	}
	for frame := 0; frame < trace.Frames(); frame++ {
		fmt.Fprintf(&builder, "%3d : ", frame)
		realOnes := 0
		for input := 0; input < realInputs; input++ {
			value, _ := trace.Input(frame, input)
			if value {
				realOnes++
				builder.WriteByte('1')
			} else {
				builder.WriteByte('0')
			}
		}
		fmt.Fprintf(&builder, " %3d ", realOnes)
		auxiliaryOnes := 0
		for input := realInputs; input < trace.Inputs(); input++ {
			value, _ := trace.Input(frame, input)
			if value {
				auxiliaryOnes++
				builder.WriteByte('1')
			} else {
				builder.WriteByte('0')
			}
		}
		fmt.Fprintf(&builder, " %3d\n", auxiliaryOnes)
	}
	return builder.String(), nil
}

// UnrollCounterexampleProperty builds the initialized transition unrolling
// through the trace's last frame and retains only its selected final property.
// Input values are not constrained, matching Bmc_CexPerformUnrolling.
func (m *GiaManager) UnrollCounterexampleProperty(trace *Counterexample, options CounterexampleToolOptions) (*GiaManager, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return nil, err
	}
	initial := make([]LatchInit, m.RegisterCount())
	for register := range initial {
		value, _ := trace.Initial(register)
		if value {
			initial[register] = LatchInitOne
		}
	}
	unrolled, err := m.UnrollWithOptions(GiaUnrollOptions{
		Frames: trace.Frames(), Initial: initial,
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		Stop: options.Stop,
	})
	if err != nil {
		return nil, err
	}
	index := trace.LastFrame()*m.PrimaryOutputCount() + trace.Property
	return unrolled.Graph().SelectOutputs([]int{index}, GiaCofactorOptions{
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		Stop: options.Stop,
	})
}

func cexFrameValuesAndImplications(m *GiaManager, ciValues, ciImplied []bool, options CounterexampleToolOptions) ([]bool, []bool, error) {
	if len(ciValues) != m.InputCount() || len(ciImplied) != m.InputCount() {
		return nil, nil, ErrDimension
	}
	values, err := m.evaluateValues(ciValues)
	if err != nil {
		return nil, nil, err
	}
	implied := make([]bool, m.ObjectCount())
	for index, id := range m.cis {
		implied[id] = ciImplied[index]
	}
	readValue := func(literal GiaLit) bool { return values[literal.ID()] != literal.Complemented() }
	for index, object := range m.objects {
		if cexToolStopped(options) {
			return nil, nil, ErrCancelled
		}
		id := GiaID(index)
		if object.IsCO() || object.IsBuffer() {
			fanin, _ := object.Fanin0(id)
			implied[id] = implied[fanin.ID()]
			continue
		}
		if !object.IsAnd() {
			continue
		}
		first, _ := object.Fanin0(id)
		second, _ := object.Fanin1(id)
		if control, mux := m.MuxControl(id); mux {
			controlValue := readValue(control)
			if implied[control.ID()] {
				if controlValue {
					implied[id] = implied[second.ID()]
				} else {
					implied[id] = implied[first.ID()]
				}
			} else {
				implied[id] = readValue(first) == readValue(second) && implied[first.ID()] && implied[second.ID()]
			}
		} else if object.IsXor() {
			implied[id] = implied[first.ID()] && implied[second.ID()]
		} else {
			left, right := readValue(first), readValue(second)
			switch {
			case left && right:
				implied[id] = implied[first.ID()] && implied[second.ID()]
			case !left && !right:
				implied[id] = implied[first.ID()] || implied[second.ID()]
			case !left:
				implied[id] = implied[first.ID()]
			default:
				implied[id] = implied[second.ID()]
			}
		}
	}
	return values, implied, nil
}

// CounterexampleInnerStates expands a PI trace into frame-major PI/RO values
// and a parallel mask of values implied by the initialized transition system.
func (m *GiaManager) CounterexampleInnerStates(trace *Counterexample, options CounterexampleToolOptions) (*Counterexample, *Counterexample, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, nil, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return nil, nil, err
	}
	bits, err := checkedSize(0, m.InputCount(), trace.Frames(), options.MaxBits)
	if err != nil {
		return nil, nil, err
	}
	_ = bits
	states, err := NewCounterexample(0, m.InputCount(), trace.Frames())
	if err != nil {
		return nil, nil, err
	}
	implications, err := NewCounterexample(0, m.InputCount(), trace.Frames())
	if err != nil {
		return nil, nil, err
	}
	states.Property, implications.Property = trace.Property, trace.Property
	state := make([]bool, m.RegisterCount())
	stateImplied := make([]bool, m.RegisterCount())
	for register := range state {
		state[register], _ = trace.Initial(register)
		stateImplied[register] = true
	}
	ciValues := make([]bool, m.InputCount())
	ciImplied := make([]bool, m.InputCount())
	for frame := 0; frame < trace.Frames(); frame++ {
		if cexToolStopped(options) {
			return nil, nil, ErrCancelled
		}
		for input := 0; input < m.PrimaryInputCount(); input++ {
			ciValues[input], _ = trace.Input(frame, input)
			ciImplied[input] = false
		}
		copy(ciValues[m.PrimaryInputCount():], state)
		copy(ciImplied[m.PrimaryInputCount():], stateImplied)
		for input := range ciValues {
			_ = states.SetInput(frame, input, ciValues[input])
			_ = implications.SetInput(frame, input, ciImplied[input])
		}
		values, implied, evalErr := cexFrameValuesAndImplications(m, ciValues, ciImplied, options)
		if evalErr != nil {
			return nil, nil, evalErr
		}
		if frame == trace.LastFrame() && !values[m.cos[trace.Property]] {
			return nil, nil, ErrInvariant
		}
		for register, id := range m.cos[m.PrimaryOutputCount():] {
			state[register] = values[id]
		}
		for register, id := range m.cos[m.PrimaryOutputCount():] {
			stateImplied[register] = implied[id]
		}
	}
	return states, implications, nil
}

func cexReadFrame(trace *Counterexample, frame int, values []bool) {
	for input := range values {
		values[input], _ = trace.Input(frame, input)
	}
}

func (m *GiaManager) counterexampleCareBits(states, implications, essential *Counterexample, findAll bool, options CounterexampleToolOptions) (*Counterexample, error) {
	if states == nil || states.Registers() != 0 || states.Inputs() != m.InputCount() || states.Property < 0 || states.Property >= m.PrimaryOutputCount() {
		return nil, ErrDimension
	}
	if err := states.Validate(); err != nil {
		return nil, err
	}
	for _, optional := range []*Counterexample{implications, essential} {
		if optional == nil {
			continue
		}
		if err := optional.Validate(); err != nil {
			return nil, err
		}
		if optional.Registers() != 0 || optional.Inputs() != states.Inputs() || optional.Frames() != states.Frames() || optional.Property != states.Property {
			return nil, ErrDimension
		}
	}
	if states.Frames() > options.MaxFrames || states.BitCount() > options.MaxBits || m.ObjectCount() > options.MaxObjects || m.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	result, err := NewCounterexample(0, m.InputCount(), states.Frames())
	if err != nil {
		return nil, err
	}
	result.Property = states.Property
	ciValues := make([]bool, m.InputCount())
	ciBlocked := make([]bool, m.InputCount())
	for frame := states.LastFrame(); frame >= 0; frame-- {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		cexReadFrame(states, frame, ciValues)
		for input := range ciBlocked {
			ciBlocked[input] = false
			if implications != nil {
				ciBlocked[input], _ = implications.Input(frame, input)
			}
			if essential != nil {
				value, _ := essential.Input(frame, input)
				ciBlocked[input] = ciBlocked[input] || value
			}
		}
		values, blocked, evalErr := cexFrameValuesAndImplications(m, ciValues, ciBlocked, options)
		if evalErr != nil {
			return nil, evalErr
		}
		marked := make([]bool, m.ObjectCount())
		stack := make([]GiaID, 0, m.AndCount()+m.InputCount())
		pushRoot := func(driver GiaLit) { stack = append(stack, driver.ID()) }
		if frame == states.LastFrame() {
			po := m.cos[states.Property]
			if !values[po] {
				return nil, ErrInvariant
			}
			driver, _ := m.objects[po].Fanin0(po)
			pushRoot(driver)
		} else {
			for register, ri := range m.cos[m.PrimaryOutputCount():] {
				care, _ := result.Input(frame+1, m.PrimaryInputCount()+register)
				if care {
					driver, _ := m.objects[ri].Fanin0(ri)
					pushRoot(driver)
				}
			}
		}
		read := func(literal GiaLit) bool { return values[literal.ID()] != literal.Complemented() }
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if id == 0 || blocked[id] || marked[id] {
				continue
			}
			marked[id] = true
			object := m.objects[id]
			if object.IsCI() {
				continue
			}
			if !object.IsAnd() {
				return nil, ErrInvariant
			}
			first, _ := object.Fanin0(id)
			if object.IsBuffer() {
				stack = append(stack, first.ID())
				continue
			}
			second, _ := object.Fanin1(id)
			if control, mux := m.MuxControl(id); mux {
				if read(control) {
					stack = append(stack, second.ID())
				} else {
					stack = append(stack, first.ID())
				}
				if read(first) != read(second) {
					stack = append(stack, control.ID())
				}
			} else if object.IsXor() {
				stack = append(stack, first.ID(), second.ID())
			} else {
				left, right := read(first), read(second)
				if left && right {
					stack = append(stack, first.ID(), second.ID())
				} else if findAll {
					if !left {
						stack = append(stack, first.ID())
					}
					if !right {
						stack = append(stack, second.ID())
					}
				} else if !left {
					stack = append(stack, first.ID())
				} else {
					stack = append(stack, second.ID())
				}
			}
		}
		for input, id := range m.cis {
			value := marked[id]
			if essential != nil {
				fixed, _ := essential.Input(frame, input)
				value = value || fixed
			}
			if implications != nil {
				implied, _ := implications.Input(frame, input)
				value = value && !implied
			}
			if value {
				_ = result.SetInput(frame, input, true)
			}
		}
	}
	return result, nil
}

// CounterexampleCareBits derives either all controlling justifications or one
// deterministic justification at each zero-valued AND.
func (m *GiaManager) CounterexampleCareBits(states, implications, essential *Counterexample, findAll bool, options CounterexampleToolOptions) (*Counterexample, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	return m.counterexampleCareBits(states, implications, essential, findAll, options)
}

// VerifyCounterexampleCare runs conservative ternary simulation: cared PIs use
// the concrete trace, uncared PIs are unknown, and the actual initialized state
// is retained. True means the selected final property is forced to one.
func (m *GiaManager) VerifyCounterexampleCare(trace, care *Counterexample, options CounterexampleToolOptions) (bool, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return false, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return false, err
	}
	if care == nil || care.Registers() != 0 || care.Frames() != trace.Frames() || care.Inputs() < m.PrimaryInputCount() || care.Property != trace.Property {
		return false, ErrDimension
	}
	if err = care.Validate(); err != nil {
		return false, err
	}
	state := make([]GiaTernary, m.RegisterCount())
	for register := range state {
		value, _ := trace.Initial(register)
		if value {
			state[register] = GiaTernaryOne
		} else {
			state[register] = GiaTernaryZero
		}
	}
	inputs := make([]GiaTernary, m.InputCount())
	for frame := 0; frame < trace.Frames(); frame++ {
		if cexToolStopped(options) {
			return false, ErrCancelled
		}
		for input := 0; input < m.PrimaryInputCount(); input++ {
			known, _ := care.Input(frame, input)
			if !known {
				inputs[input] = GiaTernaryUnknown
				continue
			}
			value, _ := trace.Input(frame, input)
			if value {
				inputs[input] = GiaTernaryOne
			} else {
				inputs[input] = GiaTernaryZero
			}
		}
		copy(inputs[m.PrimaryInputCount():], state)
		outputs, evalErr := m.EvaluateTernary(inputs, GiaCofactorOptions{
			MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
			Stop: options.Stop,
		})
		if evalErr != nil {
			return false, evalErr
		}
		if frame == trace.LastFrame() {
			return outputs[trace.Property] == GiaTernaryOne, nil
		}
		copy(state, outputs[m.PrimaryOutputCount():])
	}
	return false, ErrInvariant
}

// CounterexampleBitEssential checks whether flipping one PI value in an
// expanded PI/RO trace changes the selected final property from one to zero.
// State columns are not accepted: they are retained by EssentialBits.
func (m *GiaManager) CounterexampleBitEssential(states *Counterexample, bit int, options CounterexampleToolOptions) (bool, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return false, err
	}
	if m == nil || states == nil || states.Registers() != 0 || states.Inputs() != m.InputCount() || states.Property < 0 || states.Property >= m.PrimaryOutputCount() || bit < 0 || bit >= states.BitCount() {
		return false, ErrDimension
	}
	if err = states.Validate(); err != nil {
		return false, err
	}
	frame := bit / states.Inputs()
	column := bit % states.Inputs()
	if column >= m.PrimaryInputCount() || states.Frames() > options.MaxFrames || states.BitCount() > options.MaxBits || m.ObjectCount() > options.MaxObjects || m.ObjectCount() > options.MaxMappingEntries {
		if column >= m.PrimaryInputCount() {
			return false, ErrDimension
		}
		return false, ErrLimit
	}
	state := make([]bool, m.RegisterCount())
	for register := range state {
		state[register], _ = states.Input(frame, m.PrimaryInputCount()+register)
	}
	primary := make([]bool, m.PrimaryInputCount())
	for current := frame; current < states.Frames(); current++ {
		if cexToolStopped(options) {
			return false, ErrCancelled
		}
		for input := range primary {
			primary[input], _ = states.Input(current, input)
		}
		if current == frame {
			primary[column] = !primary[column]
		}
		outputs, next, stepErr := m.Step(primary, state)
		if stepErr != nil {
			return false, stepErr
		}
		if current == states.LastFrame() {
			return !outputs[states.Property], nil
		}
		state = next
	}
	return false, ErrInvariant
}

func (m *GiaManager) CounterexampleEssentialBits(states, care *Counterexample, options CounterexampleToolOptions) (*Counterexample, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil || states == nil || care == nil || states.Registers() != 0 || states.Inputs() != m.InputCount() || care.Registers() != 0 || care.Inputs() != states.Inputs() || care.Frames() != states.Frames() || care.Property != states.Property {
		return nil, ErrDimension
	}
	if err = states.Validate(); err != nil {
		return nil, err
	}
	if err = care.Validate(); err != nil {
		return nil, err
	}
	result, err := NewCounterexample(0, m.InputCount(), states.Frames())
	if err != nil {
		return nil, err
	}
	result.Property = states.Property
	for frame := 0; frame < states.Frames(); frame++ {
		for input := 0; input < m.InputCount(); input++ {
			cared, _ := care.Input(frame, input)
			if !cared {
				continue
			}
			if input >= m.PrimaryInputCount() {
				_ = result.SetInput(frame, input, true)
				continue
			}
			essential, essentialErr := m.CounterexampleBitEssential(states, frame*m.InputCount()+input, options)
			if essentialErr != nil {
				return nil, essentialErr
			}
			if essential {
				_ = result.SetInput(frame, input, true)
			}
		}
	}
	return result, nil
}

// AnalyzeCounterexample runs the complete inner-state, care, essential and
// minimized-care pipeline and verifies both published care masks.
func (m *GiaManager) AnalyzeCounterexample(trace *Counterexample, options CounterexampleToolOptions) (CounterexampleAnalysis, error) {
	var result CounterexampleAnalysis
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return result, err
	}
	result.States, result.Implications, err = m.CounterexampleInnerStates(trace, options)
	if err != nil {
		return result, err
	}
	result.Care, err = m.CounterexampleCareBits(result.States, result.Implications, nil, true, options)
	if err != nil {
		return result, err
	}
	result.CareVerified, err = m.VerifyCounterexampleCare(trace, result.Care, options)
	if err != nil {
		return result, err
	}
	if !result.CareVerified {
		return result, ErrInvariant
	}
	result.Essential, err = m.CounterexampleEssentialBits(result.States, result.Care, options)
	if err != nil {
		return result, err
	}
	result.Minimized, err = m.CounterexampleCareBits(result.States, result.Implications, result.Essential, false, options)
	if err != nil {
		return result, err
	}
	result.MinimizedVerified, err = m.VerifyCounterexampleCare(trace, result.Minimized, options)
	if err != nil {
		return result, err
	}
	if !result.MinimizedVerified {
		return result, ErrInvariant
	}
	result.Statistics, err = BuildCounterexampleToolStatistics(m, trace, result.Care, result.Essential, result.Minimized)
	return result, err
}

// BuildCounterexampleNetwork creates the source's monotone justification
// network. Extended gates are rejected because its AND-specific recurrence has
// no source-defined XOR/mux/buffer interpretation.
func (m *GiaManager) BuildCounterexampleNetwork(trace *Counterexample, options CounterexampleToolOptions) (*GiaManager, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return nil, err
	}
	for id, object := range m.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return nil, ErrUnsupported
		}
		if _, mux := m.MuxControl(GiaID(id)); mux {
			return nil, ErrUnsupported
		}
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = "unate"
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	observed := make([]bool, m.ObjectCount())
	implied := make([]bool, m.ObjectCount())
	formula := make([]GiaLit, m.ObjectCount())
	for i := range formula {
		formula[i] = InvalidGiaLit
	}
	implied[0] = true
	formula[0] = GiaConst1
	stateObserved := make([]bool, m.RegisterCount())
	stateImplied := make([]bool, m.RegisterCount())
	stateFormula := make([]GiaLit, m.RegisterCount())
	for register := range stateObserved {
		stateObserved[register], _ = trace.Initial(register)
		stateImplied[register] = true
		stateFormula[register] = GiaConst1
	}
	readObserved := func(literal GiaLit) bool { return observed[literal.ID()] != literal.Complemented() }
	for frame := 0; frame < trace.Frames(); frame++ {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		for input, id := range m.cis[:m.PrimaryInputCount()] {
			observed[id], _ = trace.Input(frame, input)
			implied[id] = false
			formula[id], err = target.AppendCI()
			if err != nil {
				return nil, err
			}
		}
		for register, id := range m.cis[m.PrimaryInputCount():] {
			observed[id] = stateObserved[register]
			implied[id] = stateImplied[register]
			formula[id] = stateFormula[register]
		}
		for index, object := range m.objects {
			if !object.IsAnd() {
				continue
			}
			id := GiaID(index)
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			left, right := readObserved(first), readObserved(second)
			observed[id] = left && right
			switch {
			case left && right:
				implied[id] = implied[first.ID()] && implied[second.ID()]
				if !implied[id] {
					formula[id], err = target.HashAnd(formula[first.ID()], formula[second.ID()])
				}
			case !left && !right:
				implied[id] = implied[first.ID()] || implied[second.ID()]
				if !implied[id] {
					formula[id], err = target.HashOr(formula[first.ID()], formula[second.ID()])
				}
			case !left:
				implied[id], formula[id] = implied[first.ID()], formula[first.ID()]
			default:
				implied[id], formula[id] = implied[second.ID()], formula[second.ID()]
			}
			if err != nil {
				return nil, err
			}
			if implied[id] {
				formula[id] = GiaConst1
			}
		}
		for _, id := range m.cos {
			driver, _ := m.objects[id].Fanin0(id)
			observed[id] = readObserved(driver)
			implied[id] = implied[driver.ID()]
			formula[id] = formula[driver.ID()]
		}
		for register, id := range m.cos[m.PrimaryOutputCount():] {
			stateObserved[register], stateImplied[register], stateFormula[register] = observed[id], implied[id], formula[id]
		}
	}
	po := m.cos[trace.Property]
	if !observed[po] || implied[po] || !formula[po].Valid() {
		return nil, ErrInvariant
	}
	if _, err = target.AppendCO(formula[po]); err != nil {
		return nil, err
	}
	target.HashStop()
	clean, _, err := target.Cleanup()
	return clean, err
}

// CareNodeStatistics computes the source's reverse controlling-path node count
// for each complete CI pattern without modifying graph marks.
func (m *GiaManager) CareNodeStatistics(patterns [][]bool, options CounterexampleToolOptions) (GiaCareNodeStatistics, error) {
	var result GiaCareNodeStatistics
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return result, err
	}
	if m == nil || len(patterns) == 0 {
		return result, ErrDimension
	}
	if err = m.Check(); err != nil {
		return result, err
	}
	if m.ObjectCount() > options.MaxObjects || m.ObjectCount() > options.MaxMappingEntries {
		return result, ErrLimit
	}
	result.GateCount = m.AndCount()
	result.PatternCounts = make([]int, len(patterns))
	for patternIndex, pattern := range patterns {
		if len(pattern) != m.InputCount() {
			return GiaCareNodeStatistics{}, ErrDimension
		}
		if cexToolStopped(options) {
			return GiaCareNodeStatistics{}, ErrCancelled
		}
		values, evalErr := m.evaluateValues(pattern)
		if evalErr != nil {
			return GiaCareNodeStatistics{}, evalErr
		}
		care := make([]bool, m.ObjectCount())
		for _, co := range m.cos {
			driver, _ := m.objects[co].Fanin0(co)
			care[driver.ID()] = true
		}
		read := func(literal GiaLit) bool { return values[literal.ID()] != literal.Complemented() }
		for index := len(m.objects) - 1; index > 0; index-- {
			id := GiaID(index)
			object := m.objects[id]
			if !object.IsAnd() || !care[id] {
				continue
			}
			first, _ := object.Fanin0(id)
			if object.IsBuffer() {
				care[first.ID()] = true
				continue
			}
			second, _ := object.Fanin1(id)
			if control, mux := m.MuxControl(id); mux {
				if read(control) {
					care[second.ID()] = true
				} else {
					care[first.ID()] = true
				}
				if read(first) != read(second) {
					care[control.ID()] = true
				}
			} else if object.IsXor() || values[id] {
				care[first.ID()], care[second.ID()] = true, true
			} else if !read(second) {
				care[second.ID()] = true
			} else if !read(first) {
				care[first.ID()] = true
			}
		}
		for id, object := range m.objects {
			if object.IsAnd() && care[id] {
				result.PatternCounts[patternIndex]++
			}
		}
		result.Average += float64(result.PatternCounts[patternIndex])
	}
	result.Average /= float64(len(patterns))
	return result, nil
}

// DecodeMNISTImages decodes the standard unsigned-byte image file into
// little-bit-order Boolean patterns. maxImages zero selects every image.
func DecodeMNISTImages(data []byte, maxImages, maxBits int) ([][]bool, error) {
	if maxImages < 0 || maxBits < 0 || len(data) < 16 || binary.BigEndian.Uint32(data[:4]) != 2051 {
		return nil, ErrDimension
	}
	images := uint64(binary.BigEndian.Uint32(data[4:8]))
	rows := uint64(binary.BigEndian.Uint32(data[8:12]))
	columns := uint64(binary.BigEndian.Uint32(data[12:16]))
	if rows == 0 || columns == 0 || rows > uint64(math.MaxInt) || columns > uint64(math.MaxInt) {
		return nil, ErrDimension
	}
	pixels := rows * columns
	if columns != 0 && pixels/columns != rows {
		return nil, ErrLimit
	}
	if pixels != 0 && images > (math.MaxUint64-16)/pixels {
		return nil, ErrLimit
	}
	bytesNeeded := uint64(16) + images*pixels
	if bytesNeeded != uint64(len(data)) {
		return nil, ErrDimension
	}
	if maxImages > 0 && images > uint64(maxImages) {
		images = uint64(maxImages)
	}
	bits := pixels * 8
	if bits > uint64(math.MaxInt) || bits != 0 && images > uint64(math.MaxInt)/bits {
		return nil, ErrLimit
	}
	totalBits := images * bits
	if maxBits > 0 && totalBits > uint64(maxBits) {
		return nil, ErrLimit
	}
	result := make([][]bool, int(images))
	for image := range result {
		result[image] = make([]bool, int(bits))
		base := 16 + image*int(pixels)
		for pixel := 0; pixel < int(pixels); pixel++ {
			for bit := 0; bit < 8; bit++ {
				result[image][pixel*8+bit] = data[base+pixel]>>uint(bit)&1 != 0
			}
		}
	}
	return result, nil
}
