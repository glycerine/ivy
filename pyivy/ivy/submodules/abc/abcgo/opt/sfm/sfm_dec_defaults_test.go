package sfm

import "testing"

func TestDecompositionDefaults(t *testing.T) {
	p := Sfm_Par_t{Verbose: true, Area: true, BTLimit: 99}
	Sfm_ParSetDefault3(&p)
	if p.TfoLevMax != 100 || p.TfiLevMax != 100 || p.FanoutMax != 10 || p.MffcMin != 1 || p.MffcMax != 3 || p.VarMax != 6 || p.DecMax != 1 || p.TimeWin != 1 || p.Verbose || p.Area || p.BTLimit != 0 {
		t.Fatalf("defaults %+v", p)
	}
}
