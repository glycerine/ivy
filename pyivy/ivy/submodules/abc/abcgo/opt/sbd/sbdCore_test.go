package sbd

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Sbd_NtkPerform keeps the function of every output for each LUT size and
// cut mode, with C's debug-only checks (cut-set invariants, delay
// improvement) enabled where they hold.
func TestSbdNtkPerformPreservesFunction(t *testing.T) {
	sbdCheckCuts = true
	defer func() { sbdCheckCuts = false }()
	for _, c := range []struct {
		K, moreCuts, findDivs, mapping int32
	}{{4, 0, 0, 1}, {4, 1, 0, 1}, {4, 0, 1, 1}, {4, 1, 0, 0}, {2, 0, 0, 1}, {3, 0, 0, 1}, {5, 0, 0, 1}, {6, 0, 0, 0}, {5, 1, 0, 1}} {
		p, err := gia.Gia_AigerRead("../../base/abci/test_vectors/native_mul4_gen.aig", 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		orig, err := gia.Gia_AigerRead("../../base/abci/test_vectors/native_mul4_gen.aig", 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		var pars Sbd_Par_t
		Sbd_ParSetDefault(&pars)
		pars.NLutSize, pars.FMoreCuts, pars.FFindDivs, pars.FMapping = c.K, c.moreCuts, c.findDivs, c.mapping
		res := Sbd_NtkPerform(p, &pars)
		if gia.Gia_ManHasMapping(res) != 0 {
			gia.Gia_ManMappingVerify(res)
		}
		nIns := int(Gia_ManCiNum(orig))
		for m := 0; m < 1<<nIns; m++ {
			in := make([]bool, nIns)
			for i := range in {
				in[i] = m>>i&1 != 0
			}
			want, err := orig.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := res.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("%+v: output %d differs on input %d", c, i, m)
				}
			}
		}
	}
}
