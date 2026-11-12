package bsat

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

/**CFile****************************************************************

  FileName    [satUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [C-language MiniSat solver.]

  Synopsis    [Additional SAT solver procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satUtil.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

func writeDimacsClause(writer io.Writer, literals []lit, increment int32) error {
	for _, l := range literals {
		sign := ""
		if lit_sign(l) != 0 {
			sign = "-"
		}
		if _, err := fmt.Fprintf(writer, "%s%d ", sign, lit_var(l)+increment); err != nil {
			return err
		}
	}
	if increment != 0 {
		if _, err := fmt.Fprint(writer, "0"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(writer)
	return err
}
func Sat_SolverClauseWriteDimacs(file *os.File, c *clause, increment int32) error {
	return writeDimacsClause(file, c.lits[:c.size], solver2Bool(increment != 0))
}

// WriteSolverDIMACS includes compact binary watches as well as memory clauses.
// The header is computed from the actual emitted clauses, including assumptions.
func WriteSolverDIMACS(writer io.Writer, s *sat_solver, assumptions []lit, increment bool) error {
	clauses := [][]lit{}
	binary := map[[2]lit]bool{}
	for learned := int32(0); learned < 2; learned++ {
		for _, h := range s.Mem.handles(learned) {
			if h == s.hBinary && s.binary != nil {
				continue
			}
			c := Sat_MemClauseHand(&s.Mem, h)
			literals := append([]lit(nil), c.lits[:c.size]...)
			clauses = append(clauses, literals)
			if len(literals) == 2 {
				a, b := literals[0], literals[1]
				if a > b {
					a, b = b, a
				}
				binary[[2]lit{a, b}] = true
			}
		}
	}
	for index, w := range s.wlists {
		for _, h := range veci_begin(&w) {
			if clause_is_lit(h) == 0 {
				continue
			}
			a, b := lit_neg(lit(index)), clause_read_lit(h)
			if a > b {
				a, b = b, a
			}
			key := [2]lit{a, b}
			if !binary[key] {
				clauses = append(clauses, []lit{a, b})
				binary[key] = true
			}
		}
	}
	for v := int32(0); v < s.size; v++ {
		if s.levels[v] == 0 && s.assigns[v] != satUnassigned {
			clauses = append(clauses, []lit{toLitCond(v, s.assigns[v])})
		}
	}
	for _, l := range assumptions {
		clauses = append(clauses, []lit{l})
	}
	if _, err := fmt.Fprintf(writer, "p cnf %d %d\n", s.size, len(clauses)); err != nil {
		return err
	}
	for _, c := range clauses {
		if err := writeDimacsClause(writer, c, solver2Bool(increment)); err != nil {
			return err
		}
	}
	return nil
}
func writeSolverFile(s *sat_solver, name string, assumptions []lit, increment int32) error {
	writer := os.Stdout
	var err error
	if name != "" {
		writer, err = os.Create(name)
		if err != nil {
			return err
		}
	}
	err = WriteSolverDIMACS(writer, s, assumptions, increment != 0)
	if name != "" {
		if closeErr := writer.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}
func Sat_SolverWriteDimacs(s *sat_solver, name string, assumptions []lit, increment int32) error {
	return writeSolverFile(s, name, assumptions, increment)
}
func Sat_Solver2WriteDimacs(s *sat_solver2, name string, assumptions []lit, increment int32) error {
	return writeSolverFile(&s.sat_solver_t, name, assumptions, increment)
}
func Sat_Wrd2Dbl(w word) float64 { return float64(w) }
func Sat_SolverPrintStats(file *os.File, s *sat_solver) {
	fmt.Fprintf(file, "starts        : %16d\nconflicts     : %16d\ndecisions     : %16d\npropagations  : %16d\n", s.stats.starts, s.stats.conflicts, s.stats.decisions, s.stats.propagations)
}
func Sat_Solver2PrintStats(file *os.File, s *sat_solver2) {
	Sat_SolverPrintStats(file, &s.sat_solver_t)
}
func Sat_Solver2GetVarMem(s *sat_solver2) int32 {
	return int32(2*reflect.TypeOf(veci{}).Size() + 8 + 11*4 + 1)
}
func Sat_SolverGetModel(s *sat_solver, variables []int32, n int32) []int32 {
	model := make([]int32, n+1)
	for i, v := range variables[:n] {
		model[i] = sat_solver_var_value(s, v)
	}
	return model
}
func Sat_Solver2GetModel(s *sat_solver2, variables []int32, n int32) []int32 {
	return Sat_SolverGetModel(&s.sat_solver_t, variables, n)
}

// The original C entry point unconditionally asserts: its obsolete implementation
// refers to a clause representation that is no longer used by ABC.
func Sat_SolverDoubleClauses(s *sat_solver, variable int32) {
	panic("ABC Sat_SolverDoubleClauses is disabled in the original source")
}
