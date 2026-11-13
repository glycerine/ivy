package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"time"
)

// Gia_ManFaultPrepare constructs the fault miter, solver, and prior-test constraints.
func Gia_ManFaultPrepare(p, reference *gia.Gia_Man_t, options *Bmc_ParFf_t, nFuncVars int32, faultMap, tests, scratch *vec.Vec_Int_t, miterOut **gia.Gia_Man_t, cnfOut **cnf.Cnf_Dat_t, solverOut **bsat.Solver, warmup int32) int32 {
	if p == nil || reference == nil || options == nil || nFuncVars <= 0 || vec.Vec_IntSize(tests)%int(nFuncVars) != 0 {
		return 0
	}
	var original, faulty *gia.Gia_Man_t
	switch options.Algo {
	case 0:
		faulty = Gia_ManFormulaUnfold(p, options.pFormStr, options.fFfOnly)
	case 1:
		if gia.Gia_ManRegNum(p) == 0 {
			return 0
		}
		original = Gia_ManFaultUnfold(reference, 0, options.fFfOnly)
		faulty = Gia_ManFaultUnfold(p, 1, options.fFfOnly)
	case 2:
		faulty = Gia_ManStuckAtUnfold(p, faultMap)
	case 3:
		faulty = Gia_ManFlipUnfold(p, faultMap)
	case 4:
		faulty = Gia_ManFOFUnfold(p, faultMap)
	default:
		return 0
	}
	if original == nil {
		original = Gia_ManDeriveDup(reference, gia.Gia_ManCiNum(faulty)-gia.Gia_ManCiNum(reference))
	}
	m := gia.Gia_ManMiter(original, faulty, 0, 0, 0, 0, 0)
	if m == nil {
		return 0
	}
	enc, err := gia.Mf_ManGenerateCnfResult(m, 6, true, false, false)
	if err != nil {
		return 0
	}
	data := &cnf.Cnf_Dat_t{Variables: enc.Variables, Clauses: enc.Clauses, VarNums: enc.VarNums}
	solver := bsat.NewSolver()
	for i := 0; i < enc.Variables; i++ {
		solver.NewVariable()
	}
	if options.nTimeOut > 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(options.nTimeOut) * time.Second))
	}
	for _, clause := range enc.Clauses {
		if !solver.AddClause(clause...) {
			return 0
		}
	}
	var out []bsat.Literal
	for i := int32(0); i < gia.Gia_ManCoNum(m); i++ {
		id := gia.Gia_ObjId(m, gia.Gia_ManCo(m, i))
		out = append(out, bsat.Positive(data.VarNums[id]))
	}
	if !solver.AddClause(out...) {
		return 0
	}
	*miterOut, *cnfOut, *solverOut = m, data, solver
	vec.Vec_IntClear(scratch)
	for i := nFuncVars; i < gia.Gia_ManPiNum(m); i++ {
		id := gia.Gia_ObjId(m, gia.Gia_ManCi(m, i))
		vec.Vec_IntPush(scratch, data.VarNums[id])
	}
	if options.fBasic != 0 {
		Cnf_AddCardinConstr(solver, scratch)
	} else if options.nCardConstr != 0 {
		strict := int32(1)
		if options.fNonStrict != 0 {
			strict = 0
		}
		Cnf_AddCardinConstrGeneral(solver, scratch, options.nCardConstr, strict)
	}
	if vec.Vec_IntSize(tests) == 0 && options.fStartPats != 0 {
		for value := int32(0); value < 2; value++ {
			for i := int32(0); i < nFuncVars; i++ {
				vec.Vec_IntPush(tests, value)
			}
		}
	}
	patterns := vec.Vec_IntArray(tests)
	for start := 0; start < len(patterns); start += int(nFuncVars) {
		if warmup != 0 && solver.Solve() != bsat.Satisfiable {
			vec.Vec_IntShrink(tests, start)
			return 0
		}
		vec.Vec_IntClear(scratch)
		for _, v := range patterns[start : start+int(nFuncVars)] {
			vec.Vec_IntPush(scratch, v)
		}
		if Gia_ManFaultAddOne(m, data, solver, scratch, nFuncVars, 0, m) == 0 {
			vec.Vec_IntShrink(tests, start)
			return 0
		}
	}
	fmt.Printf("Using miter with: AIG nodes = %d. CNF variables = %d. CNF clauses = %d.\n", gia.Gia_ManAndNum(m), data.Variables, len(data.Clauses))
	return 1
}
