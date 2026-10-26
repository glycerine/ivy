package abc

// Port of the per-output bounded SAT loop in src/sat/bmc/bmcBmc.c:
// Saig_ManBmcSimple. It currently uses the independent Tseitin encoder rather
// than ABC's optimized cut CNF, and keeps initial/constraint semantics explicit.

import (
	"errors"
	"fmt"
	"slices"
)

type BmcStatus uint8

const (
	BmcUnknown BmcStatus = iota
	BmcNoCounterexample
	BmcCounterexample
)

type BmcStopReason uint8

const (
	BmcBoundReached BmcStopReason = iota
	BmcSatLimit
	BmcCancelled
	BmcResourceLimit
	BmcWitnessFound
)

// BmcOptions checks frames 0 through Frames-1. Nil Properties selects all
// ordinary/bad outputs, excluding constraints. Explicit properties retain the
// caller's order and must be unique, valid ordinary/bad output indexes.
// SAT budgets apply separately to each query, as in Saig_ManBmcSimple;
// preprocessing is bounded by object, mapping and variable caps instead.
// Stop is polled between phases, unrolling nodes, CNF clauses and SAT work.
// It must not mutate the model or reenter this operation.
type BmcOptions struct {
	Frames, MaxObjects, MaxMappingEntries, MaxVariables int
	Properties                                          []int
	SAT                                                 SatLimits
}

// BmcResult never claims unbounded safety. NoCounterexample means no selected
// property fails in the requested finite prefix. CheckedFrames counts fully
// excluded frames; Frame is the current/last frame (-1 before queries).
type BmcResult struct {
	Status                         BmcStatus
	Reason                         BmcStopReason
	SatReason                      SatStopReason
	CheckedFrames, Frame, Property int
	Counterexample                 *Counterexample
	Stats                          SatStats
}

// BoundedCheck unrolls the transition system, adds environment constraints a
// frame at a time, and checks each selected property in increasing frame order.
// Every SAT result is converted to an original-model trace and replayed before
// it is returned. Errors and exhausted limits never become proof answers.
func (m *AigerModel) BoundedCheck(options BmcOptions) (result BmcResult, err error) {
	result.Frame = -1
	result.Property = -1
	if err := m.Validate(); err != nil {
		return result, err
	}
	if options.Frames < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 {
		return result, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	// Validate the variable cap even when no SAT query would be necessary.
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return result, err
	}
	defer func() { result.Stats = solver.Statistics() }()
	propertyCount := m.outputs + m.bad
	properties := slices.Clone(options.Properties)
	if properties == nil {
		properties = make([]int, propertyCount)
		for i := range properties {
			properties[i] = i
		}
	}
	selected := make([]bool, propertyCount)
	for _, property := range properties {
		if property < 0 || property >= propertyCount || selected[property] {
			return result, ErrDimension
		}
		selected[property] = true
	}
	stopped := func() bool { return options.SAT.Stop != nil && options.SAT.Stop() }
	interrupted := func(cause error) (BmcResult, error) {
		switch {
		case errors.Is(cause, ErrCancelled):
			result.Reason = BmcCancelled
			result.SatReason = SatCancelled
			return result, nil
		case errors.Is(cause, ErrLimit):
			result.Reason = BmcResourceLimit
			return result, nil
		default:
			return result, cause
		}
	}
	if stopped() {
		return interrupted(ErrCancelled)
	}
	gia, _, err := GiaFromAig(m.graph)
	if err != nil {
		return interrupted(err)
	}
	if stopped() {
		return interrupted(ErrCancelled)
	}
	unrolled, err := gia.UnrollWithOptions(GiaUnrollOptions{Frames: options.Frames, MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Initial: m.initial, Stop: options.SAT.Stop})
	if err != nil {
		return interrupted(err)
	}
	if stopped() {
		return interrupted(ErrCancelled)
	}
	aig, giaToAig, err := unrolled.graph.ToAig()
	if err != nil {
		return interrupted(err)
	}
	if stopped() {
		return interrupted(ErrCancelled)
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return interrupted(err)
	}
	ok, err := relation.loadIntoWithStop(solver, 0, options.SAT.Stop)
	if err != nil {
		return interrupted(err)
	}
	if !ok {
		return result, fmt.Errorf("%w: pure transition relation is inconsistent", ErrInvariant)
	}
	outputCount := m.graph.PrimaryOutputCount()
	complete := func() BmcResult {
		result.Status = BmcNoCounterexample
		result.Reason = BmcBoundReached
		result.CheckedFrames = options.Frames
		result.Frame = options.Frames - 1
		return result
	}
	for frame := 0; frame < options.Frames; frame++ {
		result.Frame = frame
		if stopped() {
			return interrupted(ErrCancelled)
		}
		// Add only the current constraint prefix. Constraining future frames here
		// could hide a real early bug whose trace has no admissible extension.
		for constraint := propertyCount; constraint < outputCount; constraint++ {
			if stopped() {
				return interrupted(ErrCancelled)
			}
			l := relation.outputs[frame*outputCount+constraint]
			ok, err := solver.AddClause([]SatLit{l})
			if err != nil {
				return interrupted(err)
			}
			// An inconsistent prefix cannot extend to any later admissible trace.
			if !ok {
				return complete(), nil
			}
		}
		for _, property := range properties {
			query := relation.outputs[frame*outputCount+property]
			answer, err := solver.Solve([]SatLit{query}, options.SAT)
			if err != nil {
				return result, err
			}
			switch answer.Status {
			case SatUnknown:
				result.SatReason = answer.Reason
				result.Reason = BmcSatLimit
				if answer.Reason == SatCancelled {
					result.Reason = BmcCancelled
				}
				return result, nil
			case SatUnsatisfiable:
				continue
			case SatSatisfiable:
				c, err := NewCounterexample(m.graph.RegisterCount(), m.graph.PrimaryInputCount(), frame+1)
				if err != nil {
					return interrupted(err)
				}
				c.Property = property
				value := func(l GiaLit) (bool, error) {
					if !l.Valid() || uint64(l.ID()) >= uint64(len(giaToAig)) {
						return false, ErrInvariant
					}
					a := giaToAig[l.ID()].NotIf(l.Complemented())
					sat, err := relation.ObjectLiteral(a.ID())
					if err != nil {
						return false, err
					}
					sat ^= SatLit(a & 1)
					if uint64(sat.Var()) >= uint64(len(answer.Model)) {
						return false, ErrInvariant
					}
					return answer.Model[sat.Var()] != sat.Negative(), nil
				}
				for reg, l := range unrolled.initial {
					v, err := value(l)
					if err != nil {
						return result, err
					}
					c.set(reg, v)
				}
				for f := 0; f <= frame; f++ {
					for input := 0; input < c.inputs; input++ {
						v, err := value(unrolled.frameInputs[f*c.inputs+input])
						if err != nil {
							return result, err
						}
						c.set(c.registers+f*c.inputs+input, v)
					}
				}
				if err := m.ValidateCounterexample(c); err != nil {
					return result, fmt.Errorf("%w: bounded SAT witness failed replay: %v", ErrInvariant, err)
				}
				result.Status = BmcCounterexample
				result.Reason = BmcWitnessFound
				result.Property = property
				result.Counterexample = c
				return result, nil
			default:
				return result, ErrInvariant
			}
		}
		result.CheckedFrames = frame + 1
	}
	return complete(), nil
}
