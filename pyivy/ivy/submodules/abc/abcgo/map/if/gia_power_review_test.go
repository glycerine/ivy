package if_

import "testing"

// Switching vectors store float32 bit patterns in integer slots, not counts.
func TestMappingPowerUsesProbabilities(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
	defer If_ManStop(p)
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	root.CutBest = If_Cut_t{nLimit: 4, nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
	If_ManCreateCo(p, root)
	If_ManCreateCo(p, If_Not(root))
	p.SetSwitchingProbabilities([]float32{0, 0.25, 0.5, 0.375, 0.375, 0.375})
	for pass := 0; pass < 2; pass++ {
		If_ManMarkMapping(p)
		if p.dPower != 0.75 {
			t.Fatalf("pass %d: power = %g, want 0.75", pass, p.dPower)
		}
		if root.nRefs != 2 || a.nRefs != 1 || b.nRefs != 1 || p.nNets != 2 {
			t.Fatal("shared output cone was counted more than once")
		}
	}
}
