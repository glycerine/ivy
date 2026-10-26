package abc

// Priority-selected counterexample reasons from src/sat/bmc/bmcCexMin1.c.

import "slices"

type AigCexReasonOptions struct {
	MaxObjects, MaxVisits, MaxFrames, MaxEntries int
	Stop                                         func() bool
	RegistersHighPriority                        bool
}

type AigCexPhasePriority struct {
	Priority int
	Value    bool
}

// AigCexReasonLiteral fixes one PI or register to Value. Index is local to the
// selected terminal class, matching the source's frame-local literals.
type AigCexReasonLiteral struct {
	Index int
	Value bool
}

type AigCexReasonAnalysis struct {
	PIReasons  bool
	FrameCIs   [][]AigID
	Priorities [][]AigCexPhasePriority
	Reasons    [][]AigCexReasonLiteral
	entries    [][]AigCexPhasePriority
	structure  []AigObject
	cis, cos   []AigID
	registers  int
}

type AigCexMinResult struct {
	Analysis  AigCexReasonAnalysis
	CubeGraph *AigManager
}

func normalizeAigCexReasonOptions(options AigCexReasonOptions) (AigCexReasonOptions, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxFrames < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 8 << 20
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 4 << 20
	}
	return options, nil
}

func aigCexStopped(options AigCexReasonOptions) bool {
	return options.Stop != nil && options.Stop()
}

func checkAigCounterexample(m *AigManager, trace *Counterexample, options AigCexReasonOptions) error {
	if m == nil || trace == nil {
		return ErrDimension
	}
	if err := m.Check(); err != nil {
		return err
	}
	if err := trace.Validate(); err != nil {
		return err
	}
	if trace.Registers() != m.RegisterCount() || trace.Inputs() != m.PrimaryInputCount() || trace.Property < 0 || trace.Property >= m.PrimaryOutputCount() {
		return ErrDimension
	}
	if trace.Frames() > options.MaxFrames || m.ObjectCount() > options.MaxObjects {
		return ErrLimit
	}
	return nil
}

func (m *AigManager) cexFrameRoots(trace *Counterexample, next []AigID) ([]AigID, error) {
	if next == nil {
		return []AigID{m.cos[trace.Property]}, nil
	}
	ciIndex := make([]int, len(m.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for index, id := range m.cis {
		ciIndex[id] = index
	}
	roots := make([]AigID, 0, len(next))
	for _, id := range next {
		if uint64(id) >= uint64(len(ciIndex)) || ciIndex[id] < m.PrimaryInputCount() {
			continue
		}
		register := ciIndex[id] - m.PrimaryInputCount()
		roots = append(roots, m.cos[m.PrimaryOutputCount()+register])
	}
	return roots, nil
}

func (m *AigManager) collectCexFrameCIs(roots []AigID, options AigCexReasonOptions, visits *int) ([]AigID, error) {
	seen := make([]bool, len(m.objects))
	result := make([]AigID, 0)
	for _, root := range roots {
		stack := []AigID{root}
		for len(stack) > 0 {
			if aigCexStopped(options) {
				return nil, ErrCancelled
			}
			if *visits == options.MaxVisits {
				return nil, ErrLimit
			}
			*visits++
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if uint64(id) >= uint64(len(m.objects)) || seen[id] {
				continue
			}
			seen[id] = true
			object := m.objects[id]
			switch object.Type {
			case AigConstant1:
			case AigCI:
				result = append(result, id)
			case AigCO, AigBuffer:
				stack = append(stack, object.Fanin0.ID())
			case AigAND:
				stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
			default:
				return nil, ErrInvariant
			}
		}
	}
	return result, nil
}

// CollectAigCounterexampleFrameTerms returns frame-major CI cones. Collection
// starts at the selected final PO and follows only next-frame ROs backward
// through their matching RIs, preserving native root and fanin order.
func (m *AigManager) CollectAigCounterexampleFrameTerms(trace *Counterexample, options AigCexReasonOptions) ([][]AigID, error) {
	options, err := normalizeAigCexReasonOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkAigCounterexample(m, trace, options); err != nil {
		return nil, err
	}
	entries, err := checkedSize(0, trace.Frames(), m.ObjectCount(), options.MaxEntries)
	if err != nil {
		return nil, err
	}
	_ = entries
	result := make([][]AigID, trace.Frames())
	visits := 0
	for frame := trace.LastFrame(); frame >= 0; frame-- {
		var next []AigID
		if frame < trace.LastFrame() {
			next = result[frame+1]
		}
		roots, rootErr := m.cexFrameRoots(trace, next)
		if rootErr != nil {
			return nil, rootErr
		}
		result[frame], err = m.collectCexFrameCIs(roots, options, &visits)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func aigReadPriority(entries []AigCexPhasePriority, literal AigLit) (AigCexPhasePriority, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(entries)) || entries[literal.ID()].Priority < 0 {
		return AigCexPhasePriority{}, ErrInvariant
	}
	result := entries[literal.ID()]
	result.Value = result.Value != literal.Complemented()
	return result, nil
}

func (m *AigManager) deriveCexPriorityRoots(roots []AigID, entries []AigCexPhasePriority, options AigCexReasonOptions, visits *int) error {
	type event struct {
		id   AigID
		exit bool
	}
	seen := make([]bool, len(m.objects))
	for _, root := range roots {
		stack := []event{{id: root}}
		for len(stack) > 0 {
			if aigCexStopped(options) {
				return ErrCancelled
			}
			if *visits == options.MaxVisits {
				return ErrLimit
			}
			*visits++
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			id := current.id
			if current.exit {
				object := m.objects[id]
				switch object.Type {
				case AigCO, AigBuffer:
					child, err := aigReadPriority(entries, object.Fanin0)
					if err != nil {
						return err
					}
					entries[id] = child
				case AigAND:
					left, err := aigReadPriority(entries, object.Fanin0)
					if err != nil {
						return err
					}
					right, err := aigReadPriority(entries, object.Fanin1)
					if err != nil {
						return err
					}
					result := AigCexPhasePriority{Value: left.Value && right.Value}
					switch {
					case left.Value && right.Value:
						result.Priority = min(left.Priority, right.Priority)
					case !left.Value && right.Value:
						result.Priority = left.Priority
					case left.Value && !right.Value:
						result.Priority = right.Priority
					default:
						result.Priority = max(left.Priority, right.Priority)
					}
					entries[id] = result
				}
				continue
			}
			if seen[id] || entries[id].Priority >= 0 {
				continue
			}
			seen[id] = true
			object := m.objects[id]
			switch object.Type {
			case AigConstant1, AigCI:
				return ErrInvariant
			case AigCO, AigBuffer:
				stack = append(stack, event{id: id, exit: true}, event{id: object.Fanin0.ID()})
			case AigAND:
				stack = append(stack, event{id: id, exit: true}, event{id: object.Fanin1.ID()}, event{id: object.Fanin0.ID()})
			default:
				return ErrInvariant
			}
		}
	}
	return nil
}

// DeriveAigCounterexamplePriorities returns entries aligned with frameTerms.
// The internal complete object tables are retained by Compute... for reason
// collection but never written into AIG iData scratch.
func (m *AigManager) deriveAigCounterexamplePriorities(trace *Counterexample, frameTerms [][]AigID, options AigCexReasonOptions) ([][]AigCexPhasePriority, [][]AigCexPhasePriority, error) {
	if len(frameTerms) != trace.Frames() {
		return nil, nil, ErrDimension
	}
	priorityBase := m.RegisterCount()
	if options.RegistersHighPriority {
		priorityBase = 0
	}
	priorityBound, err := checkedSize(priorityBase, trace.Frames(), m.PrimaryInputCount(), int(MaxLiteralID))
	if err != nil {
		return nil, nil, err
	}
	if options.RegistersHighPriority {
		priorityBound, err = checkedSize(priorityBound, 1, m.RegisterCount(), int(MaxLiteralID))
	}
	if err != nil {
		return nil, nil, err
	}
	if priorityBound == int(MaxLiteralID) {
		return nil, nil, ErrLimit
	}
	ciIndex := make([]int, len(m.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for index, id := range m.cis {
		ciIndex[id] = index
	}
	all := make([][]AigCexPhasePriority, trace.Frames())
	aligned := make([][]AigCexPhasePriority, trace.Frames())
	visits := 0
	for frame := 0; frame < trace.Frames(); frame++ {
		entries := make([]AigCexPhasePriority, len(m.objects))
		for i := range entries {
			entries[i].Priority = -1
		}
		entries[0] = AigCexPhasePriority{Priority: priorityBound, Value: true}
		piCount := 0
		aligned[frame] = make([]AigCexPhasePriority, len(frameTerms[frame]))
		for position, id := range frameTerms[frame] {
			index := ciIndex[id]
			if index < 0 {
				return nil, nil, ErrInvariant
			}
			var entry AigCexPhasePriority
			if index < m.PrimaryInputCount() {
				if options.RegistersHighPriority {
					entry.Priority = (frame+1)*m.PrimaryInputCount() - piCount
				} else {
					entry.Priority = m.RegisterCount() + (frame+1)*m.PrimaryInputCount() - 1 - piCount
				}
				entry.Value, _ = trace.Input(frame, index)
				piCount++
			} else {
				register := index - m.PrimaryInputCount()
				if frame == 0 {
					if options.RegistersHighPriority {
						entry.Priority = trace.Frames()*m.PrimaryInputCount() + register
					} else {
						entry.Priority = register
					}
					entry.Value, _ = trace.Initial(register)
				} else {
					entry = all[frame-1][m.cos[m.PrimaryOutputCount()+register]]
					if entry.Priority < 0 {
						return nil, nil, ErrInvariant
					}
				}
			}
			entries[id] = entry
			aligned[frame][position] = entry
		}
		var next []AigID
		if frame < trace.LastFrame() {
			next = frameTerms[frame+1]
		}
		roots, rootErr := m.cexFrameRoots(trace, next)
		if rootErr != nil {
			return nil, nil, rootErr
		}
		if err = m.deriveCexPriorityRoots(roots, entries, options, &visits); err != nil {
			return nil, nil, err
		}
		all[frame] = entries
	}
	selected := all[trace.LastFrame()][m.cos[trace.Property]]
	if selected.Priority < 0 || !selected.Value {
		return nil, nil, ErrInvariant
	}
	return aligned, all, nil
}

func (m *AigManager) DeriveAigCounterexamplePriorities(trace *Counterexample, frameTerms [][]AigID, options AigCexReasonOptions) ([][]AigCexPhasePriority, error) {
	options, err := normalizeAigCexReasonOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkAigCounterexample(m, trace, options); err != nil {
		return nil, err
	}
	aligned, _, err := m.deriveAigCounterexamplePriorities(trace, frameTerms, options)
	return aligned, err
}

func (m *AigManager) collectCexReasonRoots(roots []AigID, entries []AigCexPhasePriority, piReasons bool, options AigCexReasonOptions, visits *int) ([]AigCexReasonLiteral, error) {
	ciIndex := make([]int, len(m.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for index, id := range m.cis {
		ciIndex[id] = index
	}
	seen := make([]bool, len(m.objects))
	result := make([]AigCexReasonLiteral, 0)
	for _, root := range roots {
		stack := []AigID{root}
		for len(stack) > 0 {
			if aigCexStopped(options) {
				return nil, ErrCancelled
			}
			if *visits == options.MaxVisits {
				return nil, ErrLimit
			}
			*visits++
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			object := m.objects[id]
			switch object.Type {
			case AigConstant1:
			case AigCI:
				index := ciIndex[id]
				if index < 0 || entries[id].Priority < 0 {
					return nil, ErrInvariant
				}
				if piReasons && index < m.PrimaryInputCount() {
					result = append(result, AigCexReasonLiteral{Index: index, Value: entries[id].Value})
				} else if !piReasons && index >= m.PrimaryInputCount() {
					result = append(result, AigCexReasonLiteral{Index: index - m.PrimaryInputCount(), Value: entries[id].Value})
				}
			case AigCO, AigBuffer:
				stack = append(stack, object.Fanin0.ID())
			case AigAND:
				left, err := aigReadPriority(entries, object.Fanin0)
				if err != nil {
					return nil, err
				}
				right, err := aigReadPriority(entries, object.Fanin1)
				if err != nil {
					return nil, err
				}
				node := entries[id]
				if node.Value {
					if !left.Value || !right.Value {
						return nil, ErrInvariant
					}
					stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
				} else if !left.Value && right.Value {
					stack = append(stack, object.Fanin0.ID())
				} else if left.Value && !right.Value {
					stack = append(stack, object.Fanin1.ID())
				} else if left.Priority >= right.Priority {
					stack = append(stack, object.Fanin0.ID())
				} else {
					stack = append(stack, object.Fanin1.ID())
				}
			default:
				return nil, ErrInvariant
			}
		}
	}
	return result, nil
}

func (m *AigManager) collectAigCounterexampleReasons(trace *Counterexample, frameTerms [][]AigID, all [][]AigCexPhasePriority, piReasons bool, options AigCexReasonOptions) ([][]AigCexReasonLiteral, error) {
	if len(frameTerms) != trace.Frames() || len(all) != trace.Frames() {
		return nil, ErrDimension
	}
	result := make([][]AigCexReasonLiteral, trace.Frames())
	visits := 0
	for frame := trace.LastFrame(); frame >= 0; frame-- {
		var next []AigID
		if frame < trace.LastFrame() {
			next = frameTerms[frame+1]
		}
		roots, err := m.cexFrameRoots(trace, next)
		if err != nil {
			return nil, err
		}
		result[frame], err = m.collectCexReasonRoots(roots, all[frame], piReasons, options, &visits)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// ComputeAigCounterexampleReasons owns the complete frame-term,
// phase/priority and selected-reason pipeline.
func (m *AigManager) ComputeAigCounterexampleReasons(trace *Counterexample, piReasons bool, options AigCexReasonOptions) (AigCexReasonAnalysis, error) {
	var result AigCexReasonAnalysis
	options, err := normalizeAigCexReasonOptions(options)
	if err != nil {
		return result, err
	}
	if err = checkAigCounterexample(m, trace, options); err != nil {
		return result, err
	}
	result.FrameCIs, err = m.CollectAigCounterexampleFrameTerms(trace, options)
	if err != nil {
		return result, err
	}
	result.Priorities, result.entries, err = m.deriveAigCounterexamplePriorities(trace, result.FrameCIs, options)
	if err != nil {
		return result, err
	}
	result.Reasons, err = m.collectAigCounterexampleReasons(trace, result.FrameCIs, result.entries, piReasons, options)
	if err != nil {
		return result, err
	}
	result.PIReasons = piReasons
	result.structure = slices.Clone(m.objects)
	result.cis, result.cos = slices.Clone(m.cis), slices.Clone(m.cos)
	result.registers = m.registers
	return result, nil
}

func (m *AigManager) validateCexReasonAnalysis(analysis AigCexReasonAnalysis) error {
	if !slices.Equal(analysis.structure, m.objects) || !slices.Equal(analysis.cis, m.cis) || !slices.Equal(analysis.cos, m.cos) || analysis.registers != m.registers {
		return ErrInvariant
	}
	if len(analysis.Reasons) != len(analysis.FrameCIs) {
		return ErrDimension
	}
	return nil
}

// DuplicateAigWithRegisterCubes emits one primary output for every reason row
// after frame zero, followed by the original RIs, and preserves register count.
func (m *AigManager) DuplicateAigWithRegisterCubes(reasons [][]AigCexReasonLiteral, options AigCexReasonOptions) (*AigManager, error) {
	options, err := normalizeAigCexReasonOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil || len(reasons) == 0 || m.ConstraintCount() != 0 {
		return nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	if m.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	totalLiterals := 0
	for _, row := range reasons {
		for _, literal := range row {
			if literal.Index < 0 || literal.Index >= m.RegisterCount() {
				return nil, ErrDimension
			}
			totalLiterals++
		}
	}
	estimate, err := checkedSize(m.ObjectCount(), len(reasons)-1, totalLiterals, options.MaxObjects)
	if err != nil {
		return nil, err
	}
	if _, err = checkedSize(estimate, 1, m.RegisterCount(), options.MaxObjects); err != nil {
		return nil, err
	}
	target := NewAigManager()
	target.Name = m.Name
	mapping := make([]AigLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	mapLiteral := func(literal AigLit) AigLit { return mapping[literal.ID()].NotIf(literal.Complemented()) }
	for id, object := range m.objects {
		if aigCexStopped(options) {
			return nil, ErrCancelled
		}
		if id == 0 || object.Type == AigNone || object.Type == AigCO {
			continue
		}
		switch object.Type {
		case AigCI:
			mapping[id], err = target.CreateCI()
		case AigBuffer:
			mapping[id], err = target.CreateBuffer(mapLiteral(object.Fanin0))
		case AigAND:
			mapping[id], err = target.And(mapLiteral(object.Fanin0), mapLiteral(object.Fanin1))
		default:
			return nil, ErrUnsupported
		}
		if err != nil {
			return nil, err
		}
	}
	for frame := 1; frame < len(reasons); frame++ {
		cube := AigConst1
		for _, literal := range reasons[frame] {
			ri := m.cos[m.PrimaryOutputCount()+literal.Index]
			driver := mapLiteral(m.objects[ri].Fanin0).NotIf(!literal.Value)
			cube, err = target.And(cube, driver)
			if err != nil {
				return nil, err
			}
		}
		if _, err = target.CreateCO(cube); err != nil {
			return nil, err
		}
	}
	for _, ri := range m.cos[m.PrimaryOutputCount():] {
		if _, err = target.CreateCO(mapLiteral(m.objects[ri].Fanin0)); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(m.RegisterCount()); err != nil {
		return nil, err
	}
	target.Cleanup()
	if err = target.Check(); err != nil {
		return nil, err
	}
	return target, nil
}

// DuplicateAigWithReasonAnalysis validates structural provenance and accepts
// only register reasons before constructing the cube graph.
func (m *AigManager) DuplicateAigWithReasonAnalysis(analysis AigCexReasonAnalysis, options AigCexReasonOptions) (*AigManager, error) {
	if m == nil || analysis.PIReasons {
		return nil, ErrDimension
	}
	if err := m.validateCexReasonAnalysis(analysis); err != nil {
		return nil, err
	}
	return m.DuplicateAigWithRegisterCubes(analysis.Reasons, options)
}

// MinimizeAigCounterexampleReasons replaces the source's print-and-write
// driver. It returns the selected state reasons and the owned cube graph; like
// the C routine, it does not claim to have produced a reduced trace.
func (m *AigManager) MinimizeAigCounterexampleReasons(trace *Counterexample, options AigCexReasonOptions) (AigCexMinResult, error) {
	var result AigCexMinResult
	analysis, err := m.ComputeAigCounterexampleReasons(trace, false, options)
	if err != nil {
		return result, err
	}
	result.Analysis = analysis
	result.CubeGraph, err = m.DuplicateAigWithReasonAnalysis(analysis, options)
	return result, err
}
