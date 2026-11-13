package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Bmc_EcoSolve checks whether the divisor variables distinguish the root's
// onset from its offset, blocking each proven distinguishing cube.
func Bmc_EcoSolve(s *bsat.Solver, root int32, vars *vec.Vec_Int_t) int32 {
	if s == nil || vars == nil || root < 0 || root >= s.NumVars() {
		return -1
	}
	gate := s.NewVariable()
	s.SetLimits(1000000, 0)
	for iteration := 0; ; iteration++ {
		status := s.Solve(bsat.Positive(root), bsat.Positive(gate))
		if status == bsat.Unknown {
			return -1
		}
		if status == bsat.Unsatisfiable {
			return 1
		}
		assumptions := []bsat.Literal{bsat.Negative(root)}
		for _, v := range vec.Vec_IntArray(vars) {
			value, ok := s.Value(v)
			if !ok {
				return -1
			}
			lit := bsat.Positive(v)
			if !value {
				lit = bsat.Negative(v)
			}
			assumptions = append(assumptions, lit)
		}
		status = s.Solve(assumptions...)
		if status == bsat.Unknown {
			return -1
		}
		if status == bsat.Satisfiable {
			return 0
		}
		block := []bsat.Literal{bsat.Negative(gate)}
		for _, lit := range s.UnsatCore() {
			if lit != bsat.Positive(root) {
				block = append(block, lit)
			}
		}
		if !s.AddClause(block...) {
			return 1
		}
		_ = iteration
	}
}

// Bmc_EcoPatch derives the patch feasibility query from the ECO miter.
func Bmc_EcoPatch(p *gia.Gia_Man_t, nIns, nOuts int32) int32 {
	if p == nil || nIns < 0 || nOuts < 0 || gia.Gia_ManPoNum(p) != nOuts+1+nIns || gia.Gia_ManPiNum(p) == 0 {
		return -1
	}
	g := gia.Gia_ManToAigSimple(p)
	formula := cnf.Cnf_DeriveSimple(g, int32(g.OutputCount()))
	if formula == nil {
		return -1
	}
	s := cnf.Cnf_DataWriteIntoSolver(formula, 1, 0)
	if s == nil {
		return -1
	}
	outputs := g.Outputs()
	for i := int32(0); i < nOuts; i++ {
		v := formula.VarNums[outputs[i].NodeID()]
		if v < 0 || !s.AddClause(bsat.Negative(v)) {
			return -1
		}
	}
	v := formula.VarNums[outputs[nOuts].NodeID()]
	if v < 0 || !s.AddClause(bsat.Positive(v)) {
		return -1
	}
	vars := vec.Vec_IntAlloc(int(nIns))
	for i := int32(0); i < nIns; i++ {
		v := formula.VarNums[outputs[nOuts+1+i].NodeID()]
		if v < 0 {
			return -1
		}
		vec.Vec_IntPush(vars, v)
	}
	inputs := g.Inputs()
	root := formula.VarNums[inputs[len(inputs)-1].NodeID()]
	return Bmc_EcoSolve(s, root, vars)
}
func Bmc_EcoMiterTest() {
	gold, err := gia.Gia_AigerRead("eco_gold.aig", 0, 0, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	old, err := gia.Gia_AigerRead("eco_old.aig", 0, 0, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fans := vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(old)))
	for i := int32(0); i < gia.Gia_ManCiNum(old); i++ {
		vec.Vec_IntPush(fans, gia.Gia_ObjId(old, gia.Gia_ManCi(old, i)))
	}
	miter := Bmc_EcoMiter(gold, old, fans)
	if miter == nil {
		fmt.Println("Cannot construct ECO miter.")
		return
	}
	if err := gia.Gia_AigerWrite(miter, "eco_miter.aig", 0, 0, 0); err != nil {
		fmt.Println(err)
		return
	}
	switch Bmc_EcoPatch(miter, gia.Gia_ManCiNum(gold), gia.Gia_ManCoNum(gold)) {
	case 1:
		fmt.Println("Patch is computed.")
	case 0:
		fmt.Println("Cannot be patched.")
	default:
		fmt.Println("Resource limit exceeded.")
	}
}
