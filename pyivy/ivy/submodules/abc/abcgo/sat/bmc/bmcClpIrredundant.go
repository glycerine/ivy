package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func sopCubes(sop *vec.Vec_Str_t, count, variables int32) [][]byte {
	if sop == nil || count < 0 || variables < 0 || vec.Vec_StrSize(sop) != int(count*(variables+3)+1) {
		panic("bmc: invalid SOP cover")
	}
	data := vec.Vec_StrArray(sop)
	cubes := make([][]byte, count)
	for i := range cubes {
		start := i * int(variables+3)
		cubes[i] = append([]byte(nil), data[start:start+int(variables+3)]...)
		for _, v := range cubes[i][:variables] {
			if v != '0' && v != '1' && v != '-' {
				panic("bmc: invalid SOP literal")
			}
		}
	}
	return cubes
}
func sopReplace(sop *vec.Vec_Str_t, cubes [][]byte, keep []bool) {
	vec.Vec_StrClear(sop)
	for i, cube := range cubes {
		if keep[i] {
			for _, b := range cube {
				vec.Vec_StrPush(sop, b)
			}
		}
	}
	vec.Vec_StrPush(sop, 0)
}
func sopAssumptions(cube []byte, variables int32) []bsat.Literal {
	var lits []bsat.Literal
	for i := int32(0); i < variables; i++ {
		switch cube[i] {
		case '0':
			lits = append(lits, bsat.Positive(i))
		case '1':
			lits = append(lits, bsat.Negative(i))
		}
	}
	return lits
}

// Bmc_CollapseIrredundant removes cubes covered by later cubes.
func Bmc_CollapseIrredundant(sop *vec.Vec_Str_t, count, variables int32) int32 {
	cubes := sopCubes(sop, count, variables)
	keep := make([]bool, len(cubes))
	solver := bsat.NewSolver()
	for i := int32(0); i < variables; i++ {
		solver.NewVariable()
	}
	for i := len(cubes) - 1; i >= 0; i-- {
		assumptions := sopAssumptions(cubes[i], variables)
		status := solver.Solve(assumptions...)
		if status == bsat.Unknown {
			return 0
		}
		if status == bsat.Unsatisfiable {
			continue
		}
		keep[i] = true
		clause := make([]bsat.Literal, len(assumptions))
		for j, lit := range assumptions {
			clause[j] = lit ^ 1
		}
		solver.AddClause(clause...)
	}
	sopReplace(sop, cubes, keep)
	return 1
}

// Bmc_CollapseIrredundantFull checks each cube against all other retained cubes.
func Bmc_CollapseIrredundantFull(sop *vec.Vec_Str_t, count, variables int32) int32 {
	cubes := sopCubes(sop, count, variables)
	keep := make([]bool, len(cubes))
	solver := bsat.NewSolver()
	for i := int32(0); i < variables+count; i++ {
		solver.NewVariable()
	}
	for i, cube := range cubes {
		keep[i] = true
		clause := []bsat.Literal{bsat.Negative(variables + int32(i))}
		for _, lit := range sopAssumptions(cube, variables) {
			clause = append(clause, lit^1)
		}
		if !solver.AddClause(clause...) {
			panic("bmc: inconsistent cover encoding")
		}
	}
	for i, cube := range cubes {
		assumptions := sopAssumptions(cube, variables)
		for j, active := range keep {
			if active && j != i {
				assumptions = append(assumptions, bsat.Positive(variables+int32(j)))
			}
		}
		status := solver.Solve(assumptions...)
		if status == bsat.Unknown {
			return 0
		}
		if status == bsat.Unsatisfiable {
			keep[i] = false
		}
	}
	sopReplace(sop, cubes, keep)
	return 1
}
