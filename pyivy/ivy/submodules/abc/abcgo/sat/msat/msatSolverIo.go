// Go port of solver reports and DIMACS output in msatSolverIo.c.
package msat

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Msat_SolverPrintStats(p *Msat_Solver_t) {
	fmt.Printf("MSAT solver (%d vars; %d clauses; %d learned):\nstarts: %d\nconflicts: %d\ndecisions: %d\npropagations: %d\ninspects: %d\n", p.nVars, len(p.vClauses.data), len(p.vLearned.data), p.Stats.nStarts, p.Stats.nConflicts, p.Stats.nDecisions, p.Stats.nPropagations, p.Stats.nInspects)
}
func Msat_SolverPrintAssignment(p *Msat_Solver_t) {
	for i, lit := range p.pAssigns[:p.nVars] {
		fmt.Printf("%d: %d\n", i, lit)
	}
}
func Msat_SolverPrintClauses(p *Msat_Solver_t) {
	for _, c := range p.vClauses.data {
		Msat_ClausePrint(c)
	}
}
func (p *Msat_Solver_t) WriteDIMACS(w io.Writer) error {
	units := []int{}
	for _, lit := range p.vTrail.data {
		if p.pLevel[lit>>1] == 0 {
			units = append(units, lit)
		}
	}
	if _, err := fmt.Fprintf(w, "p cnf %d %d\n", p.nVars, len(p.vClauses.data)+len(units)); err != nil {
		return err
	}
	for _, c := range p.vClauses.data {
		for _, lit := range c.lits {
			v := (lit >> 1) + 1
			if lit&1 != 0 {
				v = -v
			}
			if _, err := fmt.Fprintf(w, "%d ", v); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "0"); err != nil {
			return err
		}
	}
	for _, lit := range units {
		v := (lit >> 1) + 1
		if lit&1 != 0 {
			v = -v
		}
		if _, err := fmt.Fprintf(w, "%d 0\n", v); err != nil {
			return err
		}
	}
	return nil
}
func Msat_SolverWriteDimacs(p *Msat_Solver_t, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	count := len(p.vClauses.data) + len(p.vLearned.data)
	for i := 0; i < p.nVars; i++ {
		if p.pLevel[i] == 0 {
			count++
		}
	}
	fmt.Fprintf(f, "c Produced by Msat_SolverWriteDimacs() on %s\n", time.Now().Format(time.ANSIC))
	fmt.Fprintf(f, "p cnf %d %d\n", p.nVars, count)
	for _, c := range p.vClauses.data {
		Msat_ClauseWriteDimacs(f, c, 1)
	}
	for _, c := range p.vLearned.data {
		Msat_ClauseWriteDimacs(f, c, 1)
	}
	for i := 0; i < p.nVars; i++ {
		if p.pLevel[i] == 0 {
			sign := ""
			if p.pAssigns[i]&1 != 0 {
				sign = "-"
			}
			fmt.Fprintf(f, "%s%d 0\n", sign, i+1)
		}
	}
	fmt.Fprintln(f)
}
