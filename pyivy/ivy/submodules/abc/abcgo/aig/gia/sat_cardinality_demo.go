package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Sbm_AddCardinConstrTest() {
	var sorted *Vec_Int_t
	solver := Sbm_AddCardinSolver(3, &sorted)
	bound := bsat.Negative(vec.Vec_IntEntry(sorted, 2))
	fmt.Printf("LogN = 3. N = 8. Vars = %d. Clauses = %d.\n", solver.NumVars(), solver.Statistics().Clauses)
	for count := 1; solver.Solve(bound) == bsat.Satisfiable; count++ {
		block := make([]bsat.Literal, 0, 8)
		fmt.Printf("%3d : ", count)
		for i := int32(0); i < 8; i++ {
			value, _ := solver.Value(i)
			if value {
				fmt.Print("1")
				block = append(block, bsat.Negative(i))
			} else {
				fmt.Print("0")
				block = append(block, bsat.Positive(i))
			}
		}
		fmt.Println()
		if !solver.AddClause(block...) {
			break
		}
	}
}
