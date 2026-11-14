package gia

import (
	"testing"

	"github.com/glycerine/abcgo/opt/sfm"
)

func TestGiaMfsBridge(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	var mp Jf_Par_t
	Mf_ManSetDefaultPars(&mp)
	mp.FCoarsen = 0
	mapped := Mf_ManPerformMapping(g, &mp)
	n := Gia_ManExtractMfs(mapped)
	for input := 0; input < 4; input++ {
		got, err := n.Evaluate([]bool{input&1 != 0, input&2 != 0})
		if err != nil || len(got) != 1 || got[0] != (input == 3) {
			t.Fatalf("input %d: %v, %v", input, got, err)
		}
	}
	inserted := Gia_ManInsertMfs(mapped, n, 0)
	Gia_ManMappingVerify(inserted)
	pars := sfm.DefaultParams()
	result := Gia_ManPerformMfs(mapped, &pars)
	if result == nil || Gia_ManCiNum(result) != 2 || Gia_ManCoNum(result) != 1 {
		t.Fatalf("invalid optimized network: %v", result)
	}
	Gia_ManMappingVerify(result)
}
