package abc

// Scalable solve-all bounded checking from src/sat/bmc/bmcBmc3.c.

import (
	"errors"
	"fmt"
)

type AigScalableBMCOptions struct {
	StartFrame, Frames, FramesJump              int
	SolveAll, DropUnknownOutputs, StoreWitness  bool
	MaxObjects, MaxMappingEntries, MaxVariables int
	MaxClauses, MaxLiterals                     int
	SAT, JumpSAT                                SatLimits
	OnProgress                                  func(property, depth int)
	OnFailure                                   func(property int, witness *Counterexample) bool
}

type AigScalableOutput struct {
	Status  BmcStatus
	Frame   int
	Dropped bool
	Witness *Counterexample
}

type AigScalableBMCResult struct {
	Status                         BmcStatus
	Reason                         BmcStopReason
	SatReason                      SatStopReason
	Complete                       bool
	CheckedFrames, CheckedTargets  int
	SkippedFrames, Failed, Dropped int
	Frame, Property                int
	Outputs                        []AigScalableOutput
	Stats                          SatStats
}

func normalizeAigScalableBMCOptions(options AigScalableBMCOptions) (AigScalableBMCOptions, error) {
	if options.StartFrame < 0 || options.Frames < 1 || options.StartFrame > options.Frames || options.FramesJump < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
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
	if options.JumpSAT.Stop == nil {
		options.JumpSAT.Stop = options.SAT.Stop
	}
	return options, nil
}

func aigScalableWitness(source, relationGraph *AigManager, model *AigerModel, unrolled *GiaUnrolling, giaToAig []AigLit, relation *AigCNF, assignment []bool, frame, property int) (*Counterexample, error) {
	witness, err := NewCounterexample(source.RegisterCount(), source.PrimaryInputCount(), frame+1)
	if err != nil {
		return nil, err
	}
	witness.Property = property
	read := func(literal GiaLit) (bool, error) {
		if !literal.Valid() || uint64(literal.ID()) >= uint64(len(giaToAig)) {
			return false, ErrInvariant
		}
		mapped := giaToAig[literal.ID()].NotIf(literal.Complemented())
		satLiteral, err := relation.ObjectLiteral(mapped.ID())
		if err != nil {
			return false, err
		}
		satLiteral ^= SatLit(mapped & 1)
		if int(satLiteral.Var()) >= len(assignment) {
			return false, ErrInvariant
		}
		return assignment[satLiteral.Var()] != satLiteral.Negative(), nil
	}
	for current := 0; current <= frame; current++ {
		inputs, err := unrolled.FrameInputs(current)
		if err != nil {
			return nil, err
		}
		for input, literal := range inputs {
			value, err := read(literal)
			if err != nil {
				return nil, err
			}
			if err = witness.SetInput(current, input, value); err != nil {
				return nil, err
			}
		}
	}
	if relationGraph == nil {
		return nil, ErrInvariant
	}
	if err := model.ValidateCounterexample(witness); err != nil {
		return nil, fmt.Errorf("%w: scalable BMC witness failed replay: %v", ErrInvariant, err)
	}
	return witness, nil
}

// ScalableBoundedCheck preserves BMC3's frame/output order, optional solve-all
// mode, same-model discovery, entailed UNSAT units, start frame, jump after an
// unknown query, finite-state completion and callback points. The three native
// SAT backends are consolidated into the concrete Go CDCL solver.
func (graph *AigManager) ScalableBoundedCheck(options AigScalableBMCOptions) (result AigScalableBMCResult, err error) {
	result.Frame, result.Property = -1, -1
	if graph == nil || graph.PrimaryOutputCount() == 0 {
		return result, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return result, err
	}
	for _, object := range graph.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return result, ErrUnsupported
		}
	}
	options, err = normalizeAigScalableBMCOptions(options)
	if err != nil {
		return result, err
	}
	outputCount := graph.PrimaryOutputCount()
	result.Outputs = make([]AigScalableOutput, outputCount)
	for index := range result.Outputs {
		result.Outputs[index].Frame = -1
	}
	stopped := func() bool { return options.SAT.Stop != nil && options.SAT.Stop() }
	interrupted := func(cause error) (AigScalableBMCResult, error) {
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
	owned := graph.Clone()
	model, err := NewAigerModel(owned, outputCount, 0, 0, make([]LatchInit, owned.RegisterCount()))
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
	relationGraph, giaToAig, err := unrolled.graph.ToAig()
	if err != nil {
		return interrupted(err)
	}
	relation, err := EncodeAigCNF(relationGraph)
	if err != nil {
		return interrupted(err)
	}
	clauses := relation.Clauses()
	literalCount := 0
	for _, clause := range clauses {
		literalCount += len(clause)
	}
	units, err := checkedSize(0, options.Frames-options.StartFrame, outputCount, options.MaxClauses)
	if err != nil {
		return interrupted(err)
	}
	if relation.VariableCount() > options.MaxVariables || len(clauses) > options.MaxClauses-units || literalCount > options.MaxLiterals-units {
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
		return result, fmt.Errorf("%w: scalable transition relation is inconsistent", ErrInvariant)
	}
	markFailure := func(property, frame int, assignment []bool, notify bool) (bool, error) {
		witness, err := aigScalableWitness(graph, relationGraph, model, unrolled, giaToAig, relation, assignment, frame, property)
		if err != nil {
			return false, err
		}
		result.Outputs[property].Status = BmcCounterexample
		result.Outputs[property].Frame = frame
		if options.StoreWitness {
			result.Outputs[property].Witness = witness
		}
		result.Failed++
		result.Status = BmcCounterexample
		if notify && options.OnFailure != nil {
			callbackWitness := (*Counterexample)(nil)
			if options.StoreWitness {
				callbackWitness = witness.Clone()
			}
			return options.OnFailure(property, callbackWitness), nil
		}
		return false, nil
	}
	nextFrame := options.StartFrame
	limits := options.SAT
	for frame := 0; frame < options.Frames; frame++ {
		if graph.RegisterCount() < 30 && options.FramesJump == 0 && frame == 1<<graph.RegisterCount() {
			result.Complete = result.Dropped == 0
			for property := range result.Outputs {
				if result.Outputs[property].Status == BmcUnknown && !result.Outputs[property].Dropped {
					result.Outputs[property].Status = BmcNoCounterexample
					result.Outputs[property].Frame = frame - 1
				}
			}
			if result.Dropped != 0 {
				result.Status, result.Reason = BmcUnknown, BmcSatLimit
				return result, nil
			}
			if result.Failed == 0 {
				result.Status = BmcNoCounterexample
			}
			result.Reason = BmcBoundReached
			return result, nil
		}
		if frame < nextFrame {
			result.SkippedFrames++
			continue
		}
		if stopped() {
			return interrupted(ErrCancelled)
		}
		result.Frame = frame
		jumped := false
		for property := 0; property < outputCount; property++ {
			if result.Outputs[property].Status == BmcCounterexample || result.Outputs[property].Dropped {
				continue
			}
			result.Property = property
			query := relation.outputs[frame*outputCount+property]
			answer, solveErr := solver.Solve([]SatLit{query}, limits)
			if solveErr != nil {
				return result, solveErr
			}
			result.CheckedTargets++
			switch answer.Status {
			case SatUnsatisfiable:
				ok, addErr := solver.AddClause([]SatLit{query.Not()})
				if addErr != nil {
					return interrupted(addErr)
				}
				if !ok {
					return result, ErrInvariant
				}
				if options.OnProgress != nil {
					options.OnProgress(property, frame)
				}
			case SatUnknown:
				result.SatReason = answer.Reason
				if options.FramesJump > 0 {
					nextFrame = frame + options.FramesJump
					limits = options.JumpSAT
					jumped = true
					break
				}
				if options.SolveAll && options.DropUnknownOutputs {
					result.Outputs[property].Dropped = true
					result.Outputs[property].Frame = frame
					result.Dropped++
					continue
				}
				result.Status, result.Reason = BmcUnknown, BmcSatLimit
				if answer.Reason == SatCancelled {
					result.Reason = BmcCancelled
				}
				return result, nil
			case SatSatisfiable:
				stop, markErr := markFailure(property, frame, answer.Model, true)
				if markErr != nil {
					return result, markErr
				}
				result.Property = property
				if !options.SolveAll || stop {
					result.Reason = BmcWitnessFound
					return result, nil
				}
				for other := 0; other < outputCount; other++ {
					if result.Outputs[other].Status == BmcCounterexample || result.Outputs[other].Dropped {
						continue
					}
					literal := relation.outputs[frame*outputCount+other]
					if int(literal.Var()) < len(answer.Model) && answer.Model[literal.Var()] != literal.Negative() {
						if _, markErr = markFailure(other, frame, answer.Model, false); markErr != nil {
							return result, markErr
						}
					}
				}
			}
			if result.Failed+result.Dropped == outputCount {
				if result.Failed != 0 {
					result.Status, result.Reason = BmcCounterexample, BmcWitnessFound
				} else {
					result.Status, result.Reason = BmcUnknown, BmcSatLimit
				}
				return result, nil
			}
		}
		if jumped {
			continue
		}
		result.CheckedFrames++
	}
	if result.Dropped != 0 {
		result.Status, result.Reason = BmcUnknown, BmcSatLimit
		return result, nil
	}
	if result.Failed == 0 {
		result.Status = BmcNoCounterexample
	}
	result.Reason = BmcBoundReached
	return result, nil
}
