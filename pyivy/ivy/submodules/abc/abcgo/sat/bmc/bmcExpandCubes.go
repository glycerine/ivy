package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Abc_ObjExpandCubesTry enlarges each on-set cube while avoiding the supplied
// off-set solver.
func Abc_ObjExpandCubesTry(sop *vec.Vec_Str_t, solver *bsat.Solver, vars *vec.Vec_Int_t) int32 {
	if sop == nil || solver == nil || vars == nil {
		return -1
	}
	n := vec.Vec_IntSize(vars)
	stride := n + 3
	data := vec.Vec_StrArray(sop)
	if len(data) == 0 || (len(data)-1)%stride != 0 || data[len(data)-1] != 0 {
		panic("bmc: invalid SOP cover")
	}
	count := (len(data) - 1) / stride
	lits, temp := vec.Vec_IntAlloc(n), vec.Vec_IntAlloc(n)
	for cube := 0; cube < count; cube++ {
		vec.Vec_IntFill(lits, n, -1)
		start := cube * stride
		for k := 0; k < n; k++ {
			switch data[start+k] {
			case '0':
				vec.Vec_IntWriteEntry(lits, k, bsat.Negative(vec.Vec_IntEntry(vars, k)))
			case '1':
				vec.Vec_IntWriteEntry(lits, k, bsat.Positive(vec.Vec_IntEntry(vars, k)))
			case '-':
			default:
				panic("bmc: invalid cube literal")
			}
		}
		if Bmc_CollapseExpandRound(solver, nil, lits, nil, temp, 0, 0, -1) < 0 {
			return -1
		}
		for k, lit := range vec.Vec_IntArray(lits) {
			data[start+k] = '-'
			if lit >= 0 {
				data[start+k] = '1' - byte(lit&1)
			}
		}
	}
	return int32(count)
}
func Abc_ObjExpandCubes(sop *vec.Vec_Str_t, p *gia.Gia_Man_t, nVars int32) int32 {
	if sop == nil || p == nil || nVars != gia.Gia_ManCiNum(p) || gia.Gia_ManCoNum(p) != 1 {
		return -1
	}
	g := gia.Gia_ManToAigSimple(p)
	formula := cnf.Cnf_DeriveSimple(g, int32(g.OutputCount()))
	if formula == nil {
		return -1
	}
	solver := cnf.Cnf_DataWriteIntoSolver(formula, 1, 0)
	if solver == nil {
		return -1
	}
	output := formula.VarNums[g.Outputs()[0].NodeID()]
	if output < 0 {
		return -1
	}
	for side := int32(0); side < 2; side++ {
		status := solver.Solve(bsat.Positive(output) ^ side)
		if status == bsat.Unknown {
			return -1
		}
		if status == bsat.Unsatisfiable {
			vec.Vec_StrClear(sop)
			label := byte('0')
			if side == 1 {
				label = '1'
			}
			for _, c := range []byte{' ', label, '\n', 0} {
				vec.Vec_StrPush(sop, c)
			}
			return 1
		}
	}
	if !solver.AddClause(bsat.Negative(output)) {
		return -1
	}
	vars := vec.Vec_IntAlloc(int(nVars))
	for _, e := range g.Inputs() {
		vec.Vec_IntPush(vars, formula.VarNums[e.NodeID()])
	}
	count := Abc_ObjExpandCubesTry(sop, solver, vars)
	if count < 0 {
		return -1
	}
	if count > 1 {
		Bmc_CollapseIrredundantFull(sop, count, nVars)
	}
	return 0
}
