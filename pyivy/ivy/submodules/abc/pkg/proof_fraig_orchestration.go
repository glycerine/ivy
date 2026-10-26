package abc

// Derived from ABC src/proof/fraig/fraigApi.c, fraigChoice.c, fraigFeed.c,
// fraigMan.c, fraigSat.c, fraigUtil.c, and module.make.

import (
	"math/bits"
	"slices"
)

// FraigParameters records the source manager controls that have concrete Go
// meaning. Logical work limits are applied per SAT query instead of using a
// process-global timer.
type FraigParameters struct {
	RandomPatterns, DynamicPatterns int
	BacktrackLimit, Seconds         int
	FunctionalReduce, Feedback      bool
	DistanceOne, Sparse, Choices    bool
	TryProve                        bool
}

func DefaultFraigParameters() FraigParameters {
	return FraigParameters{2048, 2048, 99, 20, true, true, true, false, false, true}
}

func FullFraigParameters() FraigParameters {
	p := DefaultFraigParameters()
	p.BacktrackLimit = -1
	p.Sparse = true
	p.TryProve = false
	return p
}

type FraigOrchestrationOptions struct {
	Manager                 FraigOptions
	Parameters              FraigParameters
	MaxPatterns, MaxChoices int
	MaxClauses, MaxVisits   int
	Stop                    func() bool
}

func (o FraigOrchestrationOptions) normalized() (FraigOrchestrationOptions, error) {
	if o.MaxPatterns < 0 || o.MaxChoices < 0 || o.MaxClauses < 0 || o.MaxVisits < 0 {
		return o, ErrDimension
	}
	if o.Parameters == (FraigParameters{}) {
		o.Parameters = DefaultFraigParameters()
	}
	if o.Parameters.RandomPatterns < 0 || o.Parameters.DynamicPatterns < 0 || o.Parameters.BacktrackLimit < -1 || o.Parameters.Seconds < 0 {
		return o, ErrDimension
	}
	if o.Parameters.RandomPatterns < 128 {
		o.Parameters.RandomPatterns = 128
	}
	if o.Parameters.RandomPatterns > 32768 {
		o.Parameters.RandomPatterns = 32768
	}
	if o.Parameters.DynamicPatterns < 128 {
		o.Parameters.DynamicPatterns = 128
	}
	if o.Parameters.DynamicPatterns > 32768 {
		o.Parameters.DynamicPatterns = 32768
	}
	if !o.Parameters.FunctionalReduce {
		o.Parameters.RandomPatterns, o.Parameters.DynamicPatterns = 128, 128
	}
	if o.MaxPatterns == 0 {
		o.MaxPatterns = o.Parameters.DynamicPatterns
	}
	if o.MaxChoices == 0 {
		o.MaxChoices = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 24
	}
	if o.Manager.Stop == nil {
		o.Manager.Stop = o.Stop
	}
	o.Manager.FunctionalReduce = o.Parameters.FunctionalReduce
	return o, nil
}

type FraigRelationStatus uint8

const (
	FraigRelationUnknown FraigRelationStatus = iota
	FraigRelationProved
	FraigRelationDisproved
)

type FraigSATLimits struct {
	Conflicts, Propagations, Decisions uint64
	Stop                               func() bool
}

type FraigRelationResult struct {
	Status FraigRelationStatus
	Model  []bool
	Stats  SatStats
}

type FraigChoiceClass struct {
	Representative FraigEdge
	Alternatives   []FraigEdge
}

type FraigOrchestrationStats struct {
	Manager           FraigStats
	ChoiceClasses     int
	Choices, Patterns int
	Clauses           int
	SATQueries        uint64
	SATProved         uint64
	SATDisproved      uint64
	SATUnknown        uint64
}

// FraigOrchestrator owns manager policy, equivalence choices, feedback
// patterns, names, and additional user clauses. The underlying graph remains
// the concrete FraigManager used by the foundation.
type FraigOrchestrator struct {
	manager     *FraigManager
	options     FraigOrchestrationOptions
	inputNames  []string
	outputNames []string
	choices     map[FraigID][]FraigEdge
	represented map[FraigID]FraigID
	choiceRoots map[FraigID]FraigEdge
	patterns    [][]bool
	clauses     [][]FraigEdge
	lastModel   []bool
	queries     [3]uint64
	closed      bool
}

func NewFraigOrchestrator(options FraigOrchestrationOptions) (*FraigOrchestrator, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	m, err := NewFraigManager(o.Manager)
	if err != nil {
		return nil, err
	}
	return &FraigOrchestrator{manager: m, options: o, choices: make(map[FraigID][]FraigEdge), represented: make(map[FraigID]FraigID), choiceRoots: make(map[FraigID]FraigEdge)}, nil
}

type fraigManagerCheckpoint struct {
	nodes      []fraigNode
	inputs     []FraigEdge
	outputs    []FraigEdge
	structural map[fraigPair]FraigEdge
	functional map[fraigFunction]FraigEdge
}

func captureFraigManager(manager *FraigManager) fraigManagerCheckpoint {
	checkpoint := fraigManagerCheckpoint{
		nodes:      make([]fraigNode, len(manager.nodes)),
		inputs:     slices.Clone(manager.inputs),
		outputs:    slices.Clone(manager.outputs),
		structural: make(map[fraigPair]FraigEdge, len(manager.structural)),
		functional: make(map[fraigFunction]FraigEdge, len(manager.functional)),
	}
	copy(checkpoint.nodes, manager.nodes)
	for i := range checkpoint.nodes {
		checkpoint.nodes[i].fanouts = slices.Clone(manager.nodes[i].fanouts)
		checkpoint.nodes[i].randomSim = slices.Clone(manager.nodes[i].randomSim)
		checkpoint.nodes[i].dynamicSim = slices.Clone(manager.nodes[i].dynamicSim)
	}
	for key, edge := range manager.structural {
		checkpoint.structural[key] = edge
	}
	for key, edge := range manager.functional {
		checkpoint.functional[key] = edge
	}
	return checkpoint
}

func restoreFraigManager(manager *FraigManager, checkpoint fraigManagerCheckpoint) {
	manager.nodes = checkpoint.nodes
	manager.inputs = checkpoint.inputs
	manager.outputs = checkpoint.outputs
	manager.structural = checkpoint.structural
	manager.functional = checkpoint.functional
}

func (o *FraigOrchestrator) check() error {
	if o == nil || o.closed || o.manager == nil {
		return ErrInvariant
	}
	if o.options.Stop != nil && o.options.Stop() {
		return ErrCancelled
	}
	return o.manager.check()
}

func (o *FraigOrchestrator) Manager() *FraigManager { return o.manager }

func (o *FraigOrchestrator) NewInput(name string) (FraigEdge, error) {
	if err := o.check(); err != nil {
		return 0, err
	}
	edge, err := o.manager.NewInput()
	if err != nil {
		return 0, err
	}
	o.inputNames = append(o.inputNames, name)
	return edge, nil
}

func (o *FraigOrchestrator) Input(index int) (FraigEdge, error) {
	if err := o.check(); err != nil {
		return 0, err
	}
	if index < 0 || index >= o.manager.opts.MaxInputs {
		return 0, ErrDimension
	}
	checkpoint := captureFraigManager(o.manager)
	nameCount := len(o.inputNames)
	for len(o.manager.inputs) <= index {
		if _, err := o.NewInput(""); err != nil {
			restoreFraigManager(o.manager, checkpoint)
			o.inputNames = o.inputNames[:nameCount]
			return 0, err
		}
	}
	return o.manager.inputs[index], nil
}

func (o *FraigOrchestrator) And(a, b FraigEdge) (FraigEdge, error) {
	if err := o.check(); err != nil {
		return 0, err
	}
	return o.manager.And(a, b)
}

func (o *FraigOrchestrator) Or(a, b FraigEdge) (FraigEdge, error) {
	edge, err := o.And(a.Not(), b.Not())
	if err != nil {
		return 0, err
	}
	return edge.Not(), err
}

func (o *FraigOrchestrator) Xor(a, b FraigEdge) (FraigEdge, error) {
	return o.Mux(a, b.Not(), b)
}

func (o *FraigOrchestrator) Mux(control, whenTrue, whenFalse FraigEdge) (FraigEdge, error) {
	if err := o.check(); err != nil {
		return 0, err
	}
	checkpoint := captureFraigManager(o.manager)
	first, err := o.And(control, whenTrue)
	if err != nil {
		restoreFraigManager(o.manager, checkpoint)
		return 0, err
	}
	second, err := o.And(control.Not(), whenFalse)
	if err != nil {
		restoreFraigManager(o.manager, checkpoint)
		return 0, err
	}
	result, err := o.Or(first, second)
	if err != nil {
		restoreFraigManager(o.manager, checkpoint)
		return 0, err
	}
	return result, nil
}

func (o *FraigOrchestrator) AddOutput(edge FraigEdge, name string) error {
	if err := o.check(); err != nil {
		return err
	}
	if err := o.manager.AddOutput(edge); err != nil {
		return err
	}
	o.outputNames = append(o.outputNames, name)
	return nil
}

func (o *FraigOrchestrator) InputNames() []string  { return slices.Clone(o.inputNames) }
func (o *FraigOrchestrator) OutputNames() []string { return slices.Clone(o.outputNames) }
func (o *FraigOrchestrator) LastModel() []bool     { return slices.Clone(o.lastModel) }

func (o *FraigOrchestrator) AddClause(edges []FraigEdge) error {
	if err := o.check(); err != nil {
		return err
	}
	if len(edges) == 0 {
		return ErrDimension
	}
	if len(o.clauses) >= o.options.MaxClauses {
		return ErrLimit
	}
	for _, edge := range edges {
		if !o.manager.validEdge(edge) {
			return ErrDimension
		}
	}
	o.clauses = append(o.clauses, slices.Clone(edges))
	return nil
}

func fraigSATLiteral(edge FraigEdge) (SatLit, error) {
	return MakeSatLit(SatVar(edge.ID()), edge.Complemented())
}

func (o *FraigOrchestrator) combinedStop(local func() bool) func() bool {
	global := o.options.Stop
	switch {
	case global == nil:
		return local
	case local == nil:
		return global
	default:
		return func() bool { return global() || local() }
	}
}

func (o *FraigOrchestrator) newSolver(limits FraigSATLimits) (*SatSolver, error) {
	if err := o.check(); err != nil {
		return nil, err
	}
	// A caller stop is a SAT-query budget and is reported as an unknown solve.
	// The manager stop also governs solver construction and must be polled here.
	stop := o.options.Stop
	visits := 0
	tick := func() error {
		visits++
		if visits > o.options.MaxVisits {
			return ErrLimit
		}
		if stop != nil && stop() {
			return ErrCancelled
		}
		return nil
	}
	solver, err := NewSatSolverWithLimit(len(o.manager.nodes))
	if err != nil {
		return nil, err
	}
	if err = solver.SetVariableCount(len(o.manager.nodes)); err != nil {
		return nil, err
	}
	one, _ := fraigSATLiteral(o.manager.ConstantOne())
	if ok, addErr := solver.AddClause([]SatLit{one}); addErr != nil || !ok {
		if addErr != nil {
			return nil, addErr
		}
		return nil, ErrInvariant
	}
	for id := 1; id < len(o.manager.nodes); id++ {
		if err = tick(); err != nil {
			return nil, err
		}
		node := o.manager.nodes[id]
		if node.kind != FraigNodeAnd {
			continue
		}
		z, _ := fraigSATLiteral(makeFraigEdge(FraigID(id), false))
		a, _ := fraigSATLiteral(node.fanin[0])
		b, _ := fraigSATLiteral(node.fanin[1])
		for _, clause := range [][]SatLit{{a.Not(), b.Not(), z}, {a, z.Not()}, {b, z.Not()}} {
			ok, addErr := solver.AddClause(clause)
			if addErr != nil {
				return nil, addErr
			}
			if !ok {
				return nil, ErrInvariant
			}
		}
	}
	for _, source := range o.clauses {
		if err = tick(); err != nil {
			return nil, err
		}
		clause := make([]SatLit, len(source))
		for i, edge := range source {
			clause[i], _ = fraigSATLiteral(edge)
		}
		ok, addErr := solver.AddClause(clause)
		if addErr != nil {
			return nil, addErr
		}
		if !ok {
			return solver, nil
		}
	}
	return solver, nil
}

func (o *FraigOrchestrator) satLimits(limits FraigSATLimits) SatLimits {
	conflicts := limits.Conflicts
	if conflicts == 0 && o.options.Parameters.BacktrackLimit > 0 {
		conflicts = uint64(o.options.Parameters.BacktrackLimit)
	}
	return SatLimits{Conflicts: conflicts, Propagations: limits.Propagations, Decisions: limits.Decisions, Stop: o.combinedStop(limits.Stop)}
}

func (o *FraigOrchestrator) modelFrom(result SatResult) []bool {
	if result.Status != SatSatisfiable {
		return nil
	}
	model := make([]bool, len(o.manager.inputs))
	for i, edge := range o.manager.inputs {
		if int(edge.ID()) < len(result.Model) {
			model[i] = result.Model[edge.ID()] != edge.Complemented()
		}
	}
	return model
}

func (o *FraigOrchestrator) solveAssumptions(solver *SatSolver, assumptions []FraigEdge, limits FraigSATLimits) (FraigRelationResult, error) {
	literals := make([]SatLit, len(assumptions))
	for i, edge := range assumptions {
		if !o.manager.validEdge(edge) {
			return FraigRelationResult{}, ErrDimension
		}
		literals[i], _ = fraigSATLiteral(edge)
	}
	result, err := solver.Solve(literals, o.satLimits(limits))
	if err != nil {
		return FraigRelationResult{}, err
	}
	relation := FraigRelationResult{Stats: result.Stats}
	switch result.Status {
	case SatSatisfiable:
		relation.Status = FraigRelationDisproved
		relation.Model = o.modelFrom(result)
		o.lastModel = slices.Clone(relation.Model)
		o.queries[1]++
	case SatUnsatisfiable:
		relation.Status = FraigRelationProved
		o.queries[0]++
	default:
		relation.Status = FraigRelationUnknown
		o.queries[2]++
	}
	return relation, nil
}

func addFraigStats(a, b SatStats) SatStats {
	a.Conflicts += b.Conflicts
	a.Propagations += b.Propagations
	a.Decisions += b.Decisions
	a.Inspections += b.Inspections
	a.Restarts += b.Restarts
	a.Learnt += b.Learnt
	a.Deleted += b.Deleted
	return a
}

func (o *FraigOrchestrator) Equivalent(a, b FraigEdge, limits FraigSATLimits) (FraigRelationResult, error) {
	if err := o.check(); err != nil {
		return FraigRelationResult{}, err
	}
	if !o.manager.validEdge(a) || !o.manager.validEdge(b) {
		return FraigRelationResult{}, ErrDimension
	}
	if a == b {
		return FraigRelationResult{Status: FraigRelationProved}, nil
	}
	solver, err := o.newSolver(limits)
	if err != nil {
		return FraigRelationResult{}, err
	}
	first, err := o.solveAssumptions(solver, []FraigEdge{a, b.Not()}, limits)
	if err != nil || first.Status != FraigRelationProved {
		return first, err
	}
	second, err := o.solveAssumptions(solver, []FraigEdge{a.Not(), b}, limits)
	second.Stats = addFraigStats(first.Stats, second.Stats)
	return second, err
}

func (o *FraigOrchestrator) Implies(a, b FraigEdge, limits FraigSATLimits) (FraigRelationResult, error) {
	solver, err := o.newSolver(limits)
	if err != nil {
		return FraigRelationResult{}, err
	}
	return o.solveAssumptions(solver, []FraigEdge{a, b.Not()}, limits)
}

// CheckMiter proves that every output is zero or returns a replayable model.
func (o *FraigOrchestrator) CheckMiter(limits FraigSATLimits) (FraigRelationResult, error) {
	solver, err := o.newSolver(limits)
	if err != nil {
		return FraigRelationResult{}, err
	}
	var total SatStats
	for _, output := range o.manager.outputs {
		result, solveErr := o.solveAssumptions(solver, []FraigEdge{output}, limits)
		total = addFraigStats(total, result.Stats)
		result.Stats = total
		if solveErr != nil || result.Status != FraigRelationProved {
			return result, solveErr
		}
	}
	return FraigRelationResult{Status: FraigRelationProved, Stats: total}, nil
}

func (o *FraigOrchestrator) AddChoice(representative, alternative FraigEdge, limits FraigSATLimits) error {
	if err := o.check(); err != nil {
		return err
	}
	if representative.ID() == alternative.ID() || !o.manager.validEdge(representative) || !o.manager.validEdge(alternative) {
		return ErrDimension
	}
	if _, exists := o.represented[alternative.ID()]; exists {
		return ErrInvariant
	}
	if len(o.choices[alternative.ID()]) != 0 {
		return ErrInvariant
	}
	root := representative.ID()
	rootEdge := representative
	if old, ok := o.represented[root]; ok {
		matched := false
		for _, edge := range o.choices[old] {
			if edge.ID() == representative.ID() {
				matched = edge == representative
				break
			}
		}
		if !matched {
			return ErrInvariant
		}
		root = old
		rootEdge = o.choiceRoots[root]
	} else if existing, ok := o.choiceRoots[root]; ok {
		if existing != representative {
			return ErrInvariant
		}
		rootEdge = existing
	}
	count := 0
	for _, alternatives := range o.choices {
		count += len(alternatives)
	}
	if count >= o.options.MaxChoices {
		return ErrLimit
	}
	relation, err := o.Equivalent(representative, alternative, limits)
	if err != nil {
		return err
	}
	if relation.Status == FraigRelationUnknown {
		return ErrLimit
	}
	if relation.Status != FraigRelationProved {
		return ErrInvariant
	}
	if in, checkErr := o.InTFI(root, alternative.ID()); checkErr != nil || in {
		if checkErr != nil {
			return checkErr
		}
		return ErrInvariant
	}
	o.choices[root] = append(o.choices[root], alternative)
	o.represented[alternative.ID()] = root
	o.choiceRoots[root] = rootEdge
	return nil
}

func (o *FraigOrchestrator) ChoiceClasses() []FraigChoiceClass {
	classes := make([]FraigChoiceClass, 0, len(o.choices))
	for id := FraigID(0); int(id) < len(o.manager.nodes); id++ {
		if choices := o.choices[id]; len(choices) != 0 {
			classes = append(classes, FraigChoiceClass{Representative: o.choiceRoots[id], Alternatives: slices.Clone(choices)})
		}
	}
	return classes
}

func (o *FraigOrchestrator) AddAssociativeChoices(limit int, satLimits FraigSATLimits) (int, error) {
	if err := o.check(); err != nil {
		return 0, err
	}
	if limit < 0 {
		return 0, ErrDimension
	}
	checkpoint := captureFraigManager(o.manager)
	choiceCheckpoint := make(map[FraigID][]FraigEdge, len(o.choices))
	for id, choices := range o.choices {
		choiceCheckpoint[id] = slices.Clone(choices)
	}
	representedCheckpoint := make(map[FraigID]FraigID, len(o.represented))
	for id, representative := range o.represented {
		representedCheckpoint[id] = representative
	}
	rootCheckpoint := make(map[FraigID]FraigEdge, len(o.choiceRoots))
	for id, root := range o.choiceRoots {
		rootCheckpoint[id] = root
	}
	queriesCheckpoint, modelCheckpoint := o.queries, slices.Clone(o.lastModel)
	rollback := func() {
		restoreFraigManager(o.manager, checkpoint)
		o.choices, o.represented, o.choiceRoots = choiceCheckpoint, representedCheckpoint, rootCheckpoint
		o.queries, o.lastModel = queriesCheckpoint, modelCheckpoint
	}
	initial := len(o.manager.nodes)
	added := 0
	for id := 1; id < initial && (limit == 0 || added < limit); id++ {
		if o.options.Stop != nil && o.options.Stop() {
			rollback()
			return added, ErrCancelled
		}
		node := o.manager.nodes[id]
		if node.kind != FraigNodeAnd {
			continue
		}
		for side := 0; side < 2 && (limit == 0 || added < limit); side++ {
			innerEdge := node.fanin[side]
			if innerEdge.Complemented() || o.manager.nodes[innerEdge.ID()].kind != FraigNodeAnd {
				continue
			}
			inner := o.manager.nodes[innerEdge.ID()]
			outer := node.fanin[1-side]
			for pivot := 0; pivot < 2 && (limit == 0 || added < limit); pivot++ {
				temporary, makeErr := o.And(inner.fanin[1-pivot], outer)
				if makeErr != nil {
					rollback()
					return added, makeErr
				}
				alternative, makeErr := o.And(inner.fanin[pivot], temporary)
				if makeErr != nil {
					rollback()
					return added, makeErr
				}
				if alternative.ID() == FraigID(id) {
					continue
				}
				if addErr := o.AddChoice(makeFraigEdge(FraigID(id), false), alternative, satLimits); addErr == nil {
					added++
				} else if addErr != ErrInvariant {
					rollback()
					return added, addErr
				}
			}
		}
	}
	return added, nil
}

// DFS returns an owned postorder. Constants and inputs are optional; choices
// follow representatives after their ordinary cone, matching the source's
// equivalence-aware traversal order.
func (o *FraigOrchestrator) DFS(roots []FraigEdge, includeChoices, includeTerminals bool) ([]FraigID, error) {
	if err := o.check(); err != nil {
		return nil, err
	}
	for _, root := range roots {
		if !o.manager.validEdge(root) {
			return nil, ErrDimension
		}
	}
	type frame struct {
		id       FraigID
		expanded bool
	}
	seen := make([]bool, len(o.manager.nodes))
	result := make([]FraigID, 0)
	visits := 0
	for _, root := range roots {
		stack := []frame{{root.ID(), false}}
		for len(stack) != 0 {
			last := len(stack) - 1
			item := stack[last]
			stack = stack[:last]
			if item.expanded {
				node := o.manager.nodes[item.id]
				if includeTerminals || node.kind == FraigNodeAnd {
					result = append(result, item.id)
				}
				continue
			}
			if seen[item.id] {
				continue
			}
			seen[item.id] = true
			visits++
			if visits > o.options.MaxVisits {
				return nil, ErrLimit
			}
			if o.options.Stop != nil && o.options.Stop() {
				return nil, ErrCancelled
			}
			stack = append(stack, frame{item.id, true})
			if includeChoices {
				choices := o.choices[item.id]
				for i := len(choices) - 1; i >= 0; i-- {
					stack = append(stack, frame{choices[i].ID(), false})
				}
			}
			node := o.manager.nodes[item.id]
			if node.kind == FraigNodeAnd {
				stack = append(stack, frame{node.fanin[1].ID(), false}, frame{node.fanin[0].ID(), false})
			}
		}
	}
	return result, nil
}

func (o *FraigOrchestrator) InTFI(needle, root FraigID) (bool, error) {
	if err := o.check(); err != nil {
		return false, err
	}
	if int(needle) >= len(o.manager.nodes) || int(root) >= len(o.manager.nodes) {
		return false, ErrDimension
	}
	seen := make([]bool, len(o.manager.nodes))
	stack := []FraigID{root}
	for visits := 0; len(stack) != 0; visits++ {
		if visits >= o.options.MaxVisits {
			return false, ErrLimit
		}
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]
		if id == needle {
			return true, nil
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		if o.options.Stop != nil && o.options.Stop() {
			return false, ErrCancelled
		}
		node := o.manager.nodes[id]
		if node.kind == FraigNodeAnd {
			stack = append(stack, node.fanin[1].ID(), node.fanin[0].ID())
		}
	}
	return false, nil
}

func (o *FraigOrchestrator) RecognizeMux(root FraigEdge) (control, whenTrue, whenFalse FraigEdge, ok bool) {
	if o == nil || o.manager == nil || !o.manager.validEdge(root) || !root.Complemented() {
		return 0, 0, 0, false
	}
	node := o.manager.nodes[root.ID()]
	if node.kind != FraigNodeAnd || !node.fanin[0].Complemented() || !node.fanin[1].Complemented() {
		return 0, 0, 0, false
	}
	left := o.manager.nodes[node.fanin[0].ID()]
	right := o.manager.nodes[node.fanin[1].ID()]
	if left.kind != FraigNodeAnd || right.kind != FraigNodeAnd {
		return 0, 0, 0, false
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if left.fanin[i] == right.fanin[j].Not() {
				control, whenTrue, whenFalse = left.fanin[i], left.fanin[1-i], right.fanin[1-j]
				if control.Complemented() {
					control = control.Not()
					whenTrue, whenFalse = whenFalse, whenTrue
				}
				return control, whenTrue, whenFalse, true
			}
		}
	}
	return 0, 0, 0, false
}

func (o *FraigOrchestrator) RecognizeXor(root FraigEdge) (a, b FraigEdge, ok bool) {
	c, t, e, mux := o.RecognizeMux(root)
	if !mux || t != e.Not() {
		return 0, 0, false
	}
	return c, e, true
}

func (o *FraigOrchestrator) CollectSupergate(root FraigEdge, stopAtMux bool) ([]FraigEdge, error) {
	if err := o.check(); err != nil {
		return nil, err
	}
	if !o.manager.validEdge(root) {
		return nil, ErrDimension
	}
	type item struct {
		edge  FraigEdge
		first bool
	}
	result := make([]FraigEdge, 0)
	seenLeaves := make(map[FraigEdge]struct{})
	stack := []item{{root, true}}
	for visits := 0; len(stack) != 0; visits++ {
		if visits >= o.options.MaxVisits {
			return nil, ErrLimit
		}
		last := len(stack) - 1
		entry := stack[last]
		stack = stack[:last]
		edge := entry.edge
		if o.options.Stop != nil && o.options.Stop() {
			return nil, ErrCancelled
		}
		node := o.manager.nodes[edge.ID()]
		_, _, _, mux := o.RecognizeMux(edge)
		if !mux {
			_, _, _, mux = o.RecognizeMux(edge.Not())
		}
		if edge.Complemented() || node.kind != FraigNodeAnd || !entry.first && node.refs > 1 || stopAtMux && mux {
			if _, exists := seenLeaves[edge]; !exists {
				seenLeaves[edge] = struct{}{}
				result = append(result, edge)
			}
			continue
		}
		stack = append(stack, item{node.fanin[1], false}, item{node.fanin[0], false})
	}
	return result, nil
}

func (o *FraigOrchestrator) AddFeedbackPattern(model []bool) (bool, error) {
	if err := o.check(); err != nil {
		return false, err
	}
	if len(model) != len(o.manager.inputs) {
		return false, ErrDimension
	}
	for _, old := range o.patterns {
		if slices.Equal(old, model) {
			return false, nil
		}
	}
	if len(o.patterns) >= o.options.MaxPatterns {
		return false, ErrLimit
	}
	patterns := append(slices.Clone(o.patterns), slices.Clone(model))
	words := (len(patterns) + 31) / 32
	simulations := make([][]uint32, len(o.manager.nodes))
	for id := range simulations {
		simulations[id] = make([]uint32, words)
	}
	visits := 0
	for pattern, values := range patterns {
		if o.options.Stop != nil && o.options.Stop() {
			return false, ErrCancelled
		}
		nodeValues := make([]bool, len(o.manager.nodes))
		nodeValues[0] = true
		for id := 1; id < len(o.manager.nodes); id++ {
			visits++
			if visits > o.options.MaxVisits {
				return false, ErrLimit
			}
			if o.options.Stop != nil && o.options.Stop() {
				return false, ErrCancelled
			}
			node := o.manager.nodes[id]
			switch node.kind {
			case FraigNodeInput:
				nodeValues[id] = values[node.input]
			case FraigNodeAnd:
				a := nodeValues[node.fanin[0].ID()] != node.fanin[0].Complemented()
				b := nodeValues[node.fanin[1].ID()] != node.fanin[1].Complemented()
				nodeValues[id] = a && b
			}
			if nodeValues[id] {
				simulations[id][pattern/32] |= uint32(1) << uint(pattern&31)
			}
		}
		simulations[0][pattern/32] |= uint32(1) << uint(pattern&31)
	}
	for id := range o.manager.nodes {
		o.manager.nodes[id].dynamicSim = simulations[id]
		o.manager.nodes[id].simInverted = false
	}
	o.patterns = patterns
	return true, nil
}

func (o *FraigOrchestrator) FeedbackPatterns() [][]bool {
	result := make([][]bool, len(o.patterns))
	for i := range result {
		result[i] = slices.Clone(o.patterns[i])
	}
	return result
}

func (o *FraigOrchestrator) FeedbackOnes(id FraigID) (int, error) {
	if err := o.check(); err != nil {
		return 0, err
	}
	if int(id) >= len(o.manager.nodes) {
		return 0, ErrDimension
	}
	count := 0
	for _, word := range o.manager.nodes[id].dynamicSim {
		count += bits.OnesCount32(word)
	}
	return count, nil
}

func (o *FraigOrchestrator) Validate() error {
	if err := o.check(); err != nil {
		return err
	}
	if len(o.inputNames) != len(o.manager.inputs) || len(o.outputNames) != len(o.manager.outputs) {
		return ErrInvariant
	}
	if err := o.manager.Validate(); err != nil {
		return err
	}
	for representative, alternatives := range o.choices {
		if int(representative) >= len(o.manager.nodes) {
			return ErrInvariant
		}
		root, ok := o.choiceRoots[representative]
		if !ok || root.ID() != representative || !o.manager.validEdge(root) {
			return ErrInvariant
		}
		for _, alternative := range alternatives {
			if !o.manager.validEdge(alternative) || o.represented[alternative.ID()] != representative {
				return ErrInvariant
			}
		}
	}
	return nil
}

func (o *FraigOrchestrator) Stats() FraigOrchestrationStats {
	if o == nil || o.manager == nil {
		return FraigOrchestrationStats{}
	}
	choices := 0
	for _, class := range o.choices {
		choices += len(class)
	}
	return FraigOrchestrationStats{Manager: o.manager.Stats(), ChoiceClasses: len(o.choices), Choices: choices, Patterns: len(o.patterns), Clauses: len(o.clauses), SATQueries: o.queries[0] + o.queries[1] + o.queries[2], SATProved: o.queries[0], SATDisproved: o.queries[1], SATUnknown: o.queries[2]}
}

func (o *FraigOrchestrator) Close() {
	if o == nil || o.closed {
		return
	}
	o.closed = true
	if o.manager != nil {
		o.manager.Close()
	}
	o.manager = nil
	o.inputNames = nil
	o.outputNames = nil
	o.choices = nil
	o.represented = nil
	o.choiceRoots = nil
	o.patterns = nil
	o.clauses = nil
	o.lastModel = nil
}
