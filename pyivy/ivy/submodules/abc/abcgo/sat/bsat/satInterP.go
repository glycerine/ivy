package bsat

import (
	"fmt"
	"io"
	"os"
)

/**CFile****************************************************************

  FileName    [satInter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Interpolation package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satInter.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

type Intp_Man_t = Intp_Man_t_
type Intp_Man_t_ struct {
	pCnf                               *Sto_Man_t
	fVerbose, fProofVerif, fProofWrite int32
	Counter                            int32
	dependencies                       [][]int32
	lastErr                            error
}

func Intp_ManAlloc() *Intp_Man_t { return &Intp_Man_t{fProofVerif: 1} }
func Intp_ManFree(p *Intp_Man_t) {
	if p != nil {
		*p = Intp_Man_t{}
	}
}
func Intp_ManError(p *Intp_Man_t) error { return p.lastErr }
func Intp_ManUnsatCore(p *Intp_Man_t, cnf *Sto_Man_t, learned, verbose int32) []int32 {
	p.pCnf, p.fVerbose = cnf, verbose
	var writer io.Writer
	var file *os.File
	if p.fProofWrite != 0 {
		file, p.lastErr = os.Create("proof.cnf_")
		if p.lastErr != nil {
			return nil
		}
		writer = file
	}
	trace, err := reconstructResolution(cnf, nil, writer, 0)
	if file != nil {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}
	p.lastErr = err
	if err != nil {
		return nil
	}
	p.Counter, p.dependencies = trace.counter, trace.dependencies
	var core []int32
	Intp_ManUnsatCore_rec(p.dependencies, cnf.pEmpty.Id, &core, cnf.nRoots, make([]bool, cnf.nClauses), learned)
	if verbose != 0 {
		fmt.Printf("Root clauses = %d. Learned clauses = %d. UNSAT core size = %d.\n", cnf.nRoots, cnf.nClauses-cnf.nRoots, len(core))
	}
	return core
}
func Intp_ManUnsatCore_rec(antecedents [][]int32, clause int32, core *[]int32, roots int32, visited []bool, learned int32) {
	if visited[clause] {
		return
	}
	visited[clause] = true
	if clause < roots {
		if learned == 0 {
			*core = append(*core, clause)
		}
		return
	}
	for _, parent := range antecedents[clause] {
		Intp_ManUnsatCore_rec(antecedents, parent, core, roots, visited, learned)
	}
	if learned != 0 {
		*core = append(*core, clause)
	}
}
func Intp_ManUnsatCoreVerify(cnf *Sto_Man_t, core []int32) {
	clauses := make([]*Sto_Cls_t, cnf.nClauses)
	for c := cnf.pHead; c != nil; c = c.pNext {
		clauses[c.Id] = c
	}
	solver := NewSolver()
	for _, id := range core {
		if !solver.AddClause(clauses[id].pLits...) {
			return
		}
	}
	if solver.Solve() != Unsatisfiable {
		panic("extracted clause core is satisfiable")
	}
}
func Intp_ManUnsatCorePrintForBmc(file *os.File, cnf *Sto_Man_t, core, variableMap []int32) {
	clauses := make([]*Sto_Cls_t, cnf.nClauses)
	for c := cnf.pHead; c != nil; c = c.pNext {
		clauses[c.Id] = c
	}
	fmt.Fprintf(file, "UNSAT contains %d learned clauses:\n", len(core))
	for index, id := range core {
		c := clauses[id]
		fmt.Fprintf(file, "%6d : %6d : ", index, id-cnf.nRoots)
		for _, l := range c.pLits {
			prefix := ""
			if lit_sign(l) != 0 {
				prefix = "!"
			}
			v := 2 * lit_var(l)
			fmt.Fprintf(file, "%s%d(%d) ", prefix, variableMap[v], variableMap[v+1])
		}
		if c.nLits == 0 {
			fmt.Fprint(file, "Empty")
		}
		fmt.Fprintln(file)
	}
}
