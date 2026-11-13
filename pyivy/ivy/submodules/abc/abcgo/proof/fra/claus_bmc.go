package fra

import "github.com/glycerine/abcgo/sat/bsat"

// clausFilterBMC removes candidates falsifiable in any selected timeframe,
// preserving the original base-frame clause literals and candidate ordering.
func clausFilterBMC(s *bsat.Solver, clauses [][]int32, costs []int32, prefix, frames, variables int32) int {
	if s == nil || len(clauses) != len(costs) || prefix < 0 || frames < 0 || variables <= 0 {
		panic("invalid clause BMC input")
	}
	removed := 0
	for frame := int32(0); frame < frames; frame++ {
		offset := (prefix + frame) * 2 * variables
		for i, clause := range clauses {
			if costs[i] == -1 {
				continue
			}
			if costs[i] <= 0 {
				panic("invalid clause cost")
			}
			assumptions := make([]int32, len(clause))
			for j, lit := range clause {
				assumptions[j] = (lit + offset) ^ 1
			}
			// Both a counterexample and an unknown SAT result discard the candidate,
			// matching the C routine's conservative proof behavior.
			if s.Solve(assumptions...) != bsat.Unsatisfiable {
				costs[i] = -1
				removed++
			}
		}
	}
	return removed
}
