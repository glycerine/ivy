package abc

// Port of src/base/abci/abcDarUnfold2.c. The source file contains the
// Abc_NtkDarUnfold2 and Abc_NtkDarFold2 network/AIG wrappers.

import "fmt"

// DarUnfold2Options contains the functionally classified constraints consumed
// by the source wrapper's Saig_ManDupUnfoldConstrsFunc2 dependency. TypeI and
// TypeII literals belong to Source. The explicit owned lists keep constraint
// discovery separate from this wrapper and make the transformation reusable.
type DarUnfold2Options struct {
	TypeI, TypeII []AigLit
	MaxObjects    int // zero: one million source/result objects
	Stop          func() bool
}

// DarUnfold2Result owns the unfolded graph and reports the number of Type-II
// registers added by the transformation.
type DarUnfold2Result struct {
	Graph       *AigManager
	TypeICount  int
	TypeIICount int
}

// DarFold2Options controls Abc_NtkDarFold2. Complement preserves the source's
// fCompl polarity switch. Verbose source printing has no Go equivalent.
type DarFold2Options struct {
	Complement bool
	MaxObjects int // zero: one million source/result objects
	Stop       func() bool
}

type darUnfold2Budget struct {
	maximum int
	stop    func() bool
}

func newDarUnfold2Budget(maximum int, stop func() bool) (darUnfold2Budget, error) {
	if maximum < 0 {
		return darUnfold2Budget{}, ErrDimension
	}
	if maximum == 0 {
		maximum = 1 << 20
	}
	return darUnfold2Budget{maximum: maximum, stop: stop}, nil
}

func (b darUnfold2Budget) step() error {
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	return nil
}

func (b darUnfold2Budget) checkSize(graph *AigManager) error {
	if graph == nil || graph.ObjectCount() > b.maximum {
		return ErrLimit
	}
	return nil
}

// darUnfold2CopyLogic performs the AIG half of Abc_NtkToDar followed by the
// source dependency's duplication-without-POs. It creates every CI and live
// internal object, but no CO. The returned map translates regular source IDs.
func darUnfold2CopyLogic(source *AigManager, budget darUnfold2Budget) (*AigManager, []AigLit, error) {
	if source == nil {
		return nil, nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	if source.equivs != nil {
		return nil, nil, ErrUnsupported
	}
	if source.ObjectCount() > budget.maximum {
		return nil, nil, ErrLimit
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := make([]AigLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for _, id := range source.cis {
		if err := budget.step(); err != nil {
			return nil, nil, err
		}
		literal, err := target.CreateCI()
		if err != nil {
			return nil, nil, err
		}
		mapping[id] = literal
		if err = budget.checkSize(target); err != nil {
			return nil, nil, err
		}
	}
	copyLiteral := func(literal AigLit) (AigLit, error) {
		if uint64(literal.ID()) >= uint64(len(mapping)) || mapping[literal.ID()] == InvalidAigLit {
			return InvalidAigLit, ErrInvariant
		}
		return mapping[literal.ID()].NotIf(literal.Complemented()), nil
	}
	for id, object := range source.objects {
		if err := budget.step(); err != nil {
			return nil, nil, err
		}
		var literal AigLit
		var err error
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigBuffer:
			var fanin AigLit
			fanin, err = copyLiteral(object.Fanin0)
			if err == nil {
				literal, err = target.CreateBuffer(fanin)
			}
		case AigAND:
			var left, right AigLit
			left, err = copyLiteral(object.Fanin0)
			if err == nil {
				right, err = copyLiteral(object.Fanin1)
			}
			if err == nil {
				literal, err = target.And(left, right)
			}
		default:
			return nil, nil, ErrUnsupported
		}
		if err != nil {
			return nil, nil, err
		}
		mapping[id] = literal
		if err = budget.checkSize(target); err != nil {
			return nil, nil, err
		}
	}
	return target, mapping, nil
}

func darUnfold2Mapped(source *AigManager, mapping []AigLit, literal AigLit) (AigLit, error) {
	if source == nil || !source.validSignal(literal) || uint64(literal.ID()) >= uint64(len(mapping)) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	mapped := mapping[literal.ID()]
	if mapped == InvalidAigLit {
		return InvalidAigLit, ErrInvariant
	}
	return mapped.NotIf(literal.Complemented()), nil
}

func darUnfold2CheckCandidates(source *AigManager, typeI, typeII []AigLit) error {
	seen := make(map[AigID]struct{}, len(typeI)+len(typeII))
	for _, list := range [][]AigLit{typeI, typeII} {
		for _, literal := range list {
			if source == nil || literal.ID() == 0 || !source.validSignal(literal) {
				return ErrInvalidLiteral
			}
			if _, duplicate := seen[literal.ID()]; duplicate {
				return fmt.Errorf("%w: duplicate classified object %d", ErrInvariant, literal.ID())
			}
			seen[literal.ID()] = struct{}{}
		}
	}
	return nil
}

// DarUnfold2 implements the graph transformation performed by
// Abc_NtkDarUnfold2 after functional constraint discovery. Existing primary
// outputs remain in order. Type-I constraints are direct candidate outputs.
// Each Type-II constraint receives a new latch whose output is conjoined with
// the candidate, and whose next-state input is the candidate itself. The
// source dependency fixes fCompl to zero for both conjunction inputs.
// Name and Spec are owned strings in Go and are copied by value.
func DarUnfold2(source *AigManager, options DarUnfold2Options) (DarUnfold2Result, error) {
	budget, err := newDarUnfold2Budget(options.MaxObjects, options.Stop)
	if err != nil || source == nil {
		if err == nil {
			err = ErrDimension
		}
		return DarUnfold2Result{}, err
	}
	if err = source.Check(); err != nil {
		return DarUnfold2Result{}, err
	}
	if err = darUnfold2CheckCandidates(source, options.TypeI, options.TypeII); err != nil {
		return DarUnfold2Result{}, err
	}
	target, mapping, err := darUnfold2CopyLogic(source, budget)
	if err != nil {
		return DarUnfold2Result{}, err
	}
	primaryOutputs := source.PrimaryOutputCount()
	appendMappedOutput := func(literal AigLit) error {
		if err := budget.step(); err != nil {
			return err
		}
		mapped, err := darUnfold2Mapped(source, mapping, literal)
		if err != nil {
			return err
		}
		if _, err = target.CreateCO(mapped); err != nil {
			return err
		}
		return budget.checkSize(target)
	}
	for _, id := range source.cos[:primaryOutputs] {
		if err = appendMappedOutput(source.objects[id].Fanin0); err != nil {
			return DarUnfold2Result{}, err
		}
	}
	for _, candidate := range options.TypeI {
		if err = appendMappedOutput(candidate); err != nil {
			return DarUnfold2Result{}, err
		}
	}
	typeIILatches := make([]AigLit, len(options.TypeII))
	for index, candidate := range options.TypeII {
		if err = budget.step(); err != nil {
			return DarUnfold2Result{}, err
		}
		if typeIILatches[index], err = target.CreateCI(); err != nil {
			return DarUnfold2Result{}, err
		}
		mapped, mapErr := darUnfold2Mapped(source, mapping, candidate)
		if mapErr != nil {
			return DarUnfold2Result{}, mapErr
		}
		constraint, andErr := target.And(typeIILatches[index], mapped)
		if andErr != nil {
			return DarUnfold2Result{}, andErr
		}
		if _, err = target.CreateCO(constraint); err != nil {
			return DarUnfold2Result{}, err
		}
		if err = budget.checkSize(target); err != nil {
			return DarUnfold2Result{}, err
		}
	}
	for _, id := range source.cos[primaryOutputs:] {
		if err = appendMappedOutput(source.objects[id].Fanin0); err != nil {
			return DarUnfold2Result{}, err
		}
	}
	for _, candidate := range options.TypeII {
		if err = appendMappedOutput(candidate); err != nil {
			return DarUnfold2Result{}, err
		}
	}
	if err = target.SetRegisterCount(source.registers + len(options.TypeII)); err != nil {
		return DarUnfold2Result{}, err
	}
	if err = target.SetConstraintCount(source.constraints + len(options.TypeI) + len(options.TypeII)); err != nil {
		return DarUnfold2Result{}, err
	}
	if err = target.Check(); err != nil {
		return DarUnfold2Result{}, err
	}
	return DarUnfold2Result{Graph: target, TypeICount: len(options.TypeI), TypeIICount: len(options.TypeII)}, nil
}

// DarFold2 implements Abc_NtkDarFold2 and its active dependency behavior. The
// source dependency resets its Type-II count before use, so every constraint
// is folded identically. A sequential graph gains a sticky violation latch;
// ordinary outputs are disabled and old state is held once violation occurs.
func DarFold2(source *AigManager, options DarFold2Options) (*AigManager, error) {
	budget, err := newDarUnfold2Budget(options.MaxObjects, options.Stop)
	if err != nil || source == nil {
		if err == nil {
			err = ErrDimension
		}
		return nil, err
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.constraints == 0 {
		target, mapping, copyErr := darUnfold2CopyLogic(source, budget)
		if copyErr != nil {
			return nil, copyErr
		}
		for _, id := range source.cos {
			if copyErr = budget.step(); copyErr != nil {
				return nil, copyErr
			}
			mapped, mapErr := darUnfold2Mapped(source, mapping, source.objects[id].Fanin0)
			if mapErr != nil {
				return nil, mapErr
			}
			if _, copyErr = target.CreateCO(mapped); copyErr != nil {
				return nil, copyErr
			}
			if copyErr = budget.checkSize(target); copyErr != nil {
				return nil, copyErr
			}
		}
		if err = target.SetRegisterCount(source.registers); err != nil {
			return nil, err
		}
		target.Cleanup()
		if err = target.Check(); err != nil {
			return nil, err
		}
		return target, nil
	}
	primaryOutputs := source.PrimaryOutputCount()
	ordinaryOutputs := primaryOutputs - source.constraints
	if ordinaryOutputs <= 0 {
		return nil, ErrDimension
	}
	target, mapping, err := darUnfold2CopyLogic(source, budget)
	if err != nil {
		return nil, err
	}
	violation := AigConst0
	for _, id := range source.cos[ordinaryOutputs:primaryOutputs] {
		if err = budget.step(); err != nil {
			return nil, err
		}
		constraint, mapErr := darUnfold2Mapped(source, mapping, source.objects[id].Fanin0)
		if mapErr != nil {
			return nil, mapErr
		}
		violation, err = target.Or(violation, constraint.NotIf(options.Complement))
		if err != nil {
			return nil, err
		}
	}
	stickyNext := violation
	if source.registers > 0 {
		var stickyOutput AigLit
		stickyOutput, err = target.CreateCI()
		if err == nil {
			stickyNext, err = target.Or(violation, stickyOutput)
		}
		if err != nil {
			return nil, err
		}
	}
	for _, id := range source.cos[:ordinaryOutputs] {
		driver, mapErr := darUnfold2Mapped(source, mapping, source.objects[id].Fanin0)
		if mapErr != nil {
			return nil, mapErr
		}
		gated, gateErr := target.And(driver, stickyNext.Not())
		if gateErr != nil {
			return nil, gateErr
		}
		if _, err = target.CreateCO(gated); err != nil {
			return nil, err
		}
	}
	for index, id := range source.cos[primaryOutputs:] {
		next, mapErr := darUnfold2Mapped(source, mapping, source.objects[id].Fanin0)
		if mapErr != nil {
			return nil, mapErr
		}
		oldOutput := mapping[source.cis[source.PrimaryInputCount()+index]]
		held, muxErr := target.Mux(stickyNext.Not(), next, oldOutput)
		if muxErr != nil {
			return nil, muxErr
		}
		if _, err = target.CreateCO(held); err != nil {
			return nil, err
		}
	}
	registers := source.registers
	if source.registers > 0 {
		if _, err = target.CreateCO(stickyNext); err != nil {
			return nil, err
		}
		registers++
	}
	if err = target.SetRegisterCount(registers); err != nil {
		return nil, err
	}
	if err = budget.checkSize(target); err != nil {
		return nil, err
	}
	target.Cleanup()
	cleaned, err := target.SequentialCleanup(AigSequentialCleanupOptions{MaxObjects: budget.maximum, Stop: budget.stop})
	if err != nil {
		return nil, err
	}
	if err = cleaned.Graph.Check(); err != nil {
		return nil, err
	}
	return cleaned.Graph, nil
}
