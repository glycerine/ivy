package abc

// Three-valued prefix analysis and interval bounded checking from
// src/sat/bmc/bmcBmc2.c.

import (
	"errors"
	"fmt"
)

type AigTernaryPrefixOptions struct {
	Frames, MaxObjects, MaxEntries int
	Stop                           func() bool
}

type AigTernaryPrefixResult struct {
	Values                                [][]AigTernary
	FirstNonzeroFrame, FirstNonzeroOutput int
	Complete                              bool
}

// SequentialTernaryPrefix starts every register at zero and every primary
// input at unknown. It retains object-major values for each completed frame
// and stops at the first PO that cannot be proved zero.
func (graph *AigManager) SequentialTernaryPrefix(options AigTernaryPrefixOptions) (AigTernaryPrefixResult, error) {
	result := AigTernaryPrefixResult{FirstNonzeroFrame: -1, FirstNonzeroOutput: -1}
	if graph == nil || graph.RegisterCount() == 0 {
		return result, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return result, err
	}
	if options.Frames < 1 || options.MaxObjects < 0 || options.MaxEntries < 0 {
		return result, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	if graph.ObjectCount() > options.MaxObjects {
		return result, ErrLimit
	}
	if _, err := checkedSize(0, options.Frames, graph.ObjectCount(), options.MaxEntries); err != nil {
		return result, err
	}
	state := make([]AigTernary, graph.RegisterCount())
	for register := range state {
		state[register] = AigTernaryZero
	}
	read := func(values []AigTernary, literal AigLit) AigTernary {
		value := values[literal.ID()]
		if literal.Complemented() {
			value = value.Not()
		}
		return value
	}
	for frame := 0; frame < options.Frames; frame++ {
		if options.Stop != nil && options.Stop() {
			return AigTernaryPrefixResult{}, ErrCancelled
		}
		values := make([]AigTernary, graph.ObjectCount())
		values[0] = AigTernaryOne
		for input, id := range graph.cis[:graph.PrimaryInputCount()] {
			_ = input
			values[id] = AigTernaryUnknown
		}
		for register, id := range graph.cis[graph.PrimaryInputCount():] {
			values[id] = state[register]
		}
		for id, object := range graph.objects {
			if options.Stop != nil && options.Stop() {
				return AigTernaryPrefixResult{}, ErrCancelled
			}
			switch object.Type {
			case AigAND:
				values[id] = aigTernaryAnd(read(values, object.Fanin0), read(values, object.Fanin1))
			case AigBuffer, AigCO:
				values[id] = read(values, object.Fanin0)
			}
		}
		result.Values = append(result.Values, values)
		for output, id := range graph.cos[:graph.PrimaryOutputCount()] {
			if values[id] != AigTernaryZero {
				result.FirstNonzeroFrame, result.FirstNonzeroOutput = frame, output
				return result, nil
			}
		}
		for register, id := range graph.cos[graph.PrimaryOutputCount():] {
			state[register] = values[id]
		}
	}
	result.Complete = true
	return result, nil
}

type AigIntervalBMCOptions struct {
	StartFrame, Frames, NodeBatch               int
	MaxObjects, MaxMappingEntries, MaxVariables int
	MaxClauses, MaxLiterals                     int
	GlobalConflicts                             uint64
	SAT                                         SatLimits
	OnTargetDone                                func(frame, property int, status SatStatus)
}

type AigIntervalBMCResult struct {
	Status                          BmcStatus
	Reason                          BmcStopReason
	SatReason                       SatStopReason
	CheckedFrames, CheckedTargets   int
	SkippedTargets, Frame, Property int
	Counterexample                  *Counterexample
	Stats                           SatStats
}

func normalizeAigIntervalBMCOptions(options AigIntervalBMCOptions) (AigIntervalBMCOptions, error) {
	if options.StartFrame < 0 || options.Frames < 1 || options.StartFrame > options.Frames || options.NodeBatch < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
	}
	if options.NodeBatch == 0 {
		options.NodeBatch = 10_000
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 4 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 16 << 20
	}
	return options, nil
}

// IntervalBoundedCheck preserves bmc2's frame-major target order, start-frame
// skipping, per-target and cumulative SAT limits, UNSAT target blocking and
// first replayed witness. NodeBatch remains a construction-performance hint;
// the owned Go unroller builds the same finite relation in one arena.
func (graph *AigManager) IntervalBoundedCheck(options AigIntervalBMCOptions) (result AigIntervalBMCResult, err error) {
	result.Frame, result.Property = -1, -1
	if graph == nil {
		return result, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return result, err
	}
	options, err = normalizeAigIntervalBMCOptions(options)
	if err != nil {
		return result, err
	}
	if graph.PrimaryOutputCount() == 0 {
		return result, ErrDimension
	}
	stopped := func() bool { return options.SAT.Stop != nil && options.SAT.Stop() }
	interrupted := func(cause error) (AigIntervalBMCResult, error) {
		switch {
		case errors.Is(cause, ErrCancelled):
			result.Status, result.Reason, result.SatReason = BmcUnknown, BmcCancelled, SatCancelled
			return result, nil
		case errors.Is(cause, ErrLimit):
			result.Status, result.Reason = BmcUnknown, BmcResourceLimit
			return result, nil
		default:
			return result, cause
		}
	}
	if stopped() {
		return interrupted(ErrCancelled)
	}
	// bmc2 treats every PO as a target. Clone before clearing AIGER constraint
	// metadata so this adapter never mutates its caller.
	owned := graph.Clone()
	model, err := NewAigerModel(owned, owned.PrimaryOutputCount(), 0, 0, make([]LatchInit, owned.RegisterCount()))
	if err != nil {
		return result, err
	}
	gia, _, err := GiaFromAig(owned)
	if err != nil {
		return interrupted(err)
	}
	unrolled, err := gia.UnrollWithOptions(GiaUnrollOptions{Frames: options.Frames, MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.SAT.Stop})
	if err != nil {
		return interrupted(err)
	}
	aig, giaToAig, err := unrolled.graph.ToAig()
	if err != nil {
		return interrupted(err)
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return interrupted(err)
	}
	clauses, literalCount := relation.Clauses(), 0
	for _, clause := range clauses {
		literalCount += len(clause)
	}
	targetUnits, err := checkedSize(0, options.Frames-options.StartFrame, graph.PrimaryOutputCount(), options.MaxClauses)
	if err != nil {
		return interrupted(err)
	}
	if relation.VariableCount() > options.MaxVariables || len(clauses) > options.MaxClauses-targetUnits || literalCount > options.MaxLiterals-targetUnits {
		return interrupted(ErrLimit)
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return interrupted(err)
	}
	defer func() { result.Stats = solver.Statistics() }()
	ok, err := relation.loadIntoWithStop(solver, 0, options.SAT.Stop)
	if err != nil {
		return interrupted(err)
	}
	if !ok {
		return result, fmt.Errorf("%w: pure unrolled relation is inconsistent", ErrInvariant)
	}
	outputCount := graph.PrimaryOutputCount()
	value := func(answer SatResult, literal GiaLit) (bool, error) {
		if !literal.Valid() || uint64(literal.ID()) >= uint64(len(giaToAig)) {
			return false, ErrInvariant
		}
		aigLiteral := giaToAig[literal.ID()].NotIf(literal.Complemented())
		satLiteral, valueErr := relation.ObjectLiteral(aigLiteral.ID())
		if valueErr != nil {
			return false, valueErr
		}
		satLiteral ^= SatLit(aigLiteral & 1)
		if uint64(satLiteral.Var()) >= uint64(len(answer.Model)) {
			return false, ErrInvariant
		}
		return answer.Model[satLiteral.Var()] != satLiteral.Negative(), nil
	}
	for frame := 0; frame < options.Frames; frame++ {
		result.Frame = frame
		for property := 0; property < outputCount; property++ {
			result.Property = property
			if frame < options.StartFrame {
				result.SkippedTargets++
				continue
			}
			if stopped() {
				return interrupted(ErrCancelled)
			}
			limits := options.SAT
			if options.GlobalConflicts > 0 {
				used := solver.Statistics().Conflicts
				if used >= options.GlobalConflicts {
					result.Status, result.Reason, result.SatReason = BmcUnknown, BmcSatLimit, SatConflictLimit
					return result, nil
				}
				remaining := options.GlobalConflicts - used
				if limits.Conflicts == 0 || remaining < limits.Conflicts {
					limits.Conflicts = remaining
				}
			}
			query := relation.outputs[frame*outputCount+property]
			answer, solveErr := solver.Solve([]SatLit{query}, limits)
			if solveErr != nil {
				return result, solveErr
			}
			result.CheckedTargets++
			if options.OnTargetDone != nil {
				options.OnTargetDone(frame, property, answer.Status)
			}
			switch answer.Status {
			case SatUnknown:
				result.Status, result.Reason, result.SatReason = BmcUnknown, BmcSatLimit, answer.Reason
				if answer.Reason == SatCancelled {
					result.Reason = BmcCancelled
				}
				return result, nil
			case SatUnsatisfiable:
				ok, addErr := solver.AddClause([]SatLit{query.Not()})
				if addErr != nil {
					return interrupted(addErr)
				}
				if !ok {
					return result, fmt.Errorf("%w: entailed target block contradicted relation", ErrInvariant)
				}
			case SatSatisfiable:
				counterexample, cexErr := NewCounterexample(graph.RegisterCount(), graph.PrimaryInputCount(), frame+1)
				if cexErr != nil {
					return interrupted(cexErr)
				}
				counterexample.Property = property
				for register, literal := range unrolled.initial {
					bit, valueErr := value(answer, literal)
					if valueErr != nil {
						return result, valueErr
					}
					counterexample.set(register, bit)
				}
				for f := 0; f <= frame; f++ {
					for input := 0; input < counterexample.inputs; input++ {
						bit, valueErr := value(answer, unrolled.frameInputs[f*counterexample.inputs+input])
						if valueErr != nil {
							return result, valueErr
						}
						counterexample.set(counterexample.registers+f*counterexample.inputs+input, bit)
					}
				}
				if cexErr = model.ValidateCounterexample(counterexample); cexErr != nil {
					return result, fmt.Errorf("%w: interval BMC witness failed replay: %v", ErrInvariant, cexErr)
				}
				result.Status, result.Reason = BmcCounterexample, BmcWitnessFound
				result.Counterexample = counterexample
				return result, nil
			default:
				return result, ErrInvariant
			}
		}
		if frame >= options.StartFrame {
			result.CheckedFrames++
		}
	}
	result.Status, result.Reason, result.SatReason = BmcNoCounterexample, BmcBoundReached, SatFinished
	return result, nil
}
