package abc

// Bounded DIMACS loading from src/sat/msat/msatRead.c.

func ParseMsatDIMACS(data []byte, options CNFDIMACSReadOptions, config MsatSolverConfig) (*MsatSolver, bool, error) {
	formula, err := ParseCNFDIMACS(data, options)
	if err != nil {
		return nil, false, err
	}
	return newMsatSolverFromFormula(formula, config)
}

func newMsatSolverFromFormula(formula *CNFFormula, config MsatSolverConfig) (*MsatSolver, bool, error) {
	solver, err := NewMsatSolver(formula.variables, config)
	if err != nil {
		return nil, false, err
	}
	if err = solver.Clean(formula.variables); err != nil {
		solver.Close()
		return nil, false, err
	}
	consistent := true
	for _, clause := range formula.clauses {
		if consistent {
			consistent, err = solver.AddClause(clause)
			if err != nil {
				solver.Close()
				return nil, false, err
			}
		}
	}
	if consistent {
		consistent, err = solver.Simplify()
		if err != nil {
			solver.Close()
			return nil, false, err
		}
	}
	return solver, consistent, nil
}

func ReadMsatDIMACSFile(path string, options CNFDIMACSReadOptions, config MsatSolverConfig) (*MsatSolver, bool, error) {
	formula, err := ReadCNFDIMACSFile(path, options)
	if err != nil {
		return nil, false, err
	}
	return newMsatSolverFromFormula(formula, config)
}
