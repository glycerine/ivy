package abc

// Stand-alone Go port of src/opt/res: bounded windowing, divisor collection,
// bit-parallel filtering, SAT validation, window strashing, and resynthesis.

import "slices"

type ResynthesisParameters struct {
	Window, GrowthLevel, SimulationWords, Candidates int
	Area, Delay                                      bool
}

type ResynthesisLimits struct {
	MaxObjects, MaxWork, MaxDivisors, MaxCandidates, MaxSimulationWords int
	SATConflicts, SATPropagations, SATDecisions                         uint64
	Stop                                                                func() bool
}

type ResynthesisWindow struct {
	Pivot                          AigID
	TFILevels, TFOLevels           int
	Leaves, Branches, Nodes, Roots []AigID
	Divisors                       []AigID
}

type ResynthesisCandidateKind uint8

const (
	ResynthesisConstant0 ResynthesisCandidateKind = iota
	ResynthesisConstant1
	ResynthesisLiteral
	ResynthesisAND
	ResynthesisOR
	ResynthesisXOR
)

type ResynthesisCandidate struct {
	Kind        ResynthesisCandidateKind
	Inputs      []AigLit
	Replacement AigLit
	Gain        int
	Simulated   bool
	Proved      bool
}

type ResynthesisSimulation struct {
	Words, Patterns int
	Inputs          []uint64
	Values          [][]uint64
}

type ResynthesisStats struct {
	Windows, TrivialWindows, WindowNodes, Divisors int
	Candidates, Simulated, SATCalls, Proved        int
	Replacements, NodesBefore, NodesAfter          int
}

type ResynthesisResult struct {
	Graph        *AigManager
	Stats        ResynthesisStats
	Replacements []ResynthesisCandidate
}

type resynthesisBudget struct {
	limits ResynthesisLimits
	work   int
}

func normalizeResynthesis(parameters ResynthesisParameters, limits ResynthesisLimits) (ResynthesisParameters, ResynthesisLimits, error) {
	if parameters.Window < 0 || parameters.GrowthLevel < 0 || parameters.SimulationWords < 0 || parameters.Candidates < 0 ||
		limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxDivisors < 0 || limits.MaxCandidates < 0 || limits.MaxSimulationWords < 0 {
		return parameters, limits, ErrDimension
	}
	if parameters.Window == 0 {
		parameters.Window = 52
	}
	if parameters.SimulationWords == 0 {
		parameters.SimulationWords = 4
	}
	if parameters.Candidates == 0 {
		parameters.Candidates = 20
	}
	if !parameters.Area && !parameters.Delay {
		parameters.Area = true
	}
	// Res_ManAlloc asserts that both encoded window and candidate count are
	// strictly below 100. Keep the checked Go entry point within that domain.
	if parameters.Window/10 < 1 || parameters.Window/10 > 9 || parameters.Window%10 > 9 || parameters.Candidates >= 100 {
		return parameters, limits, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if limits.MaxWork == 0 {
		limits.MaxWork = 1 << 24
	}
	if limits.MaxDivisors == 0 {
		limits.MaxDivisors = 4096
	}
	if limits.MaxCandidates == 0 {
		limits.MaxCandidates = parameters.Candidates
	}
	if limits.MaxSimulationWords == 0 {
		limits.MaxSimulationWords = 1 << 20
	}
	if parameters.SimulationWords > limits.MaxSimulationWords {
		return parameters, limits, ErrLimit
	}
	if limits.MaxCandidates > parameters.Candidates {
		limits.MaxCandidates = parameters.Candidates
	}
	return parameters, limits, nil
}

func (b *resynthesisBudget) step() error {
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	if b.work == b.limits.MaxWork {
		return ErrLimit
	}
	b.work++
	return nil
}

func resAppendUnique(ids *[]AigID, present []bool, id AigID) {
	if !present[id] {
		present[id] = true
		*ids = append(*ids, id)
	}
}

// ComputeResynthesisWindow follows the level-bounded TFI/TFO construction of
// Res_WinCompute and collects deterministic MFFC-external divisors.
func ComputeResynthesisWindow(graph *AigManager, pivot AigID, parameters ResynthesisParameters, limits ResynthesisLimits) (ResynthesisWindow, error) {
	parameters, limits, err := normalizeResynthesis(parameters, limits)
	if err != nil {
		return ResynthesisWindow{}, err
	}
	if graph == nil {
		return ResynthesisWindow{}, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return ResynthesisWindow{}, err
	}
	if len(graph.objects) > limits.MaxObjects {
		return ResynthesisWindow{}, ErrLimit
	}
	if uint64(pivot) >= uint64(len(graph.objects)) || graph.objects[pivot].Type != AigAND {
		return ResynthesisWindow{}, ErrInvalidLiteral
	}
	b := &resynthesisBudget{limits: limits}
	w := ResynthesisWindow{Pivot: pivot, TFILevels: parameters.Window / 10, TFOLevels: parameters.Window % 10}
	pivotLevel := graph.objects[pivot].Level
	seen := make([]bool, len(graph.objects))
	leafSeen := make([]bool, len(graph.objects))
	type frame struct {
		id   AigID
		exit bool
	}
	stack := []frame{{pivot, false}}
	for len(stack) != 0 {
		if err = b.step(); err != nil {
			return ResynthesisWindow{}, err
		}
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if item.exit {
			w.Nodes = append(w.Nodes, item.id)
			continue
		}
		if seen[item.id] {
			continue
		}
		seen[item.id] = true
		obj := graph.objects[item.id]
		if obj.Type != AigAND {
			return ResynthesisWindow{}, ErrInvariant
		}
		stack = append(stack, frame{item.id, true})
		for side := 1; side >= 0; side-- {
			fanin := obj.Fanin0.ID()
			if side == 1 {
				fanin = obj.Fanin1.ID()
			}
			fobj := graph.objects[fanin]
			if fobj.Type == AigCI || int(pivotLevel-fobj.Level) > w.TFILevels {
				resAppendUnique(&w.Leaves, leafSeen, fanin)
			} else if fobj.Type == AigAND && !seen[fanin] {
				stack = append(stack, frame{fanin, false})
			}
		}
	}
	if len(w.Leaves) == 0 {
		return ResynthesisWindow{}, ErrInvariant
	}
	fanouts, err := BuildAigFanoutIndex(graph, AigTraversalOptions{MaxObjects: limits.MaxObjects, MaxVisits: limits.MaxWork, Stop: limits.Stop})
	if err != nil {
		return ResynthesisWindow{}, err
	}
	rootSeen := make([]bool, len(graph.objects))
	tfoSeen := make([]bool, len(graph.objects))
	queue := []AigID{pivot}
	maxLevel := pivotLevel + uint32(w.TFOLevels)
	for len(queue) != 0 {
		if err = b.step(); err != nil {
			return ResynthesisWindow{}, err
		}
		id := queue[0]
		queue = queue[1:]
		if tfoSeen[id] {
			continue
		}
		tfoSeen[id] = true
		edges, e := fanouts.Fanouts(id)
		if e != nil {
			return ResynthesisWindow{}, e
		}
		boundary := len(edges) == 0 || len(edges) > 10
		for _, edge := range edges {
			fo := graph.objects[edge.Object]
			if fo.Type == AigCO || fo.Level > maxLevel {
				boundary = true
				continue
			}
			if fo.Type == AigAND {
				queue = append(queue, edge.Object)
			}
		}
		if boundary {
			resAppendUnique(&w.Roots, rootSeen, id)
		}
	}
	if len(w.Roots) == 0 {
		w.Roots = append(w.Roots, pivot)
		rootSeen[pivot] = true
	}
	branchSeen := make([]bool, len(graph.objects))
	for id, inTFO := range tfoSeen {
		if !inTFO || AigID(id) == pivot || graph.objects[id].Type != AigAND {
			continue
		}
		obj := graph.objects[id]
		for _, lit := range []AigLit{obj.Fanin0, obj.Fanin1} {
			if !tfoSeen[lit.ID()] && !seen[lit.ID()] {
				resAppendUnique(&w.Branches, branchSeen, lit.ID())
			}
		}
	}
	mffc, err := graph.MFFCLabel(pivot, nil, AigMFFCOptions{MaxVisits: limits.MaxWork, Stop: limits.Stop})
	if err != nil {
		return ResynthesisWindow{}, err
	}
	direct := make([]bool, len(graph.objects))
	pobj := graph.objects[pivot]
	direct[pobj.Fanin0.ID()], direct[pobj.Fanin1.ID()] = true, true
	for id, obj := range graph.objects {
		if err = b.step(); err != nil {
			return ResynthesisWindow{}, err
		}
		if obj.Type != AigCI && obj.Type != AigAND {
			continue
		}
		if AigID(id) == pivot || direct[id] || tfoSeen[id] || (id < len(mffc.Members) && mffc.Members[id]) || obj.Level >= pivotLevel {
			continue
		}
		w.Divisors = append(w.Divisors, AigID(id))
		if len(w.Divisors) == limits.MaxDivisors {
			break
		}
	}
	return w, nil
}

func resRandomPatterns(inputs, words int) []uint64 {
	patterns := make([]uint64, inputs*words)
	state := uint64(0x9e3779b97f4a7c15)
	for i := range patterns {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		patterns[i] = state
	}
	return patterns
}

// SimulateResynthesis performs owned bit-parallel simulation. Input words are
// CI-major; nil inputs request deterministic pseudo-random patterns.
func SimulateResynthesis(graph *AigManager, words int, inputs []uint64, limits ResynthesisLimits) (ResynthesisSimulation, error) {
	if graph == nil || words <= 0 || limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxSimulationWords < 0 {
		return ResynthesisSimulation{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return ResynthesisSimulation{}, err
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if limits.MaxWork == 0 {
		limits.MaxWork = 1 << 24
	}
	if limits.MaxSimulationWords == 0 {
		limits.MaxSimulationWords = 1 << 20
	}
	if words > limits.MaxSimulationWords {
		return ResynthesisSimulation{}, ErrLimit
	}
	if len(graph.objects) > limits.MaxObjects {
		return ResynthesisSimulation{}, ErrLimit
	}
	if _, err := checkedSize(0, len(graph.objects), words, limits.MaxWork); err != nil {
		return ResynthesisSimulation{}, err
	}
	expected, err := checkedSize(0, len(graph.cis), words, limits.MaxWork)
	if err != nil {
		return ResynthesisSimulation{}, err
	}
	if inputs == nil {
		inputs = resRandomPatterns(len(graph.cis), words)
	}
	if len(inputs) != expected {
		return ResynthesisSimulation{}, ErrDimension
	}
	b := &resynthesisBudget{limits: limits}
	values := make([][]uint64, len(graph.objects))
	values[0] = make([]uint64, words)
	for i := range values[0] {
		values[0][i] = ^uint64(0)
	}
	ci := 0
	read := func(l AigLit, word int) uint64 {
		v := values[l.ID()][word]
		if l.Complemented() {
			return ^v
		}
		return v
	}
	for id, obj := range graph.objects {
		if err := b.step(); err != nil {
			return ResynthesisSimulation{}, err
		}
		switch obj.Type {
		case AigNone, AigConstant1:
		case AigCI:
			values[id] = slices.Clone(inputs[ci*words : (ci+1)*words])
			ci++
		case AigBuffer, AigCO:
			values[id] = make([]uint64, words)
			for word := range words {
				values[id][word] = read(obj.Fanin0, word)
			}
		case AigAND:
			values[id] = make([]uint64, words)
			for word := range words {
				values[id][word] = read(obj.Fanin0, word) & read(obj.Fanin1, word)
			}
		default:
			return ResynthesisSimulation{}, ErrUnsupported
		}
	}
	return ResynthesisSimulation{Words: words, Patterns: words * 64, Inputs: slices.Clone(inputs), Values: values}, nil
}

func resExpressionWords(candidate ResynthesisCandidate, simulation ResynthesisSimulation) ([]uint64, error) {
	read := func(l AigLit, word int) uint64 {
		if uint64(l.ID()) >= uint64(len(simulation.Values)) || simulation.Values[l.ID()] == nil {
			return 0
		}
		v := simulation.Values[l.ID()][word]
		if l.Complemented() {
			return ^v
		}
		return v
	}
	result := make([]uint64, simulation.Words)
	for word := range result {
		switch candidate.Kind {
		case ResynthesisConstant0:
		case ResynthesisConstant1:
			result[word] = ^uint64(0)
		case ResynthesisLiteral:
			if len(candidate.Inputs) != 1 {
				return nil, ErrInvariant
			}
			result[word] = read(candidate.Inputs[0], word)
		case ResynthesisAND, ResynthesisOR, ResynthesisXOR:
			if len(candidate.Inputs) != 2 {
				return nil, ErrInvariant
			}
			a, c := read(candidate.Inputs[0], word), read(candidate.Inputs[1], word)
			if candidate.Kind == ResynthesisAND {
				result[word] = a & c
			} else if candidate.Kind == ResynthesisOR {
				result[word] = a | c
			} else {
				result[word] = a ^ c
			}
		default:
			return nil, ErrInvariant
		}
	}
	return result, nil
}

func resSimulationEquivalent(graph *AigManager, pivot AigID, replacement []uint64, simulation ResynthesisSimulation) bool {
	values := make([][]uint64, len(simulation.Values))
	copy(values, simulation.Values)
	values[pivot] = replacement
	read := func(l AigLit, word int) uint64 {
		v := values[l.ID()][word]
		if l.Complemented() {
			return ^v
		}
		return v
	}
	for id := int(pivot) + 1; id < len(graph.objects); id++ {
		obj := graph.objects[id]
		if obj.Type != AigAND && obj.Type != AigBuffer && obj.Type != AigCO {
			continue
		}
		values[id] = make([]uint64, simulation.Words)
		for word := range simulation.Words {
			values[id][word] = read(obj.Fanin0, word)
			if obj.Type == AigAND {
				values[id][word] &= read(obj.Fanin1, word)
			}
		}
	}
	for _, id := range graph.cos {
		if !slices.Equal(values[id], simulation.Values[id]) {
			return false
		}
	}
	return true
}

// FilterResynthesisCandidates mirrors the simulation filter before SAT.
func FilterResynthesisCandidates(graph *AigManager, window ResynthesisWindow, simulation ResynthesisSimulation, maximum int, limits ResynthesisLimits) ([]ResynthesisCandidate, error) {
	if graph == nil || maximum < 0 || maximum >= 100 || limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxDivisors < 0 || limits.MaxCandidates < 0 || limits.MaxSimulationWords < 0 || uint64(window.Pivot) >= uint64(len(graph.objects)) || simulation.Words <= 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if len(graph.objects) > limits.MaxObjects {
		return nil, ErrLimit
	}
	if graph.objects[window.Pivot].Type != AigAND || len(simulation.Values) != len(graph.objects) || len(simulation.Inputs) != len(graph.cis)*simulation.Words {
		return nil, ErrInvariant
	}
	for id, object := range graph.objects {
		if object.Type != AigNone && len(simulation.Values[id]) != simulation.Words {
			return nil, ErrInvariant
		}
	}
	for _, id := range window.Divisors {
		if uint64(id) >= uint64(len(graph.objects)) || (graph.objects[id].Type != AigCI && graph.objects[id].Type != AigAND) {
			return nil, ErrInvalidLiteral
		}
	}
	if maximum == 0 {
		maximum = 20
	}
	if limits.MaxWork == 0 {
		limits.MaxWork = 1 << 24
	}
	if limits.MaxCandidates > 0 && maximum > limits.MaxCandidates {
		maximum = limits.MaxCandidates
	}
	b := &resynthesisBudget{limits: limits}
	candidates := []ResynthesisCandidate{{Kind: ResynthesisConstant0}, {Kind: ResynthesisConstant1}}
	pivotObject := graph.objects[window.Pivot]
	pool := []AigID{pivotObject.Fanin0.ID(), pivotObject.Fanin1.ID()}
	for _, id := range window.Divisors {
		if !slices.Contains(pool, id) {
			pool = append(pool, id)
		}
	}
	for _, id := range pool {
		candidates = append(candidates, ResynthesisCandidate{Kind: ResynthesisLiteral, Inputs: []AigLit{AigLit(id) << 1}}, ResynthesisCandidate{Kind: ResynthesisLiteral, Inputs: []AigLit{(AigLit(id) << 1).Not()}})
	}
	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			a, c := AigLit(pool[i])<<1, AigLit(pool[j])<<1
			for _, kind := range []ResynthesisCandidateKind{ResynthesisAND, ResynthesisOR, ResynthesisXOR} {
				candidates = append(candidates, ResynthesisCandidate{Kind: kind, Inputs: []AigLit{a, c}})
			}
			if len(candidates) >= maximum*16 {
				break
			}
		}
		if len(candidates) >= maximum*16 {
			break
		}
	}
	result := make([]ResynthesisCandidate, 0, maximum)
	for _, candidate := range candidates {
		if err := b.step(); err != nil {
			return nil, err
		}
		words, err := resExpressionWords(candidate, simulation)
		if err != nil {
			return nil, err
		}
		if resSimulationEquivalent(graph, window.Pivot, words, simulation) {
			candidate.Simulated = true
			result = append(result, candidate)
			if len(result) == maximum {
				break
			}
		}
	}
	return result, nil
}

func resBuildCandidate(graph *AigManager, candidate ResynthesisCandidate) (AigLit, error) {
	switch candidate.Kind {
	case ResynthesisConstant0:
		return AigConst0, nil
	case ResynthesisConstant1:
		return AigConst1, nil
	case ResynthesisLiteral:
		if len(candidate.Inputs) != 1 {
			return InvalidAigLit, ErrInvariant
		}
		return candidate.Inputs[0], nil
	case ResynthesisAND:
		if len(candidate.Inputs) != 2 {
			return InvalidAigLit, ErrInvariant
		}
		return graph.And(candidate.Inputs[0], candidate.Inputs[1])
	case ResynthesisOR:
		if len(candidate.Inputs) != 2 {
			return InvalidAigLit, ErrInvariant
		}
		return graph.Or(candidate.Inputs[0], candidate.Inputs[1])
	case ResynthesisXOR:
		if len(candidate.Inputs) != 2 {
			return InvalidAigLit, ErrInvariant
		}
		return graph.Xor(candidate.Inputs[0], candidate.Inputs[1])
	default:
		return InvalidAigLit, ErrInvariant
	}
}

func resShiftSat(l SatLit, base SatVar) SatLit { return l + SatLit(base)<<1 }

func resGraphsEquivalent(first, second *AigManager, limits ResynthesisLimits) (bool, error) {
	if first == nil || second == nil || limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxDivisors < 0 || limits.MaxCandidates < 0 || limits.MaxSimulationWords < 0 {
		return false, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if first.ObjectCount() > limits.MaxObjects || second.ObjectCount() > limits.MaxObjects {
		return false, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return false, ErrCancelled
	}
	if first.InputCount() != second.InputCount() || first.OutputCount() != second.OutputCount() || first.RegisterCount() != second.RegisterCount() {
		return false, ErrDimension
	}
	a, err := EncodeAigCNF(first)
	if err != nil {
		return false, err
	}
	c, err := EncodeAigCNF(second)
	if err != nil {
		return false, err
	}
	solver := NewSatSolver()
	if ok, e := a.LoadInto(solver, 0); e != nil || !ok {
		return false, e
	}
	base := SatVar(a.VariableCount())
	if ok, e := c.LoadInto(solver, base); e != nil || !ok {
		return false, e
	}
	for i, left := range a.inputs {
		right := resShiftSat(c.inputs[i], base)
		if ok, e := solver.AddClause([]SatLit{left.Not(), right}); e != nil || !ok {
			return false, e
		}
		if ok, e := solver.AddClause([]SatLit{left, right.Not()}); e != nil || !ok {
			return false, e
		}
	}
	satLimits := SatLimits{Conflicts: limits.SATConflicts, Propagations: limits.SATPropagations, Decisions: limits.SATDecisions, Stop: limits.Stop}
	for i, left := range a.outputs {
		right := resShiftSat(c.outputs[i], base)
		for _, assumptions := range [][]SatLit{{left, right.Not()}, {left.Not(), right}} {
			result, e := solver.Solve(assumptions, satLimits)
			if e != nil {
				return false, e
			}
			if result.Status == SatSatisfiable {
				return false, nil
			}
			if result.Status == SatUnknown {
				if result.Reason == SatCancelled {
					return false, ErrCancelled
				}
				return false, ErrLimit
			}
		}
	}
	return true, nil
}

// ProveResynthesisCandidate builds the candidate on a private clone, rewrites
// the pivot, and SAT-proves all outputs equivalent under shared inputs.
func ProveResynthesisCandidate(graph *AigManager, pivot AigID, candidate ResynthesisCandidate, limits ResynthesisLimits) (*AigManager, ResynthesisCandidate, error) {
	if graph == nil || uint64(pivot) >= uint64(len(graph.objects)) || graph.objects[pivot].Type != AigAND {
		return nil, candidate, ErrInvalidLiteral
	}
	if limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxDivisors < 0 || limits.MaxCandidates < 0 || limits.MaxSimulationWords < 0 {
		return nil, candidate, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if len(graph.objects) > limits.MaxObjects {
		return nil, candidate, ErrLimit
	}
	if err := graph.Check(); err != nil {
		return nil, candidate, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, candidate, ErrCancelled
	}
	private := graph.Clone()
	replacement, err := resBuildCandidate(private, candidate)
	if err != nil {
		return nil, candidate, err
	}
	if replacement.ID() == pivot && !replacement.Complemented() {
		return nil, candidate, ErrInvariant
	}
	rewrite, err := private.ReplaceObject(pivot, replacement, AigRewriteOptions{MaxObjects: limits.MaxObjects, MaxVisits: limits.MaxWork, Stop: limits.Stop})
	if err != nil {
		return nil, candidate, err
	}
	equivalent, err := resGraphsEquivalent(graph, rewrite.Graph, limits)
	if err != nil {
		return nil, candidate, err
	}
	if !equivalent {
		return nil, candidate, nil
	}
	candidate.Replacement, candidate.Gain, candidate.Proved = rewrite.ObjectMap[pivot], graph.AndCount()-rewrite.Graph.AndCount(), true
	return rewrite.Graph, candidate, nil
}

// StrashResynthesisWindow creates the local AIG used by simulation/filtering.
// Outputs are care, pivot, pivot fanins, then divisors; care is the OR of root
// differences when the pivot is complemented.
func StrashResynthesisWindow(graph *AigManager, window ResynthesisWindow, limits ResynthesisLimits) (*AigManager, []AigLit, error) {
	if graph == nil || uint64(window.Pivot) >= uint64(len(graph.objects)) {
		return nil, nil, ErrInvalidLiteral
	}
	if limits.MaxObjects < 0 || limits.MaxWork < 0 || limits.MaxDivisors < 0 || limits.MaxCandidates < 0 || limits.MaxSimulationWords < 0 {
		return nil, nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, nil, err
	}
	if graph.objects[window.Pivot].Type != AigAND || len(window.Roots) == 0 {
		return nil, nil, ErrInvariant
	}
	for _, group := range [][]AigID{window.Leaves, window.Branches, window.Divisors} {
		for _, id := range group {
			if uint64(id) >= uint64(len(graph.objects)) {
				return nil, nil, ErrInvalidLiteral
			}
			if kind := graph.objects[id].Type; kind != AigCI && kind != AigAND {
				return nil, nil, ErrInvariant
			}
		}
	}
	for _, group := range [][]AigID{window.Nodes, window.Roots} {
		for _, id := range group {
			if uint64(id) >= uint64(len(graph.objects)) {
				return nil, nil, ErrInvalidLiteral
			}
			if graph.objects[id].Type != AigAND {
				return nil, nil, ErrInvariant
			}
		}
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if limits.MaxWork == 0 {
		limits.MaxWork = 1 << 24
	}
	dst := NewAigManager()
	normal := make([]AigLit, len(graph.objects))
	altered := make([]AigLit, len(graph.objects))
	for i := range normal {
		normal[i], altered[i] = InvalidAigLit, InvalidAigLit
	}
	normal[0], altered[0] = AigConst1, AigConst1
	boundary := append(slices.Clone(window.Leaves), window.Branches...)
	for _, id := range boundary {
		if normal[id] != InvalidAigLit {
			continue
		}
		lit, err := dst.CreateCI()
		if err != nil {
			return nil, nil, err
		}
		normal[id], altered[id] = lit, lit
	}
	b := &resynthesisBudget{limits: limits}
	var build func(AigID, bool) (AigLit, error)
	build = func(id AigID, changed bool) (AigLit, error) {
		if err := b.step(); err != nil {
			return InvalidAigLit, err
		}
		table := normal
		if changed {
			table = altered
		}
		if table[id] != InvalidAigLit {
			return table[id], nil
		}
		obj := graph.objects[id]
		if obj.Type == AigCI {
			lit, e := dst.CreateCI()
			if e != nil {
				return InvalidAigLit, e
			}
			normal[id], altered[id] = lit, lit
			return lit, nil
		}
		if obj.Type != AigAND {
			return InvalidAigLit, ErrUnsupported
		}
		left, e := build(obj.Fanin0.ID(), changed)
		if e != nil {
			return InvalidAigLit, e
		}
		right, e := build(obj.Fanin1.ID(), changed)
		if e != nil {
			return InvalidAigLit, e
		}
		left, right = left.NotIf(obj.Fanin0.Complemented()), right.NotIf(obj.Fanin1.Complemented())
		lit, e := dst.And(left, right)
		if e != nil {
			return InvalidAigLit, e
		}
		if changed && id == window.Pivot {
			lit = lit.Not()
		}
		table[id] = lit
		return lit, nil
	}
	care := AigConst0
	for _, root := range window.Roots {
		a, e := build(root, false)
		if e != nil {
			return nil, nil, e
		}
		c, e := build(root, true)
		if e != nil {
			return nil, nil, e
		}
		x, e := dst.Xor(a, c)
		if e != nil {
			return nil, nil, e
		}
		care, e = dst.Or(care, x)
		if e != nil {
			return nil, nil, e
		}
	}
	pivot, err := build(window.Pivot, false)
	if err != nil {
		return nil, nil, err
	}
	outputs := []AigLit{care, pivot}
	pobj := graph.objects[window.Pivot]
	for _, original := range []AigLit{pobj.Fanin0, pobj.Fanin1} {
		lit, e := build(original.ID(), false)
		if e != nil {
			return nil, nil, e
		}
		outputs = append(outputs, lit.NotIf(original.Complemented()))
	}
	for _, id := range window.Divisors {
		lit, e := build(id, false)
		if e != nil {
			continue
		}
		outputs = append(outputs, lit)
	}
	for _, output := range outputs {
		if _, err = dst.CreateCO(output); err != nil {
			return nil, nil, err
		}
	}
	if dst.ObjectCount() > limits.MaxObjects {
		return nil, nil, ErrLimit
	}
	if err = dst.Check(); err != nil {
		return nil, nil, err
	}
	return dst, slices.Clone(outputs), nil
}

// ResynthesizeAIG performs deterministic, atomic area/delay resynthesis.
func ResynthesizeAIG(source *AigManager, parameters ResynthesisParameters, limits ResynthesisLimits) (ResynthesisResult, error) {
	parameters, limits, err := normalizeResynthesis(parameters, limits)
	if err != nil {
		return ResynthesisResult{}, err
	}
	if source == nil {
		return ResynthesisResult{}, ErrDimension
	}
	if err = source.Check(); err != nil {
		return ResynthesisResult{}, err
	}
	if source.equivs != nil {
		return ResynthesisResult{}, ErrUnsupported
	}
	if len(source.objects) > limits.MaxObjects {
		return ResynthesisResult{}, ErrLimit
	}
	working := source.Clone()
	result := ResynthesisResult{Stats: ResynthesisStats{NodesBefore: source.AndCount()}}
	budget := &resynthesisBudget{limits: limits}
	for {
		changed := false
		for id := 1; id < len(working.objects); id++ {
			if err = budget.step(); err != nil {
				return ResynthesisResult{}, err
			}
			if working.objects[id].Type != AigAND || working.objects[id].RefCount == 0 {
				continue
			}
			window, e := ComputeResynthesisWindow(working, AigID(id), parameters, limits)
			if e != nil {
				if e == ErrInvariant {
					continue
				}
				return ResynthesisResult{}, e
			}
			result.Stats.Windows++
			result.Stats.WindowNodes += len(window.Nodes)
			result.Stats.Divisors += len(window.Divisors)
			if len(window.Roots) == 1 && window.Roots[0] == AigID(id) {
				result.Stats.TrivialWindows++
			}
			simulation, e := SimulateResynthesis(working, parameters.SimulationWords, nil, limits)
			if e != nil {
				return ResynthesisResult{}, e
			}
			candidates, e := FilterResynthesisCandidates(working, window, simulation, limits.MaxCandidates, limits)
			if e != nil {
				return ResynthesisResult{}, e
			}
			result.Stats.Candidates += len(candidates)
			result.Stats.Simulated += len(candidates)
			for _, candidate := range candidates {
				result.Stats.SATCalls++
				trial, proved, e := ProveResynthesisCandidate(working, AigID(id), candidate, limits)
				if e != nil {
					if e == ErrInvariant {
						continue
					}
					return ResynthesisResult{}, e
				}
				if trial == nil {
					continue
				}
				if parameters.Area && proved.Gain <= 0 {
					continue
				}
				if parameters.Delay && trial.objects[proved.Replacement.ID()].Level > working.objects[id].Level+uint32(parameters.GrowthLevel) {
					continue
				}
				working = trial
				result.Stats.Proved++
				result.Stats.Replacements++
				result.Replacements = append(result.Replacements, proved)
				changed = true
				break
			}
			if changed {
				break
			}
		}
		if !changed {
			break
		}
	}
	if err = working.Check(); err != nil {
		return ResynthesisResult{}, err
	}
	result.Graph = working
	result.Stats.NodesAfter = working.AndCount()
	return result, nil
}
