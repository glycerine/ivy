package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Bmc_CollapseOne_int3 enumerates covers with independent enumeration and
// clean equivalence solvers for each output polarity.
func Bmc_CollapseOne_int3(on, off, cleanOn, cleanOff *bsat.Solver, nVars, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	solvers := [2]*bsat.Solver{on, off}
	clean := [2]*bsat.Solver{cleanOn, cleanOff}
	if on == nil || off == nil || cleanOn == nil || cleanOff == nil || nVars < 0 || cubeLimit < 0 || conflictLimit < 0 || on.NumVars()-nVars < 2 {
		return nil
	}
	base := on.NumVars() - nVars
	for side := int32(0); side < 2; side++ {
		if solvers[side].NumVars() != on.NumVars() || clean[side].NumVars() != on.NumVars() {
			return nil
		}
		lit := bsat.Positive(1) ^ side
		status := collapseSolve(solvers[side], []int32{lit}, conflictLimit)
		if status == bsat.Unknown {
			return nil
		}
		if status == bsat.Unsatisfiable {
			if side == 0 {
				return constantCover('0')
			}
			return constantCover('1')
		}
		if !solvers[side].AddClause(lit) || !clean[side].AddClause(lit) {
			return nil
		}
	}
	covers := [2]*vec.Vec_Str_t{vec.Vec_StrAlloc(100), vec.Vec_StrAlloc(100)}
	lits, nums, temp := vec.Vec_IntAlloc(int(nVars)), vec.Vec_IntAlloc(int(nVars)), vec.Vec_IntAlloc(int(nVars))
	for round := int32(0); cubeLimit == 0 || round < cubeLimit; round++ {
		for side := int32(0); side < 2; side++ {
			s := solvers[side]
			status := bsat.Unknown
			if canonical != 0 {
				order := vec.Vec_IntAlloc(int(nVars))
				for i := int32(0); i < nVars; i++ {
					idx := i
					if reverse != 0 {
						idx = nVars - 1 - i
					}
					vec.Vec_IntPush(order, bsat.Negative(base+idx))
				}
				status = bsat.Status(Bmc_ComputeCanonical(s, order, temp, conflictLimit))
			} else {
				status = collapseSolve(s, nil, conflictLimit)
			}
			if status == bsat.Unknown {
				return nil
			}
			if status == bsat.Unsatisfiable {
				vec.Vec_StrPush(covers[side], 0)
				return covers[side]
			}
			vec.Vec_IntClear(lits)
			for i := int32(0); i < nVars; i++ {
				idx := i
				if reverse != 0 {
					idx = nVars - 1 - i
				}
				variable := base + idx
				value, ok := s.Value(variable)
				if !ok {
					return nil
				}
				lit := bsat.Positive(variable)
				if !value {
					lit = bsat.Negative(variable)
				}
				vec.Vec_IntPush(lits, lit)
			}
			if Bmc_CollapseExpand(clean[1-side], s, lits, nums, temp, conflictLimit, canonical, -1) < 0 {
				return nil
			}
			cube := make([]byte, int(nVars)+3)
			for i := int32(0); i < nVars; i++ {
				cube[i] = '-'
			}
			cube[nVars] = ' '
			cube[nVars+1] = '1' - byte(side)
			cube[nVars+2] = '\n'
			var block []bsat.Literal
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
	_ = verbose
	return nil
}
