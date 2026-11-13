package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// Build a SAT representation directly from a truth table, independent of the
// BMC AIG-to-CNF conversion and SOP extraction implementations.
func collapseTruthSolver(table int) *bsat.Solver {
	s := bsat.NewSolver()
	for i := 0; i < 5; i++ {
		s.NewVariable()
	}
	for bits := 0; bits < 4; bits++ {
		clause := make([]bsat.Literal, 0, 3)
		for i := 0; i < 2; i++ {
			lit := bsat.Positive(int32(i + 3))
			if bits&(1<<i) != 0 {
				lit ^= 1
			}
			clause = append(clause, lit)
		}
		value := bsat.Positive(2)
		if table&(1<<bits) == 0 {
			value ^= 1
		}
		clause = append(clause, value)
		if !s.AddClause(clause...) {
			panic("inconsistent truth encoding")
		}
	}
	return s
}
func TestCollapseEveryTwoInputFunction(t *testing.T) {
	for table := 0; table < 16; table++ {
		for _, separate := range []bool{false, true} {
			for _, canonical := range []int32{0, 1} {
				for _, reverse := range []int32{0, 1} {
					s := collapseTruthSolver(table)
					var cover *vec.Vec_Str_t
					if separate {
						cover = Bmc_CollapseOne_int2(s, collapseTruthSolver(table), 2, 0, 0, canonical, reverse, 0)
					} else {
						cover = Bmc_CollapseOne_int(s, 2, 0, 0, canonical, reverse, 0)
					}
					if cover == nil {
						t.Fatalf("function=%x separate=%v canonical=%d reverse=%d: no cover", table, separate, canonical, reverse)
					}
					for bits := 0; bits < 4; bits++ {
						inputs := []bool{bits&1 != 0, bits&2 != 0}
						got := algorithmCoverValue(vec.Vec_StrArray(cover), inputs)
						if want := table&(1<<bits) != 0; got != want {
							t.Fatalf("function=%x input=%d separate=%v canonical=%d reverse=%d cover=%q got=%v want=%v", table, bits, separate, canonical, reverse, vec.Vec_StrArray(cover), got, want)
						}
					}
				}
			}
		}
	}
}

func TestConditionalCollapseExpansion(t *testing.T) {
	s := bsat.NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	// condition c selects an offset in which x must be false; y is irrelevant.
	s.AddClause(bsat.Negative(2), bsat.Negative(0))
	lits := vec.Vec_IntAllocArrayCopy([]int32{bsat.Positive(0), bsat.Positive(1)}, 2)
	temp := vec.Vec_IntAlloc(2)
	if got := Bmc_CollapseExpandRound2(s, lits, temp, 0, bsat.Positive(2)); got != 0 {
		t.Fatalf("status %d", got)
	}
	if got := vec.Vec_IntArray(lits); got[0] != bsat.Positive(0) || got[1] != -1 {
		t.Fatalf("expanded cube %v", got)
	}
}

func algorithmCoverValue(sop []byte, in []bool) bool {
	if len(sop) >= 3 && sop[0] == ' ' {
		return sop[1] == '1'
	}
	def := len(sop) > len(in)+1 && sop[len(in)+1] == '0'
	for len(sop) > 1 && sop[0] != 0 {
		match := true
		for i, v := range in {
			if sop[i] != '-' && (sop[i] == '1') != v {
				match = false
			}
		}
		if match {
			return !def
		}
		sop = sop[len(in)+3:]
	}
	return def
}

// A patch is possible exactly when no two input assignments agree on all
// selected divisors while producing different values of the target function.
func TestEcoEveryTwoInputFunctionAndDivisorSet(t *testing.T) {
	for table := 0; table < 16; table++ {
		for subset := 0; subset < 4; subset++ {
			s := bsat.NewSolver()
			for i := 0; i < 3; i++ {
				s.NewVariable()
			}
			for assignment := 0; assignment < 4; assignment++ {
				clause := make([]bsat.Literal, 0, 3)
				for i := 0; i < 2; i++ {
					lit := bsat.Positive(int32(i + 1))
					if assignment&(1<<i) != 0 {
						lit ^= 1
					}
					clause = append(clause, lit)
				}
				target := bsat.Positive(0)
				if table&(1<<assignment) == 0 {
					target ^= 1
				}
				clause = append(clause, target)
				if !s.AddClause(clause...) {
					t.Fatal("truth table unexpectedly inconsistent")
				}
			}
			divisors := vec.Vec_IntAlloc(2)
			for i := 0; i < 2; i++ {
				if subset&(1<<i) != 0 {
					vec.Vec_IntPush(divisors, int32(i+1))
				}
			}
			want := int32(1)
			for a := 0; a < 4; a++ {
				for b := 0; b < 4; b++ {
					if (a^b)&subset == 0 && (table>>a)&1 != (table>>b)&1 {
						want = 0
					}
				}
			}
			if got := Bmc_EcoSolve(s, 0, divisors); got != want {
				t.Fatalf("function=%04b divisors=%02b: got %d want %d", table, subset, got, want)
			}
		}
	}
}

func TestExpandCubesEveryTwoInputFunction(t *testing.T) {
	for table := 0; table < 16; table++ {
		g := gia.Gia_ManStart(32)
		x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gia.Gia_ManHashStart(g)
		out := int32(0)
		var raw []byte
		for bits := 0; bits < 4; bits++ {
			if table&(1<<bits) == 0 {
				continue
			}
			a, b := x, y
			ca, cb := byte('0'), byte('0')
			if bits&1 == 0 {
				a ^= 1
			} else {
				ca = '1'
			}
			if bits&2 == 0 {
				b ^= 1
			} else {
				cb = '1'
			}
			out = gia.Gia_ManHashOr(g, out, gia.Gia_ManHashAnd(g, a, b))
			raw = append(raw, ca, cb, ' ', '1', '\n')
		}
		gia.Gia_ManAppendCo(g, out)
		gia.Gia_ManHashStop(g)
		raw = append(raw, 0)
		sop := vec.Vec_StrAllocArrayCopy(raw, len(raw))
		if status := Abc_ObjExpandCubes(sop, g, 2); status < 0 {
			t.Fatalf("function=%04b status=%d", table, status)
		}
		for bits := 0; bits < 4; bits++ {
			inputs := []bool{bits&1 != 0, bits&2 != 0}
			got := algorithmCoverValue(vec.Vec_StrArray(sop), inputs)
			want := table&(1<<bits) != 0
			if got != want {
				t.Fatalf("function=%04b input=%d sop=%q got=%v want=%v", table, bits, vec.Vec_StrArray(sop), got, want)
			}
		}
	}
}
