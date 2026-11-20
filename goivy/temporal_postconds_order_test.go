package goivy

import "testing"

// Python NormalProgram.fmlas walks self.postconds.values(), a dict, in
// insertion order; goivy must not reorder postconditions by Go map order.
func TestNormalProgramFmlasPostcondOrder(t *testing.T) {
	acfg := NewAstConfig()
	pcs := NewInsMap[string, []*LabeledFormula]()
	var want []*LabeledFormula
	for _, name := range []string{"zeta", "alpha", "mid", "beta", "omega", "_idle", "gamma", "delta"} {
		lf := acfg.NewLabeledFormula(acfg.NewAtom(name), True)
		pcs.Set(name, []*LabeledFormula{lf})
		want = append(want, lf)
	}
	np := &NormalProgram{Postconds: pcs}
	for run := 0; run < 10; run++ {
		got := np.Fmlas()
		if len(got) != len(want) {
			t.Fatalf("got %d fmlas, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != Node(want[i]) {
				t.Fatalf("run %d: fmla %d out of insertion order", run, i)
			}
		}
	}
}
