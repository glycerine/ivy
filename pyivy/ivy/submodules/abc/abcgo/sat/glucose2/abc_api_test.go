package glucose2

import "testing"

func TestABCAndEncodingTruthTable(t *testing.T) {
	for _, phases := range []struct{ a, b, out bool }{
		{}, {true, false, false}, {false, true, true}, {true, true, true},
	} {
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				for out := int32(0); out < 2; out++ {
					s := NewABCSolver()
					if !s.AddABCAnd(2, 0, 1, phases.a, phases.b, phases.out) {
						t.Fatal("encoding unexpectedly made the formula inconsistent")
					}
					want := ((a ^ boolBit(phases.a)) & (b ^ boolBit(phases.b))) ^ boolBit(phases.out)
					assumptions := []int32{abcLit(0, a == 0), abcLit(1, b == 0), abcLit(2, out == 0)}
					got := s.SolveABCLiterals(assumptions)
					if (got == GLUCOSE_SAT) != (out == want) {
						t.Fatalf("AND phases %+v, inputs (%d,%d,%d): status %d, want output %d", phases, a, b, out, got, want)
					}
				}
			}
		}
	}
}

func TestABCXorEncodingTruthTable(t *testing.T) {
	for complement := 0; complement < 2; complement++ {
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				for out := int32(0); out < 2; out++ {
					s := NewABCSolver()
					if !s.AddABCXor(2, 0, 1, complement != 0) {
						t.Fatal("encoding unexpectedly made the formula inconsistent")
					}
					want := a ^ b ^ int32(complement)
					assumptions := []int32{abcLit(0, a == 0), abcLit(1, b == 0), abcLit(2, out == 0)}
					got := s.SolveABCLiterals(assumptions)
					if (got == GLUCOSE_SAT) != (out == want) {
						t.Fatalf("XOR complement %d, inputs (%d,%d,%d): status %d, want output %d", complement, a, b, out, got, want)
					}
				}
			}
		}
	}
}

func TestMinimizeABCAssumptionsRetainsFixedPrefix(t *testing.T) {
	s := NewABCSolver()
	s.EnsureABCVariables(3)
	// a and b may not both be true. c is unrelated and should be removed.
	if !s.AddABCClause([]int32{abcLit(0, true), abcLit(1, true)}) {
		t.Fatal("could not add base clause")
	}
	input := []int32{abcLit(0, false), abcLit(1, false), abcLit(2, false)}
	got := s.MinimizeABCAssumptions(input, 1)
	want := []int32{input[0], input[1]}
	if !equalABCInts(got, want) {
		t.Fatalf("minimized assumptions %v, want %v", got, want)
	}
	if !equalABCInts(input, []int32{abcLit(0, false), abcLit(1, false), abcLit(2, false)}) {
		t.Fatalf("input assumptions were mutated: %v", input)
	}
}

func TestSolveABCLiteralsReturnsStatusAndConflictCopy(t *testing.T) {
	s := NewABCSolver()
	s.EnsureABCVariables(1)
	if !s.AddABCClause([]int32{abcLit(0, false)}) {
		t.Fatal("could not add unit clause")
	}
	if got := s.SolveABCLiterals([]int32{abcLit(0, true)}); got != GLUCOSE_UNSAT {
		t.Fatalf("status = %d, want UNSAT", got)
	}
	core := s.ABCConflict()
	if len(core) != 1 || core[0] != 0 {
		t.Fatalf("conflict = %v, want the solver's unit conflict clause", core)
	}
	core[0] = 123
	if int32(s.Conflict[0]) != 0 {
		t.Fatal("ABCConflict returned storage owned by the solver")
	}
}

func boolBit(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func equalABCInts(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
