package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
	"testing"
)

func TestCardinalityConstraints(t *testing.T) {
	for n := 2; n <= 6; n++ {
		for k := 1; k < n; k++ {
			for _, strict := range []int32{0, 1} {
				for _, pairwise := range []bool{false, true} {
					s := bsat.NewSolver()
					v := vec.Vec_IntAlloc(n)
					for i := 0; i < n; i++ {
						vec.Vec_IntPush(v, s.NewVariable())
					}
					if pairwise {
						Cnf_AddCardinConstrPairWise(s, v, int32(k), strict)
					} else {
						Cnf_AddCardinConstrGeneral(s, v, int32(k), strict)
					}
					for mask := 0; mask < 1<<n; mask++ {
						assumptions := make([]bsat.Literal, n)
						for i := 0; i < n; i++ {
							assumptions[i] = bsat.Negative(int32(i))
							if mask&(1<<i) != 0 {
								assumptions[i] = bsat.Positive(int32(i))
							}
						}
						want := bits.OnesCount(uint(mask)) <= k
						if strict != 0 {
							want = bits.OnesCount(uint(mask)) == k
						}
						got := s.Solve(assumptions...) == bsat.Satisfiable
						if got != want {
							t.Fatalf("n=%d k=%d strict=%d pair=%v mask=%b", n, k, strict, pairwise, mask)
						}
					}
				}
			}
		}
	}
}
func TestCardinalityAtMostOne(t *testing.T) {
	s := bsat.NewSolver()
	v := vec.Vec_IntAlloc(4)
	for i := 0; i < 4; i++ {
		vec.Vec_IntPush(v, s.NewVariable())
	}
	Cnf_AddCardinConstr(s, v)
	if vec.Vec_IntSize(v) != 1 {
		t.Fatal("missing aggregate variable")
	}
	for mask := 0; mask < 16; mask++ {
		a := make([]bsat.Literal, 4)
		for i := range a {
			a[i] = bsat.Negative(int32(i))
			if mask&(1<<i) != 0 {
				a[i] = bsat.Positive(int32(i))
			}
		}
		if got := s.Solve(a...) == bsat.Satisfiable; got != (bits.OnesCount(uint(mask)) <= 1) {
			t.Fatalf("mask %04b", mask)
		}
	}
}
