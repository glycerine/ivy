package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func collapseSolve(s *bsat.Solver, lits []int32, limit int32) bsat.Status {
	if s == nil || limit < 0 {
		panic("bmc: invalid collapse solver")
	}
	s.SetLimits(int64(limit), 0)
	return s.Solve(lits...)
}
func collapseAssumptions(lits *vec.Vec_Int_t, condition int32, skip int, flip bool) []int32 {
	var result []int32
	if condition >= 0 {
		result = append(result, condition)
	}
	for i, lit := range vec.Vec_IntArray(lits) {
		if lit < 0 || i == skip && !flip {
			continue
		}
		if i == skip && flip {
			lit ^= 1
		}
		result = append(result, lit)
	}
	return result
}
func Bmc_CollapseExpandRound2(s *bsat.Solver, lits, temp *vec.Vec_Int_t, limit, condition int32) int32 {
	return Bmc_CollapseExpandRound(s, nil, lits, nil, temp, limit, 0, condition)
}
func Bmc_CollapseExpandRound(s, on *bsat.Solver, lits, nums, temp *vec.Vec_Int_t, limit, canonical, condition int32) int32 {
	for k := vec.Vec_IntSize(lits) - 1; k >= 0; k-- {
		saved := vec.Vec_IntEntry(lits, k)
		if saved < 0 {
			continue
		}
		if on != nil {
			if condition >= 0 {
				panic("bmc: conditional onset expansion")
			}
			if canonical != 0 && saved&1 == 0 {
				continue
			}
			status := collapseSolve(on, collapseAssumptions(lits, -1, k, true), limit)
			if status == bsat.Unknown {
				return -1
			}
			if status == bsat.Unsatisfiable {
				continue
			}
		}
		status := collapseSolve(s, collapseAssumptions(lits, condition, k, false), limit)
		if status == bsat.Unknown {
			return -1
		}
		if status == bsat.Unsatisfiable {
			vec.Vec_IntWriteEntry(lits, k, -1)
		}
	}
	_ = nums
	_ = temp
	return 0
}
func Bmc_CollapseExpand(s, on *bsat.Solver, lits, nums, temp *vec.Vec_Int_t, limit, canonical, condition int32) int32 {
	if canonical == 0 {
		status := collapseSolve(s, collapseAssumptions(lits, condition, -1, false), limit)
		if status == bsat.Unknown {
			return -1
		}
		if status != bsat.Unsatisfiable {
			panic("bmc: cube is not disjoint from offset")
		}
		core := make(map[int32]bool)
		for _, lit := range s.UnsatCore() {
			core[lit^1] = true
		}
		for i, lit := range vec.Vec_IntArray(lits) {
			if lit >= 0 && !core[lit] {
				vec.Vec_IntWriteEntry(lits, i, -1)
			}
		}
		if Bmc_CollapseExpandRound(s, nil, lits, nums, temp, limit, canonical, condition) < 0 {
			return -1
		}
	} else {
		if Bmc_CollapseExpandRound(s, on, lits, nums, temp, limit, canonical, -1) < 0 || Bmc_CollapseExpandRound(s, nil, lits, nums, temp, limit, canonical, -1) < 0 {
			return -1
		}
	}
	vec.Vec_IntClear(nums)
	for i, lit := range vec.Vec_IntArray(lits) {
		if lit >= 0 {
			vec.Vec_IntPush(nums, int32(i))
		}
	}
	return 0
}
func Bmc_CollapseExpand2(s, on *bsat.Solver, lits, nums, temp *vec.Vec_Int_t, limit, canonical, condition int32) int32 {
	if canonical != 0 {
		return Bmc_CollapseExpand(s, on, lits, nums, temp, limit, canonical, condition)
	}
	copy := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(lits), vec.Vec_IntSize(lits))
	return Bmc_CollapseExpand(s, on, copy, nums, temp, limit, canonical, condition)
}
func Bmc_ComputeCanonical2(s *bsat.Solver, lits, temp *vec.Vec_Int_t, limit int32) int32 {
	status := bsat.Unknown
	for i := 0; i < vec.Vec_IntSize(lits); i++ {
		prefix := append([]int32(nil), vec.Vec_IntArray(lits)[:i+1]...)
		status = collapseSolve(s, prefix, limit)
		if status == bsat.Unknown {
			return int32(status)
		}
		if status == bsat.Satisfiable {
			continue
		}
		lit := vec.Vec_IntEntry(lits, i)
		if lit&1 == 0 {
			return int32(bsat.Unsatisfiable)
		}
		vec.Vec_IntWriteEntry(lits, i, lit^1)
		for j := i + 1; j < vec.Vec_IntSize(lits); j++ {
			vec.Vec_IntWriteEntry(lits, j, vec.Vec_IntEntry(lits, j)|1)
		}
		i--
	}
	_ = temp
	return int32(status)
}
func Bmc_ComputeCanonical(s *bsat.Solver, lits, temp *vec.Vec_Int_t, limit int32) int32 {
	s.SetLimits(int64(limit), 0)
	_ = temp
	return int32(s.SolveLexicographic(vec.Vec_IntArray(lits)))
}
