package abs

import "github.com/glycerine/abcgo/sat/bsat"

// ga2Solve runs sat_solver2_solve without limits under the given assumptions.
func ga2Solve(s *bsat.Sat_solver2, lits ...int32) bsat.Status {
	for _, l := range lits {
		if l>>1 >= bsat.Sat_solver2_nvars(s) {
			bsat.Sat_solver2_setnvars(s, l>>1+1)
		}
	}
	return bsat.Status(bsat.Sat_solver2_solve(s, lits, 0, 0, 0, 0))
}

// ga2NewVar adds one variable and returns its index.
func ga2NewVar(s *bsat.Sat_solver2) int32 {
	v := bsat.Sat_solver2_nvars(s)
	bsat.Sat_solver2_setnvars(s, v+1)
	return v
}
