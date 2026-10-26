package abc

// Lazy tuple-cached Exa3 minterm CNF from src/sat/bmc/bmcMaj.c.

import "slices"

type exactLazyKey struct {
	assignment, node int
	count            uint8
	fanins           [6]int
}

type ExactLazyStatistics struct {
	ExactCNFStatistics
	TriedTuples, AddedTuples, CacheHits int
}

type ExactLazySynthesisSession struct {
	Topology       ExactSynthesisTopology
	Solver         *SatSolver
	Options        ExactCNFOptions
	EvaluationBase int
	Tried, Added   int
	cache          map[exactLazyKey]struct{}
}

func NewExactLazySynthesisSession(topology ExactSynthesisTopology, options ExactCNFOptions) (*ExactLazySynthesisSession, ExactCNFStatistics, error) {
	base, statistics, err := NewExactSynthesisSession(topology, options)
	if err != nil {
		return nil, statistics, err
	}
	if topology.ConstantObjects != 0 || topology.LUTSize < 2 || topology.LUTSize > 3 || topology.Inputs >= 63 {
		return nil, ExactCNFStatistics{}, ErrDimension
	}
	assignments := uint64(1) << topology.Inputs
	if assignments > uint64(base.Options.MaxAssignments) {
		return nil, ExactCNFStatistics{}, ErrLimit
	}
	extra := uint64(topology.InternalNodes) * assignments
	if extra > uint64(base.Options.MaxVariables-base.Solver.VariableCount()) {
		return nil, ExactCNFStatistics{}, ErrLimit
	}
	trial := base.Solver.cloneState()
	trial.bookmark = base.Solver.bookmark
	if err = trial.SetVariableCount(base.Solver.VariableCount() + int(extra)); err != nil {
		return nil, ExactCNFStatistics{}, err
	}
	result := &ExactLazySynthesisSession{
		Topology: topology.Clone(), Solver: &trial, Options: base.Options,
		EvaluationBase: base.Solver.VariableCount(), cache: make(map[exactLazyKey]struct{}),
	}
	statistics.Variables += int(extra)
	return result, statistics, nil
}

func exactSelectedFanins(topology ExactSynthesisTopology, model []bool) ([][]int, error) {
	if len(model) < topology.VariableCount {
		return nil, ErrDimension
	}
	result := make([][]int, len(topology.Nodes))
	for row, node := range topology.Nodes {
		result[row] = make([]int, len(node.Pins))
		for pinIndex, pin := range node.Pins {
			found := -1
			for i, variable := range pin.Variables {
				if int(variable) >= len(model) {
					return nil, ErrDimension
				}
				if model[variable] {
					if found != -1 {
						return nil, ErrInvariant
					}
					found = pin.Candidates[i]
				}
			}
			if found == -1 {
				return nil, ErrInvariant
			}
			result[row][pinIndex] = found
		}
	}
	return result, nil
}

func (s *ExactLazySynthesisSession) evaluationVariable(assignment, object int) (SatVar, error) {
	row := object - s.Topology.Inputs
	if assignment < 0 || uint64(assignment) >= uint64(1)<<s.Topology.Inputs || row < 0 || row >= s.Topology.InternalNodes {
		return 0, ErrDimension
	}
	return SatVar(s.EvaluationBase + assignment*s.Topology.InternalNodes + row), nil
}

func (s *ExactLazySynthesisSession) AddLazyMinterm(assignment int, target bool, model []bool) (ExactLazyStatistics, error) {
	result := ExactLazyStatistics{}
	if s == nil || s.Solver == nil || assignment < 0 || s.Topology.Inputs >= 63 || uint64(assignment) >= uint64(1)<<s.Topology.Inputs {
		return result, ErrDimension
	}
	fanins, err := exactSelectedFanins(s.Topology, model)
	if err != nil {
		return result, err
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	newKeys := make([]exactLazyKey, 0, len(s.Topology.Nodes))
	for row, node := range s.Topology.Nodes {
		key := exactLazyKey{assignment: assignment, node: row, count: uint8(len(fanins[row]))}
		copy(key.fanins[:], fanins[row])
		result.TriedTuples++
		if _, exists := s.cache[key]; exists {
			result.CacheHits++
			continue
		}
		newKeys = append(newKeys, key)
		result.AddedTuples++
		final := row == len(s.Topology.Nodes)-1
		mask := (1 << s.Topology.LUTSize) - 1
		for value := 0; value < 2; value++ {
			if final && (value != 0) == target {
				continue
			}
			for pattern := 0; pattern <= mask; pattern++ {
				if pattern == 0 && value == 1 {
					continue
				}
				clause := make([]SatLit, 0, 2*s.Topology.LUTSize+2)
				valid := true
				for pin := 0; pin < s.Topology.LUTSize; pin++ {
					source := fanins[row][pin]
					selection, selected := SatVar(0), false
					for i, candidate := range node.Pins[pin].Candidates {
						if candidate == source {
							selection = node.Pins[pin].Variables[i]
							selected = true
							break
						}
					}
					if !selected {
						return ExactLazyStatistics{}, ErrInvariant
					}
					clause = append(clause, satCardinalityLit(selection, true))
					bit := pattern&(1<<pin) != 0
					if source < s.Topology.Inputs {
						if (assignment&(1<<source) != 0) != bit {
							valid = false
							break
						}
					} else {
						variable, variableErr := s.evaluationVariable(assignment, source)
						if variableErr != nil {
							return ExactLazyStatistics{}, variableErr
						}
						clause = append(clause, satCardinalityLit(variable, bit))
					}
				}
				if !valid {
					continue
				}
				if !final {
					variable, variableErr := s.evaluationVariable(assignment, node.Object)
					if variableErr != nil {
						return ExactLazyStatistics{}, variableErr
					}
					clause = append(clause, satCardinalityLit(variable, value == 0))
				}
				if pattern > 0 {
					clause = append(clause, satCardinalityLit(node.FunctionVariables[pattern-1], value != 0))
				}
				if err = builder.add(clause...); err != nil {
					return ExactLazyStatistics{}, err
				}
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
		ok, addErr := trial.AddClause(clause)
		if addErr != nil {
			return ExactLazyStatistics{}, addErr
		}
		inconsistent = !ok
	}
	*s.Solver = trial
	for _, key := range newKeys {
		s.cache[key] = struct{}{}
	}
	s.Tried += result.TriedTuples
	s.Added += result.AddedTuples
	result.ExactCNFStatistics = ExactCNFStatistics{Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return result, nil
}

func (s *ExactLazySynthesisSession) Decode(model []bool, complementOutput bool) (ExactNetwork, error) {
	if s == nil {
		return ExactNetwork{}, ErrDimension
	}
	base := ExactSynthesisSession{Topology: s.Topology}
	return base.Decode(model, complementOutput)
}

type ExactLazySynthesisResult struct {
	ExactSynthesisResult
	TriedTuples, AddedTuples int
}

func SynthesizeExactNetworkLazy(topology ExactSynthesisTopology, target []bool, options ExactSynthesisOptions) (ExactLazySynthesisResult, error) {
	result := ExactLazySynthesisResult{ExactSynthesisResult: ExactSynthesisResult{Status: SatUnknown}}
	if topology.Inputs < 1 || topology.Inputs >= 63 || len(target) != 1<<topology.Inputs || options.MaxIterations < 0 {
		return result, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 10000
	}
	options.CNF.Stop = faultCombinedStop(options.Stop, options.CNF.Stop)
	options.SAT.Stop = faultCombinedStop(options.Stop, options.SAT.Stop)
	normalized := slices.Clone(target)
	if normalized[0] {
		result.ComplementedTarget = true
		for i := range normalized {
			normalized[i] = !normalized[i]
		}
	}
	session, _, err := NewExactLazySynthesisSession(topology, options.CNF)
	if err != nil {
		return result, err
	}
	answer, err := session.Solver.Solve(nil, options.SAT)
	if err != nil {
		return result, err
	}
	result.Status, result.Reason = answer.Status, answer.Reason
	if answer.Status != SatSatisfiable {
		return result, nil
	}
	assignment := 1
	for result.Iterations < options.MaxIterations {
		if options.Stop != nil && options.Stop() {
			return ExactLazySynthesisResult{}, ErrCancelled
		}
		added, addErr := session.AddLazyMinterm(assignment, normalized[assignment], answer.Model)
		if addErr != nil {
			return ExactLazySynthesisResult{}, addErr
		}
		result.Minterms = append(result.Minterms, assignment)
		result.Iterations++
		if added.Inconsistent {
			result.Status, result.Reason = SatUnsatisfiable, SatFinished
			break
		}
		answer, err = session.Solver.Solve(nil, options.SAT)
		if err != nil {
			return ExactLazySynthesisResult{}, err
		}
		result.Status, result.Reason = answer.Status, answer.Reason
		if answer.Status != SatSatisfiable {
			break
		}
		network, decodeErr := session.Decode(answer.Model, false)
		if decodeErr != nil {
			return ExactLazySynthesisResult{}, decodeErr
		}
		limits := options.Limits
		limits.MaxAssignments = max(limits.MaxAssignments, len(target))
		limits.Stop = faultCombinedStop(options.Stop, limits.Stop)
		mismatch, mismatchErr := network.FirstMismatch(normalized, false, limits)
		if mismatchErr != nil {
			return ExactLazySynthesisResult{}, mismatchErr
		}
		if mismatch == -1 {
			network.ComplementOutput = result.ComplementedTarget
			result.Network = &network
			break
		}
		assignment = mismatch
	}
	if result.Network == nil && result.Status == SatSatisfiable && result.Iterations >= options.MaxIterations {
		result.Status, result.IterationLimitReached = SatUnknown, true
	}
	result.Variables, result.Clauses = session.Solver.VariableCount(), session.Solver.OriginalClauseCount()
	result.Conflicts = session.Solver.ConflictCount()
	result.TriedTuples, result.AddedTuples = session.Tried, session.Added
	return result, nil
}
