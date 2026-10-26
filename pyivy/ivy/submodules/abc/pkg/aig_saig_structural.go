package abc

// Ported from ABC src/aig/saig/saig.h, saigCone.c, and saigDup.c.

import (
	"fmt"
	"slices"
)

// SaigMiterStatus is the owned counterpart of Sec_MtrStatus_t.
type SaigMiterStatus struct {
	Inputs, Nodes, Outputs int
	Unsatisfiable          int
	Satisfiable            int
	Undecided              int
	SatisfiedOutput        int
}

// SaigBDDParameters records the public BBR parameters declared by saig.h.
// The CUDD-backed verification engine is outside this structural source slice.
type SaigBDDParameters struct {
	TimeLimitSeconds  int
	MaximumBDDNodes   int
	MaximumIterations int
	Partition         bool
	Reorder           bool
	ReorderImage      bool
	Verbose           bool
	Silent            bool
	SkipOutputCheck   bool
	ExploredFrame     int
}

type SaigStructuralOptions struct {
	MaxObjects int
	MaxFrames  int
	Stop       func() bool
}

type saigStructuralBudget struct {
	options SaigStructuralOptions
}

func newSaigStructuralBudget(graph *AigManager, options SaigStructuralOptions) (*saigStructuralBudget, error) {
	if graph == nil || options.MaxObjects < 0 || options.MaxFrames < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 24
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if len(graph.objects) > options.MaxObjects {
		return nil, ErrLimit
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return &saigStructuralBudget{options: options}, nil
}

func (budget *saigStructuralBudget) step() error {
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (budget *saigStructuralBudget) checkTarget(target *AigManager) error {
	if len(target.objects) > budget.options.MaxObjects {
		return ErrLimit
	}
	return budget.step()
}

type SaigObjectRole uint8

const (
	SaigObjectOther SaigObjectRole = iota
	SaigObjectPrimaryInput
	SaigObjectPrimaryOutput
	SaigObjectLatchOutput
	SaigObjectLatchInput
)

func (graph *AigManager) SequentialObjectRole(id AigID) SaigObjectRole {
	if graph == nil {
		return SaigObjectOther
	}
	for index, candidate := range graph.cis {
		if candidate != id {
			continue
		}
		if index < graph.PrimaryInputCount() {
			return SaigObjectPrimaryInput
		}
		return SaigObjectLatchOutput
	}
	for index, candidate := range graph.cos {
		if candidate != id {
			continue
		}
		if index < graph.PrimaryOutputCount() {
			return SaigObjectPrimaryOutput
		}
		return SaigObjectLatchInput
	}
	return SaigObjectOther
}

func (graph *AigManager) PrimaryInputs() []AigID {
	if graph == nil {
		return nil
	}
	return slices.Clone(graph.cis[:graph.PrimaryInputCount()])
}

func (graph *AigManager) PrimaryOutputs() []AigID {
	if graph == nil {
		return nil
	}
	return slices.Clone(graph.cos[:graph.PrimaryOutputCount()])
}

func (graph *AigManager) LatchOutputs() []AigID {
	if graph == nil {
		return nil
	}
	return slices.Clone(graph.cis[graph.PrimaryInputCount():])
}

func (graph *AigManager) LatchInputs() []AigID {
	if graph == nil {
		return nil
	}
	return slices.Clone(graph.cos[graph.PrimaryOutputCount():])
}

func (graph *AigManager) LatchInputForOutput(id AigID) (AigID, error) {
	if graph == nil {
		return InvalidAigID, ErrDimension
	}
	for index, candidate := range graph.cis[graph.PrimaryInputCount():] {
		if candidate == id {
			return graph.cos[graph.PrimaryOutputCount()+index], nil
		}
	}
	return InvalidAigID, ErrDimension
}

func (graph *AigManager) LatchOutputForInput(id AigID) (AigID, error) {
	if graph == nil {
		return InvalidAigID, ErrDimension
	}
	for index, candidate := range graph.cos[graph.PrimaryOutputCount():] {
		if candidate == id {
			return graph.cis[graph.PrimaryInputCount()+index], nil
		}
	}
	return InvalidAigID, ErrDimension
}

func saigMappedLiteral(mapping []AigLit, literal AigLit) (AigLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	mapped := mapping[literal.ID()]
	if mapped == InvalidAigLit {
		return InvalidAigLit, ErrInvariant
	}
	return mapped.NotIf(literal.Complemented()), nil
}

func saigNewMapping(source *AigManager) []AigLit {
	mapping := make([]AigLit, len(source.objects))
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	mapping[0] = AigConst1
	return mapping
}

func saigCopyLogic(source, target *AigManager, mapping []AigLit, needed []bool, budget *saigStructuralBudget) error {
	for id := 1; id < len(source.objects); id++ {
		if needed != nil && !needed[id] {
			continue
		}
		object := source.objects[id]
		if object.Type == AigNone || object.Type == AigCI || object.Type == AigCO {
			continue
		}
		if err := budget.step(); err != nil {
			return err
		}
		first, err := saigMappedLiteral(mapping, object.Fanin0)
		if err != nil {
			return err
		}
		var result AigLit
		switch object.Type {
		case AigAND:
			second, mapErr := saigMappedLiteral(mapping, object.Fanin1)
			if mapErr != nil {
				return mapErr
			}
			result, err = target.And(first, second)
		case AigBuffer:
			result, err = target.CreateBuffer(first)
		default:
			return ErrUnsupported
		}
		if err != nil {
			return err
		}
		mapping[id] = result
		if err := budget.checkTarget(target); err != nil {
			return err
		}
	}
	return nil
}

func saigCopyAllInputs(source, target *AigManager, mapping []AigLit, budget *saigStructuralBudget) error {
	for _, id := range source.cis {
		literal, err := target.CreateCI()
		if err != nil {
			return err
		}
		mapping[id] = literal
		if err := budget.checkTarget(target); err != nil {
			return err
		}
	}
	return nil
}

func saigCreateMappedOutput(source, target *AigManager, mapping []AigLit, co AigID, budget *saigStructuralBudget) error {
	driver, err := saigMappedLiteral(mapping, source.objects[co].Fanin0)
	if err != nil {
		return err
	}
	if _, err = target.CreateCO(driver); err != nil {
		return err
	}
	return budget.checkTarget(target)
}

// DuplicateOrPrimaryOutputs implements Saig_ManDupOrpos and leaves latch
// transition outputs after the single ORed property output.
func (source *AigManager) DuplicateOrPrimaryOutputs(options SaigStructuralOptions) (*AigManager, error) {
	budget, err := newSaigStructuralBudget(source, options)
	if err != nil {
		return nil, err
	}
	if source.constraints != 0 {
		return nil, ErrDimension
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := saigNewMapping(source)
	if err = saigCopyAllInputs(source, target, mapping, budget); err != nil {
		return nil, err
	}
	if err = saigCopyLogic(source, target, mapping, nil, budget); err != nil {
		return nil, err
	}
	merged := AigConst0
	for _, co := range source.cos[:source.PrimaryOutputCount()] {
		driver, mapErr := saigMappedLiteral(mapping, source.objects[co].Fanin0)
		if mapErr != nil {
			return nil, mapErr
		}
		merged, err = target.Or(merged, driver)
		if err != nil {
			return nil, err
		}
	}
	if _, err = target.CreateCO(merged); err != nil {
		return nil, err
	}
	for _, co := range source.cos[source.PrimaryOutputCount():] {
		if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
			return nil, err
		}
	}
	target.Cleanup()
	if err = target.SetRegisterCount(source.registers); err != nil {
		return nil, err
	}
	return target, target.Check()
}

// CreateSequentialEquivalenceMiter implements Saig_ManCreateEquivMiter.
// Each pair contains source object IDs. Phase normalization follows the C code.
func (source *AigManager) CreateSequentialEquivalenceMiter(pairs [][2]AigID, addLatchOutputs bool, options SaigStructuralOptions) (*AigManager, error) {
	budget, err := newSaigStructuralBudget(source, options)
	if err != nil {
		return nil, err
	}
	if source.constraints != 0 {
		return nil, ErrDimension
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := saigNewMapping(source)
	if err = saigCopyAllInputs(source, target, mapping, budget); err != nil {
		return nil, err
	}
	if err = saigCopyLogic(source, target, mapping, nil, budget); err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		if uint64(pair[0]) >= uint64(len(mapping)) || uint64(pair[1]) >= uint64(len(mapping)) || mapping[pair[0]] == InvalidAigLit || mapping[pair[1]] == InvalidAigLit {
			return nil, ErrDimension
		}
		result, xorErr := target.Xor(mapping[pair[0]], mapping[pair[1]])
		if xorErr != nil {
			return nil, xorErr
		}
		result = result.NotIf(source.objects[pair[0]].Phase != source.objects[pair[1]].Phase)
		if _, err = target.CreateCO(result); err != nil {
			return nil, err
		}
	}
	if addLatchOutputs {
		for _, co := range source.cos[source.PrimaryOutputCount():] {
			if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
				return nil, err
			}
		}
		if err = target.SetRegisterCount(source.registers); err != nil {
			return nil, err
		}
	}
	target.Cleanup()
	return target, target.Check()
}

type SaigTrimResult struct {
	Graph       *AigManager
	OriginalCIs []int
}

// TrimUnusedPrimaryInputs implements Saig_ManTrimPis. It preserves all PIs
// when every PI is unused, matching the source's intentional corner case.
func (source *AigManager) TrimUnusedPrimaryInputs(options SaigStructuralOptions) (SaigTrimResult, error) {
	var result SaigTrimResult
	budget, err := newSaigStructuralBudget(source, options)
	if err != nil {
		return result, err
	}
	allUnused := true
	for _, id := range source.cis[:source.PrimaryInputCount()] {
		if source.objects[id].RefCount != 0 {
			allUnused = false
			break
		}
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := saigNewMapping(source)
	for index, id := range source.cis {
		if index < source.PrimaryInputCount() && !allUnused && source.objects[id].RefCount == 0 {
			continue
		}
		literal, createErr := target.CreateCI()
		if createErr != nil {
			return SaigTrimResult{}, createErr
		}
		mapping[id] = literal
		result.OriginalCIs = append(result.OriginalCIs, index)
	}
	if err = saigCopyLogic(source, target, mapping, nil, budget); err != nil {
		return SaigTrimResult{}, err
	}
	for _, co := range source.cos {
		if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
			return SaigTrimResult{}, err
		}
	}
	if err = target.SetRegisterCount(source.registers); err != nil {
		return SaigTrimResult{}, err
	}
	if err = target.SetConstraintCount(source.constraints); err != nil {
		return SaigTrimResult{}, err
	}
	result.Graph = target
	return result, target.Check()
}

// DuplicateSequentialAbstraction preserves only the selected registers.
// Unselected latch outputs become primary inputs, as in Saig_ManDupAbstraction.
func (source *AigManager) DuplicateSequentialAbstraction(registers []int, options SaigStructuralOptions) (*AigManager, error) {
	budget, err := newSaigStructuralBudget(source, options)
	if err != nil {
		return nil, err
	}
	selected := make([]bool, source.registers)
	for _, register := range registers {
		if register < 0 || register >= source.registers || selected[register] {
			return nil, ErrDimension
		}
		selected[register] = true
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := saigNewMapping(source)
	piCount := source.PrimaryInputCount()
	selectedOrder := make([]int, 0, len(registers))
	for register, keep := range selected {
		if keep {
			selectedOrder = append(selectedOrder, register)
		}
	}
	for index, id := range source.cis {
		if index >= piCount && selected[index-piCount] {
			continue
		}
		literal, createErr := target.CreateCI()
		if createErr != nil {
			return nil, createErr
		}
		mapping[id] = literal
	}
	for _, register := range selectedOrder {
		id := source.cis[piCount+register]
		literal, createErr := target.CreateCI()
		if createErr != nil {
			return nil, createErr
		}
		mapping[id] = literal
	}
	if err = saigCopyLogic(source, target, mapping, nil, budget); err != nil {
		return nil, err
	}
	for _, co := range source.cos[:source.PrimaryOutputCount()] {
		if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
			return nil, err
		}
	}
	for _, register := range selectedOrder {
		co := source.cos[source.PrimaryOutputCount()+register]
		if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(len(registers)); err != nil {
		return nil, err
	}
	target.Cleanup()
	return target, target.Check()
}

type saigTraceResult struct {
	outputs, state, next []bool
}

func (graph *AigManager) simulateSaigTrace(trace *Counterexample, budget *saigStructuralBudget) (saigTraceResult, error) {
	var result saigTraceResult
	if trace == nil {
		return result, ErrDimension
	}
	if err := trace.Validate(); err != nil {
		return result, err
	}
	if trace.registers != graph.registers || trace.inputs != graph.PrimaryInputCount() {
		return result, ErrDimension
	}
	if trace.frames > budget.options.MaxFrames {
		return result, ErrLimit
	}
	state := make([]bool, trace.registers)
	for index := range state {
		state[index] = trace.bit(index)
	}
	primary := make([]bool, trace.inputs)
	for frame := 0; frame < trace.frames; frame++ {
		if err := budget.step(); err != nil {
			return saigTraceResult{}, err
		}
		for index := range primary {
			primary[index] = trace.bit(trace.registers + frame*trace.inputs + index)
		}
		outputs, next, err := graph.Step(primary, state)
		if err != nil {
			return saigTraceResult{}, err
		}
		result.outputs = outputs
		result.state = slices.Clone(state)
		result.next = slices.Clone(next)
		state = next
	}
	return result, nil
}

func (graph *AigManager) VerifySequentialCounterexample(trace *Counterexample, options SaigStructuralOptions) (bool, error) {
	budget, err := newSaigStructuralBudget(graph, options)
	if err != nil {
		return false, err
	}
	if trace == nil || trace.Property < 0 || trace.Property >= graph.PrimaryOutputCount() {
		return false, ErrDimension
	}
	result, err := graph.simulateSaigTrace(trace, budget)
	if err != nil {
		return false, err
	}
	return result.outputs[trace.Property], nil
}

// SequentialFailingState returns the latch-output state in the failing frame,
// or the latch-input state for the following frame when next is true.
func (graph *AigManager) SequentialFailingState(trace *Counterexample, next bool, options SaigStructuralOptions) ([]bool, error) {
	budget, err := newSaigStructuralBudget(graph, options)
	if err != nil {
		return nil, err
	}
	if trace == nil || trace.Property < 0 || trace.Property >= graph.PrimaryOutputCount() {
		return nil, ErrDimension
	}
	result, err := graph.simulateSaigTrace(trace, budget)
	if err != nil {
		return nil, err
	}
	if !result.outputs[trace.Property] {
		return nil, fmt.Errorf("%w: counterexample does not fail", ErrInvariant)
	}
	if next {
		return result.next, nil
	}
	return result.state, nil
}

// ExtendSequentialCounterexample replaces the initial-state-plus-PI layout by
// the source routine's frame-major PI-plus-latch-output layout.
func (graph *AigManager) ExtendSequentialCounterexample(trace *Counterexample, options SaigStructuralOptions) (*Counterexample, error) {
	budget, err := newSaigStructuralBudget(graph, options)
	if err != nil {
		return nil, err
	}
	if trace == nil || trace.Property < 0 || trace.Property >= graph.PrimaryOutputCount() {
		return nil, ErrDimension
	}
	if err = trace.Validate(); err != nil || trace.registers != graph.registers || trace.inputs != graph.PrimaryInputCount() {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if trace.frames > budget.options.MaxFrames {
		return nil, ErrLimit
	}
	extended, err := NewCounterexample(0, graph.InputCount(), trace.frames)
	if err != nil {
		return nil, err
	}
	extended.Property = trace.Property
	state := make([]bool, graph.registers)
	for index := range state {
		state[index] = trace.bit(index)
	}
	primary := make([]bool, graph.PrimaryInputCount())
	for frame := 0; frame < trace.frames; frame++ {
		if err = budget.step(); err != nil {
			return nil, err
		}
		for index := range primary {
			primary[index] = trace.bit(trace.registers + frame*trace.inputs + index)
			extended.set(frame*graph.InputCount()+index, primary[index])
		}
		for index, value := range state {
			extended.set(frame*graph.InputCount()+len(primary)+index, value)
		}
		outputs, next, stepErr := graph.Step(primary, state)
		if stepErr != nil {
			return nil, stepErr
		}
		if frame == trace.LastFrame() && !outputs[trace.Property] {
			return nil, fmt.Errorf("%w: counterexample does not fail", ErrInvariant)
		}
		state = next
	}
	return extended, nil
}

func (graph *AigManager) FindSequentialFailedOutput(trace *Counterexample, options SaigStructuralOptions) (int, error) {
	budget, err := newSaigStructuralBudget(graph, options)
	if err != nil {
		return -1, err
	}
	result, err := graph.simulateSaigTrace(trace, budget)
	if err != nil {
		return -1, err
	}
	for index, value := range result.outputs {
		if value {
			return index, nil
		}
	}
	return -1, nil
}

// DuplicateWithInitialPhase implements Saig_ManDupWithPhase. The returned
// graph assumes an all-zero initial state representing the supplied source
// initial state.
func (source *AigManager) DuplicateWithInitialPhase(initial []bool, options SaigStructuralOptions) (*AigManager, error) {
	budget, err := newSaigStructuralBudget(source, options)
	if err != nil {
		return nil, err
	}
	if len(initial) < source.registers {
		return nil, ErrDimension
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := saigNewMapping(source)
	if err = saigCopyAllInputs(source, target, mapping, budget); err != nil {
		return nil, err
	}
	for index, id := range source.cis[source.PrimaryInputCount():] {
		mapping[id] = mapping[id].NotIf(initial[index])
	}
	if err = saigCopyLogic(source, target, mapping, nil, budget); err != nil {
		return nil, err
	}
	for _, co := range source.cos[:source.PrimaryOutputCount()] {
		if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
			return nil, err
		}
	}
	for index, co := range source.cos[source.PrimaryOutputCount():] {
		driver, mapErr := saigMappedLiteral(mapping, source.objects[co].Fanin0)
		if mapErr != nil {
			return nil, mapErr
		}
		if _, err = target.CreateCO(driver.NotIf(initial[index])); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(source.registers); err != nil {
		return nil, err
	}
	if err = target.SetConstraintCount(source.constraints); err != nil {
		return nil, err
	}
	target.Cleanup()
	return target, target.Check()
}

type SaigConeDuplicate struct {
	Graph         *AigManager
	PrimaryInputs []int
	Registers     []int
	Outputs       []int
}

// DuplicateSequentialCones copies selected primary-output cones and closes
// them through latch transitions. Terminal order is source order, making the
// reduced graph deterministic even when multiple roots discover a terminal.
func (source *AigManager) DuplicateSequentialCones(outputs []int, options SaigStructuralOptions) (SaigConeDuplicate, error) {
	var result SaigConeDuplicate
	budget, err := newSaigStructuralBudget(source, options)
	if err != nil {
		return result, err
	}
	if len(outputs) == 0 {
		return result, ErrDimension
	}
	for _, output := range outputs {
		if output < 0 || output >= source.PrimaryOutputCount() {
			return result, ErrDimension
		}
	}
	ciIndex := make([]int, len(source.objects))
	for index := range ciIndex {
		ciIndex[index] = -1
	}
	for index, id := range source.cis {
		ciIndex[id] = index
	}
	needed := make([]bool, len(source.objects))
	neededPI := make([]bool, source.PrimaryInputCount())
	neededReg := make([]bool, source.registers)
	stack := make([]AigLit, 0, len(outputs))
	for _, output := range outputs {
		co := source.cos[output]
		stack = append(stack, source.objects[co].Fanin0)
	}
	for len(stack) != 0 {
		if err = budget.step(); err != nil {
			return SaigConeDuplicate{}, err
		}
		last := len(stack) - 1
		literal := stack[last]
		stack = stack[:last]
		id := literal.ID()
		if needed[id] {
			continue
		}
		needed[id] = true
		object := source.objects[id]
		switch object.Type {
		case AigConstant1:
		case AigCI:
			index := ciIndex[id]
			if index < 0 {
				return SaigConeDuplicate{}, ErrInvariant
			}
			if index < source.PrimaryInputCount() {
				neededPI[index] = true
			} else {
				register := index - source.PrimaryInputCount()
				if !neededReg[register] {
					neededReg[register] = true
					co := source.cos[source.PrimaryOutputCount()+register]
					stack = append(stack, source.objects[co].Fanin0)
				}
			}
		case AigBuffer:
			stack = append(stack, object.Fanin0)
		case AigAND:
			stack = append(stack, object.Fanin1, object.Fanin0)
		default:
			return SaigConeDuplicate{}, ErrUnsupported
		}
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := saigNewMapping(source)
	for index, keep := range neededPI {
		if !keep {
			continue
		}
		literal, createErr := target.CreateCI()
		if createErr != nil {
			return SaigConeDuplicate{}, createErr
		}
		mapping[source.cis[index]] = literal
		result.PrimaryInputs = append(result.PrimaryInputs, index)
	}
	for index, keep := range neededReg {
		if !keep {
			continue
		}
		literal, createErr := target.CreateCI()
		if createErr != nil {
			return SaigConeDuplicate{}, createErr
		}
		mapping[source.cis[source.PrimaryInputCount()+index]] = literal
		result.Registers = append(result.Registers, index)
	}
	if err = saigCopyLogic(source, target, mapping, needed, budget); err != nil {
		return SaigConeDuplicate{}, err
	}
	for _, output := range outputs {
		if err = saigCreateMappedOutput(source, target, mapping, source.cos[output], budget); err != nil {
			return SaigConeDuplicate{}, err
		}
		result.Outputs = append(result.Outputs, output)
	}
	for _, register := range result.Registers {
		co := source.cos[source.PrimaryOutputCount()+register]
		if err = saigCreateMappedOutput(source, target, mapping, co, budget); err != nil {
			return SaigConeDuplicate{}, err
		}
	}
	if err = target.SetRegisterCount(len(result.Registers)); err != nil {
		return SaigConeDuplicate{}, err
	}
	result.Graph = target
	return result, target.Check()
}

type SaigConeStage struct {
	Frame, Total, Current int
	New, Previous, Older  int
}

func (graph *AigManager) saigDirectRegisterSupport(roots []AigID, budget *saigStructuralBudget) ([]int, error) {
	seen := make([]bool, len(graph.objects))
	ciIndex := make([]int, len(graph.objects))
	for index := range ciIndex {
		ciIndex[index] = -1
	}
	for index, id := range graph.cis {
		ciIndex[id] = index
	}
	registerSeen := make([]bool, graph.registers)
	stack := slices.Clone(roots)
	for len(stack) != 0 {
		if err := budget.step(); err != nil {
			return nil, err
		}
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]
		if uint64(id) >= uint64(len(graph.objects)) || seen[id] {
			continue
		}
		seen[id] = true
		object := graph.objects[id]
		switch object.Type {
		case AigCO, AigBuffer:
			stack = append(stack, object.Fanin0.ID())
		case AigAND:
			stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
		case AigCI:
			index := ciIndex[id]
			if index >= graph.PrimaryInputCount() {
				registerSeen[index-graph.PrimaryInputCount()] = true
			}
		case AigConstant1:
		default:
			return nil, ErrUnsupported
		}
	}
	result := make([]int, 0)
	for index, present := range registerSeen {
		if present {
			result = append(result, index)
		}
	}
	return result, nil
}

// SequentialConeProfile returns the information formerly printed by
// Saig_ManPrintConeOne, without mutating traversal IDs or writing stdout.
func (graph *AigManager) SequentialConeProfile(output int, options SaigStructuralOptions) ([]SaigConeStage, error) {
	budget, err := newSaigStructuralBudget(graph, options)
	if err != nil {
		return nil, err
	}
	if output < 0 || output >= graph.PrimaryOutputCount() {
		return nil, ErrDimension
	}
	current, err := graph.saigDirectRegisterSupport([]AigID{graph.cos[output]}, budget)
	if err != nil {
		return nil, err
	}
	previous := make([]bool, graph.registers)
	total := make([]bool, graph.registers)
	result := make([]SaigConeStage, 0)
	for frame := 0; frame < budget.options.MaxFrames; frame++ {
		newCount, previousCount, olderCount := 0, 0, 0
		for _, register := range current {
			if !total[register] {
				total[register] = true
				newCount++
			} else if previous[register] {
				previousCount++
			} else {
				olderCount++
			}
		}
		totalCount := 0
		for _, present := range total {
			if present {
				totalCount++
			}
		}
		result = append(result, SaigConeStage{Frame: frame, Total: totalCount, Current: len(current), New: newCount, Previous: previousCount, Older: olderCount})
		if newCount == 0 {
			return result, nil
		}
		for index := range previous {
			previous[index] = false
		}
		roots := make([]AigID, len(current))
		for index, register := range current {
			previous[register] = true
			roots[index] = graph.cos[graph.PrimaryOutputCount()+register]
		}
		current, err = graph.saigDirectRegisterSupport(roots, budget)
		if err != nil {
			return nil, err
		}
	}
	return nil, ErrLimit
}

// SequentialConeProfiles is the owned replacement for Saig_ManPrintCones.
// Profiles are returned in primary-output order and no process output occurs.
func (graph *AigManager) SequentialConeProfiles(options SaigStructuralOptions) ([][]SaigConeStage, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	profiles := make([][]SaigConeStage, graph.PrimaryOutputCount())
	for output := range profiles {
		profile, err := graph.SequentialConeProfile(output, options)
		if err != nil {
			return nil, err
		}
		profiles[output] = profile
	}
	return profiles, nil
}

// VerifyUsingBDDs dispositions the no-CUDD tail in saigDup.c explicitly.
// The structural port has no hidden native BDD backend.
func (graph *AigManager) VerifyUsingBDDs(parameters SaigBDDParameters) (int, error) {
	if graph == nil {
		return 0, ErrDimension
	}
	return 0, ErrUnsupported
}
