package abc

// Invariant validation for the port of src/proof/pdr/pdrInv.c. This checker
// rebuilds Boolean miters from the original AIG; it does not reuse PDR frame
// solvers, frame clauses, or cube-to-SAT mappings.

import (
	"errors"
	"slices"
)

type PdrInvariant struct{ clauses []*PdrCube }

// NewPdrInvariant treats each cube as a blocked conjunction: its negation is
// an invariant clause. No cubes means true; one empty cube means false.
func NewPdrInvariant(cubes []*PdrCube) (*PdrInvariant, error) {
	result := &PdrInvariant{clauses: make([]*PdrCube, len(cubes))}
	for i, c := range cubes {
		if err := c.Check(); err != nil {
			return nil, err
		}
		copy, err := NewPdrCube(c.state, nil)
		if err != nil {
			return nil, err
		}
		result.clauses[i] = copy
	}
	return result, nil
}
func (i *PdrInvariant) Clauses() []*PdrCube { return slices.Clone(i.clauses) }

// Holds evaluates the state predicate independently of any SAT solver.
func (i *PdrInvariant) Holds(state []bool) (bool, error) {
	if i == nil {
		return false, ErrDimension
	}
	for _, cube := range i.clauses {
		matches, err := cube.Matches(state)
		if err != nil {
			return false, err
		}
		if matches {
			return false, nil
		}
	}
	return true, nil
}

type PdrInvariantStatus uint8

const (
	PdrInvariantUnknown PdrInvariantStatus = iota
	PdrInvariantValid
	PdrInvariantInvalid
)

type PdrInvariantPhase uint8

const (
	PdrInvariantComplete PdrInvariantPhase = iota
	PdrInvariantInitiation
	PdrInvariantInduction
	PdrInvariantSafety
)

type PdrVerifyOptions struct {
	MaxObjects, MaxVariables int
	// Zero selects all three conditions. A smaller positive cap can interrupt
	// verification, but cannot validate a certificate from a partial check.
	MaxQueries int
	SAT        SatLimits
}
type PdrInvariantResult struct {
	Status     PdrInvariantStatus
	Phase      PdrInvariantPhase
	Reason     SatStopReason
	QueryLimit bool
	Queries    uint64
	Stats      SatStats
}

// Verify checks initiation, consecution and selected-property safety. Positive
// environment constraints guard each step, including frame zero; thus a model
// with no admissible initial input has no valid trace. No proof is returned
// from an incomplete, interrupted or budget-limited certificate check.
func (invariant *PdrInvariant) Verify(model *AigerModel, property int, options PdrVerifyOptions) (result PdrInvariantResult, err error) {
	if invariant == nil {
		return result, ErrDimension
	}
	if err := model.Validate(); err != nil {
		return result, err
	}
	if property < 0 || property >= model.outputs+model.bad || options.MaxObjects < 0 || options.MaxVariables < 0 || options.MaxQueries < 0 {
		return result, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxQueries == 0 {
		options.MaxQueries = 3
	}
	stopped := func() bool { return options.SAT.Stop != nil && options.SAT.Stop() }
	if stopped() {
		result.Reason = SatCancelled
		return result, nil
	}
	original := model.graph
	if original.ObjectCount() > options.MaxObjects {
		return result, ErrLimit
	}
	for _, cube := range invariant.clauses {
		if err := cube.Check(); err != nil {
			return result, err
		}
		for _, l := range cube.state {
			if uint64(l.Index()) >= uint64(original.RegisterCount()) {
				return result, ErrDimension
			}
		}
	}
	g := original.Clone()
	if err := g.SetConstraintCount(0); err != nil {
		return result, err
	}
	if err := g.SetRegisterCount(0); err != nil {
		return result, err
	}
	and := func(a, b AigLit) (AigLit, error) {
		if stopped() {
			return InvalidAigLit, ErrCancelled
		}
		l, err := g.And(a, b)
		if err != nil {
			return InvalidAigLit, err
		}
		if g.ObjectCount() > options.MaxObjects {
			return InvalidAigLit, ErrLimit
		}
		return l, nil
	}
	or := func(a, b AigLit) (AigLit, error) {
		l, err := and(a.Not(), b.Not())
		if err != nil {
			return InvalidAigLit, err
		}
		return l.Not(), nil
	}
	current, next := make([]AigLit, original.registers), make([]AigLit, original.registers)
	for r := range current {
		current[r] = AigLit(original.cis[original.PrimaryInputCount()+r]) << 1
		next[r] = original.objects[original.cos[original.PrimaryOutputCount()+r]].Fanin0
	}
	predicate := func(signals []AigLit) (AigLit, error) {
		total := AigConst1
		for _, cube := range invariant.clauses {
			clause := AigConst0
			for _, l := range cube.state {
				var err error
				clause, err = or(clause, signals[l.Index()].NotIf(!l.Negative()))
				if err != nil {
					return InvalidAigLit, err
				}
			}
			var err error
			total, err = and(total, clause)
			if err != nil {
				return InvalidAigLit, err
			}
		}
		return total, nil
	}
	build := func() ([]AigLit, error) {
		now, err := predicate(current)
		if err != nil {
			return nil, err
		}
		later, err := predicate(next)
		if err != nil {
			return nil, err
		}
		initial, environment := AigConst1, AigConst1
		for r, v := range model.initial {
			if v == LatchInitUnknown {
				continue
			}
			initial, err = and(initial, current[r].NotIf(v == LatchInitZero))
			if err != nil {
				return nil, err
			}
		}
		for out := model.outputs + model.bad; out < original.PrimaryOutputCount(); out++ {
			environment, err = and(environment, original.objects[original.cos[out]].Fanin0)
			if err != nil {
				return nil, err
			}
		}
		initAllowed, err := and(initial, environment)
		if err != nil {
			return nil, err
		}
		initiation, err := and(initAllowed, now.Not())
		if err != nil {
			return nil, err
		}
		currentAllowed, err := and(now, environment)
		if err != nil {
			return nil, err
		}
		induction, err := and(currentAllowed, later.Not())
		if err != nil {
			return nil, err
		}
		safety, err := and(currentAllowed, original.objects[original.cos[property]].Fanin0)
		if err != nil {
			return nil, err
		}
		return []AigLit{initiation, induction, safety}, nil
	}
	miters, err := build()
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			result.Reason = SatCancelled
			return result, nil
		}
		return result, err
	}
	for _, miter := range miters {
		if _, err := g.CreateCO(miter); err != nil {
			return result, err
		}
		if g.ObjectCount() > options.MaxObjects {
			return result, ErrLimit
		}
	}
	if stopped() {
		result.Reason = SatCancelled
		return result, nil
	}
	cnf, err := EncodeAigCNF(g)
	if err != nil {
		return result, err
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return result, err
	}
	defer func() { result.Stats = solver.Statistics() }()
	_, err = cnf.loadIntoWithStop(solver, 0, options.SAT.Stop)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			result.Reason = SatCancelled
			return result, nil
		}
		return result, err
	}
	checks := cnf.outputs[len(cnf.outputs)-3:]
	for index, l := range checks {
		result.Phase = PdrInvariantPhase(index + 1)
		if int(result.Queries) >= options.MaxQueries {
			result.QueryLimit = true
			return result, nil
		}
		result.Queries++
		answer, err := solver.Solve([]SatLit{l}, options.SAT)
		if err != nil {
			return result, err
		}
		if answer.Status == SatUnknown {
			result.Reason = answer.Reason
			return result, nil
		}
		if answer.Status == SatSatisfiable {
			result.Status = PdrInvariantInvalid
			return result, nil
		}
	}
	result.Status = PdrInvariantValid
	result.Phase = PdrInvariantComplete
	return result, nil
}
