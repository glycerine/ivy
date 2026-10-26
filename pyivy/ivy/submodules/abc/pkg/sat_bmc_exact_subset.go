package abc

// Generic subset-topology exact synthesis from src/sat/bmc/bmcMaj3.c (Zyx manager).

import "slices"

type SubsetExactTopology struct {
	Inputs, InternalNodes, LUTSize int
	Majority                       bool
	Objects, LUTMask               int
	TopologyBase, EvaluationBase   int
	VariableCount                  int
	SymmetricInputPairs            [][2]int
}

func (t SubsetExactTopology) Clone() SubsetExactTopology {
	r := t
	r.SymmetricInputPairs = slices.Clone(t.SymmetricInputPairs)
	return r
}

func (t SubsetExactTopology) FunctionVariable(node, pattern int) (SatVar, error) {
	if node < t.Inputs || node >= t.Objects || pattern < 0 || pattern > t.LUTMask {
		return 0, ErrDimension
	}
	return SatVar((t.LUTMask+1)*(node-t.Inputs) + pattern), nil
}

func (t SubsetExactTopology) SelectionVariable(node, source int) (SatVar, error) {
	if node < t.Inputs || node >= t.Objects || source < 0 || source >= t.Objects {
		return 0, ErrDimension
	}
	return SatVar(t.TopologyBase + t.Objects*(node-t.Inputs) + source), nil
}

func (t SubsetExactTopology) EvaluationVariable(assignment, object int) (SatVar, error) {
	if t.Inputs >= 63 || assignment < 0 || uint64(assignment) >= uint64(1)<<t.Inputs || object < 0 || object >= t.Objects {
		return 0, ErrDimension
	}
	return SatVar(t.EvaluationBase + t.Objects*assignment + object), nil
}

func exactSymmetricInputPairs(target []bool, inputs int, stop func() bool) ([][2]int, error) {
	if inputs < 1 || inputs >= 63 || len(target) != 1<<inputs {
		return nil, ErrDimension
	}
	var result [][2]int
	for left := 0; left < inputs; left++ {
		for right := left + 1; right < inputs; right++ {
			if stop != nil && stop() {
				return nil, ErrCancelled
			}
			symmetric := true
			for assignment := 0; assignment < len(target); assignment++ {
				if assignment&(1<<left) != 0 || assignment&(1<<right) == 0 {
					continue
				}
				swapped := assignment | 1<<left
				swapped &^= 1 << right
				if target[assignment] != target[swapped] {
					symmetric = false
					break
				}
			}
			if symmetric {
				result = append(result, [2]int{left, right})
			}
		}
	}
	return result, nil
}

func buildSubsetExactTopology(inputs, nodes, lutSize int, majority bool, target []bool, options ExactCNFOptions) (SubsetExactTopology, error) {
	if inputs < lutSize || inputs > 10 || nodes < 1 || lutSize < 2 || lutSize > 6 || majority && lutSize != 3 || len(target) != 1<<inputs || inputs+nodes > 32 {
		return SubsetExactTopology{}, ErrDimension
	}
	assignments := 1 << inputs
	if assignments > options.MaxAssignments {
		return SubsetExactTopology{}, ErrLimit
	}
	objects, mask := inputs+nodes, (1<<lutSize)-1
	topologyBase := (mask + 1) * nodes
	evaluationBase := topologyBase + nodes*objects
	count64 := uint64(evaluationBase) + uint64(assignments)*uint64(objects)
	if count64 > uint64(options.MaxVariables) || count64 > uint64(MaxLiteralID)+1 {
		return SubsetExactTopology{}, ErrLimit
	}
	pairs, err := exactSymmetricInputPairs(target, inputs, options.Stop)
	if err != nil {
		return SubsetExactTopology{}, err
	}
	return SubsetExactTopology{
		Inputs: inputs, InternalNodes: nodes, LUTSize: lutSize, Majority: majority,
		Objects: objects, LUTMask: mask, TopologyBase: topologyBase,
		EvaluationBase: evaluationBase, VariableCount: int(count64), SymmetricInputPairs: pairs,
	}, nil
}

type SubsetExactStatistics struct {
	ExactCNFStatistics
	LazyTopology, TriedTuples, AddedTuples, CacheHits int
}

type subsetExactKey struct {
	Assignment, Node int
	Count            uint8
	Fanins           [6]int
}

type SubsetExactSession struct {
	Topology     SubsetExactTopology
	Target       []bool
	Solver       *SatSolver
	Options      ExactCNFOptions
	OnlyAND      bool
	cache        map[subsetExactKey]struct{}
	Counts       []int
	Tried, Added int
}

func addSubsetUnit(builder *exactClauseBuilder, variable SatVar, value bool) error {
	return builder.add(satCardinalityLit(variable, !value))
}

func NewSubsetExactSession(inputs, nodes, lutSize int, majority bool, target []bool, onlyAND bool, options ExactCNFOptions) (*SubsetExactSession, SubsetExactStatistics, error) {
	statistics := SubsetExactStatistics{}
	var err error
	options, err = exactCNFLimits(options)
	if err != nil {
		return nil, statistics, err
	}
	if majority {
		target, err = ExactMajorityTruth(inputs, ExactSynthesisLimits{MaxAssignments: options.MaxAssignments, Stop: options.Stop})
		if err != nil {
			return nil, statistics, err
		}
	} else {
		target = slices.Clone(target)
	}
	topology, err := buildSubsetExactTopology(inputs, nodes, lutSize, majority, target, options)
	if err != nil {
		return nil, statistics, err
	}
	builder := exactClauseBuilder{maxClauses: options.MaxClauses, maxLiterals: options.MaxLiterals, stop: options.Stop}
	// Fixed function bit zero and invalid/future topology variables.
	for node := inputs; node < topology.Objects; node++ {
		variable, _ := topology.FunctionVariable(node, 0)
		if err = addSubsetUnit(&builder, variable, false); err != nil {
			return nil, statistics, err
		}
		for source := node; source < topology.Objects; source++ {
			variable, _ = topology.SelectionVariable(node, source)
			if err = addSubsetUnit(&builder, variable, false); err != nil {
				return nil, statistics, err
			}
		}
	}
	// The output consumes the preceding node.
	variable, _ := topology.SelectionVariable(topology.Objects-1, topology.Objects-2)
	if err = addSubsetUnit(&builder, variable, true); err != nil {
		return nil, statistics, err
	}
	// Majority fixes the first gate to inputs 0, 1 and 2.
	if majority {
		for source := 0; source < inputs; source++ {
			variable, _ = topology.SelectionVariable(inputs, source)
			if err = addSubsetUnit(&builder, variable, source < 3); err != nil {
				return nil, statistics, err
			}
		}
	}
	// All primary input values and the final target are preallocated constants.
	for assignment, output := range target {
		for input := 0; input < inputs; input++ {
			variable, _ = topology.EvaluationVariable(assignment, input)
			if err = addSubsetUnit(&builder, variable, assignment&(1<<input) != 0); err != nil {
				return nil, statistics, err
			}
		}
		variable, _ = topology.EvaluationVariable(assignment, topology.Objects-1)
		if err = addSubsetUnit(&builder, variable, output); err != nil {
			return nil, statistics, err
		}
	}
	// Every node has a fanin and every non-output object has a fanout.
	for node := inputs; node < topology.Objects; node++ {
		clause := make([]SatLit, node)
		for source := 0; source < node; source++ {
			variable, _ = topology.SelectionVariable(node, source)
			clause[source] = satCardinalityLit(variable, false)
		}
		if err = builder.add(clause...); err != nil {
			return nil, statistics, err
		}
	}
	for source := 0; source < topology.Objects-1; source++ {
		clause := make([]SatLit, 0, nodes)
		for node := inputs; node < topology.Objects; node++ {
			variable, _ = topology.SelectionVariable(node, source)
			clause = append(clause, satCardinalityLit(variable, false))
		}
		if err = builder.add(clause...); err != nil {
			return nil, statistics, err
		}
	}
	if lutSize == 2 {
		for node := inputs; node < topology.Objects; node++ {
			f1, _ := topology.FunctionVariable(node, 1)
			f2, _ := topology.FunctionVariable(node, 2)
			f3, _ := topology.FunctionVariable(node, 3)
			for k := 0; k < 3; k++ {
				if err = builder.add(satCardinalityLit(f1, k == 1), satCardinalityLit(f2, k == 2), satCardinalityLit(f3, k != 0)); err != nil {
					return nil, statistics, err
				}
			}
			if onlyAND {
				if err = builder.add(satCardinalityLit(f1, true), satCardinalityLit(f2, true), satCardinalityLit(f3, false)); err != nil {
					return nil, statistics, err
				}
			}
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
	return &SubsetExactSession{
		Topology: topology, Target: target, Solver: solver, Options: options, OnlyAND: onlyAND,
		cache: make(map[subsetExactKey]struct{}), Counts: make([]int, len(target)),
	}, statistics, nil
}

func (s *SubsetExactSession) selectedFanins(node int, model []bool) ([]int, []SatVar, []SatVar, error) {
	if s == nil || node < s.Topology.Inputs || node >= s.Topology.Objects || len(model) < s.Topology.VariableCount {
		return nil, nil, nil, ErrDimension
	}
	var fanins []int
	var on, off []SatVar
	for source := 0; source < node; source++ {
		variable, _ := s.Topology.SelectionVariable(node, source)
		if model[variable] {
			fanins = append(fanins, source)
			on = append(on, variable)
		} else {
			off = append(off, variable)
		}
	}
	return fanins, on, off, nil
}

func subsetModelBlockingLiteral(variable SatVar, value bool) SatLit {
	return satCardinalityLit(variable, value)
}

// RefineTopologyModel preserves Zyx_ManAddCnfLazyTopo's cardinality,
// adjacent-node ordering and symmetric-input first-use refinements.
func (s *SubsetExactSession) RefineTopologyModel(model []bool, orderNodes bool) (SubsetExactStatistics, error) {
	result := SubsetExactStatistics{}
	if s == nil || s.Solver == nil || len(model) < s.Topology.VariableCount {
		return result, ErrDimension
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	allFanins := make([][]int, s.Topology.InternalNodes)
	for row := range allFanins {
		node := s.Topology.Inputs + row
		fanins, on, off, err := s.selectedFanins(node, model)
		if err != nil {
			return result, err
		}
		allFanins[row] = fanins
		if len(fanins) == s.Topology.LUTSize {
			continue
		}
		result.LazyTopology++
		var clause []SatLit
		if len(fanins) > s.Topology.LUTSize {
			count := min(s.Topology.LUTSize+1, len(on))
			clause = make([]SatLit, count)
			for i := range clause {
				clause[i] = satCardinalityLit(on[i], true)
			}
		} else {
			clause = make([]SatLit, len(off))
			for i := range clause {
				clause[i] = satCardinalityLit(off[i], false)
			}
		}
		if len(clause) == 0 {
			return result, ErrInvariant
		}
		if err = builder.add(clause...); err != nil {
			return SubsetExactStatistics{}, err
		}
	}
	if result.LazyTopology == 0 && orderNodes {
		for row := 1; row < len(allFanins); row++ {
			prior, current := allFanins[row-1], allFanins[row]
			difference := -1
			for k := s.Topology.LUTSize - 1; k >= 0; k-- {
				if prior[k] != current[k] {
					difference = k
					break
				}
			}
			priorNode, currentNode := s.Topology.Inputs+row-1, s.Topology.Inputs+row
			if difference == -1 {
				if s.Topology.Majority {
					continue
				}
				bit := -1
				for k := s.Topology.LUTMask; k >= 0; k-- {
					a, _ := s.Topology.FunctionVariable(priorNode, k)
					b, _ := s.Topology.FunctionVariable(currentNode, k)
					if model[a] != model[b] {
						bit = k
						break
					}
				}
				if bit < 0 {
					continue
				}
				a, _ := s.Topology.FunctionVariable(priorNode, bit)
				b, _ := s.Topology.FunctionVariable(currentNode, bit)
				if !model[a] || model[b] {
					continue
				}
				var clause []SatLit
				for k := s.Topology.LUTMask; k >= bit; k-- {
					a, _ = s.Topology.FunctionVariable(priorNode, k)
					b, _ = s.Topology.FunctionVariable(currentNode, k)
					clause = append(clause, subsetModelBlockingLiteral(a, model[a]), subsetModelBlockingLiteral(b, model[b]))
				}
				if err := builder.add(clause...); err != nil {
					return SubsetExactStatistics{}, err
				}
				result.LazyTopology++
				continue
			}
			if prior[difference] < current[difference] {
				continue
			}
			var clause []SatLit
			for source := prior[difference]; source < s.Topology.Objects-1; source++ {
				a, _ := s.Topology.SelectionVariable(priorNode, source)
				b, _ := s.Topology.SelectionVariable(currentNode, source)
				clause = append(clause, subsetModelBlockingLiteral(a, model[a]), subsetModelBlockingLiteral(b, model[b]))
			}
			if err := builder.add(clause...); err != nil {
				return SubsetExactStatistics{}, err
			}
			result.LazyTopology++
		}
		for _, pair := range s.Topology.SymmetricInputPairs {
			firstUse := [2]int{-1, -1}
			for side, input := range pair {
				for node := s.Topology.Inputs; node < s.Topology.Objects; node++ {
					variable, _ := s.Topology.SelectionVariable(node, input)
					if model[variable] {
						firstUse[side] = node
						break
					}
				}
				if firstUse[side] < 0 {
					return SubsetExactStatistics{}, ErrInvariant
				}
			}
			if firstUse[0] <= firstUse[1] {
				continue
			}
			var clause []SatLit
			for node := s.Topology.Inputs; node <= firstUse[1]; node++ {
				for _, input := range pair {
					variable, _ := s.Topology.SelectionVariable(node, input)
					clause = append(clause, subsetModelBlockingLiteral(variable, model[variable]))
				}
			}
			if err := builder.add(clause...); err != nil {
				return SubsetExactStatistics{}, err
			}
			result.LazyTopology++
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
			return SubsetExactStatistics{}, err
		}
		inconsistent = !ok
	}
	*s.Solver = trial
	result.ExactCNFStatistics = ExactCNFStatistics{Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return result, nil
}
