package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"sort"
)

// Bmc_FxSolve enumerates and minimizes onset cubes, guarded by auxVar.
// Returns 1 when complete, 0 for an ambiguous projection, -1 on timeout.
func Bmc_FxSolve(s *bsat.Solver, out, auxVar int32, vars *vec.Vec_Int_t, dumpPLA, verbose int32, counter *int32, cubes *vec.Vec_Wec_t) int32 {
	variables := vec.Vec_IntArray(vars)
	if cubes != nil {
		vec.Vec_WecClear(cubes)
	}
	if dumpPLA != 0 {
		fmt.Printf(".i %d\n.o 1\n", len(variables))
	}
	conflicts, propagations := s.Limits()
	s.SetLimits(1000000, 0)
	defer s.SetLimits(conflicts, propagations)
	count, reduced, total := int32(0), 0, 0
	result := int32(1)
	for {
		status := s.Solve(bsat.Positive(out+1), bsat.Positive(auxVar))
		if status == bsat.Unknown {
			result = -1
			break
		}
		if status == bsat.Unsatisfiable {
			break
		}
		assumptions := []bsat.Literal{bsat.Negative(out + 1)}
		for _, v := range variables {
			value, ok := s.Value(v)
			if !ok {
				panic("bmc: missing cube model")
			}
			if value {
				assumptions = append(assumptions, bsat.Positive(v))
			} else {
				assumptions = append(assumptions, bsat.Negative(v))
			}
		}
		before := len(assumptions)
		status = s.Solve(assumptions...)
		if status == bsat.Unknown {
			result = -1
			break
		}
		if status == bsat.Satisfiable {
			result = 0
			break
		}
		for i := 1; i < len(assumptions); {
			trial := append([]bsat.Literal(nil), assumptions[:i]...)
			trial = append(trial, assumptions[i+1:]...)
			status = s.Solve(trial...)
			if status == bsat.Unknown {
				result = -1
				break
			}
			if status == bsat.Unsatisfiable {
				assumptions = trial
			} else {
				i++
			}
		}
		if result < 0 {
			break
		}
		status = s.Solve(assumptions...)
		if status == bsat.Unknown {
			result = -1
			break
		}
		if status != bsat.Unsatisfiable {
			result = 0
			break
		}
		core := s.UnsatCore()
		block := []bsat.Literal{bsat.Negative(auxVar)}
		pattern := make([]byte, len(variables))
		for i := range pattern {
			pattern[i] = '-'
		}
		positions := make([]int32, 0, len(core))
		for _, lit := range core {
			if lit == bsat.Positive(out+1) {
				continue
			}
			index := -1
			for j, v := range variables {
				if v == lit>>1 {
					index = j
					break
				}
			}
			if index < 0 {
				panic("bmc: unexpected cube core literal")
			}
			block = append(block, lit)
			bit := int32(0)
			if lit&1 == 0 {
				bit = 1
				pattern[index] = '0'
			} else {
				pattern[index] = '1'
			}
			positions = append(positions, int32(index)*2+bit)
		}
		sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })
		if cubes != nil {
			level := vec.Vec_WecPushLevel(cubes)
			for _, lit := range positions {
				vec.Vec_IntPush(level, lit)
			}
		}
		if dumpPLA != 0 {
			fmt.Printf("%s 1\n", pattern)
		}
		reduced += before - len(core)
		total += len(positions)
		count++
		if !s.AddClause(block...) {
			break
		}
	}
	if dumpPLA != 0 {
		fmt.Printf(".e\n.p %d\n\n", count)
	}
	if verbose != 0 {
		fmt.Printf("Cubes = %d.  Reduced = %d.  Lits = %d\n", count, reduced, total)
	}
	if counter != nil {
		*counter = count
	}
	return result
}
