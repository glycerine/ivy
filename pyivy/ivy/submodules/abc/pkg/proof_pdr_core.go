package abc

// Port of the basic IC3/PDR blocking, generalization, propagation and fixed
// point loop in src/proof/pdr/pdrCore.c. Advanced ABC scheduling, abstraction,
// ternary shrinking and lazy/optimized CNF remain separate port work.

import (
	"errors"
	"fmt"
)

type PdrStatus uint8

const (
	PdrUnknown PdrStatus = iota
	PdrProved
	PdrCounterexample
)

type PdrStopReason uint8

const (
	PdrProof PdrStopReason = iota
	PdrWitnessFound
	PdrFrameLimit
	PdrQueryLimit
	PdrClauseLimit
	PdrObligationLimit
	PdrSatLimit
	PdrCancelled
	PdrResourceLimit
)

// PdrOptions selects one ordinary/bad output and finite search resources.
// Zero caps select defaults: 256 frontier frames, 100K queries, 1M clause
// insertions/obligation pushes, and 1M variables per solver and across frames.
// MaxFrames permits one additional lookahead frame for clause propagation.
// SAT limits apply per query. Stop must not mutate/reenter the model.
type PdrOptions struct {
	StructuralPriorities bool
	MaxPriorityObjects   int
	// RecycleVariables defaults to ABC's 300 retired activation variables.
	RecycleVariables                                     int
	Property, MaxFrames, MaxVariables, MaxFrameVariables int
	MaxQueries, MaxClauses, MaxObligations               uint64
	RandomSeed                                           float64
	DisableGeneralization                                bool
	// OrderAssumptions prefers high-priority registers in next-state SAT queries.
	OrderAssumptions bool
	SAT              SatLimits
}
type PdrStats struct {
	ContainedClauses, SubsumedClauses                                               uint64
	SolverRecycles                                                                  uint64
	Frames                                                                          int
	Queries, Obligations, ClauseAdds, Generalizations, ClausePushes, CoreReductions uint64
	PeakQueue                                                                       int
	SAT                                                                             SatStats
}

// PdrResult reports unbounded safety only with a checked inductive invariant.
// Frame is the frontier reached, or the failing trace's last frame. A witness
// is valid but is not promised to be shortest.
type PdrResult struct {
	Status          PdrStatus
	Reason          PdrStopReason
	SatReason       SatStopReason
	Frame, Property int
	Invariant       *PdrInvariant
	Counterexample  *Counterexample
	Stats           PdrStats
}

func (o PdrOptions) normalized() (PdrOptions, error) {
	if o.MaxPriorityObjects < 0 || o.RecycleVariables < 0 || o.MaxFrames < 0 || o.MaxVariables < 0 || o.MaxFrameVariables < 0 {
		return o, ErrDimension
	}
	if o.MaxPriorityObjects == 0 {
		o.MaxPriorityObjects = 1 << 20
	}
	if o.RecycleVariables == 0 {
		o.RecycleVariables = 300
	}
	if o.MaxFrames == 0 {
		o.MaxFrames = 256
	}
	if o.MaxVariables == 0 {
		o.MaxVariables = 1 << 20
	}
	if o.MaxFrameVariables == 0 {
		o.MaxFrameVariables = 1 << 20
	}
	if o.MaxQueries == 0 {
		o.MaxQueries = 100000
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 20
	}
	if o.MaxObligations == 0 {
		o.MaxObligations = 1 << 20
	}
	if o.RandomSeed == 0 {
		o.RandomSeed = 91648253
	}
	probe, err := NewSatSolverWithLimit(o.MaxVariables)
	if err != nil {
		return o, err
	}
	if err := probe.SetRandomSeed(o.RandomSeed); err != nil {
		return o, err
	}
	return o, nil
}
func (p *pdrEngine) pushObligation(queue *PdrQueue, o *PdrObligation) error {
	if p.stats.Obligations >= p.options.MaxObligations {
		return p.stop(PdrObligationLimit, SatFinished)
	}
	if err := queue.Push(o); err != nil {
		return err
	}
	p.stats.Obligations++
	p.stats.PeakQueue = max(p.stats.PeakQueue, queue.Len())
	return nil
}
func (p *pdrEngine) generalize(frame int, cube *PdrCube, conflict []SatLit) (*PdrCube, error) {
	current, err := NewPdrCube(cube.state, nil)
	if err != nil {
		return nil, err
	}
	if p.options.DisableGeneralization {
		return current, nil
	}
	current, err = p.reduceClause(cube, conflict)
	if err != nil {
		return nil, err
	}
	order, err := p.orderCube(current)
	if err != nil {
		return nil, err
	}
	for _, literal := range order {
		index := -1
		for i, l := range current.state {
			if l == literal {
				index = i
				break
			}
		}
		if index < 0 {
			continue
		}
		if err := p.checkStop(); err != nil {
			return nil, err
		}
		candidate, err := current.Without(index)
		if err != nil {
			return nil, err
		}
		intersects, err := candidate.IntersectsInitial(p.model.initial)
		if err != nil {
			return nil, err
		}
		if intersects {
			continue
		}
		answer, err := p.queryRelativeCube(frame, candidate)
		if err != nil {
			return nil, err
		}
		if answer.Status == SatUnsatisfiable {
			current = candidate
			p.stats.Generalizations++
		}
	}
	return current, nil
}
func (p *pdrEngine) trace(obligation *PdrObligation) (*Counterexample, error) {
	count := 0
	for o := obligation; o != nil; o = o.next {
		count++
		if uint64(count) > p.options.MaxObligations {
			return nil, ErrLimit
		}
	}
	if count == 0 {
		return nil, ErrInvariant
	}
	c, err := NewCounterexample(len(p.current), len(p.inputs), count)
	if err != nil {
		return nil, err
	}
	c.Property = p.options.Property
	for _, l := range obligation.state.state {
		if uint64(l.Index()) >= uint64(c.registers) {
			return nil, ErrInvariant
		}
		c.set(int(l.Index()), !l.Negative())
	}
	frame := 0
	for o := obligation; o != nil; o = o.next {
		if len(o.state.state) != c.registers || len(o.state.inputs) != c.inputs {
			return nil, ErrInvariant
		}
		for _, l := range o.state.inputs {
			if uint64(l.Index()) >= uint64(c.inputs) {
				return nil, ErrInvariant
			}
			c.set(c.registers+frame*c.inputs+int(l.Index()), !l.Negative())
		}
		frame++
	}
	if err := p.model.ValidateCounterexample(c); err != nil {
		return nil, fmt.Errorf("%w: PDR witness failed original-model replay: %v", ErrInvariant, err)
	}
	return c, nil
}
func (p *pdrEngine) block(frontier int, cube *PdrCube) (*Counterexample, error) {
	var queue PdrQueue
	defer queue.Clear()
	root, err := NewPdrObligation(frontier, p.priority, cube, nil)
	if err != nil {
		return nil, err
	}
	p.priority--
	if err := p.pushObligation(&queue, root); err != nil {
		return nil, err
	}
	for !queue.Empty() {
		if err := p.checkStop(); err != nil {
			return nil, err
		}
		current := queue.Pop()
		initial, err := current.state.IntersectsInitial(p.model.initial)
		if err != nil {
			return nil, err
		}
		// Obligations contain complete states and PI assignments. An initial state
		// anywhere in the chain gives a concrete shorter trace, even if its frame
		// label is nonzero (important when the transition has no predecessor).
		if initial {
			return p.trace(current)
		}
		if current.frame == 0 {
			return nil, ErrInvariant
		}
		if p.alreadyBlocked(current.frame, current.state) {
			continue
		}
		answer, err := p.queryCube(current.frame-1, current.state, true)
		if err != nil {
			return nil, err
		}
		if answer.Status == SatSatisfiable {
			predecessor, err := p.modelCube(answer.Model)
			if err != nil {
				return nil, err
			}
			if err := current.Reschedule(current.frame, p.priority); err != nil {
				return nil, err
			}
			p.priority--
			if err := p.pushObligation(&queue, current); err != nil {
				return nil, err
			}
			next, err := NewPdrObligation(current.frame-1, p.priority, predecessor, current)
			if err != nil {
				return nil, err
			}
			p.priority--
			if err := p.pushObligation(&queue, next); err != nil {
				return nil, err
			}
			continue
		}
		generalized, err := p.generalize(current.frame-1, current.state, answer.Conflict)
		if err != nil {
			return nil, err
		}
		if err := p.bumpCubePriority(generalized); err != nil {
			return nil, err
		}
		// F_i implies F_(i+1): learning at level j strengthens every frame 1..j.
		for frame := 1; frame <= current.frame; frame++ {
			if err := p.addBlocked(frame, generalized); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}
func (p *pdrEngine) propagate(frontier int) (int, error) {
	for frame := 1; frame <= frontier; frame++ {
		// A pushed core may strengthen this frame and remove entries. Iterate
		// an owned snapshot rather than mutating an active dictionary iterator.
		cubes := make([]*PdrCube, 0, p.frames[frame].clauses.Len())
		for _, cube := range p.frames[frame].clauses.All() {
			cubes = append(cubes, cube)
		}
		for _, cube := range cubes {
			if p.alreadyBlocked(frame+1, cube) {
				continue
			}
			answer, err := p.queryCube(frame, cube, true)
			if err != nil {
				return -1, err
			}
			if answer.Status == SatUnsatisfiable {
				reduced, err := p.reduceClause(cube, answer.Conflict)
				if err != nil {
					return -1, err
				}
				// The stronger pushed clause must also strengthen earlier cumulative
				// frames, preserving F_i => F_(i+1) after core reduction.
				for earlier := 1; earlier <= frame+1; earlier++ {
					if err := p.addBlocked(earlier, reduced); err != nil {
						return -1, err
					}
				}
				p.stats.ClausePushes++
			}
		}
		if p.equalFrames(frame, frame+1) {
			return frame, nil
		}
	}
	return -1, nil
}
func (p *pdrEngine) invariant(frame int) (*PdrInvariant, error) {
	cubes := make([]*PdrCube, 0, p.frames[frame].clauses.Len())
	for _, cube := range p.frames[frame].clauses.All() {
		cubes = append(cubes, cube)
	}
	return NewPdrInvariant(cubes)
}

// Prove runs a concrete IC3/PDR safety search. It is not bounded model
// checking: proof requires clause-frame convergence and independent invariant
// validation; bugs require a replayed original-model trace. Unsupported ABC
// optimization/abstraction/portfolio modes are not exposed by this API.
func (m *AigerModel) Prove(options PdrOptions) (result PdrResult, err error) {
	result.Frame = -1
	result.Property = options.Property
	if err := m.Validate(); err != nil {
		return result, err
	}
	options, err = options.normalized()
	if err != nil {
		return result, err
	}
	if options.Property < 0 || options.Property >= m.outputs+m.bad {
		return result, ErrDimension
	}
	if options.SAT.Stop != nil && options.SAT.Stop() {
		result.Reason = PdrCancelled
		result.SatReason = SatCancelled
		return result, nil
	}
	p, err := newPdrEngine(m, options)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			result.Reason = PdrCancelled
			result.SatReason = SatCancelled
			return result, nil
		}
		if errors.Is(err, ErrLimit) {
			result.Reason = PdrResourceLimit
			return result, nil
		}
		return result, err
	}
	var verifierStats SatStats
	defer func() {
		result.Stats = p.stats
		for _, frame := range p.frames {
			result.Stats.SAT = pdrAddSatStats(result.Stats.SAT, frame.solver.Statistics())
		}
		result.Stats.SAT = pdrAddSatStats(result.Stats.SAT, verifierStats)
	}()
	interrupted := func(cause error) (PdrResult, error) {
		switch {
		case errors.Is(cause, errPdrStopped):
			result.Reason = p.reason
			result.SatReason = p.satReason
			return result, nil
		case errors.Is(cause, ErrCancelled):
			result.Reason = PdrCancelled
			result.SatReason = SatCancelled
			return result, nil
		case errors.Is(cause, ErrLimit):
			result.Reason = PdrResourceLimit
			return result, nil
		default:
			return result, cause
		}
	}
	proof := func(invariant *PdrInvariant) (PdrResult, error) {
		if p.stats.Queries >= options.MaxQueries {
			return interrupted(p.stop(PdrQueryLimit, SatFinished))
		}
		allowed := min(options.MaxQueries-p.stats.Queries, uint64(3))
		check, err := invariant.Verify(m, options.Property, PdrVerifyOptions{MaxVariables: options.MaxVariables, MaxQueries: int(allowed), SAT: options.SAT})
		verifierStats = check.Stats
		p.stats.Queries += check.Queries
		if err != nil {
			return interrupted(err)
		}
		if check.Status == PdrInvariantUnknown {
			if check.QueryLimit {
				result.Reason = PdrQueryLimit
				return result, nil
			}
			result.Reason = PdrSatLimit
			result.SatReason = check.Reason
			if check.Reason == SatCancelled {
				result.Reason = PdrCancelled
			}
			return result, nil
		}
		if check.Status != PdrInvariantValid {
			return result, fmt.Errorf("%w: PDR invariant failed phase %d", ErrInvariant, check.Phase)
		}
		result.Status = PdrProved
		result.Reason = PdrProof
		result.Invariant = invariant
		return result, nil
	}
	if err := p.newFrame(true); err != nil {
		return interrupted(err)
	}
	result.Frame = 0
	answer, err := p.query(0, []SatLit{p.properties[options.Property]})
	if err != nil {
		return interrupted(err)
	}
	if answer.Status == SatSatisfiable {
		cube, err := p.modelCube(answer.Model)
		if err != nil {
			return result, err
		}
		o, err := NewPdrObligation(0, 0, cube, nil)
		if err != nil {
			return result, err
		}
		trace, err := p.trace(o)
		if err != nil {
			return result, err
		}
		result.Status = PdrCounterexample
		result.Reason = PdrWitnessFound
		result.Counterexample = trace
		return result, nil
	}
	// No admissible initial input means no valid trace. The false invariant is
	// valid under the explicit environment assumptions and is still checked.
	if p.frames[0].solver.inconsistent {
		empty, err := NewPdrCube(nil, nil)
		if err != nil {
			return result, err
		}
		invariant, err := NewPdrInvariant([]*PdrCube{empty})
		if err != nil {
			return result, err
		}
		return proof(invariant)
	}
	if err := p.newFrame(false); err != nil {
		return interrupted(err)
	}
	for frontier := 1; frontier <= options.MaxFrames; frontier++ {
		result.Frame = frontier
		for {
			answer, err := p.query(frontier, []SatLit{p.properties[options.Property]})
			if err != nil {
				return interrupted(err)
			}
			if answer.Status == SatUnsatisfiable {
				break
			}
			cube, err := p.modelCube(answer.Model)
			if err != nil {
				return result, err
			}
			trace, err := p.block(frontier, cube)
			if err != nil {
				return interrupted(err)
			}
			if trace != nil {
				result.Status = PdrCounterexample
				result.Reason = PdrWitnessFound
				result.Frame = trace.LastFrame()
				result.Counterexample = trace
				return result, nil
			}
		}
		if err := p.newFrame(false); err != nil {
			return interrupted(err)
		}
		fixed, err := p.propagate(frontier)
		if err != nil {
			return interrupted(err)
		}
		if fixed >= 0 {
			invariant, err := p.invariant(fixed)
			if err != nil {
				return result, err
			}
			return proof(invariant)
		}
	}
	result.Reason = PdrFrameLimit
	return result, nil
}
