package cudd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

type DdTlcInfo struct {
	clauses []twoLiteralClause
}

type twoLiteralClause struct {
	var1, var2     uint32
	phase1, phase2 int32
}

// Cudd_FindEssential returns a cube containing the literals implied by f.
func Cudd_FindEssential(dd *DdManager, f *DdNode) *DdNode {
	if dd == nil || f == nil {
		return nil
	}
	result := Cudd_ReadOne(dd)
	for index := range dd.Vars {
		variable := core.Cudd_bddIthVar(dd, index)
		if core.Cudd_bddAnd(dd, f, Cudd_Not(variable)) == Cudd_ReadLogicZero(dd) {
			result = core.Cudd_bddAnd(dd, result, variable)
		} else if core.Cudd_bddAnd(dd, f, variable) == Cudd_ReadLogicZero(dd) {
			result = core.Cudd_bddAnd(dd, result, Cudd_Not(variable))
		}
	}
	return result
}

// Cudd_bddIsVarEssential reports whether f implies the selected phase of x.
func Cudd_bddIsVarEssential(dd *DdManager, f *DdNode, id, phase int32) int32 {
	if dd == nil || f == nil || id < 0 || int(id) >= len(dd.Vars) || (phase != 0 && phase != 1) {
		return 0
	}
	x := core.Cudd_bddIthVar(dd, int(id))
	if phase == 0 {
		x = Cudd_Not(x)
	}
	if core.Cudd_bddAnd(dd, f, Cudd_Not(x)) == Cudd_ReadLogicZero(dd) {
		return 1
	}
	return 0
}

// Cudd_FindTwoLiteralClauses enumerates all prime implicates of f with at
// most two literals. Phase 1 denotes a positive literal; phase 0 is negated.
func Cudd_FindTwoLiteralClauses(dd *DdManager, f *DdNode) *DdTlcInfo {
	if dd == nil || f == nil {
		return nil
	}
	info := &DdTlcInfo{}
	variables := bddSupportIndices(f)
	sort.Ints(variables)
	unit := make(map[[2]uint32]bool)
	for _, i := range variables {
		for _, phase := range []int32{0, 1} {
			if Cudd_bddIsVarEssential(dd, f, int32(i), phase) != 0 {
				unit[[2]uint32{uint32(i), uint32(phase)}] = true
				info.clauses = append(info.clauses, twoLiteralClause{var1: uint32(i), var2: CUDD_MAXINDEX, phase1: phase})
			}
		}
	}
	for iPos, i := range variables {
		for _, j := range variables[iPos+1:] {
			for _, phaseI := range []int32{0, 1} {
				for _, phaseJ := range []int32{0, 1} {
					if unit[[2]uint32{uint32(i), uint32(phaseI)}] || unit[[2]uint32{uint32(j), uint32(phaseJ)}] {
						continue
					}
					li, lj := core.Cudd_bddIthVar(dd, i), core.Cudd_bddIthVar(dd, j)
					if phaseI == 0 {
						li = Cudd_Not(li)
					}
					if phaseJ == 0 {
						lj = Cudd_Not(lj)
					}
					counterexample := core.Cudd_bddAnd(dd, f, Cudd_Not(li))
					counterexample = core.Cudd_bddAnd(dd, counterexample, Cudd_Not(lj))
					if counterexample == Cudd_ReadLogicZero(dd) {
						info.clauses = append(info.clauses, twoLiteralClause{uint32(i), uint32(j), phaseI, phaseJ})
					}
				}
			}
		}
	}
	return info
}

func Cudd_ReadIthClause(info *DdTlcInfo, i int32, var1, var2 *DdHalfWord, phase1, phase2 *int32) int32 {
	if info == nil || i < 0 || int(i) >= len(info.clauses) || var1 == nil || var2 == nil || phase1 == nil || phase2 == nil {
		return 0
	}
	clause := info.clauses[i]
	*var1, *var2 = DdHalfWord(clause.var1), DdHalfWord(clause.var2)
	*phase1, *phase2 = clause.phase1, clause.phase2
	return 1
}

func Cudd_PrintTwoLiteralClauses(dd *DdManager, f *DdNode, names []string, writer io.Writer) int32 {
	if writer == nil {
		writer = os.Stdout
	}
	info := Cudd_FindTwoLiteralClauses(dd, f)
	if info == nil {
		return 0
	}
	for _, clause := range info.clauses {
		first := literalLabel(names, clause.var1, clause.phase1)
		if clause.var2 == CUDD_MAXINDEX {
			if _, err := fmt.Fprintln(writer, first); err != nil {
				return 0
			}
			continue
		}
		second := literalLabel(names, clause.var2, clause.phase2)
		if _, err := fmt.Fprintf(writer, "%s | %s\n", first, second); err != nil {
			return 0
		}
	}
	return 1
}

func literalLabel(names []string, index uint32, phase int32) string {
	name := fmt.Sprint(index)
	if int(index) < len(names) && names[index] != "" {
		name = names[index]
	}
	if phase == 0 {
		return "~" + name
	}
	return name
}

func Cudd_tlcInfoFree(info *DdTlcInfo) {
	if info != nil {
		info.clauses = nil
	}
}
