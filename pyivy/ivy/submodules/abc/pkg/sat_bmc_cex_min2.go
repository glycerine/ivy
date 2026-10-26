package abc

// Unate counterexample-justification networks from src/sat/bmc/bmcCexMin2.c.

import "slices"

type GiaCexAnnotation struct {
	frames, objects, property               int
	inputs, outputs, registers, constraints int
	values, justified                       []bool
	structure                               []GiaObject
	muxes                                   []GiaLit
}

func (a GiaCexAnnotation) Frames() int   { return a.frames }
func (a GiaCexAnnotation) Objects() int  { return a.objects }
func (a GiaCexAnnotation) Property() int { return a.property }
func (a GiaCexAnnotation) Value(frame int, object GiaID) (bool, error) {
	if frame < 0 || frame >= a.frames || uint64(object) >= uint64(a.objects) || len(a.values) != a.frames*a.objects {
		return false, ErrDimension
	}
	return a.values[frame*a.objects+int(object)], nil
}
func (a GiaCexAnnotation) Justified(frame int, object GiaID) (bool, error) {
	if frame < 0 || frame >= a.frames || uint64(object) >= uint64(a.objects) || len(a.justified) != a.frames*a.objects {
		return false, ErrDimension
	}
	return a.justified[frame*a.objects+int(object)], nil
}
func (a GiaCexAnnotation) Clone() GiaCexAnnotation {
	a.values = slices.Clone(a.values)
	a.justified = slices.Clone(a.justified)
	a.structure = slices.Clone(a.structure)
	a.muxes = slices.Clone(a.muxes)
	return a
}

type CexUnateOptions struct {
	CounterexampleToolOptions
	StartFrame, RealPrimaryInputs int
	JustifyMaximum, UseAllObjects bool
}

type CexUnateResult struct {
	Annotation GiaCexAnnotation
	Frames     []int
	Networks   []*GiaManager
}

func checkCexMin2PlainGraph(m *GiaManager) error {
	if m == nil {
		return ErrDimension
	}
	if err := m.Check(); err != nil {
		return err
	}
	for id, object := range m.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return ErrUnsupported
		}
		if _, mux := m.MuxControl(GiaID(id)); mux {
			return ErrUnsupported
		}
	}
	return nil
}

// AnnotateCounterexampleUnrolling stores the concrete value of every object in
// every frame and a backward justification path for the selected final PO.
// justifyMaximum selects both controlling zero fanins instead of the source's
// deterministic fanin-zero choice.
func (m *GiaManager) AnnotateCounterexampleUnrolling(trace *Counterexample, justifyMaximum bool, options CounterexampleToolOptions) (GiaCexAnnotation, error) {
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
	totalEntries, err := checkedSize(0, 2, entries, options.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	if _, err = checkedSize(totalEntries, 1, m.ObjectCount(), options.MaxMappingEntries); err != nil {
		return result, err
	}
	result = GiaCexAnnotation{
		frames: trace.Frames(), objects: m.ObjectCount(), property: trace.Property,
		inputs: m.InputCount(), outputs: m.OutputCount(), registers: m.RegisterCount(), constraints: m.ConstraintCount(),
		values: make([]bool, entries), justified: make([]bool, entries),
		structure: slices.Clone(m.objects), muxes: slices.Clone(m.muxes),
	}
	state := make([]bool, m.RegisterCount())
	for register := range state {
		state[register], _ = trace.Initial(register)
	}
	inputs := make([]bool, m.InputCount())
	for frame := 0; frame < trace.Frames(); frame++ {
		if cexToolStopped(options) {
			return GiaCexAnnotation{}, ErrCancelled
		}
		for input := 0; input < m.PrimaryInputCount(); input++ {
			inputs[input], _ = trace.Input(frame, input)
		}
		copy(inputs[m.PrimaryInputCount():], state)
		values, evalErr := m.evaluateValues(inputs)
		if evalErr != nil {
			return GiaCexAnnotation{}, evalErr
		}
		copy(result.values[frame*m.ObjectCount():(frame+1)*m.ObjectCount()], values)
		for register, id := range m.cos[m.PrimaryOutputCount():] {
			state[register] = values[id]
		}
	}
	selected := m.cos[trace.Property]
	if !result.values[trace.LastFrame()*m.ObjectCount()+int(selected)] {
		return GiaCexAnnotation{}, ErrInvariant
	}
	result.justified[trace.LastFrame()*m.ObjectCount()+int(selected)] = true
	read := func(frame int, literal GiaLit) bool {
		return result.values[frame*m.ObjectCount()+int(literal.ID())] != literal.Complemented()
	}
	for frame := trace.LastFrame(); frame >= 0; frame-- {
		row := result.justified[frame*m.ObjectCount() : (frame+1)*m.ObjectCount()]
		for _, co := range m.cos {
			if !row[co] {
				continue
			}
			driver, _ := m.objects[co].Fanin0(co)
			row[driver.ID()] = true
		}
		for index := len(m.objects) - 1; index > 0; index-- {
			if cexToolStopped(options) {
				return GiaCexAnnotation{}, ErrCancelled
			}
			id := GiaID(index)
			object := m.objects[id]
			if !object.IsAnd() || !row[id] {
				continue
			}
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			left, right := read(frame, first), read(frame, second)
			if left == right {
				if left || justifyMaximum {
					row[first.ID()], row[second.ID()] = true, true
				} else {
					row[first.ID()] = true
				}
			} else if !left {
				row[first.ID()] = true
			} else {
				row[second.ID()] = true
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

func (m *GiaManager) validateCexAnnotation(trace *Counterexample, annotation GiaCexAnnotation) error {
	if trace == nil || annotation.frames != trace.Frames() || annotation.objects != m.ObjectCount() || annotation.property != trace.Property ||
		annotation.inputs != m.InputCount() || annotation.outputs != m.OutputCount() || annotation.registers != m.RegisterCount() || annotation.constraints != m.ConstraintCount() {
		return ErrDimension
	}
	entries := annotation.frames * annotation.objects
	if entries < 0 || len(annotation.values) != entries || len(annotation.justified) != entries {
		return ErrInvariant
	}
	if !slices.Equal(annotation.structure, m.objects) || !slices.Equal(annotation.muxes, m.muxes) {
		return ErrInvariant
	}
	return nil
}

// CreateCounterexampleUnate creates the monotone fixedness network from one
// start frame through the failing frame. realPrimaryInputs < 0 selects CEX
// minimization mode, whose variables are justified ROs in startFrame. A
// nonnegative value selects refinement mode, whose variables are justified PIs
// after the real-input prefix. useAllObjects retains unused candidate CIs.
func (m *GiaManager) CreateCounterexampleUnate(trace *Counterexample, annotation GiaCexAnnotation, startFrame, realPrimaryInputs int, useAllObjects bool, options CounterexampleToolOptions) (*GiaManager, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkCexMin2PlainGraph(m); err != nil {
		return nil, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return nil, err
	}
	if err = m.validateCexAnnotation(trace, annotation); err != nil {
		return nil, err
	}
	if startFrame < 0 || startFrame >= trace.Frames() || realPrimaryInputs >= m.PrimaryInputCount() {
		return nil, ErrDimension
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = "unate"
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	formula := make([]GiaLit, m.ObjectCount())
	row := func(frame int) []bool {
		return annotation.justified[frame*m.ObjectCount() : (frame+1)*m.ObjectCount()]
	}
	if realPrimaryInputs < 0 {
		path := row(startFrame)
		for _, ro := range m.cis[m.PrimaryInputCount():] {
			formula[ro] = GiaConst0
			if useAllObjects || path[ro] {
				candidate, appendErr := target.AppendCI()
				if appendErr != nil {
					return nil, appendErr
				}
				if path[ro] {
					formula[ro] = candidate
				}
			}
		}
	} else {
		path := row(startFrame)
		for _, ro := range m.cis[m.PrimaryInputCount():] {
			if path[ro] {
				formula[ro] = GiaConst1
			} else {
				formula[ro] = GiaConst0
			}
		}
	}
	for frame := startFrame; frame < trace.Frames(); frame++ {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		path := row(frame)
		if path[0] {
			formula[0] = GiaConst1
		} else {
			formula[0] = GiaConst0
		}
		for input, pi := range m.cis[:m.PrimaryInputCount()] {
			formula[pi] = GiaConst0
			if realPrimaryInputs < 0 || input < realPrimaryInputs {
				if path[pi] {
					formula[pi] = GiaConst1
				}
				continue
			}
			if useAllObjects || path[pi] {
				candidate, appendErr := target.AppendCI()
				if appendErr != nil {
					return nil, appendErr
				}
				if path[pi] {
					formula[pi] = candidate
				}
			}
		}
		values := annotation.values[frame*m.ObjectCount() : (frame+1)*m.ObjectCount()]
		for index, object := range m.objects {
			if !object.IsAnd() {
				continue
			}
			id := GiaID(index)
			formula[id] = GiaConst0
			if !path[id] {
				continue
			}
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			left, right := formula[first.ID()], formula[second.ID()]
			switch {
			case left != GiaConst0 && right != GiaConst0:
				if values[id] {
					formula[id], err = target.HashAnd(left, right)
				} else {
					formula[id], err = target.HashOr(left, right)
				}
			case left != GiaConst0:
				formula[id] = left
			case right != GiaConst0:
				formula[id] = right
			default:
				return nil, ErrInvariant
			}
			if err != nil {
				return nil, err
			}
		}
		for _, co := range m.cos {
			driver, _ := m.objects[co].Fanin0(co)
			formula[co] = formula[driver.ID()]
		}
		if frame < trace.LastFrame() {
			for register, ro := range m.cis[m.PrimaryInputCount():] {
				formula[ro] = formula[m.cos[m.PrimaryOutputCount()+register]]
			}
		}
	}
	selected := m.cos[trace.Property]
	if !annotation.justified[trace.LastFrame()*m.ObjectCount()+int(selected)] || formula[selected] == GiaConst0 {
		return nil, ErrInvariant
	}
	if _, err = target.AppendCO(formula[selected]); err != nil {
		return nil, err
	}
	target.HashStop()
	clean, _, err := target.Cleanup()
	return clean, err
}

// CounterexampleUnateNetworks is the useful owned replacement for
// Gia_ManCexMin, whose C implementation only prints graphs and returns NULL.
func (m *GiaManager) CounterexampleUnateNetworks(trace *Counterexample, options CexUnateOptions) (CexUnateResult, error) {
	var result CexUnateResult
	if m == nil {
		return result, ErrDimension
	}
	normalized, err := normalizeCounterexampleToolOptions(options.CounterexampleToolOptions)
	if err != nil {
		return result, err
	}
	if options.StartFrame < 0 || trace == nil || options.StartFrame >= trace.Frames() || options.RealPrimaryInputs >= m.PrimaryInputCount() {
		return result, ErrDimension
	}
	result.Annotation, err = m.AnnotateCounterexampleUnrolling(trace, options.JustifyMaximum, normalized)
	if err != nil {
		return result, err
	}
	if options.RealPrimaryInputs >= 0 {
		result.Frames = []int{options.StartFrame}
	} else {
		for frame := trace.LastFrame(); frame >= options.StartFrame; frame-- {
			result.Frames = append(result.Frames, frame)
		}
	}
	for _, frame := range result.Frames {
		if cexToolStopped(normalized) {
			return CexUnateResult{}, ErrCancelled
		}
		network, buildErr := m.CreateCounterexampleUnate(trace, result.Annotation, frame, options.RealPrimaryInputs, options.UseAllObjects, normalized)
		if buildErr != nil {
			return CexUnateResult{}, buildErr
		}
		result.Networks = append(result.Networks, network)
	}
	return result, nil
}
