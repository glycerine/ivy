package abc

// Exact-synthesis CEGAR loops from src/sat/bmc/bmcMaj.c.

import "slices"

type ExactSynthesisOptions struct {
	Limits        ExactSynthesisLimits
	CNF           ExactCNFOptions
	SAT           SatLimits
	MaxIterations int
	MiddleOnly    bool
	Stop          func() bool
}

type ExactSynthesisResult struct {
	Status                SatStatus
	Reason                SatStopReason
	Network               *ExactNetwork
	Minterms              []int
	Iterations            int
	ComplementedTarget    bool
	IterationLimitReached bool
	Variables, Clauses    int
	Conflicts             uint64
}

func SynthesizeExactNetwork(topology ExactSynthesisTopology, target []bool, options ExactSynthesisOptions) (ExactSynthesisResult, error) {
	result := ExactSynthesisResult{Status: SatUnknown}
	if topology.Inputs < 1 || topology.Inputs >= 63 || len(target) != 1<<topology.Inputs || options.MaxIterations < 0 {
		return result, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = len(target)
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
	session, _, err := NewExactSynthesisSession(topology, options.CNF)
	if err != nil {
		return result, err
	}
	assignment := 1
	if topology.ConstantObjects == 2 {
		assignment = 0
	}
	for result.Iterations < options.MaxIterations {
		if options.Stop != nil && options.Stop() {
			return ExactSynthesisResult{}, ErrCancelled
		}
		added, addErr := session.AddMinterm(assignment, normalized[assignment])
		if addErr != nil {
			return ExactSynthesisResult{}, addErr
		}
		result.Minterms = append(result.Minterms, assignment)
		result.Iterations++
		if added.Inconsistent {
			result.Status, result.Reason = SatUnsatisfiable, SatFinished
			break
		}
		answer, solveErr := session.Solver.Solve(nil, options.SAT)
		if solveErr != nil {
			return ExactSynthesisResult{}, solveErr
		}
		result.Status, result.Reason = answer.Status, answer.Reason
		if answer.Status != SatSatisfiable {
			break
		}
		network, decodeErr := session.Decode(answer.Model, false)
		if decodeErr != nil {
			return ExactSynthesisResult{}, decodeErr
		}
		semanticLimits := options.Limits
		semanticLimits.MaxAssignments = max(semanticLimits.MaxAssignments, len(target))
		semanticLimits.Stop = faultCombinedStop(options.Stop, semanticLimits.Stop)
		mismatch, mismatchErr := network.FirstMismatch(normalized, options.MiddleOnly, semanticLimits)
		if mismatchErr != nil {
			return ExactSynthesisResult{}, mismatchErr
		}
		if mismatch == -1 {
			network.ComplementOutput = result.ComplementedTarget
			result.Network = &network
			break
		}
		assignment = mismatch
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

func SynthesizeExactMajority(inputs, nodes int, useConstants, cascade bool, options ExactSynthesisOptions) (ExactSynthesisResult, error) {
	limits := options.Limits
	limits.MaxVariables = max(limits.MaxVariables, options.CNF.MaxVariables)
	limits.Stop = faultCombinedStop(options.Stop, limits.Stop)
	topology, err := BuildMajorityExactTopology(inputs, nodes, useConstants, cascade, limits)
	if err != nil {
		return ExactSynthesisResult{Status: SatUnknown}, err
	}
	target, err := ExactMajorityTruth(inputs, limits)
	if err != nil {
		return ExactSynthesisResult{Status: SatUnknown}, err
	}
	options.MiddleOnly = true
	return SynthesizeExactNetwork(topology, target, options)
}
