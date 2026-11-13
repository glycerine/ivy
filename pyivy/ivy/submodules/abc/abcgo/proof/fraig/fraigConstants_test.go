package fraig

import "testing"

func TestConstantConeHelpersRequireNonconstantRoot(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*Fraig_Man_t)
	}{
		{"mark", func(p *Fraig_Man_t) { Fraig_MarkTfi_rec(p, p.pConst1) }},
		{"mark2", func(p *Fraig_Man_t) { Fraig_MarkTfi2_rec(p, p.pConst1) }},
		{"mark3", func(p *Fraig_Man_t) { Fraig_MarkTfi3_rec(p, p.pConst1) }},
		{"supergate membership", func(p *Fraig_Man_t) { Fraig_NodeIsInSupergate(p.pConst1, Fraig_ManReadIthVar(p, 0)) }},
		{"supergate collection", func(p *Fraig_Man_t) { Fraig_CollectSupergate(p.pConst1, 0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
			defer Fraig_ManFree(p)
			p.nTravIds += 2
			defer func() {
				if recover() == nil {
					t.Fatal("C helper descends into the constant node's nil fanins")
				}
			}()
			tc.run(p)
		})
	}
}
