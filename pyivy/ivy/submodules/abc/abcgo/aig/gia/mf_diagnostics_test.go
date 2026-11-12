package gia

import "testing"

func TestMfCheckCuts(t *testing.T) {
	a := &Mf_Cut_t{nLeaves: 2}
	a.pLeaves[0], a.pLeaves[1] = 1, 2
	a.Sign = Mf_CutGetSign(a.pLeaves[:], 2)
	b := &Mf_Cut_t{nLeaves: 2}
	b.pLeaves[0], b.pLeaves[1] = 2, 3
	b.Sign = Mf_CutGetSign(b.pLeaves[:], 2)
	if Mf_SetCheckArray([]*Mf_Cut_t{a, b}, 2) != 1 {
		t.Fatal("valid cuts rejected")
	}
	b.pLeaves[1] = 2
	defer func() {
		if recover() == nil {
			t.Fatal("invalid cut accepted")
		}
	}()
	Mf_SetCheckArray([]*Mf_Cut_t{a, b}, 2)
}
