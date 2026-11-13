package acb

import "testing"

func TestParSetDefaultResetsAllFields(t *testing.T) {
	p := Acb_Par_t{nLutSize: -1, nTfoLevMax: -1, nTfiLevMax: -1, nFanoutMax: -1, nWinNodeMax: -1, nGrowthLevel: -1, nBTLimit: -1, nNodesMax: -1, fUseAshen: -1, iNodeOne: -1, fArea: -1, fMoreEffort: -1, fVerbose: -1, fVeryVerbose: -1}
	Acb_ParSetDefault(&p)
	want := Acb_Par_t{nLutSize: 4, nTfoLevMax: 2, nTfiLevMax: 3, nFanoutMax: 20, nWinNodeMax: 100, fArea: 1}
	if p != want {
		t.Fatalf("defaults = %+v, want %+v", p, want)
	}
}
