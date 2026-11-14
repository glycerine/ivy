// Package aiginterp owns AIG interpolation drivers above the graph, CNF and
// SAT packages, avoiding an import cycle between those implementations.
package aiginterp

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func validate(on, off *aig.Graph) {
	if on == nil || off == nil || on.InputCount() != off.InputCount() {
		panic("interpolation requires equal CI counts")
	}
	if err := on.Check(); err != nil {
		panic(err)
	}
	if err := off.Check(); err != nil {
		panic(err)
	}
}

func addCNF(s *bsat.Solver, data *cnf.Cnf_Dat_t) bool {
	for _, clause := range data.Clauses {
		literals := make([]bsat.Literal, len(clause))
		for i, l := range clause {
			literals[i] = bsat.Literal(l)
		}
		if !s.AddClause(literals...) {
			return false
		}
	}
	return true
}

// Aig_ManInter follows src/aig/aig/aigInter.c: simple CNFs, A/B partitioned
// clause storage, shared CI equalities, SAT refutation and Inta interpolation.
func Aig_ManInter(on, off *aig.Graph, relation, verbose int32) *aig.Graph {
	validate(on, off)
	if relation != 0 && on.OutputCount() == 0 {
		panic("relation interpolation needs an A output")
	}
	a, b := cnf.Cnf_DeriveSimple(on, 0), cnf.Cnf_DeriveSimple(off, 0)
	if a == nil || b == nil {
		panic("interpolation CNF derivation failed")
	}
	defer cnf.Cnf_DataFree(a)
	defer cnf.Cnf_DataFree(b)
	cnf.Cnf_DataLift(b, int32(a.Variables))
	solver := bsat.NewSolver()
	defer solver.Close()
	solver.EnableClauseStore()
	solver.SetNumVars(int32(a.Variables + b.Variables))
	if !addCNF(solver, a) {
		return nil
	}
	solver.MarkClauseStorePartitionA()
	globals := make([]int32, 0, on.InputCount()+1)
	if relation != 0 {
		globals = append(globals, solver.ChangeLastStoredClause())
	}
	if !addCNF(solver, b) {
		return nil
	}
	inputsA, inputsB := on.Inputs(), off.Inputs()
	for i, input := range inputsA {
		va, vb := a.VarNums[input.NodeID()], b.VarNums[inputsB[i].NodeID()]
		if va < 0 || vb < 0 {
			panic("interpolation CI has no CNF variable")
		}
		globals = append(globals, va)
		if !solver.AddClause(bsat.Positive(va), bsat.Negative(vb)) || !solver.AddClause(bsat.Negative(va), bsat.Positive(vb)) {
			panic("interpolation PI equality is inconsistent")
		}
	}
	solver.MarkClauseStoreRoots()
	if solver.Solve() != bsat.Unsatisfiable {
		fmt.Println("The SAT problem is not unsat.")
		return nil
	}
	store := solver.ReleaseClauseStore()
	defer bsat.Sto_ManFree(store)
	manager := bsat.Inta_ManAlloc()
	defer bsat.Inta_ManFree(manager)
	result := bsat.Inta_ManInterpolate(manager, store, 0, globals, verbose)
	if result == nil {
		panic(fmt.Sprintf("interpolation failed: %v", bsat.Inta_ManError(manager)))
	}
	return result
}
