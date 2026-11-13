package fra

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Fra_InvariantVerify checks each clause in the initial state and checks that
// their conjunction implies each clause in the next frame. For frames other
// than 1, ABC returns success without checking; callers must not rely on it.
// Literal values use
// ABC's zero-based SAT encoding and clause ends are cumulative offsets.
func Fra_InvariantVerify(graph *aig.Graph, frames int32, ends, literals *vec.Vec_Int_t) int32 {
	start := time.Now()
	if frames != 1 {
		fmt.Println("Invariant verification: Can only verify for K = 1")
		// ABC returns success despite performing no check for this depth.
		return 1
	}
	data := cnf.Cnf_DeriveSimple(graph, int32(graph.OutputCount()))
	defer cnf.Cnf_DataFree(data)
	values := vec.Vec_IntArray(literals)
	previous := int32(0)
	for _, end := range vec.Vec_IntArray(ends) {
		if end < previous || int64(end) > int64(len(values)) {
			fmt.Println("Invariant verification: invalid clause offsets.")
			return 0
		}
		previous = end
	}
	if int64(previous) != int64(len(values)) {
		fmt.Println("Invariant verification: incomplete clause offsets.")
		return 0
	}
	for _, literal := range values {
		if literal < 0 || int64(literal)>>1 >= int64(data.Variables) {
			fmt.Println("Invariant verification: invalid clause literal.")
			return 0
		}
	}
	check := func(solver *bsat.Solver, offset int32) int {
		failures, begin := 0, 0
		for _, end := range vec.Vec_IntArray(ends) {
			assumptions := make([]bsat.Literal, int(end)-begin)
			for i, literal := range values[begin:end] {
				assumptions[i] = bsat.Literal((literal + offset*2) ^ 1)
			}
			if solver.Solve(assumptions...) != bsat.Unsatisfiable {
				failures++
			}
			begin = int(end)
		}
		return failures
	}
	base := cnf.Cnf_DataWriteIntoSolver(data, 1, 1)
	if base == nil {
		fmt.Println("Invariant verification: failed to construct base solver.")
		return 0
	}
	baseFailures := check(base, 0)
	induction := cnf.Cnf_DataWriteIntoSolver(data, 2, 0)
	if induction == nil {
		fmt.Println("Invariant verification: failed to construct inductive solver.")
		return 0
	}
	begin := 0
	for _, end := range vec.Vec_IntArray(ends) {
		clause := make([]bsat.Literal, int(end)-begin)
		for i, literal := range values[begin:end] {
			clause[i] = bsat.Literal(literal)
		}
		if !induction.AddClause(clause...) {
			fmt.Println("Invariant verification: SAT solver is unsat after adding a clause.")
			return 0
		}
		begin = int(end)
	}
	indFailures := check(induction, int32(data.Variables))
	if baseFailures != 0 {
		fmt.Printf("Invariant verification: %d clauses (out of %d) FAILED the base case.\n", baseFailures, vec.Vec_IntSize(ends))
	}
	if indFailures != 0 {
		fmt.Printf("Invariant verification: %d clauses (out of %d) FAILED the inductive case.\n", indFailures, vec.Vec_IntSize(ends))
	}
	if baseFailures != 0 || indFailures != 0 {
		return 0
	}
	fmt.Printf("Invariant verification: %d clauses verified correctly.  Time = %.2f sec\n", vec.Vec_IntSize(ends), time.Since(start).Seconds())
	return 1
}
