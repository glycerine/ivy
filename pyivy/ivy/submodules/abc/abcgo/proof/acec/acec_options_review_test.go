package acec

import "testing"

// The source initializer overwrites every field, not just the conflict limit.
func TestSourceParameterDefaultsReplacePriorValues(t *testing.T) {
	p := Acec_ParCec_t{NBTLimit: 7, TimeLimit: 7, FMiter: 7, FDualOutput: 7, FTwoOutput: 7, FBooth: 7, FSilent: 7, fVeryVerbose: 7, FVerbose: 7, iOutFail: 7}
	Acec_ManCecSetDefaultParams(&p)
	want := Acec_ParCec_t{NBTLimit: 1000, iOutFail: -1}
	if p != want {
		t.Fatalf("parameters %#v, source defaults %#v", p, want)
	}
}

func TestSourceSolveIgnoresInactiveOptions(t *testing.T) {
	// These source fields are not read by Acec_Solve; only nBTLimit and fVerbose are.
	for _, mutate := range []struct {
		name string
		set  func(*Acec_ParCec_t)
	}{
		{"timeout", func(p *Acec_ParCec_t) { p.TimeLimit = 1 }},
		{"miter", func(p *Acec_ParCec_t) { p.FMiter = 1 }},
		{"dual", func(p *Acec_ParCec_t) { p.FDualOutput = 1 }},
		{"two_output", func(p *Acec_ParCec_t) { p.FTwoOutput = 1 }},
		{"booth", func(p *Acec_ParCec_t) { p.FBooth = 1 }},
		{"silent", func(p *Acec_ParCec_t) { p.FSilent = 1 }},
		{"very_verbose", func(p *Acec_ParCec_t) { p.fVeryVerbose = 1 }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			var pars Acec_ParCec_t
			Acec_ManCecSetDefaultParams(&pars)
			mutate.set(&pars)
			before := pars
			if got := Acec_Solve(reviewAdderChain(-1), reviewAdderChain(-1), &pars); got != 1 {
				t.Fatalf("identical graphs: status %d", got)
			}
			if pars != before {
				t.Fatalf("parameter block was changed: before %#v after %#v", before, pars)
			}
		})
	}
}
