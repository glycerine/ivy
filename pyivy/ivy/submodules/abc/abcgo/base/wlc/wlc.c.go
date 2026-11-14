package wlc

// CNF multiplier construction follows src/base/wlc/wlc.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

type sat_solver = bsat.Solver

func Wlc_BlastFullAdderCtrlCnf(s *sat_solver, a, ac, b, c int32, pc, ps, next *int32) {
	rows := [12][6]int32{
		{-1, 0, 0, 0, 0, 0}, {-1, 0, 0, 1, 0, 1}, {-1, 0, 1, 0, 0, 1}, {-1, 0, 1, 1, 1, 0},
		{0, -1, 0, 0, 0, 0}, {0, -1, 0, 1, 0, 1}, {0, -1, 1, 0, 0, 1}, {0, -1, 1, 1, 1, 0},
		{1, 1, 0, 0, 0, 1}, {1, 1, 0, 1, 1, 0}, {1, 1, 1, 0, 1, 0}, {1, 1, 1, 1, 1, 1},
	}
	vars := [6]int32{a, ac, b, c, *next, *next + 1}
	for _, row := range rows {
		lits := make([]bsat.Literal, 0, 6)
		skip := false
		for i, pol := range row {
			if pol < 0 {
				continue
			}
			if vars[i] == 0 {
				if pol == 1 {
					skip = true
					break
				}
				continue
			}
			if vars[i] == -1 {
				if pol == 0 {
					skip = true
					break
				}
				continue
			}
			lits = append(lits, bsat.Literal(vars[i]*2+pol))
		}
		if skip {
			continue
		}
		if len(lits) <= 2 {
			panic("source asserts CNF clause has more than two literals")
		}
		s.AddClause(lits...)
	}
	*pc = *next
	*next++
	*ps = *next
	*next++
}
func Wlc_BlastMultiplierCnf(s *sat_solver, a, b []int32, na, nb int32, temp, res *vec.Vec_Int_t, next *int32) {
	if na <= 0 || nb <= 0 {
		panic("invalid multiplier dimensions")
	}
	vec.Vec_IntFill(res, int(na+nb), 0)
	out := vec.Vec_IntArray(res)
	vec.Vec_IntFill(temp, int(2*na), 0)
	scratch := vec.Vec_IntArray(temp)
	carries, sums := scratch[:na], scratch[na:]
	for j := int32(0); j < nb; j++ {
		for i := int32(0); i < na; i++ {
			target := &out[j]
			if i != 0 {
				target = &sums[i-1]
			}
			Wlc_BlastFullAdderCtrlCnf(s, a[i], b[j], sums[i], carries[i], &carries[i], target, next)
		}
	}
	sums[na-1] = 0
	carry := int32(0)
	for i := int32(0); i < na; i++ {
		Wlc_BlastFullAdderCtrlCnf(s, -1, carries[i], sums[i], carry, &carry, &out[nb+i], next)
	}
}
func Wlc_BlastMultiplierCnfMain(n int32) *sat_solver {
	r1, r2, tmp := vec.Vec_IntAlloc(int(2*n)), vec.Vec_IntAlloc(int(2*n)), vec.Vec_IntAlloc(int(2*n))
	a, b := make([]int32, n), make([]int32, n)
	next, total := int32(1)+2*n, int32(1)+4*n+4*n*(n+1)
	s := bsat.NewSolver()
	for s.NumVars() < total {
		s.NewVariable()
	}
	for i := int32(0); i < n; i++ {
		a[i], b[i] = 1+i, 1+n+i
	}
	Wlc_BlastMultiplierCnf(s, a, b, n, n, tmp, r1, &next)
	for i := int32(0); i < n; i++ {
		a[i], b[i] = 1+n+i, 1+i
	}
	Wlc_BlastMultiplierCnf(s, a, b, n, n, tmp, r2, &next)
	clause := make([]bsat.Literal, 0, 2*n)
	for i, x := range vec.Vec_IntArray(r1) {
		y := vec.Vec_IntEntry(r2, i)
		z := next
		next++
		// z = x XOR y, encoded with the source solver's four clauses.
		s.AddClause(bsat.Literal(2*x+1), bsat.Literal(2*y+1), bsat.Literal(2*z+1))
		s.AddClause(bsat.Literal(2*x+1), bsat.Literal(2*y), bsat.Literal(2*z))
		s.AddClause(bsat.Literal(2*x), bsat.Literal(2*y+1), bsat.Literal(2*z))
		s.AddClause(bsat.Literal(2*x), bsat.Literal(2*y), bsat.Literal(2*z+1))
		clause = append(clause, bsat.Literal(2*z))
	}
	if next != total {
		panic("CNF variable count mismatch")
	}
	s.AddClause(clause...)
	return s
}
func Wlc_BlastMultiplierCnfTest(n int32) {
	start := time.Now()
	s := Wlc_BlastMultiplierCnfMain(n)
	status := s.Solve()
	_ = bsat.Sat_SolverWriteDimacs(s, "test_mult.cnf", nil, 0)
	for i := int32(0); i < s.NumVars(); i++ {
		v, _ := s.Value(i)
		value := 0
		if v {
			value = 1
		}
		fmt.Printf("%d=%d ", i, value)
	}
	label := "UNSAT"
	if status == bsat.Satisfiable {
		label = "SAT"
	}
	fmt.Printf("\nVerifying for %d bits:  %s   Time = %.2f sec\n", n, label, time.Since(start).Seconds())
}
