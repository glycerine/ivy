package bmc

import (
	"bytes"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
	"strings"
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

func TestCardinalityDiagnostic(t *testing.T) {
	var output bytes.Buffer
	if n := printCardinalityPatterns(&output); n != 28 {
		t.Fatalf("got %d patterns, want 28", n)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		parts := strings.Split(line, ": ")
		if len(parts) != 2 || len(parts[1]) != 8 || strings.Count(parts[1], "1") != 2 || seen[parts[1]] {
			t.Fatalf("invalid pattern %q", line)
		}
		seen[parts[1]] = true
	}
	if len(seen) != 28 {
		t.Fatalf("got %d unique patterns", len(seen))
	}
}

func TestCardinalityAggregateAndSize(t *testing.T) {
	for n := 1; n <= 8; n++ {
		s := bsat.NewSolver()
		vars := vec.Vec_IntAlloc(n)
		for i := 0; i < n; i++ {
			vec.Vec_IntPush(vars, s.NewVariable())
		}
		Cnf_AddCardinConstr(s, vars)
		if s.NumVars() != int32(2*n-1) {
			t.Fatalf("n=%d: got %d variables", n, s.NumVars())
		}
		out := vec.Vec_IntEntry(vars, 0)
		for mask := 0; mask < 1<<n; mask++ {
			for z := 0; z < 2; z++ {
				assumptions := make([]bsat.Literal, n+1)
				for i := 0; i < n; i++ {
					assumptions[i] = bsat.Negative(int32(i))
					if mask&(1<<i) != 0 {
						assumptions[i] = bsat.Positive(int32(i))
					}
				}
				assumptions[n] = bsat.Negative(out)
				if z != 0 {
					assumptions[n] = bsat.Positive(out)
				}
				want := bits.OnesCount(uint(mask)) <= 1 && (z != 0) == (mask != 0)
				if got := s.Solve(assumptions...) == bsat.Satisfiable; got != want {
					t.Fatalf("n=%d mask=%d out=%d got=%v", n, mask, z, got)
				}
			}
		}
	}
}
