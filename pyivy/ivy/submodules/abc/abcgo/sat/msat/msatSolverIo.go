// Go port of solver reports and DIMACS output in msatSolverIo.c.
package msat

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Msat_SolverPrintStats(p *Msat_Solver_t) {
	fmt.Printf("C solver (%d vars; %d clauses; %d learned):\n", p.nVars, len(p.vClauses.data), len(p.vLearned.data))
	fmt.Printf("starts        : %d\nconflicts     : %d\ndecisions     : %d\npropagations  : %d\ninspects      : %d\n", int32(p.Stats.nStarts), int32(p.Stats.nConflicts), int32(p.Stats.nDecisions), int32(p.Stats.nPropagations), int32(p.Stats.nInspects))
}
func Msat_SolverPrintAssignment(p *Msat_Solver_t) {
	fmt.Print("Current assignments are: \n")
	for i := 0; i < p.nVars; i++ {
		fmt.Printf("%d", i%10)
	}
	fmt.Println()
	for i := 0; i < p.nVars; i++ {
		lit := p.pAssigns[i]
		if lit == MSAT_VAR_UNASSIGNED {
			fmt.Print(".")
			continue
		}
		assert(i == MSAT_LIT2VAR(lit), "assignment variable matches index")
		if MSAT_LITSIGN(lit) != 0 {
			fmt.Print("0")
		} else {
			fmt.Print("1")
		}
	}
	fmt.Println()
}
func Msat_SolverPrintClauses(p *Msat_Solver_t) {
	fmt.Println("Original clauses: ")
	for i, c := range p.vClauses.data {
		fmt.Printf("%3d: ", i)
		Msat_ClausePrint(c)
	}
	fmt.Println("Learned clauses: ")
	for i, c := range p.vLearned.data {
		fmt.Printf("%3d: ", i)
		Msat_ClausePrint(c)
	}
	fmt.Println("Variable activity: ")
	for i := 0; i < p.nVars; i++ {
		fmt.Printf("%3d : %.4f\n", i, p.pdActivity[i])
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
