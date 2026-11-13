package abs

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Ga2_ManCnfAddStatic emits both ISOP covers as clauses labeled by proofID.
func Ga2_ManCnfAddStatic(solver *bsat.LabeledSolver2, cnf0, cnf1 *vec.Vec_Int_t, lits []int32, out, proofID int32) {
	for polarity, cover := range []*vec.Vec_Int_t{cnf0, cnf1} {
		for i := 0; i < vec.Vec_IntSize(cover); i++ {
			cube := uint32(vec.Vec_IntEntry(cover, i))
			output := out
			if polarity == 1 {
				output ^= 1
			}
			clause := []bsat.Literal{bsat.Literal(output)}
			for b := 0; b < 5; b++ {
				value := (cube >> uint(2*b)) & 3
				if value == 0 {
					continue
				}
				if value == 3 || b >= len(lits) {
					panic("abs: invalid CNF cover cube")
				}
				literal := lits[b]
				if value == 2 {
					literal ^= 1
				}
				clause = append(clause, bsat.Literal(literal))
			}
			solver.AddClause(proofID, clause...)
		}
	}
}
