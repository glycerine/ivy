package acec

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// The source initializer overwrites every field, not just the conflict limit.
func TestSourceParameterDefaultsReplacePriorValues(t *testing.T) {
	p := Acec_ParCec_t{nBTLimit: 7, TimeLimit: 7, fMiter: 7, fDualOutput: 7, fTwoOutput: 7, fBooth: 7, fSilent: 7, fVeryVerbose: 7, fVerbose: 7, iOutFail: 7}
	Acec_ManCecSetDefaultParams(&p)
	want := Acec_ParCec_t{nBTLimit: 1000, iOutFail: -1}
	if !reflect.DeepEqual(p, want) {
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
		{"miter", func(p *Acec_ParCec_t) { p.fMiter = 1 }},
		{"dual", func(p *Acec_ParCec_t) { p.fDualOutput = 1 }},
		{"two_output", func(p *Acec_ParCec_t) { p.fTwoOutput = 1 }},
		{"booth", func(p *Acec_ParCec_t) { p.fBooth = 1 }},
		{"silent", func(p *Acec_ParCec_t) { p.fSilent = 1 }},
		{"very_verbose", func(p *Acec_ParCec_t) { p.fVeryVerbose = 1 }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			makeGraph := func() *Gia_Man_t {
				g := gia.Gia_ManStart(4)
				x := gia.Gia_ManAppendCi(g)
				gia.Gia_ManAppendCo(g, x)
				return g
			}
			var pars Acec_ParCec_t
			Acec_ManCecSetDefaultParams(&pars)
			mutate.set(&pars)
			before := pars
			if got := Acec_Solve(makeGraph(), makeGraph(), &pars); got != 1 {
				t.Fatalf("identical graphs: status %d", got)
			}
			if pars != before {
				t.Fatalf("parameter block was changed: before %#v after %#v", before, pars)
			}
		})
	}
}
