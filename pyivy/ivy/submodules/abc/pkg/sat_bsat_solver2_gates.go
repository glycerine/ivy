package abc

// Port of the public gate-clause constructors in src/sat/bsat/satSolver2.h.

import "slices"

type SatSolver2ClauseSpec struct {
	Literals []SatLit
	PartA    bool
	ID       int
}

func satSolver2Literal(variable int, negative bool) (SatLit, error) {
	if variable < 0 {
		return 0, ErrDimension
	}
	return MakeSatLit(SatVar(variable), negative)
}

func satSolver2ClauseSpec(id int, partA bool, literals ...SatLit) (SatSolver2ClauseSpec, error) {
	if id == 0 || len(literals) == 0 {
		return SatSolver2ClauseSpec{}, ErrDimension
	}
	for _, literal := range literals {
		if !literal.Valid() {
			return SatSolver2ClauseSpec{}, ErrInvalidLiteral
		}
	}
	return SatSolver2ClauseSpec{Literals: slices.Clone(literals), PartA: partA, ID: id}, nil
}

func SatSolver2ConstantClauses(variable int, complement, partA bool, id int) ([]SatSolver2ClauseSpec, error) {
	literal, err := satSolver2Literal(variable, complement)
	if err != nil {
		return nil, err
	}
	clause, err := satSolver2ClauseSpec(id, partA, literal)
	if err != nil {
		return nil, err
	}
	return []SatSolver2ClauseSpec{clause}, nil
}

func SatSolver2BufferClauses(variableA, variableB int, complement, partA bool, id int) ([]SatSolver2ClauseSpec, error) {
	a, err := satSolver2Literal(variableA, false)
	if err != nil {
		return nil, err
	}
	b, err := satSolver2Literal(variableB, !complement)
	if err != nil {
		return nil, err
	}
	first, err := satSolver2ClauseSpec(id, partA, a, b)
	if err != nil {
		return nil, err
	}
	a, _ = satSolver2Literal(variableA, true)
	b, _ = satSolver2Literal(variableB, complement)
	second, _ := satSolver2ClauseSpec(id, partA, a, b)
	return []SatSolver2ClauseSpec{first, second}, nil
}

func SatSolver2AndClauses(output, input0, input1 int, complement0, complement1, partA bool, id int) ([]SatSolver2ClauseSpec, error) {
	variables := []int{output, input0, input1}
	for _, variable := range variables {
		if _, err := satSolver2Literal(variable, false); err != nil {
			return nil, err
		}
	}
	outNegative, _ := satSolver2Literal(output, true)
	firstInput, _ := satSolver2Literal(input0, complement0)
	secondInput, _ := satSolver2Literal(input1, complement1)
	first, err := satSolver2ClauseSpec(id, partA, outNegative, firstInput)
	if err != nil {
		return nil, err
	}
	second, _ := satSolver2ClauseSpec(id, partA, outNegative, secondInput)
	outPositive, _ := satSolver2Literal(output, false)
	third, _ := satSolver2ClauseSpec(id, partA, outPositive, firstInput.Not(), secondInput.Not())
	return []SatSolver2ClauseSpec{first, second, third}, nil
}

func SatSolver2XorClauses(variableA, variableB, variableC int, complement, partA bool, id int) ([]SatSolver2ClauseSpec, error) {
	for _, variable := range []int{variableA, variableB, variableC} {
		if _, err := satSolver2Literal(variable, false); err != nil {
			return nil, err
		}
	}
	patterns := [][3]bool{
		{!complement, true, true},
		{!complement, false, false},
		{complement, true, false},
		{complement, false, true},
	}
	result := make([]SatSolver2ClauseSpec, 0, len(patterns))
	for _, pattern := range patterns {
		a, _ := satSolver2Literal(variableA, pattern[0])
		b, _ := satSolver2Literal(variableB, pattern[1])
		c, _ := satSolver2Literal(variableC, pattern[2])
		clause, err := satSolver2ClauseSpec(id, partA, a, b, c)
		if err != nil {
			return nil, err
		}
		result = append(result, clause)
	}
	return result, nil
}

func SatSolver2ConstraintClauses(variable, auxiliary int, complement, partA bool, id int) ([]SatSolver2ClauseSpec, error) {
	literal, err := satSolver2Literal(variable, complement)
	if err != nil {
		return nil, err
	}
	positive, err := satSolver2Literal(auxiliary, false)
	if err != nil {
		return nil, err
	}
	first, err := satSolver2ClauseSpec(id, partA, literal, positive)
	if err != nil {
		return nil, err
	}
	second, _ := satSolver2ClauseSpec(id, partA, literal, positive.Not())
	return []SatSolver2ClauseSpec{first, second}, nil
}

func CloneSatSolver2ClauseSpecs(clauses []SatSolver2ClauseSpec) []SatSolver2ClauseSpec {
	result := make([]SatSolver2ClauseSpec, len(clauses))
	for i, clause := range clauses {
		result[i] = clause
		result[i].Literals = slices.Clone(clause.Literals)
	}
	return result
}

// AddSatSolver2ClauseSpecs atomically loads the logical clauses into the
// concrete solver. PartA and ID remain owned descriptive metadata for proof and
// interpolation consumers; SatSolver's CDCL core needs only the literals.
func AddSatSolver2ClauseSpecs(solver *SatSolver, clauses []SatSolver2ClauseSpec) (bool, error) {
	if solver == nil {
		return false, ErrDimension
	}
	if solver.solving || len(solver.clauses) == 0 {
		return false, ErrInvariant
	}
	for _, clause := range clauses {
		if clause.ID == 0 || len(clause.Literals) == 0 {
			return false, ErrDimension
		}
		for _, literal := range clause.Literals {
			if !literal.Valid() {
				return false, ErrInvalidLiteral
			}
		}
	}
	trial := solver.cloneState()
	trial.bookmark = solver.bookmark
	for _, clause := range clauses {
		ok, err := trial.AddClause(clause.Literals)
		if err != nil {
			return false, err
		}
		if !ok {
			*solver = trial
			return false, nil
		}
	}
	*solver = trial
	return true, nil
}
