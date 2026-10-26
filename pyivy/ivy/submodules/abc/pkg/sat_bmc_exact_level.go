package abc

// Level-constrained lazy MAJ3 synthesis from src/sat/bmc/bmcMaj3.c.

import "slices"

type LevelMajorityChoice struct {
	Source   int
	Variable SatVar
}

type LevelMajorityNode struct {
	Object     int
	Level      int
	Fixed      []int
	Selectable []LevelMajorityChoice
}

type LevelMajorityTopology struct {
	Inputs, VariableCount int
	Levels                []int // top level to bottom level, as in Maj3_ManMarkup
	Nodes                 []LevelMajorityNode
}

func (t LevelMajorityTopology) Clone() LevelMajorityTopology {
	r := t
	r.Levels = slices.Clone(t.Levels)
	r.Nodes = make([]LevelMajorityNode, len(t.Nodes))
	for i, node := range t.Nodes {
		r.Nodes[i] = node
		r.Nodes[i].Fixed = slices.Clone(node.Fixed)
		r.Nodes[i].Selectable = slices.Clone(node.Selectable)
	}
	return r
}

// BuildLevelMajorityTopology preserves Maj3_ManFirstAndLevel and
// Maj3_ManMarkup. Levels are supplied from the single output level downward.
func BuildLevelMajorityTopology(inputs int, levels []int, limits ExactSynthesisLimits) (LevelMajorityTopology, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || inputs < 3 || len(levels) < 2 || levels[0] != 1 {
		return LevelMajorityTopology{}, ErrDimension
	}
	nodes := 0
	for _, count := range levels {
		if count < 1 || count > 3 {
			return LevelMajorityTopology{}, ErrDimension
		}
		nodes += count
	}
	objects := inputs + nodes
	if objects > 32 || objects > limits.MaxObjects {
		return LevelMajorityTopology{}, ErrLimit
	}
	first := make([]int, len(levels)+1)
	objectLevel := make([]int, objects)
	nextObject := inputs
	for bottom := 1; bottom <= len(levels); bottom++ {
		first[bottom] = nextObject
		count := levels[len(levels)-bottom]
		for k := 0; k < count; k++ {
			objectLevel[nextObject] = bottom
			nextObject++
		}
	}
	if nextObject != objects {
		return LevelMajorityTopology{}, ErrInvariant
	}
	fixed := make([][]int, objects)
	fixed[inputs] = []int{0, 1, 2}
	secondFromTop := levels[1]
	for k := 0; k < secondFromTop; k++ {
		fixed[objects-1] = append(fixed[objects-1], objects-2-k)
	}
	for level := 2; level < len(levels); level++ {
		fixed[first[level]] = append(fixed[first[level]], first[level-1])
	}
	result := LevelMajorityTopology{Inputs: inputs, Levels: slices.Clone(levels)}
	nextVariable := 2
	for object := inputs; object < objects; object++ {
		node := LevelMajorityNode{Object: object, Level: objectLevel[object], Fixed: slices.Clone(fixed[object])}
		if len(node.Fixed) > 3 {
			return LevelMajorityTopology{}, ErrDimension
		}
		assignVariables := object > inputs && !(object == objects-1 && secondFromTop == 3)
		if assignVariables {
			candidateLimit := first[objectLevel[object]]
			for source := 0; source < candidateLimit; source++ {
				if slices.Contains(node.Fixed, source) {
					continue
				}
				if limits.Stop != nil && limits.Stop() {
					return LevelMajorityTopology{}, ErrCancelled
				}
				variable, variableErr := exactTopologyVariable(&nextVariable, limits.MaxVariables)
				if variableErr != nil {
					return LevelMajorityTopology{}, variableErr
				}
				node.Selectable = append(node.Selectable, LevelMajorityChoice{Source: source, Variable: variable})
			}
		}
		if len(node.Fixed) < 3 && len(node.Selectable) == 0 {
			return LevelMajorityTopology{}, ErrDimension
		}
		result.Nodes = append(result.Nodes, node)
	}
	result.VariableCount = nextVariable
	return result, nil
}

type LevelMajorityStatistics struct {
	ExactCNFStatistics
	LazyConstraints int
}

type LevelMajoritySession struct {
	Topology LevelMajorityTopology
	Solver   *SatSolver
	Options  ExactCNFOptions
	Minterms []int
}

func checkLevelMajorityTopology(topology LevelMajorityTopology, options ExactCNFOptions) error {
	if topology.Inputs < 3 || len(topology.Nodes) < 2 || topology.VariableCount < 2 || topology.VariableCount > options.MaxVariables {
		return ErrDimension
	}
	seen := make([]bool, topology.VariableCount)
	seen[0], seen[1] = true, true
	for row, node := range topology.Nodes {
		if node.Object != topology.Inputs+row || node.Level < 1 || len(node.Fixed) > 3 {
			return ErrDimension
		}
		sources := make([]bool, node.Object)
		for _, source := range node.Fixed {
			if source < 0 || source >= node.Object || sources[source] {
				return ErrDimension
			}
			sources[source] = true
		}
		for _, choice := range node.Selectable {
			if choice.Source < 0 || choice.Source >= node.Object || sources[choice.Source] || int(choice.Variable) < 2 || int(choice.Variable) >= len(seen) || seen[choice.Variable] {
				return ErrDimension
			}
			sources[choice.Source], seen[choice.Variable] = true, true
		}
		if len(node.Fixed) < 3 && len(node.Selectable) == 0 {
			return ErrDimension
		}
	}
	for _, used := range seen {
		if !used {
			return ErrDimension
		}
	}
	return nil
}

func NewLevelMajoritySession(topology LevelMajorityTopology, options ExactCNFOptions) (*LevelMajoritySession, LevelMajorityStatistics, error) {
	statistics := LevelMajorityStatistics{}
	var err error
	options, err = exactCNFLimits(options)
	if err != nil {
		return nil, statistics, err
	}
	if err = checkLevelMajorityTopology(topology, options); err != nil {
		return nil, statistics, err
	}
	builder := exactClauseBuilder{maxClauses: options.MaxClauses, maxLiterals: options.MaxLiterals, stop: options.Stop}
	objects := topology.Inputs + len(topology.Nodes)
	fixedFanout := make([]bool, objects)
	selectableFanout := make([][]SatLit, objects)
	for _, node := range topology.Nodes {
		for _, source := range node.Fixed {
			fixedFanout[source] = true
		}
		for _, choice := range node.Selectable {
			selectableFanout[choice.Source] = append(selectableFanout[choice.Source], satCardinalityLit(choice.Variable, false))
		}
		if len(node.Fixed) < 3 {
			clause := make([]SatLit, len(node.Selectable))
			for i, choice := range node.Selectable {
				clause[i] = satCardinalityLit(choice.Variable, false)
			}
			if err = builder.add(clause...); err != nil {
				return nil, statistics, err
			}
		}
	}
	for source := 0; source < objects-1; source++ {
		if fixedFanout[source] {
			continue
		}
		if len(selectableFanout[source]) == 0 {
			return nil, statistics, ErrDimension
		}
		if err = builder.add(selectableFanout[source]...); err != nil {
			return nil, statistics, err
		}
	}
	solver := NewSatSolver()
	if err = solver.SetVariableCount(topology.VariableCount); err != nil {
		return nil, statistics, err
	}
	inconsistent := false
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		ok, addErr := solver.AddClause(clause)
		if addErr != nil {
			return nil, statistics, addErr
		}
		inconsistent = !ok
	}
	statistics.ExactCNFStatistics = ExactCNFStatistics{Variables: topology.VariableCount, Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return &LevelMajoritySession{Topology: topology.Clone(), Solver: solver, Options: options}, statistics, nil
}

func levelMajoritySelected(node LevelMajorityNode, model []bool) ([]int, []LevelMajorityChoice, []LevelMajorityChoice, error) {
	selected := slices.Clone(node.Fixed)
	var on, off []LevelMajorityChoice
	for _, choice := range node.Selectable {
		if int(choice.Variable) >= len(model) {
			return nil, nil, nil, ErrDimension
		}
		if model[choice.Variable] {
			selected = append(selected, choice.Source)
			on = append(on, choice)
		} else {
			off = append(off, choice)
		}
	}
	slices.Sort(selected)
	return selected, on, off, nil
}

// RefineTopologyModel lazily excludes every node whose selected fanin count is
// not three, matching Maj3_ManAddConstraintsLazy. It commits atomically.
func (s *LevelMajoritySession) RefineTopologyModel(model []bool) (LevelMajorityStatistics, error) {
	result := LevelMajorityStatistics{}
	if s == nil || s.Solver == nil || len(model) < s.Solver.VariableCount() {
		return result, ErrDimension
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	for _, node := range s.Topology.Nodes[1:] {
		selected, on, off, err := levelMajoritySelected(node, model)
		if err != nil {
			return result, err
		}
		if len(selected) == 3 {
			continue
		}
		result.LazyConstraints++
		if len(selected) < 3 {
			clause := make([]SatLit, len(off))
			for i, choice := range off {
				clause[i] = satCardinalityLit(choice.Variable, false)
			}
			if len(clause) == 0 {
				return result, ErrInvariant
			}
			if err = builder.add(clause...); err != nil {
				return LevelMajorityStatistics{}, err
			}
		} else {
			count := min(4-len(node.Fixed), len(on))
			if count < 1 {
				return result, ErrInvariant
			}
			clause := make([]SatLit, count)
			for i := range clause {
				clause[i] = satCardinalityLit(on[i].Variable, true)
			}
			if err = builder.add(clause...); err != nil {
				return LevelMajorityStatistics{}, err
			}
		}
	}
	trial := s.Solver.cloneState()
	trial.bookmark = s.Solver.bookmark
	inconsistent := trial.inconsistent
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		ok, err := trial.AddClause(clause)
		if err != nil {
			return LevelMajorityStatistics{}, err
		}
		inconsistent = !ok
	}
	*s.Solver = trial
	result.ExactCNFStatistics = ExactCNFStatistics{Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return result, nil
}

type levelMajorityValue struct {
	fixed    bool
	value    bool
	variable SatVar
}

func (s *LevelMajoritySession) AddMinterm(assignment int, target bool) (LevelMajorityStatistics, error) {
	result := LevelMajorityStatistics{}
	if s == nil || s.Solver == nil || s.Topology.Inputs >= 63 || assignment < 0 || uint64(assignment) >= uint64(1)<<s.Topology.Inputs {
		return result, ErrDimension
	}
	if slices.Contains(s.Minterms, assignment) {
		return result, ErrInvariant
	}
	objects := s.Topology.Inputs + len(s.Topology.Nodes)
	values := make([]levelMajorityValue, objects)
	for input := 0; input < s.Topology.Inputs; input++ {
		values[input] = levelMajorityValue{fixed: true, value: assignment&(1<<input) != 0}
	}
	values[s.Topology.Inputs] = levelMajorityValue{fixed: true, value: exactMajorityValueUnchecked(assignment, 3)}
	values[objects-1] = levelMajorityValue{fixed: true, value: target}
	first := s.Solver.VariableCount()
	additional := max(0, len(s.Topology.Nodes)-2)
	if additional > s.Options.MaxVariables-first {
		return result, ErrLimit
	}
	for object := s.Topology.Inputs + 1; object < objects-1; object++ {
		values[object] = levelMajorityValue{variable: SatVar(first + object - s.Topology.Inputs - 1)}
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	for _, node := range s.Topology.Nodes[1:] {
		type candidate struct {
			source   int
			selector *SatVar
		}
		candidates := make([]candidate, 0, len(node.Fixed)+len(node.Selectable))
		for _, source := range node.Fixed {
			candidates = append(candidates, candidate{source: source})
		}
		for i := range node.Selectable {
			candidates = append(candidates, candidate{source: node.Selectable[i].Source, selector: &node.Selectable[i].Variable})
		}
		slices.SortFunc(candidates, func(a, b candidate) int { return a.source - b.source })
		for phase := 0; phase < 2; phase++ {
			want := phase != 0
			for left := 0; left < len(candidates); left++ {
				lv := values[candidates[left].source]
				if lv.fixed && lv.value != want {
					continue
				}
				for right := left + 1; right < len(candidates); right++ {
					rv := values[candidates[right].source]
					if rv.fixed && rv.value != want {
						continue
					}
					out := values[node.Object]
					if out.fixed && out.value == want {
						continue
					}
					clause := make([]SatLit, 0, 5)
					if !lv.fixed {
						clause = append(clause, satCardinalityLit(lv.variable, want))
					}
					if candidates[left].selector != nil {
						clause = append(clause, satCardinalityLit(*candidates[left].selector, true))
					}
					if !rv.fixed {
						clause = append(clause, satCardinalityLit(rv.variable, want))
					}
					if candidates[right].selector != nil {
						clause = append(clause, satCardinalityLit(*candidates[right].selector, true))
					}
					if !out.fixed {
						clause = append(clause, satCardinalityLit(out.variable, !want))
					}
					if len(clause) == 0 {
						return result, ErrInvariant
					}
					if err := builder.add(clause...); err != nil {
						return LevelMajorityStatistics{}, err
					}
				}
			}
		}
	}
	trial := s.Solver.cloneState()
	trial.bookmark = s.Solver.bookmark
	if err := trial.SetVariableCount(first + additional); err != nil {
		return result, err
	}
	inconsistent := trial.inconsistent
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		ok, err := trial.AddClause(clause)
		if err != nil {
			return LevelMajorityStatistics{}, err
		}
		inconsistent = !ok
	}
	*s.Solver = trial
	s.Minterms = append(s.Minterms, assignment)
	result.ExactCNFStatistics = ExactCNFStatistics{FirstVariable: first, Variables: additional, Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return result, nil
}

func exactMajorityValueUnchecked(assignment, inputs int) bool {
	count := 0
	for i := 0; i < inputs; i++ {
		count += assignment >> i & 1
	}
	return count > inputs/2
}

func (s *LevelMajoritySession) Decode(model []bool) (ExactNetwork, error) {
	if s == nil || len(model) < s.Topology.VariableCount {
		return ExactNetwork{}, ErrDimension
	}
	network := ExactNetwork{Inputs: s.Topology.Inputs, Gates: make([]ExactGate, len(s.Topology.Nodes))}
	for row, node := range s.Topology.Nodes {
		selected, _, _, err := levelMajoritySelected(node, model)
		if err != nil {
			return ExactNetwork{}, err
		}
		if len(selected) != 3 {
			return ExactNetwork{}, ErrInvariant
		}
		network.Gates[row] = ExactGate{Fanins: selected, Majority: true}
	}
	if err := network.Check(ExactSynthesisLimits{}); err != nil {
		return ExactNetwork{}, err
	}
	return network, nil
}

type LevelMajoritySynthesisResult struct {
	ExactSynthesisResult
	LazyConstraints int
}

func SynthesizeLevelMajority(inputs int, levels []int, options ExactSynthesisOptions) (LevelMajoritySynthesisResult, error) {
	result := LevelMajoritySynthesisResult{ExactSynthesisResult: ExactSynthesisResult{Status: SatUnknown}}
	if options.MaxIterations < 0 {
		return result, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1 << inputs
	}
	limits := options.Limits
	limits.MaxVariables = max(limits.MaxVariables, options.CNF.MaxVariables)
	limits.Stop = faultCombinedStop(options.Stop, limits.Stop)
	topology, err := BuildLevelMajorityTopology(inputs, levels, limits)
	if err != nil {
		return result, err
	}
	options.CNF.Stop = faultCombinedStop(options.Stop, options.CNF.Stop)
	options.SAT.Stop = faultCombinedStop(options.Stop, options.SAT.Stop)
	session, _, err := NewLevelMajoritySession(topology, options.CNF)
	if err != nil {
		return result, err
	}
	target, err := ExactMajorityTruth(inputs, limits)
	if err != nil {
		return result, err
	}
	assignment := 0
	for result.Iterations < options.MaxIterations {
		stats, addErr := session.AddMinterm(assignment, target[assignment])
		if addErr != nil {
			return LevelMajoritySynthesisResult{}, addErr
		}
		result.Iterations++
		result.Minterms = append(result.Minterms, assignment)
		if stats.Inconsistent {
			result.Status = SatUnsatisfiable
			break
		}
		for {
			answer, solveErr := session.Solver.Solve(nil, options.SAT)
			if solveErr != nil {
				return LevelMajoritySynthesisResult{}, solveErr
			}
			result.Status, result.Reason = answer.Status, answer.Reason
			if answer.Status != SatSatisfiable {
				break
			}
			lazy, refineErr := session.RefineTopologyModel(answer.Model)
			if refineErr != nil {
				return LevelMajoritySynthesisResult{}, refineErr
			}
			result.LazyConstraints += lazy.LazyConstraints
			if lazy.LazyConstraints == 0 {
				network, decodeErr := session.Decode(answer.Model)
				if decodeErr != nil {
					return LevelMajoritySynthesisResult{}, decodeErr
				}
				mismatch, mismatchErr := network.FirstMismatch(target, true, limits)
				if mismatchErr != nil {
					return LevelMajoritySynthesisResult{}, mismatchErr
				}
				if mismatch == -1 {
					result.Network = &network
				} else {
					assignment = mismatch
				}
				break
			}
		}
		if result.Status != SatSatisfiable || result.Network != nil {
			break
		}
	}
	if result.Network == nil && result.Status == SatSatisfiable && result.Iterations >= options.MaxIterations {
		result.Status = SatUnknown
		result.IterationLimitReached = true
	}
	result.Variables = session.Solver.VariableCount()
	result.Clauses = session.Solver.OriginalClauseCount()
	result.Conflicts = session.Solver.ConflictCount()
	return result, nil
}
