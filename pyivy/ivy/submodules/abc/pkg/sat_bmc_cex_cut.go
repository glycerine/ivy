package abc

// Bad-state characterization from src/sat/bmc/bmcCexCut.c.

import "slices"

type CexTargetStateMode uint8

const (
	CexTargetJustifiedCube CexTargetStateMode = iota
	CexTargetTerminalCone
	CexTargetAllFrameCones
)

type CexTargetStateOptions struct {
	CounterexampleToolOptions
	Mode              CexTargetStateMode
	CombinationalOnly bool
}

type CexTargetStateResult struct {
	Annotation   GiaCexAnnotation
	InitialState []bool
	Frames       []int
	Property     *GiaManager
	Graph        *GiaManager
}

// AnnotateCounterexampleStateCut computes every unrolled value and the union
// of trace-implied nodes with the deterministic backward justification of the
// selected final property. Initial registers are variables; primary inputs and
// constants are fixed. This consolidates the recursive and iterative source
// marking routines without graph scratch or recursion.
func (m *GiaManager) AnnotateCounterexampleStateCut(trace *Counterexample, options CounterexampleToolOptions) (GiaCexAnnotation, error) {
	var result GiaCexAnnotation
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return result, err
	}
	if err = checkCexMin2PlainGraph(m); err != nil {
		return result, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return result, err
	}
	entries, err := checkedSize(0, trace.Frames(), m.ObjectCount(), options.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	if _, err = checkedSize(0, 2, entries, options.MaxMappingEntries); err != nil {
		return result, err
	}
	result = GiaCexAnnotation{
		frames: trace.Frames(), objects: m.ObjectCount(), property: trace.Property,
		inputs: m.InputCount(), outputs: m.OutputCount(), registers: m.RegisterCount(), constraints: m.ConstraintCount(),
		values: make([]bool, entries), justified: make([]bool, entries),
		structure: slices.Clone(m.objects), muxes: slices.Clone(m.muxes),
	}
	stateValues := make([]bool, m.RegisterCount())
	stateFixed := make([]bool, m.RegisterCount())
	for register := range stateValues {
		stateValues[register], _ = trace.Initial(register)
	}
	ciValues := make([]bool, m.InputCount())
	ciFixed := make([]bool, m.InputCount())
	for frame := 0; frame < trace.Frames(); frame++ {
		if cexToolStopped(options) {
			return GiaCexAnnotation{}, ErrCancelled
		}
		for input := 0; input < m.PrimaryInputCount(); input++ {
			ciValues[input], _ = trace.Input(frame, input)
			ciFixed[input] = true
		}
		copy(ciValues[m.PrimaryInputCount():], stateValues)
		copy(ciFixed[m.PrimaryInputCount():], stateFixed)
		values, fixed, evalErr := cexFrameValuesAndImplications(m, ciValues, ciFixed, options)
		if evalErr != nil {
			return GiaCexAnnotation{}, evalErr
		}
		begin := frame * m.ObjectCount()
		copy(result.values[begin:begin+m.ObjectCount()], values)
		copy(result.justified[begin:begin+m.ObjectCount()], fixed)
		for register, ri := range m.cos[m.PrimaryOutputCount():] {
			stateValues[register] = values[ri]
			stateFixed[register] = fixed[ri]
		}
	}
	selected := m.cos[trace.Property]
	last := trace.LastFrame() * m.ObjectCount()
	if !result.values[last+int(selected)] || result.justified[last+int(selected)] {
		return GiaCexAnnotation{}, ErrInvariant
	}
	result.justified[last+int(selected)] = true
	read := func(frame int, literal GiaLit) bool {
		return result.values[frame*m.ObjectCount()+int(literal.ID())] != literal.Complemented()
	}
	for frame := trace.LastFrame(); frame >= 0; frame-- {
		row := result.justified[frame*m.ObjectCount() : (frame+1)*m.ObjectCount()]
		for index := len(m.objects) - 1; index > 0; index-- {
			if cexToolStopped(options) {
				return GiaCexAnnotation{}, ErrCancelled
			}
			id := GiaID(index)
			if !row[id] {
				continue
			}
			object := m.objects[id]
			switch {
			case object.IsCO():
				driver, _ := object.Fanin0(id)
				row[driver.ID()] = true
			case object.IsAnd():
				first, _ := object.Fanin0(id)
				second, _ := object.Fanin1(id)
				left, right := read(frame, first), read(frame, second)
				if result.values[frame*m.ObjectCount()+index] {
					if !left || !right {
						return GiaCexAnnotation{}, ErrInvariant
					}
					row[first.ID()], row[second.ID()] = true, true
				} else if !left {
					row[first.ID()] = true
				} else if !right {
					row[second.ID()] = true
				} else {
					return GiaCexAnnotation{}, ErrInvariant
				}
			}
		}
		if frame == 0 {
			break
		}
		previous := result.justified[(frame-1)*m.ObjectCount() : frame*m.ObjectCount()]
		for register, ro := range m.cis[m.PrimaryInputCount():] {
			if row[ro] {
				previous[m.cos[m.PrimaryOutputCount()+register]] = true
			}
		}
	}
	return result, nil
}

func (m *GiaManager) cexStateAt(annotation GiaCexAnnotation, frame int) ([]bool, error) {
	if frame < 0 || frame >= annotation.frames || annotation.objects != m.ObjectCount() {
		return nil, ErrDimension
	}
	state := make([]bool, m.RegisterCount())
	for register, ro := range m.cis[m.PrimaryInputCount():] {
		state[register] = annotation.values[frame*m.ObjectCount()+int(ro)]
	}
	return state, nil
}

func (m *GiaManager) cexJustifiedStateCube(annotation GiaCexAnnotation, frame int, options CounterexampleToolOptions) (*GiaManager, error) {
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = m.Name
	cube := GiaConst1
	row := frame * m.ObjectCount()
	for _, ro := range m.cis[m.PrimaryInputCount():] {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		variable, appendErr := target.AppendCI()
		if appendErr != nil {
			return nil, appendErr
		}
		if !annotation.justified[row+int(ro)] {
			continue
		}
		cube, err = target.AppendAnd(cube, variable.NotIf(!annotation.values[row+int(ro)]))
		if err != nil {
			return nil, err
		}
	}
	if _, err = target.AppendCO(cube); err != nil {
		return nil, err
	}
	return target, target.Check()
}

func combineCounterexampleStateCones(cones []*GiaManager, name string, options CounterexampleToolOptions) (*GiaManager, error) {
	if len(cones) == 0 {
		return nil, ErrDimension
	}
	inputs := cones[0].InputCount()
	entries := 0
	for _, cone := range cones {
		if cone == nil || cone.RegisterCount() != 0 || cone.InputCount() != inputs || cone.OutputCount() != 1 {
			return nil, ErrDimension
		}
		if err := cone.Check(); err != nil {
			return nil, err
		}
		var err error
		entries, err = checkedSize(entries, 1, cone.ObjectCount(), options.MaxMappingEntries)
		if err != nil {
			return nil, err
		}
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = name
	shared := make([]GiaLit, inputs)
	for index := range shared {
		shared[index], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	for _, cone := range cones {
		mapping := make([]GiaLit, cone.ObjectCount())
		for index := range mapping {
			mapping[index] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		for index, ci := range cone.cis {
			mapping[ci] = shared[index]
		}
		for index, object := range cone.objects {
			if !object.IsAnd() {
				continue
			}
			if cexToolStopped(options) {
				return nil, ErrCancelled
			}
			mapping[index], err = hashGiaCopiedGate(cone, target, GiaID(index), mapping)
			if err != nil {
				return nil, err
			}
		}
		co := cone.cos[0]
		driver, _ := cone.objects[co].Fanin0(co)
		driver, err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
		if _, err = target.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	target.HashStop()
	clean, _, err := target.Cleanup()
	return clean, err
}

// CounterexampleTargetStates characterizes trace-induced bad states between
// begin and end. The result is either a justified cube, one terminal unate
// cone, or one cone for every terminal frame from end through the failure.
// Sequential mode replaces source properties and recodes InitialState to zero.
func (m *GiaManager) CounterexampleTargetStates(trace *Counterexample, begin, end int, options CexTargetStateOptions) (CexTargetStateResult, error) {
	var result CexTargetStateResult
	normalized, err := normalizeCounterexampleToolOptions(options.CounterexampleToolOptions)
	if err != nil {
		return result, err
	}
	options.CounterexampleToolOptions = normalized
	if options.Mode > CexTargetAllFrameCones || trace == nil || begin < 0 || end < 0 || begin >= end || end > trace.LastFrame() {
		return result, ErrDimension
	}
	if options.Mode == CexTargetJustifiedCube {
		result.Annotation, err = m.AnnotateCounterexampleStateCut(trace, normalized)
	} else {
		result.Annotation, err = m.AnnotateCounterexampleUnrolling(trace, false, normalized)
	}
	if err != nil {
		return CexTargetStateResult{}, err
	}
	result.InitialState, err = m.cexStateAt(result.Annotation, begin)
	if err != nil {
		return CexTargetStateResult{}, err
	}
	switch options.Mode {
	case CexTargetJustifiedCube:
		result.Frames = []int{end}
		result.Property, err = m.cexJustifiedStateCube(result.Annotation, end, normalized)
	case CexTargetTerminalCone:
		result.Frames = []int{end}
		states, _, stateErr := m.CounterexampleInnerStates(trace, normalized)
		if stateErr != nil {
			return CexTargetStateResult{}, stateErr
		}
		result.Property, err = m.BuildCounterexampleDepthCone(states, end, false, normalized)
	case CexTargetAllFrameCones:
		states, _, stateErr := m.CounterexampleInnerStates(trace, normalized)
		if stateErr != nil {
			return CexTargetStateResult{}, stateErr
		}
		cones := make([]*GiaManager, 0, trace.Frames()-end)
		for frame := end; frame < trace.Frames(); frame++ {
			cone, coneErr := m.BuildCounterexampleDepthCone(states, frame, false, normalized)
			if coneErr != nil {
				return CexTargetStateResult{}, coneErr
			}
			result.Frames = append(result.Frames, frame)
			cones = append(cones, cone)
		}
		result.Property, err = combineCounterexampleStateCones(cones, m.Name, normalized)
	}
	if err != nil {
		return CexTargetStateResult{}, err
	}
	if options.CombinationalOnly {
		return result, nil
	}
	result.Graph, err = m.WithStateProperties(result.Property, GiaCofactorOptions{
		MaxObjects: normalized.MaxObjects, MaxMappingEntries: normalized.MaxMappingEntries, Stop: normalized.Stop,
	})
	if err != nil {
		return CexTargetStateResult{}, err
	}
	result.Graph, err = result.Graph.FlipRegisters(result.InitialState, GiaCofactorOptions{
		MaxObjects: normalized.MaxObjects, MaxMappingEntries: normalized.MaxMappingEntries, Stop: normalized.Stop,
	})
	if err != nil {
		return CexTargetStateResult{}, err
	}
	return result, nil
}

// CounterexampleTargetStates converts an AIG through the validated concrete
// GIA path and converts the selected combinational or sequential result back.
func (m *AigManager) CounterexampleTargetStates(trace *Counterexample, begin, end int, options CexTargetStateOptions) (*AigManager, []bool, error) {
	gia, _, err := GiaFromAig(m)
	if err != nil {
		return nil, nil, err
	}
	result, err := gia.CounterexampleTargetStates(trace, begin, end, options)
	if err != nil {
		return nil, nil, err
	}
	selected := result.Graph
	if options.CombinationalOnly {
		selected = result.Property
	}
	aig, _, err := selected.ToAig()
	return aig, slices.Clone(result.InitialState), err
}
