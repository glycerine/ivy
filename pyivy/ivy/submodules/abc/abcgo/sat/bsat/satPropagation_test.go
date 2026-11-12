package bsat

import (
	"math/rand"
	"reflect"
	"testing"
)

func propagationFixture(n int32) (satAssignments, *Sat_Mem_t, []veci, cla, *clause) {
	a := newSatAssignments(n)
	m := Sat_MemAlloc(9)
	w := make([]veci, 2*n)
	h := Sat_MemAppend(m, nil, 2, 0, 0)
	return a, m, w, h, Sat_MemClauseHand(m, h)
}
func TestCDCLBinaryPropagationAndConflict(t *testing.T) {
	a, m, w, h, b := propagationFixture(3)
	veci_push(&w[toLit(0)], clause_from_lit(toLit(1)))
	veci_push(&w[toLit(1)], clause_from_lit(lit_neg(toLit(2))))
	a.enqueue(toLit(0), 0)
	var stats stats_t
	if c := satPropagate(&a, m, w, h, b, nil, &stats); c != 0 || a.assigns[1] != 0 || a.assigns[2] != 1 || a.reasons[1] != clause_from_lit(toLit(0)) || a.qhead != 3 {
		t.Fatal(c, a)
	}
	if stats.propagations != 3 {
		t.Fatal(stats)
	}
	a, m, w, h, b = propagationFixture(2)
	veci_push(&w[toLit(0)], clause_from_lit(toLit(1)))
	a.enqueue(lit_neg(toLit(1)), 0)
	a.enqueue(toLit(0), 0)
	if c := satPropagate(&a, m, w, h, b, nil, &stats); c != h || b.lits[0] != toLit(1) || b.lits[1] != lit_neg(toLit(0)) {
		t.Fatal(c, b)
	}
}
func TestCDCLWatchMovementAndUnit(t *testing.T) {
	a, m, w, h, b := propagationFixture(3)
	cHandle := Sat_MemAppend(m, []int32{lit_neg(toLit(0)), toLit(1), toLit(2)}, 3, 1, 0)
	veci_push(&w[toLit(0)], cHandle)
	veci_push(&w[lit_neg(toLit(1))], cHandle)
	a.enqueue(toLit(0), 0)
	var stats stats_t
	if c := satPropagate(&a, m, w, h, b, nil, &stats); c != 0 || w[toLit(0)].size != 0 || w[lit_neg(toLit(2))].size != 1 {
		t.Fatal(c, w)
	}
	a.enqueue(lit_neg(toLit(2)), 0)
	if c := satPropagate(&a, m, w, h, b, nil, &stats); c != 0 || a.assigns[1] != 0 || a.reasons[1] != cHandle {
		t.Fatal(c, a)
	}
	if Sat_MemClauseHand(m, cHandle).lbd != 1 {
		t.Fatal("LBD")
	}
}
func TestCDCLWatchConflictPreservesRemaining(t *testing.T) {
	a, m, w, h, b := propagationFixture(4)
	c := Sat_MemAppend(m, []int32{lit_neg(toLit(0)), toLit(1), toLit(2)}, 3, 0, 0)
	veci_push(&w[toLit(0)], c)
	veci_push(&w[toLit(0)], clause_from_lit(toLit(3)))
	a.enqueue(toLit(0), 0)
	a.enqueue(lit_neg(toLit(1)), 0)
	a.enqueue(lit_neg(toLit(2)), 0)
	var stats stats_t
	if result := satPropagate(&a, m, w, h, b, nil, &stats); result != c || w[toLit(0)].size != 2 || a.assigns[3] != satUnassigned {
		t.Fatal(result, a, w)
	}
}

func TestCDCLPropagationAgainstNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(427))
	for trial := 0; trial < 500; trial++ {
		a, m, w, h, b := propagationFixture(5)
		var clauses [][]lit
		for k := 0; k < 12; k++ {
			permutation := rng.Perm(5)
			size := 2 + rng.Intn(3)
			c := make([]lit, size)
			for j := range c {
				c[j] = toLitCond(int32(permutation[j]), int32(rng.Intn(2)))
			}
			clauses = append(clauses, c)
			if size == 2 {
				veci_push(&w[lit_neg(c[0])], clause_from_lit(c[1]))
				veci_push(&w[lit_neg(c[1])], clause_from_lit(c[0]))
			} else {
				handle := Sat_MemAppend(m, c, int32(size), int32(rng.Intn(2)), 0)
				veci_push(&w[lit_neg(c[0])], handle)
				veci_push(&w[lit_neg(c[1])], handle)
			}
		}
		for v := int32(0); v < 5; v++ {
			if rng.Intn(3) == 0 {
				a.enqueue(toLitCond(v, int32(rng.Intn(2))), 0)
			}
		}
		expected := append([]int32(nil), a.assigns...)
		naiveConflict := false
		changed := true
		for changed && !naiveConflict {
			changed = false
			for _, c := range clauses {
				satisfied := false
				remaining := 0
				var unit lit
				for _, l := range c {
					value := expected[lit_var(l)]
					if value == lit_sign(l) {
						satisfied = true
						break
					}
					if value == satUnassigned {
						remaining++
						unit = l
					}
				}
				if satisfied {
					continue
				}
				if remaining == 0 {
					naiveConflict = true
					break
				}
				if remaining == 1 {
					expected[lit_var(unit)] = lit_sign(unit)
					changed = true
				}
			}
		}
		var stats stats_t
		frequencies := make([]uint8, 5)
		conflict := satPropagate(&a, m, w, h, b, frequencies, &stats) != 0
		if conflict != naiveConflict {
			t.Fatalf("trial %d conflict mismatch: clauses %v", trial, clauses)
		}
		if !conflict && !reflect.DeepEqual(a.assigns, expected) {
			t.Fatalf("trial %d: got %v want %v", trial, a.assigns, expected)
		}
	}
}
