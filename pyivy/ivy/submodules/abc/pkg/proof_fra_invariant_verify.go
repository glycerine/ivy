package abc

// Inductive-invariant verification from src/proof/fra/fraIndVer.c.

import "slices"

type FRAInvariantOptions struct {
	MaxObjects, MaxVariables, MaxClauses, MaxLiterals int
	MaxWork                                           int
	SAT                                               SatLimits
	Stop                                              func() bool
}

type FRAInvariantFailure struct {
	Clause int
	Status SatStatus
	Reason SatStopReason
	Model  []bool
}

type FRAInvariantResult struct {
	Verified                   bool
	InductionPremiseConsistent bool
	VariablesPerFrame          int
	BaseFailures               []FRAInvariantFailure
	InductiveFailures          []FRAInvariantFailure
}

type fraInvariantBudget struct {
	options FRAInvariantOptions
	work    int
}

func normalizeFRAInvariantOptions(options FRAInvariantOptions) (FRAInvariantOptions, error) {
	if options.MaxObjects < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 4 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 16 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 32 << 20
	}
	return options, nil
}

func (budget *fraInvariantBudget) step(amount int) error {
	if amount < 0 || budget.work > budget.options.MaxWork-amount {
		return ErrLimit
	}
	budget.work += amount
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func fraInvariantShift(clause []SatLit, offset int) []SatLit {
	result := make([]SatLit, len(clause))
	shift := SatLit(offset) << 1
	for index, literal := range clause {
		result[index] = literal + shift
	}
	return result
}

func fraInvariantModelSatisfies(model []bool, clauses [][]SatLit, assumptions []SatLit) bool {
	value := func(literal SatLit) bool {
		return model[literal.Var()] != literal.Negative()
	}
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			satisfied = satisfied || value(literal)
		}
		if !satisfied {
			return false
		}
	}
	for _, literal := range assumptions {
		if !value(literal) {
			return false
		}
	}
	return true
}

func fraInvariantSATLimits(options FRAInvariantOptions) SatLimits {
	limits := options.SAT
	satStop := limits.Stop
	limits.Stop = func() bool {
		return options.Stop != nil && options.Stop() || satStop != nil && satStop()
	}
	return limits
}

func fraInvariantAddBounded(total *int, amount, limit int) error {
	if amount < 0 || *total > limit-amount {
		return ErrLimit
	}
	*total += amount
	return nil
}

func fraInvariantLoad(clauses [][]SatLit, variables int, options FRAInvariantOptions, budget *fraInvariantBudget) (*SatSolver, bool, error) {
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, false, err
	}
	if err = solver.SetVariableCount(variables); err != nil {
		return nil, false, err
	}
	for _, clause := range clauses {
		if err = budget.step(len(clause) + 1); err != nil {
			return nil, false, err
		}
		consistent, addErr := solver.AddClause(clause)
		if addErr != nil {
			return nil, false, addErr
		}
		if !consistent {
			return solver, false, nil
		}
	}
	consistent, err := solver.Simplify()
	return solver, consistent, err
}

func fraInvariantSolveClauses(solver *SatSolver, formula [][]SatLit, clauses [][]SatLit, frame, stride int, options FRAInvariantOptions, budget *fraInvariantBudget) ([]FRAInvariantFailure, error) {
	failures := make([]FRAInvariantFailure, 0)
	for index, clause := range clauses {
		if err := budget.step(len(clause) + 1); err != nil {
			return nil, err
		}
		shifted := fraInvariantShift(clause, frame*stride)
		assumptions := make([]SatLit, len(shifted))
		for literal := range shifted {
			assumptions[literal] = shifted[literal].Not()
		}
		answer, err := solver.Solve(assumptions, fraInvariantSATLimits(options))
		if err != nil {
			return nil, err
		}
		if answer.Status == SatUnsatisfiable {
			continue
		}
		if answer.Status == SatUnknown && answer.Reason == SatCancelled {
			return nil, ErrCancelled
		}
		failure := FRAInvariantFailure{Clause: index, Status: answer.Status, Reason: answer.Reason, Model: slices.Clone(answer.Model)}
		if answer.Status == SatSatisfiable && !fraInvariantModelSatisfies(answer.Model, formula, assumptions) {
			return nil, ErrInvariant
		}
		failures = append(failures, failure)
	}
	return failures, nil
}

// VerifyFRAInvariant preserves Fra_InvariantVerify's one-frame base and
// induction checks. Clauses are disjunctions of source AIG object literals.
// Unsupported K values return an error instead of the C routine's accidental
// success result.
func VerifyFRAInvariant(graph *AigManager, frames int, clauses [][]AigLit, options FRAInvariantOptions) (FRAInvariantResult, error) {
	options, err := normalizeFRAInvariantOptions(options)
	if err != nil {
		return FRAInvariantResult{}, err
	}
	if graph == nil {
		return FRAInvariantResult{}, ErrDimension
	}
	if frames != 1 {
		return FRAInvariantResult{}, ErrUnsupported
	}
	if err = graph.Check(); err != nil {
		return FRAInvariantResult{}, err
	}
	if graph.ObjectCount() > options.MaxObjects {
		return FRAInvariantResult{}, ErrLimit
	}
	if len(clauses) > options.MaxClauses {
		return FRAInvariantResult{}, ErrLimit
	}
	budget := &fraInvariantBudget{options: options}
	relation, err := EncodeAigCNF(graph)
	if err != nil {
		return FRAInvariantResult{}, err
	}
	stride := relation.VariableCount()
	if stride > options.MaxVariables || stride > options.MaxVariables/2 {
		return FRAInvariantResult{}, ErrLimit
	}
	converted := make([][]SatLit, len(clauses))
	invariantLiterals := 0
	for clauseIndex, clause := range clauses {
		if err = budget.step(len(clause) + 1); err != nil {
			return FRAInvariantResult{}, err
		}
		if invariantLiterals > options.MaxLiterals-len(clause) {
			return FRAInvariantResult{}, ErrLimit
		}
		invariantLiterals += len(clause)
		converted[clauseIndex] = make([]SatLit, len(clause))
		for literalIndex, literal := range clause {
			if !literal.Valid() || uint64(literal.ID()) >= uint64(graph.ObjectCount()) {
				return FRAInvariantResult{}, ErrInvalidLiteral
			}
			mapped, mapErr := relation.ObjectLiteral(literal.ID())
			if mapErr != nil {
				return FRAInvariantResult{}, mapErr
			}
			if literal.Complemented() {
				mapped = mapped.Not()
			}
			converted[clauseIndex][literalIndex] = mapped
		}
	}
	result := FRAInvariantResult{InductionPremiseConsistent: true, VariablesPerFrame: stride}
	relationClauses := relation.Clauses()
	baseClauseCount := 0
	baseLiteralCount := 0
	for _, clause := range relationClauses {
		if err = fraInvariantAddBounded(&baseClauseCount, 1, options.MaxClauses); err != nil {
			return FRAInvariantResult{}, err
		}
		if err = fraInvariantAddBounded(&baseLiteralCount, len(clause), options.MaxLiterals); err != nil {
			return FRAInvariantResult{}, err
		}
	}
	if err = fraInvariantAddBounded(&baseClauseCount, graph.RegisterCount(), options.MaxClauses); err != nil {
		return FRAInvariantResult{}, err
	}
	if err = fraInvariantAddBounded(&baseLiteralCount, graph.RegisterCount(), options.MaxLiterals); err != nil {
		return FRAInvariantResult{}, err
	}
	inductionClauseCount, inductionLiteralCount := 0, 0
	for frame := 0; frame < 2; frame++ {
		for _, clause := range relationClauses {
			if err = fraInvariantAddBounded(&inductionClauseCount, 1, options.MaxClauses); err != nil {
				return FRAInvariantResult{}, err
			}
			if err = fraInvariantAddBounded(&inductionLiteralCount, len(clause), options.MaxLiterals); err != nil {
				return FRAInvariantResult{}, err
			}
		}
	}
	if err = fraInvariantAddBounded(&inductionClauseCount, graph.RegisterCount(), options.MaxClauses); err != nil {
		return FRAInvariantResult{}, err
	}
	if graph.RegisterCount() > options.MaxLiterals/2 || fraInvariantAddBounded(&inductionLiteralCount, 2*graph.RegisterCount(), options.MaxLiterals) != nil {
		return FRAInvariantResult{}, ErrLimit
	}
	if err = fraInvariantAddBounded(&inductionClauseCount, len(converted), options.MaxClauses); err != nil {
		return FRAInvariantResult{}, err
	}
	if err = fraInvariantAddBounded(&inductionLiteralCount, invariantLiterals, options.MaxLiterals); err != nil {
		return FRAInvariantResult{}, err
	}
	baseFormula := make([][]SatLit, 0, baseClauseCount)
	for _, clause := range relationClauses {
		baseFormula = append(baseFormula, slices.Clone(clause))
	}
	inputs := relation.InputLiterals()
	for _, literal := range inputs[len(inputs)-graph.RegisterCount():] {
		baseFormula = append(baseFormula, []SatLit{literal.Not()})
	}
	baseSolver, _, err := fraInvariantLoad(baseFormula, stride, options, budget)
	if err != nil {
		return FRAInvariantResult{}, err
	}
	result.BaseFailures, err = fraInvariantSolveClauses(baseSolver, baseFormula, converted, 0, stride, options, budget)
	if err != nil {
		return FRAInvariantResult{}, err
	}

	inductionFormula := make([][]SatLit, 0, inductionClauseCount)
	for frame := 0; frame < 2; frame++ {
		for _, clause := range relationClauses {
			inductionFormula = append(inductionFormula, fraInvariantShift(clause, frame*stride))
		}
	}
	outputs := relation.OutputLiterals()
	for register := 0; register < graph.RegisterCount(); register++ {
		li := outputs[len(outputs)-graph.RegisterCount()+register]
		lo := inputs[len(inputs)-graph.RegisterCount()+register] + (SatLit(stride) << 1)
		inductionFormula = append(inductionFormula, []SatLit{li, lo.Not()}, []SatLit{li.Not(), lo})
	}
	for _, clause := range converted {
		inductionFormula = append(inductionFormula, slices.Clone(clause))
	}
	inductionSolver, consistent, err := fraInvariantLoad(inductionFormula, 2*stride, options, budget)
	if err != nil {
		return FRAInvariantResult{}, err
	}
	if !consistent {
		result.InductionPremiseConsistent = false
		return result, nil
	}
	result.InductiveFailures, err = fraInvariantSolveClauses(inductionSolver, inductionFormula, converted, 1, stride, options, budget)
	if err != nil {
		return FRAInvariantResult{}, err
	}
	result.Verified = len(result.BaseFailures) == 0 && len(result.InductiveFailures) == 0
	return result, nil
}
