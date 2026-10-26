package abc

// Lazy/eager functionality and complete Zyx synthesis from src/sat/bmc/bmcMaj3.c.

import "slices"

func subsetCombinations(limit, count int, visit func([]int) error) error {
	if count < 0 || count > limit {
		return ErrDimension
	}
	choice := make([]int, count)
	var walk func(int, int) error
	walk = func(index, start int) error {
		if index == count {
			return visit(choice)
		}
		for value := start; value <= limit-(count-index); value++ {
			choice[index] = value
			if err := walk(index+1, value+1); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(0, 0)
}

func (s *SubsetExactSession) addFunctionClauses(builder *exactClauseBuilder, assignment, node int, fanins []int, majority bool) error {
	if len(fanins) != s.Topology.LUTSize {
		return ErrDimension
	}
	if majority {
		pairs := [][2]int{{0, 1}, {0, 2}, {1, 2}}
		for _, pair := range pairs {
			for value := 0; value < 2; value++ {
				clause := make([]SatLit, 0, 5)
				for _, index := range pair {
					selection, _ := s.Topology.SelectionVariable(node, fanins[index])
					source, _ := s.Topology.EvaluationVariable(assignment, fanins[index])
					clause = append(clause, satCardinalityLit(selection, true), satCardinalityLit(source, value != 0))
				}
				output, _ := s.Topology.EvaluationVariable(assignment, node)
				clause = append(clause, satCardinalityLit(output, value == 0))
				if err := builder.add(clause...); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for pattern := 0; pattern <= s.Topology.LUTMask; pattern++ {
		function, _ := s.Topology.FunctionVariable(node, pattern)
		for value := 0; value < 2; value++ {
			clause := []SatLit{satCardinalityLit(function, value != 0)}
			for pin, sourceObject := range fanins {
				selection, _ := s.Topology.SelectionVariable(node, sourceObject)
				source, _ := s.Topology.EvaluationVariable(assignment, sourceObject)
				clause = append(clause, satCardinalityLit(selection, true), satCardinalityLit(source, pattern&(1<<pin) != 0))
			}
			output, _ := s.Topology.EvaluationVariable(assignment, node)
			clause = append(clause, satCardinalityLit(output, value == 0))
			if err := builder.add(clause...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SubsetExactSession) commitFunctionClauses(builder exactClauseBuilder) (ExactCNFStatistics, error) {
	trial := s.Solver.cloneState()
	trial.bookmark = s.Solver.bookmark
	inconsistent := trial.inconsistent
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		ok, err := trial.AddClause(clause)
		if err != nil {
			return ExactCNFStatistics{}, err
		}
		inconsistent = !ok
	}
	*s.Solver = trial
	return ExactCNFStatistics{Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}, nil
}

// AddLazyFunction adds functionality only for fanin tuples selected in model,
// preserving Zyx_ManAddCnfLazyFunc2 and its tuple caches.
func (s *SubsetExactSession) AddLazyFunction(assignment int, model []bool) (SubsetExactStatistics, error) {
	result := SubsetExactStatistics{}
	if s == nil || s.Solver == nil || assignment < 0 || assignment >= len(s.Target) || len(model) < s.Topology.VariableCount {
		return result, ErrDimension
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	var newKeys []subsetExactKey
	for node := s.Topology.Inputs; node < s.Topology.Objects; node++ {
		fanins, _, _, err := s.selectedFanins(node, model)
		if err != nil || len(fanins) != s.Topology.LUTSize {
			return result, ErrInvariant
		}
		if s.Topology.Majority {
			for _, pair := range [][2]int{{0, 1}, {0, 2}, {1, 2}} {
				key := subsetExactKey{Assignment: assignment, Node: node, Count: 2}
				key.Fanins[0], key.Fanins[1] = fanins[pair[0]], fanins[pair[1]]
				result.TriedTuples++
				if _, ok := s.cache[key]; ok {
					result.CacheHits++
					continue
				}
				newKeys = append(newKeys, key)
				result.AddedTuples++
				if err = s.addMajorityPair(&builder, assignment, node, key.Fanins[0], key.Fanins[1]); err != nil {
					return SubsetExactStatistics{}, err
				}
			}
			continue
		}
		if s.Topology.LUTSize > 3 {
			if err = s.addFunctionClauses(&builder, assignment, node, fanins, false); err != nil {
				return SubsetExactStatistics{}, err
			}
			continue
		}
		key := subsetExactKey{Assignment: assignment, Node: node, Count: uint8(len(fanins))}
		copy(key.Fanins[:], fanins)
		result.TriedTuples++
		if _, ok := s.cache[key]; ok {
			result.CacheHits++
			continue
		}
		newKeys = append(newKeys, key)
		result.AddedTuples++
		if err = s.addFunctionClauses(&builder, assignment, node, fanins, false); err != nil {
			return SubsetExactStatistics{}, err
		}
	}
	statistics, err := s.commitFunctionClauses(builder)
	if err != nil {
		return SubsetExactStatistics{}, err
	}
	for _, key := range newKeys {
		s.cache[key] = struct{}{}
	}
	s.Counts[assignment]++
	s.Tried += result.TriedTuples
	s.Added += result.AddedTuples
	result.ExactCNFStatistics = statistics
	return result, nil
}

func (s *SubsetExactSession) addMajorityPair(builder *exactClauseBuilder, assignment, node, left, right int) error {
	for value := 0; value < 2; value++ {
		clause := make([]SatLit, 0, 5)
		for _, sourceObject := range []int{left, right} {
			selection, _ := s.Topology.SelectionVariable(node, sourceObject)
			source, _ := s.Topology.EvaluationVariable(assignment, sourceObject)
			clause = append(clause, satCardinalityLit(selection, true), satCardinalityLit(source, value != 0))
		}
		output, _ := s.Topology.EvaluationVariable(assignment, node)
		clause = append(clause, satCardinalityLit(output, value == 0))
		if err := builder.add(clause...); err != nil {
			return err
		}
	}
	return nil
}

// AddEagerFunction adds functionality for every possible fanin tuple for one
// minterm, preserving Zyx_ManAddCnfLazyFunc's nonincremental mode.
func (s *SubsetExactSession) AddEagerFunction(assignment int) (SubsetExactStatistics, error) {
	result := SubsetExactStatistics{}
	if s == nil || s.Solver == nil || s.Topology.Majority || s.Topology.LUTSize > 3 || assignment < 0 || assignment >= len(s.Target) {
		return result, ErrDimension
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	for node := s.Topology.Inputs; node < s.Topology.Objects; node++ {
		if err := subsetCombinations(node, s.Topology.LUTSize, func(fanins []int) error {
			return s.addFunctionClauses(&builder, assignment, node, slices.Clone(fanins), false)
		}); err != nil {
			return SubsetExactStatistics{}, err
		}
	}
	statistics, err := s.commitFunctionClauses(builder)
	if err != nil {
		return SubsetExactStatistics{}, err
	}
	s.Counts[assignment]++
	result.ExactCNFStatistics = statistics
	return result, nil
}

// Decode returns the exact network represented by a fully refined model.
func (s *SubsetExactSession) Decode(model []bool, complementOutput bool) (ExactNetwork, error) {
	if s == nil || len(model) < s.Topology.VariableCount {
		return ExactNetwork{}, ErrDimension
	}
	network := ExactNetwork{Inputs: s.Topology.Inputs, Gates: make([]ExactGate, s.Topology.InternalNodes), ComplementOutput: complementOutput}
	for row := range network.Gates {
		node := s.Topology.Inputs + row
		fanins, _, _, err := s.selectedFanins(node, model)
		if err != nil || len(fanins) != s.Topology.LUTSize {
			return ExactNetwork{}, ErrInvariant
		}
		gate := ExactGate{Fanins: fanins, Majority: s.Topology.Majority}
		if !gate.Majority {
			for pattern := 0; pattern <= s.Topology.LUTMask; pattern++ {
				variable, _ := s.Topology.FunctionVariable(node, pattern)
				if model[variable] {
					gate.Truth |= uint64(1) << pattern
				}
			}
		}
		network.Gates[row] = gate
	}
	if err := network.Check(ExactSynthesisLimits{}); err != nil {
		return ExactNetwork{}, err
	}
	return network, nil
}

// BlockTopology blocks the selected fanin sets while allowing other function
// tables, matching Zyx_ManAddCnfBlockSolution's topology enumeration.
func (s *SubsetExactSession) BlockTopology(model []bool) (ExactCNFStatistics, error) {
	if s == nil || s.Solver == nil || len(model) < s.Topology.VariableCount {
		return ExactCNFStatistics{}, ErrDimension
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	var clause []SatLit
	for node := s.Topology.Inputs; node < s.Topology.Objects; node++ {
		fanins, _, _, err := s.selectedFanins(node, model)
		if err != nil || len(fanins) != s.Topology.LUTSize {
			return ExactCNFStatistics{}, ErrInvariant
		}
		for _, source := range fanins {
			variable, _ := s.Topology.SelectionVariable(node, source)
			clause = append(clause, satCardinalityLit(variable, true))
		}
	}
	if err := builder.add(clause...); err != nil {
		return ExactCNFStatistics{}, err
	}
	return s.commitFunctionClauses(builder)
}

type SubsetExactSynthesisOptions struct {
	Exact        ExactSynthesisOptions
	LUTSize      int
	Majority     bool
	OnlyAND      bool
	OrderNodes   bool
	Incremental  bool
	Enumerate    bool
	MaxSolutions int
}

type SubsetExactSynthesisResult struct {
	ExactSynthesisResult
	Networks                               []ExactNetwork
	LazyTopology, TriedTuples, AddedTuples int
}

func SynthesizeSubsetExact(inputs, nodes int, target []bool, options SubsetExactSynthesisOptions) (SubsetExactSynthesisResult, error) {
	result := SubsetExactSynthesisResult{ExactSynthesisResult: ExactSynthesisResult{Status: SatUnknown}}
	if options.Exact.MaxIterations < 0 || options.MaxSolutions < 0 || options.Enumerate && options.MaxSolutions == 0 {
		return result, ErrDimension
	}
	if options.LUTSize == 0 {
		options.LUTSize = 3
	}
	if options.Exact.MaxIterations == 0 {
		options.Exact.MaxIterations = 10000
	}
	options.Exact.CNF.Stop = faultCombinedStop(options.Exact.Stop, options.Exact.CNF.Stop)
	options.Exact.SAT.Stop = faultCombinedStop(options.Exact.Stop, options.Exact.SAT.Stop)
	normalized := slices.Clone(target)
	if options.Majority {
		normalized = nil
	} else if inputs < 1 || inputs >= 63 || len(normalized) != 1<<inputs {
		return result, ErrDimension
	} else if normalized[0] {
		result.ComplementedTarget = true
		for i := range normalized {
			normalized[i] = !normalized[i]
		}
	}
	session, _, err := NewSubsetExactSession(inputs, nodes, options.LUTSize, options.Majority, normalized, options.OnlyAND, options.Exact.CNF)
	if err != nil {
		return result, err
	}
	semanticTarget := session.Target
	for result.Iterations < options.Exact.MaxIterations {
		var answer SatResult
		for {
			answer, err = session.Solver.Solve(nil, options.Exact.SAT)
			if err != nil {
				return SubsetExactSynthesisResult{}, err
			}
			result.Status, result.Reason = answer.Status, answer.Reason
			if answer.Status != SatSatisfiable {
				break
			}
			refined, refineErr := session.RefineTopologyModel(answer.Model, options.OrderNodes)
			if refineErr != nil {
				return SubsetExactSynthesisResult{}, refineErr
			}
			result.LazyTopology += refined.LazyTopology
			if refined.LazyTopology == 0 {
				break
			}
		}
		if result.Status != SatSatisfiable {
			break
		}
		network, decodeErr := session.Decode(answer.Model, false)
		if decodeErr != nil {
			return SubsetExactSynthesisResult{}, decodeErr
		}
		limits := options.Exact.Limits
		limits.MaxAssignments = max(limits.MaxAssignments, len(semanticTarget))
		limits.Stop = faultCombinedStop(options.Exact.Stop, limits.Stop)
		mismatch, mismatchErr := network.FirstMismatch(semanticTarget, options.Majority, limits)
		if mismatchErr != nil {
			return SubsetExactSynthesisResult{}, mismatchErr
		}
		if mismatch == -1 {
			network.ComplementOutput = result.ComplementedTarget
			result.Network = &network
			result.Networks = append(result.Networks, network)
			if !options.Enumerate || len(result.Networks) >= options.MaxSolutions {
				break
			}
			if _, err = session.BlockTopology(answer.Model); err != nil {
				return SubsetExactSynthesisResult{}, err
			}
			continue
		}
		result.Minterms = append(result.Minterms, mismatch)
		result.Iterations++
		var added SubsetExactStatistics
		if options.Incremental {
			added, err = session.AddLazyFunction(mismatch, answer.Model)
		} else {
			added, err = session.AddEagerFunction(mismatch)
		}
		if err != nil {
			return SubsetExactSynthesisResult{}, err
		}
		result.TriedTuples += added.TriedTuples
		result.AddedTuples += added.AddedTuples
	}
	if result.Network == nil && result.Status == SatSatisfiable && result.Iterations >= options.Exact.MaxIterations {
		result.Status = SatUnknown
		result.IterationLimitReached = true
	}
	result.Variables = session.Solver.VariableCount()
	result.Clauses = session.Solver.OriginalClauseCount()
	result.Conflicts = session.Solver.ConflictCount()
	return result, nil
}
