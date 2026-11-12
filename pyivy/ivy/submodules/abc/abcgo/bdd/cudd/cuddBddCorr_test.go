package cudd

import (
	"math"
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCorrelationProbabilityAndWeights(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	if got := Cudd_bddCorrelation(dd, x, x); got != 1 {
		t.Fatalf("corr(x,x) = %g, want 1", got)
	}
	if got := Cudd_bddCorrelation(dd, x, Cudd_Not(x)); got != 0 {
		t.Fatalf("corr(x,!x) = %g, want 0", got)
	}
	if got := Cudd_bddCorrelation(dd, x, y); got != 0.5 {
		t.Fatalf("corr(x,y) = %g, want 0.5", got)
	}
	p := []float64{0.2, 0.7}
	got := Cudd_bddCorrelationWeights(dd, x, y, p)
	if math.Abs(got-0.38) > 1e-12 {
		t.Fatalf("weighted corr(x,y) = %g, want 0.38", got)
	}
}
