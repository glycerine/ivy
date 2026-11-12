package kit

import (
	"github.com/glycerine/abcgo/aig/hop"
	"math/rand"
	"testing"
)

func TestPlaAndHopAllThreeVariableFunctions(t *testing.T) {
	for f := uint32(0); f < 256; f++ {
		truth := []uint32{f * 0x01010101}
		var cover []uint32
		var buffer []byte
		pla := Kit_PlaFromTruthNew(truth, 3, &cover, &buffer)
		if uint32(Kit_PlaToTruth6(pla, 3)) != truth[0] {
			t.Fatalf("PLA f=%x text=%q", truth[0], pla)
		}
		inverse := Kit_PlaComplement(pla)
		if uint32(Kit_PlaToTruth6(inverse, 3)) != ^truth[0] {
			t.Fatal("PLA complement")
		}
		manager := hop.Hop_ManStart()
		root := Kit_TruthToHop(manager, truth, 3, nil)
		if uint32(hop.Hop_ManComputeTruth6(manager, root, 3)) != truth[0] {
			t.Fatal("Hop graph bridge", f)
		}
	}
}
func TestLutNetworkTruthAndArrival(t *testing.T) {
	p := Kit_ManDecStart(3)
	supports := []int32{0, 1, 3, 2}
	luts := []uint32{0x66666666, 0x88888888}
	result := make([]uint32, 1)
	Kit_DecComputeTruth(p, 3, supports, 2, luts, result)
	if result[0] != 0x60606060 {
		t.Fatal("LUT truth", result)
	}
	out := make([]byte, 5)
	if arrival := Kit_DecComputeOuputArrival(3, supports, 2, []byte{1, 3, 2}, out); arrival != 5 {
		t.Fatal(arrival, out)
	}
	pattern := make([]int32, 2)
	if count := Kit_DecComputePattern(3, []uint32{0xcacacaca}, 1, pattern); count != 2 || pattern[0] != 0 || pattern[1] != 1 {
		t.Fatal("cofactor pattern", count, pattern)
	}
	shared := make([]int32, 6)
	if count := Kit_DecComputeShared([]int32{0, 1, 2, 3, 4, 5, 6, 7}, 3, shared); count != 2 {
		t.Fatal("shared-variable partition", count, shared)
	}
}
func TestNpnDescriptionsAndIterators(t *testing.T) {
	names := Kit_DsdNpn4ClassNames()
	if len(names) != 222 {
		t.Fatal(len(names))
	}
	old := names[0]
	names[0] = "changed"
	if Kit_DsdNpn4ClassNames()[0] != old {
		t.Fatal("NPN table leaked mutable storage")
	}
	g := Kit_GraphCreate(3)
	g.eRoot = Kit_GraphAddNodeAnd(g, Kit_EdgeCreate(0, 0), Kit_EdgeCreate(1, 0))
	count := 0
	for range Kit_GraphForEachLeaf(g) {
		count++
		break
	}
	if count != 1 {
		t.Fatal("leaf iterator early stop")
	}
	count = 0
	for range Kit_GraphForEachNode(g) {
		count++
	}
	if count != 1 {
		t.Fatal("node iterator")
	}
}

func TestMultiwordFactoringThroughRealHop(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	for n := int32(5); n <= 9; n++ {
		for trial := 0; trial < 5; trial++ {
			truth := make([]uint32, Kit_TruthWordNum(n))
			for i := range truth {
				truth[i] = rng.Uint32()
			}
			manager := hop.Hop_ManStart()
			root := Kit_TruthToHop(manager, truth, n, nil)
			out := hop.Hop_ManConvertAigToTruth(manager, root, n, nil, 0)
			if Kit_TruthIsEqual(out, truth, n) == 0 {
				t.Fatal("multiword Hop factoring", n, trial)
			}
		}
	}
}
