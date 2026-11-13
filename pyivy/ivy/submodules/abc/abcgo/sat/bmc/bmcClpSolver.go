package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func constantCover(value byte) *vec.Vec_Str_t {
	v := vec.Vec_StrAlloc(4)
	for _, b := range []byte{' ', value, '\n', 0} {
		vec.Vec_StrPush(v, b)
	}
	return v
}

// collapseSolverPair uses gated blocking clauses so the opposite polarity
// remains available for proving that an expanded cube is valid.
func collapseSolverPair(solvers [2]*bsat.Solver, nVars, cubeLimit, conflictLimit, canonical, reverse int32, single bool) *vec.Vec_Str_t {
	if solvers[0] == nil || solvers[1] == nil || nVars < 0 || cubeLimit < 0 || conflictLimit < 0 || solvers[0].NumVars() < 3+nVars || solvers[1].NumVars() < 3+nVars {
		return nil
	}
	for side := int32(0); side < 2; side++ {
		out := bsat.Positive(2) ^ side
		status := collapseSolve(solvers[side], []int32{out}, conflictLimit)
		if status == bsat.Unknown {
			return nil
		}
		if status == bsat.Unsatisfiable {
			if side == 0 {
				return constantCover('0')
			}
			return constantCover('1')
		}
		if !single && !solvers[side].AddClause(out) {
			return nil
		}
	}
	covers := [2]*vec.Vec_Str_t{vec.Vec_StrAlloc(100), vec.Vec_StrAlloc(100)}
	lits, nums, temp := vec.Vec_IntAlloc(int(nVars)), vec.Vec_IntAlloc(int(nVars)), vec.Vec_IntAlloc(int(nVars))
	for round := int32(0); cubeLimit == 0 || round < cubeLimit; round++ {
		for side := int32(0); side < 2; side++ {
			s := solvers[side]
			assumptions := []int32{bsat.Negative(side)}
			if single {
				assumptions = append(assumptions, bsat.Positive(2)^side)
			}
			status := collapseSolve(s, assumptions, conflictLimit)
			if status == bsat.Unknown {
				return nil
			}
			if status == bsat.Unsatisfiable {
				vec.Vec_StrPush(covers[side], 0)
				return covers[side]
			}
			if canonical != 0 {
				prefix := append([]int32(nil), assumptions...)
				for i := int32(0); i < nVars; i++ {
					idx := i
					if reverse != 0 {
						idx = nVars - 1 - i
					}
					trial := append(append([]int32(nil), prefix...), bsat.Negative(3+idx))
					status = collapseSolve(s, trial, conflictLimit)
					if status == bsat.Unknown {
						return nil
					}
					lit := bsat.Negative(3 + idx)
					if status == bsat.Unsatisfiable {
						lit = bsat.Positive(3 + idx)
					}
					prefix = append(prefix, lit)
				}
				if collapseSolve(s, prefix, conflictLimit) != bsat.Satisfiable {
					return nil
				}
			}
			vec.Vec_IntClear(lits)
			for i := int32(0); i < nVars; i++ {
				idx := i
				if reverse != 0 {
					idx = nVars - 1 - i
				}
				v := int32(3) + idx
				value, ok := s.Value(v)
				if !ok {
					return nil
				}
				lit := bsat.Positive(v)
				if !value {
					lit = bsat.Negative(v)
				}
				vec.Vec_IntPush(lits, lit)
			}
			condition := int32(-1)
			if single {
				condition = bsat.Positive(2) ^ (1 - side)
			}
			if Bmc_CollapseExpand(solvers[1-side], nil, lits, nums, temp, conflictLimit, 0, condition) < 0 {
				return nil
			}
			cube := make([]byte, int(nVars)+3)
			for i := int32(0); i < nVars; i++ {
				cube[i] = '-'
			}
			cube[nVars] = ' '
			cube[nVars+1] = '1' - byte(side)
			cube[nVars+2] = '\n'
			block := []bsat.Literal{bsat.Positive(side)}
			for _, position := range vec.Vec_IntArray(nums) {
				lit := vec.Vec_IntEntry(lits, int(position))
				idx := position
				if reverse != 0 {
					idx = nVars - 1 - idx
				}
				cube[idx] = '1'
				if lit&1 != 0 {
					cube[idx] = '0'
				}
				block = append(block, lit^1)
			}
			for _, b := range cube {
				vec.Vec_StrPush(covers[side], b)
			}
			if !s.AddClause(block...) {
				vec.Vec_StrPush(covers[side], 0)
				return covers[side]
			}
		}
	}
	return nil
}
func Bmc_CollapseOne_int(s *bsat.Solver, nVars, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	_ = verbose
	return collapseSolverPair([2]*bsat.Solver{s, s}, nVars, cubeLimit, conflictLimit, canonical, reverse, true)
}
func Bmc_CollapseOne_int2(on, off *bsat.Solver, nVars, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	_ = verbose
	return collapseSolverPair([2]*bsat.Solver{on, off}, nVars, cubeLimit, conflictLimit, canonical, reverse, false)
}
