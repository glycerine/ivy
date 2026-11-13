package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func integrationMultiplier(swapped, mutated bool) *gia.Gia_Man_t {
	g := gia.Gia_ManStart(128)
	a0, a1, b0, b1 := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	and := func(a, b int32) int32 {
		if swapped {
			a, b = b, a
		}
		return gia.Gia_ManAppendAnd(g, a, b)
	}
	p00, p10, p01, p11 := and(a0, b0), and(a1, b0), and(a0, b1), and(a1, b1)
	if mutated {
		p10 ^= 1
	}
	bit1 := gia.Gia_ManAppendXor(g, p10, p01)
	carry := and(p10, p01)
	bit2 := gia.Gia_ManAppendXor(g, p11, carry)
	bit3 := and(p11, carry)
	for _, out := range []int32{p00, bit1, bit2, bit3} {
		gia.Gia_ManAppendCo(g, out)
	}
	return g
}
func TestAcecMultiplierAgainstExhaustiveOracle(t *testing.T) {
	for _, mutated := range []bool{false, true} {
		a, b := integrationMultiplier(false, false), integrationMultiplier(true, mutated)
		differs := false
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			x, e1 := a.Evaluate(inputs)
			y, e2 := b.Evaluate(inputs)
			if e1 != nil || e2 != nil {
				t.Fatal(e1, e2)
			}
			for i := range x {
				differs = differs || x[i] != y[i]
			}
		}
		if differs != mutated {
			t.Fatalf("mutated=%v fixture differs=%v", mutated, differs)
		}
		var opts Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&opts)
		status := Acec_Solve(a, b, &opts)
		want := int32(1)
		if mutated {
			want = 0
		}
		if status != want {
			t.Fatalf("mutated=%v status=%d want=%d", mutated, status, want)
		}
		if mutated {
			cex := a.CombinationalCounterexample()
			if cex == nil {
				t.Fatal("missing multiplier witness")
			}
			inputs := make([]bool, 4)
			for i := range inputs {
				inputs[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
			}
			x, _ := a.Evaluate(inputs)
			y, _ := b.Evaluate(inputs)
			diff := false
			for i := range x {
				diff = diff || x[i] != y[i]
			}
			if !diff {
				t.Fatalf("invalid multiplier witness %v", inputs)
			}
		}
	}
}
