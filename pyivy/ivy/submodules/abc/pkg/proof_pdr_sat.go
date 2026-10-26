package abc

// Port of the concrete frame-solver operations in src/proof/pdr/pdrSat.c
// and pdrCnf.c. This initial path uses a complete simple CNF relation rather
// than lazy cut CNF, and preserves literal mappings explicitly.

import (
	"encoding/binary"
	"errors"

	"github.com/glycerine/insdict"
)

var errPdrStopped = errors.New("PDR search stopped")

type pdrFrame struct {
	active  bool
	initial bool
	solver  *SatSolver
	clauses *insdict.Dict[string, *PdrCube]
}
type pdrEngine struct {
	model                                          *AigerModel
	options                                        PdrOptions
	relation                                       *AigCNF
	current, next, inputs, properties, constraints []SatLit
	frames                                         []*pdrFrame
	stats                                          PdrStats
	reason                                         PdrStopReason
	satReason                                      SatStopReason
	priority                                       int
	flopPriority                                   []uint64
	priorityStep                                   uint64
}

func pdrCubeKey(c *PdrCube) string {
	data := make([]byte, 0, 4*len(c.state))
	for _, l := range c.state {
		data = binary.LittleEndian.AppendUint32(data, uint32(l))
	}
	return string(data)
}
func (p *pdrEngine) stop(reason PdrStopReason, sat SatStopReason) error {
	p.reason = reason
	p.satReason = sat
	return errPdrStopped
}
func (p *pdrEngine) checkStop() error {
	if p.options.SAT.Stop != nil && p.options.SAT.Stop() {
		return p.stop(PdrCancelled, SatCancelled)
	}
	return nil
}
func newPdrEngine(model *AigerModel, options PdrOptions) (*pdrEngine, error) {
	relation, err := EncodeAigCNF(model.graph)
	if err != nil {
		return nil, err
	}
	if relation.variables > options.MaxVariables {
		return nil, ErrLimit
	}
	pi, po := model.graph.PrimaryInputCount(), model.graph.PrimaryOutputCount()
	p := &pdrEngine{model: model, options: options, relation: relation, priority: 1 << 30, flopPriority: make([]uint64, model.graph.RegisterCount())}
	if options.StructuralPriorities {
		if model.graph.ObjectCount() > options.MaxPriorityObjects {
			return nil, ErrLimit
		}
		if err := p.checkStop(); err != nil {
			return nil, ErrCancelled
		}
		gia, _, err := GiaFromAig(model.graph)
		if err != nil {
			return nil, err
		}
		p.flopPriority, err = gia.StructuralFlopPriorities(true, options.MaxPriorityObjects)
		if err != nil {
			return nil, err
		}
		p.priorityStep = 1
		for p.priorityStep < uint64(model.graph.RegisterCount()) {
			p.priorityStep <<= 1
		}
		if err := p.checkStop(); err != nil {
			return nil, ErrCancelled
		}
	}
	p.inputs = relation.inputs[:pi]
	p.current = relation.inputs[pi:]
	p.next = relation.outputs[po:]
	p.properties = relation.outputs[:model.outputs+model.bad]
	p.constraints = relation.outputs[model.outputs+model.bad : po]
	return p, nil
}
func (p *pdrEngine) newFrame(initial bool) error {
	if err := p.checkStop(); err != nil {
		return err
	}
	total := p.relation.variables
	for _, f := range p.frames {
		var err error
		total, err = checkedSize(total, 1, f.solver.VariableCount(), p.options.MaxFrameVariables)
		if err != nil {
			return err
		}
	}
	if total > p.options.MaxFrameVariables {
		return ErrLimit
	}
	solver, err := NewSatSolverWithLimit(p.options.MaxVariables)
	if err != nil {
		return err
	}
	if err := solver.SetRandomSeed(p.options.RandomSeed); err != nil {
		return err
	}
	frame := &pdrFrame{initial: initial, solver: solver, clauses: insdict.NewDict[string, *PdrCube]()}
	// Retain it before loading so interrupted preprocessing remains accounted for.
	p.frames = append(p.frames, frame)
	p.stats.Frames = len(p.frames)
	return p.loadFrame(frame, initial)
}

func (p *pdrEngine) loadFrame(frame *pdrFrame, initial bool) error {
	solver := frame.solver
	_, err := p.relation.loadIntoWithStop(solver, 0, p.options.SAT.Stop)
	if err != nil {
		return err
	}
	for _, l := range p.constraints {
		if err := p.checkStop(); err != nil {
			return err
		}
		if _, err := solver.AddClause([]SatLit{l}); err != nil {
			return err
		}
	}
	if initial {
		for i, init := range p.model.initial {
			if init == LatchInitUnknown {
				continue
			}
			l := p.current[i]
			if init == LatchInitZero {
				l = l.Not()
			}
			if _, err := solver.AddClause([]SatLit{l}); err != nil {
				return err
			}
		}
	}
	for _, cube := range frame.clauses.All() {
		if err := p.checkStop(); err != nil {
			return err
		}
		literals, err := p.cubeLiterals(cube, false, true)
		if err != nil {
			return err
		}
		if _, err = solver.AddClause(literals); err != nil {
			return err
		}
	}

	return nil
}
func (p *pdrEngine) cubeLiterals(cube *PdrCube, next, complement bool) ([]SatLit, error) {
	if cube == nil {
		return nil, ErrDimension
	}
	table := p.current
	if next {
		table = p.next
	}
	result := make([]SatLit, len(cube.state))
	for i, l := range cube.state {
		if !l.Valid() || uint64(l.Index()) >= uint64(len(table)) {
			return nil, ErrInvalidLiteral
		}
		mapped := table[l.Index()]
		if l.Negative() != complement {
			mapped = mapped.Not()
		}
		result[i] = mapped
	}
	return result, nil
}
func (p *pdrEngine) query(frame int, assumptions []SatLit) (SatResult, error) {
	if err := p.recycleFrame(frame); err != nil {
		return SatResult{}, err
	}
	if err := p.checkStop(); err != nil {
		return SatResult{}, err
	}
	if p.stats.Queries >= p.options.MaxQueries {
		return SatResult{}, p.stop(PdrQueryLimit, SatFinished)
	}
	p.stats.Queries++
	answer, err := p.frames[frame].solver.Solve(assumptions, p.options.SAT)
	if err != nil {
		return answer, err
	}
	if answer.Status == SatUnknown {
		reason := PdrSatLimit
		if answer.Reason == SatCancelled {
			reason = PdrCancelled
		}
		return answer, p.stop(reason, answer.Reason)
	}
	return answer, nil
}
func (p *pdrEngine) queryCube(frame int, cube *PdrCube, next bool) (SatResult, error) {
	literals, err := p.queryCubeLiterals(cube, next)
	if err != nil {
		return SatResult{}, err
	}
	return p.query(frame, literals)
}
func (p *pdrEngine) modelCube(values []bool) (*PdrCube, error) {
	value := func(l SatLit) (bool, error) {
		if uint64(l.Var()) >= uint64(len(values)) {
			return false, ErrInvariant
		}
		return values[l.Var()] != l.Negative(), nil
	}
	state, inputs := make([]PdrLit, len(p.current)), make([]PdrLit, len(p.inputs))
	for i, l := range p.current {
		v, err := value(l)
		if err != nil {
			return nil, err
		}
		state[i] = PdrLit(2 * i)
		if !v {
			state[i] = state[i].Not()
		}
	}
	for i, l := range p.inputs {
		v, err := value(l)
		if err != nil {
			return nil, err
		}
		inputs[i] = PdrLit(2 * i)
		if !v {
			inputs[i] = inputs[i].Not()
		}
	}
	return NewPdrCube(state, inputs)
}
func (p *pdrEngine) addBlocked(frame int, cube *PdrCube) error {
	if frame < 1 || frame >= len(p.frames) {
		return ErrDimension
	}
	key := pdrCubeKey(cube)
	if _, exists := p.frames[frame].clauses.Get2(key); exists {
		return nil
	}
	// Negated cubes are clauses: a smaller cube is the stronger clause.
	// Skip an insertion already implied by a retained stronger clause.
	redundant := make([]string, 0)
	for oldKey, old := range p.frames[frame].clauses.All() {
		if err := p.checkStop(); err != nil {
			return err
		}
		if cube.Contains(old) {
			p.stats.ContainedClauses++
			return nil
		}
		if old.Contains(cube) {
			redundant = append(redundant, oldKey)
		}
	}
	intersects, err := cube.IntersectsInitial(p.model.initial)
	if err != nil {
		return err
	}
	if intersects {
		return ErrInvariant
	}
	if p.stats.ClauseAdds >= p.options.MaxClauses {
		return p.stop(PdrClauseLimit, SatFinished)
	}
	literals, err := p.cubeLiterals(cube, false, true)
	if err != nil {
		return err
	}
	ok, err := p.frames[frame].solver.AddClause(literals)
	if err != nil {
		return err
	}
	// Each learned clause holds for every initial state. If initial states have
	// admissible inputs, these clauses cannot make a frame inconsistent.
	if !ok {
		return ErrInvariant
	}
	// SAT may keep the old clauses until recycling: the new clause implies
	// every removed clause, so both solver representations are equivalent.
	for _, oldKey := range redundant {
		p.frames[frame].clauses.Del(oldKey)
		p.stats.SubsumedClauses++
	}
	p.frames[frame].clauses.Put(key, cube)
	p.stats.ClauseAdds++
	return nil
}
func (p *pdrEngine) alreadyBlocked(frame int, cube *PdrCube) bool {
	for _, blocked := range p.frames[frame].clauses.All() {
		if cube.Contains(blocked) {
			return true
		}
	}
	return false
}
func (p *pdrEngine) equalFrames(a, b int) bool {
	x, y := p.frames[a].clauses, p.frames[b].clauses
	if x.Len() != y.Len() {
		return false
	}
	for key := range x.All() {
		if _, exists := y.Get2(key); !exists {
			return false
		}
	}
	return true
}
func pdrAddSatStats(a, b SatStats) SatStats {
	return SatStats{Conflicts: a.Conflicts + b.Conflicts, Propagations: a.Propagations + b.Propagations, Decisions: a.Decisions + b.Decisions, Inspections: a.Inspections + b.Inspections, Restarts: a.Restarts + b.Restarts, Learnt: a.Learnt + b.Learnt, Deleted: a.Deleted + b.Deleted}
}

// reduceClause ports Pdr_ManReduceClause: translate an UNSAT assumption
// conflict back into register literals, then retain an initial-state blocker.
// Next-state signals can alias, be complemented, or be constant expressions;
// consequently SAT-variable numbers alone cannot identify register literals.
func (p *pdrEngine) reduceClause(cube *PdrCube, conflict []SatLit) (*PdrCube, error) {
	mapped, err := p.cubeLiterals(cube, true, false)
	if err != nil {
		return nil, err
	}
	selected := make([]bool, len(mapped))
	// Lookup only: canonical cube order determines the first representative
	// of aliased literals; map iteration never determines a result.
	first := make(map[SatLit]int, len(mapped))
	for i, literal := range mapped {
		if err := p.checkStop(); err != nil {
			return nil, err
		}
		if _, exists := first[literal.Not()]; !exists {
			first[literal.Not()] = i
		}
	}
	for _, failed := range conflict {
		if err := p.checkStop(); err != nil {
			return nil, err
		}
		i, exists := first[failed]
		if !exists {
			return nil, ErrInvariant
		}
		selected[i] = true
	}
	state := make([]PdrLit, 0, len(mapped))
	for i, l := range cube.state {
		if selected[i] {
			state = append(state, l)
		}
	}
	reduced, err := NewPdrCube(state, nil)
	if err != nil {
		return nil, err
	}
	overlaps, err := reduced.IntersectsInitial(p.model.initial)
	if err != nil {
		return nil, err
	}
	if overlaps {
		// C assumes all-zero initialization and restores a positive literal.
		// Here restore the first literal excluding the actual reset predicate.
		restored := false
		for i, l := range cube.state {
			init := p.model.initial[l.Index()]
			excludes := (init == LatchInitZero && !l.Negative()) || (init == LatchInitOne && l.Negative())
			if excludes {
				if !selected[i] {
					state = append(state, l)
				}
				restored = true
				break
			}
		}
		if !restored {
			return nil, ErrInvariant
		}
		reduced, err = NewPdrCube(state, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(reduced.state) < len(cube.state) {
		p.stats.CoreReductions++
	}
	return reduced, nil
}

// queryRelativeCube checks F & !cube & T & cube'. The guarded clause is
// disabled on every exit, including cancellation and resource exhaustion.
// Activation variables never enter the register/PI mappings or learned cubes.
func (p *pdrEngine) queryRelativeCube(frame int, cube *PdrCube) (answer SatResult, err error) {
	if err = p.checkStop(); err != nil {
		return answer, err
	}
	if err = p.recycleFrame(frame); err != nil {
		return answer, err
	}
	clause, err := p.cubeLiterals(cube, false, true)
	if err != nil {
		return answer, err
	}
	assumptions, err := p.queryCubeLiterals(cube, true)
	if err != nil {
		return answer, err
	}
	solver := p.frames[frame].solver
	total := 0
	for _, f := range p.frames {
		total, err = checkedSize(total, 1, f.solver.VariableCount(), p.options.MaxFrameVariables)
		if err != nil {
			return answer, err
		}
	}
	if total >= p.options.MaxFrameVariables {
		return answer, ErrLimit
	}
	variable := solver.VariableCount()
	if err = solver.SetVariableCount(variable + 1); err != nil {
		return answer, err
	}
	activation := SatLit(2 * variable)
	clause = append(clause, activation)
	if _, err = solver.AddClause(clause); err != nil {
		return answer, err
	}
	p.frames[frame].active = true
	defer func() {
		p.frames[frame].active = false
		_, cleanup := solver.AddClause([]SatLit{activation})
		if cleanup != nil && err == nil {
			err = cleanup
		}
	}()
	assumptions = append(assumptions, activation.Not())
	return p.query(frame, assumptions)
}

// recycleFrame ports Pdr_ManFetchSolver. Rebuild the exact frame relation,
// reset conditions and ordered blocked cubes, dropping only solver-derived
// clauses and retired guards. Never rebuild while a guarded query is active.
func (p *pdrEngine) recycleFrame(index int) error {
	frame := p.frames[index]
	if frame.active || frame.solver.VariableCount()-p.relation.variables < p.options.RecycleVariables {
		return nil
	}
	if err := p.checkStop(); err != nil {
		return err
	}
	p.stats.SAT = pdrAddSatStats(p.stats.SAT, frame.solver.Statistics())
	fresh, err := NewSatSolverWithLimit(p.options.MaxVariables)
	if err != nil {
		return err
	}
	frame.solver = fresh
	if err := frame.solver.SetRandomSeed(p.options.RandomSeed); err != nil {
		return err
	}
	p.stats.SolverRecycles++
	return p.loadFrame(frame, frame.initial)
}
