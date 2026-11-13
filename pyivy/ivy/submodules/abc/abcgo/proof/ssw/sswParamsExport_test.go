package ssw

import "testing"

func TestSetInductionDepth(t *testing.T) {
	p := new(Ssw_Pars_t)
	Ssw_ParsSetFramesK(p, 4)
	if p.nFramesK != 4 {
		t.Fatal("depth not set")
	}
}
