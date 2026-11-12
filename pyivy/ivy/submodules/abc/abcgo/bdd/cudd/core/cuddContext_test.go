package core

import (
	"context"
	"testing"
	"time"
)

// A deterministic context cancels inside graph recursion, not just at entry.
type countingContext struct{ remaining int }

func (c *countingContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *countingContext) Done() <-chan struct{}       { return nil }
func (c *countingContext) Value(any) any               { return nil }
func (c *countingContext) Err() error {
	c.remaining--
	if c.remaining < 0 {
		return context.Canceled
	}
	return nil
}
func TestContextCancellationLeavesCanonicalManager(t *testing.T) {
	dd := NewBDDManager(5)
	dst := NewBDDManager(5)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[4])
	cube := dd.Vars[2]
	if BDDAndContext(&countingContext{2}, dd, f, g) != nil {
		t.Fatal("AND did not cancel inside recursion")
	}
	if BDDAndAbstractContext(&countingContext{2}, dd, f, g, cube) != nil {
		t.Fatal("abstraction did not cancel")
	}
	if BDDTransferContext(&countingContext{4}, dd, dst, f, []int{4, 3, 2, 1, 0}) != nil {
		t.Fatal("transfer did not cancel")
	}
	if BDDAndContext(context.Background(), dd, f, g) != Cudd_bddAnd(dd, f, g) {
		t.Fatal("retry and")
	}
	want := Cudd_bddExistAbstract(dd, Cudd_bddAnd(dd, f, g), cube)
	if BDDAndAbstractContext(context.Background(), dd, f, g, cube) != want {
		t.Fatal("retry abstraction")
	}
	if got := BDDTransferContext(context.Background(), dd, dst, f, nil); got == nil {
		t.Fatal("retry transfer")
	}
}
func TestFusedAndAbstractAllThreeVariableFunctions(t *testing.T) {
	dd := NewBDDManager(3)
	for a := uint64(0); a < 256; a++ {
		f := bddFromTruth(dd, uint8(a), 3)
		for b := uint64(0); b < 256; b++ {
			g := bddFromTruth(dd, uint8(b), 3)
			for mask := 0; mask < 8; mask++ {
				cube := dd.One
				for i := 0; i < 3; i++ {
					if mask&(1<<i) != 0 {
						cube = Cudd_bddAnd(dd, cube, dd.Vars[i])
					}
				}
				want := Cudd_bddExistAbstract(dd, Cudd_bddAnd(dd, f, g), cube)
				if got := Cudd_bddAndAbstract(dd, f, g, cube); got != want {
					t.Fatalf("%x %x %d", a, b, mask)
				}
			}
		}
	}
}
