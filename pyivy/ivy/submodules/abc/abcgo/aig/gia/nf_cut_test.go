package gia

import (
	"reflect"
	"testing"
)

func TestNfCutMergeAndConfiguration(t *testing.T) {
	cfg := Nf_Cfg_t{fCompl: 1, Phase: 0x55, Perm: 0x654321}
	if got := Nf_Int2Cfg(Nf_Cfg2Int(cfg)); got != cfg {
		t.Fatalf("config %v", got)
	}
	for i := int32(0); i < 6; i++ {
		if Nf_CfgVar(cfg, i) != int32((cfg.Perm>>uint(4*i))&15) || Nf_CfgCompl(cfg, i) != int32((cfg.Phase>>uint(i))&1) {
			t.Fatal("config field")
		}
	}
	unit := Nf_Cut_t{}
	Nf_CutCreateUnit(&unit, 65)
	if unit.iFunc != 2 || unit.Sign != 2 {
		t.Fatal(unit)
	}
	makeCut := func(leaves ...int32) Nf_Cut_t {
		c := Nf_Cut_t{nLeaves: uint32(len(leaves)), Sign: Nf_CutGetSign(leaves)}
		copy(c.pLeaves[:], leaves)
		return c
	}
	a, b, c := makeCut(1, 3, 5), makeCut(2, 3, 7), makeCut(1, 4)
	var out Nf_Cut_t
	if Nf_CutMergeOrder(&a, &b, &out, 5) != 1 || !reflect.DeepEqual(out.pLeaves[:out.nLeaves], []int32{1, 2, 3, 5, 7}) {
		t.Fatal(out)
	}
	before := out
	if Nf_CutMergeOrderMux(&a, &b, &c, &out, 5) != 0 || out != before {
		t.Fatal("overflow changed cut")
	}
	if Nf_CutMergeOrderMux(&a, &b, &c, &out, 6) != 1 || !reflect.DeepEqual(out.pLeaves[:out.nLeaves], []int32{1, 2, 3, 4, 5, 7}) || Nf_SetCutIsContainedOrder(&out, &a) != 1 {
		t.Fatal(out)
	}
}
