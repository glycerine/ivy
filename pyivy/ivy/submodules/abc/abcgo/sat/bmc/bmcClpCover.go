package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Bmc_CollapseOneInt2 enumerates and expands an irredundant cover of output
// zero, returning nil on conflict or cube limits.
func Bmc_CollapseOneInt2(p *gia.Gia_Man_t, cubeLimit, conflictLimit, canonical, reverse, verbose, complement int32) *vec.Vec_Str_t {
	if p == nil || gia.Gia_ManCoNum(p) != 1 || cubeLimit < 0 || conflictLimit < 0 {
		return nil
	}
	graph := gia.Gia_ManToAigSimple(p)
	formula := cnf.Cnf_DeriveSimple(graph, int32(graph.OutputCount()))
	if formula == nil {
		return nil
	}
	makeSolver := func() *bsat.Solver { return cnf.Cnf_DataWriteIntoSolver(formula, 1, 0) }
	on, off, clean := makeSolver(), makeSolver(), makeSolver()
	if on == nil || off == nil || clean == nil {
		return nil
	}
	output := formula.VarNums[graph.Outputs()[0].NodeID()]
	if output < 0 {
		return nil
	}
	onLit, offLit := bsat.Positive(output), bsat.Negative(output)
	if !on.AddClause(onLit) {
		s := vec.Vec_StrAlloc(4)
		label := byte('0')
		if complement != 0 {
			label = '1'
		}
		for _, c := range []byte{' ', label, '\n', 0} {
			vec.Vec_StrPush(s, c)
		}
		return s
	}
	if !off.AddClause(offLit) {
		s := vec.Vec_StrAlloc(4)
		label := byte('1')
		if complement != 0 {
			label = '0'
		}
		for _, c := range []byte{' ', label, '\n', 0} {
			vec.Vec_StrPush(s, c)
		}
		return s
	}
	if !clean.AddClause(onLit) {
		return nil
	}
	inputs := graph.Inputs()
	vars := make([]int32, len(inputs))
	for i, e := range inputs {
		vars[i] = formula.VarNums[e.NodeID()]
	}
	order := make([]int, len(vars))
	for i := range order {
		order[i] = i
		if reverse != 0 {
			order[i] = len(vars) - 1 - i
		}
	}
	lits, nums, temp := vec.Vec_IntAlloc(len(vars)), vec.Vec_IntAlloc(len(vars)), vec.Vec_IntAlloc(len(vars))
	sop := vec.Vec_StrAlloc(100)
	for count := int32(0); ; count++ {
		status := bsat.Unknown
		if canonical != 0 {
			canon := vec.Vec_IntAlloc(len(vars))
			for _, i := range order {
				vec.Vec_IntPush(canon, bsat.Negative(vars[i]))
			}
			status = bsat.Status(Bmc_ComputeCanonical(on, canon, temp, conflictLimit))
		} else {
			status = collapseSolve(on, nil, conflictLimit)
		}
		if status == bsat.Unknown {
			return nil
		}
		if status == bsat.Unsatisfiable {
			break
		}
		if cubeLimit > 0 && count >= cubeLimit {
			return nil
		}
		vec.Vec_IntClear(lits)
		for _, i := range order {
			value, ok := on.Value(vars[i])
			if !ok {
				return nil
			}
			lit := bsat.Positive(vars[i])
			if !value {
				lit = bsat.Negative(vars[i])
			}
			vec.Vec_IntPush(lits, lit)
		}
		if Bmc_CollapseExpand(off, clean, lits, nums, temp, conflictLimit, canonical, -1) < 0 {
			return nil
		}
		cube := make([]byte, len(vars)+3)
		for i := range vars {
			cube[i] = '-'
		}
		cube[len(vars)] = ' '
		cube[len(vars)+1] = '1'
		if complement != 0 {
			cube[len(vars)+1] = '0'
		}
		cube[len(vars)+2] = '\n'
		block := make([]bsat.Literal, 0, vec.Vec_IntSize(nums))
		for _, position := range vec.Vec_IntArray(nums) {
			lit := vec.Vec_IntEntry(lits, int(position))
			idx := order[position]
			cube[idx] = '1'
			if lit&1 != 0 {
				cube[idx] = '0'
			}
			block = append(block, lit^1)
		}
		for _, c := range cube {
			vec.Vec_StrPush(sop, c)
		}
		if !on.AddClause(block...) {
			break
		}
		if canonical != 0 {
			clean.AddClause(block...)
		}
	}
	vec.Vec_StrPush(sop, 0)
	_ = verbose
	return sop
}
