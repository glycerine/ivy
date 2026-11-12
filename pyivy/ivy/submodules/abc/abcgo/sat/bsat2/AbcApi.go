package bsat2

/**CFile****************************************************************

  FileName    [AbcApi.cpp]

  PackageName [A C++ version of SAT solver MiniSAT 2.2 developed
  by Niklas Sorensson and Niklas Een. http://minisat.se.]

  Synopsis    [Interface to the SAT solver.]

  Author      [Niklas Sorensson and Niklas Een.]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: AbcApi.cpp,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

// SolveCNF runs the actual MiniSAT 2.2 solver and returns an owned model.
func SolveCNF(clauses [][]Lit) (LBool, []LBool) {
	solver := NewSolver()
	for _, clause := range clauses {
		solver.AddClause(clause...)
	}
	result := solver.Solve()
	return result, append([]LBool(nil), solver.Model...)
}

// Abc_CallMiniSat22 uses ABC's SAT=0, UNSAT=1, unknown=-1 convention.
func Abc_CallMiniSat22(clauses [][]Lit) int {
	result, _ := SolveCNF(clauses)
	switch result {
	case True:
		return 0
	case False:
		return 1
	default:
		return -1
	}
}
