package kit

import (
	"math/rand"
	"testing"
)

func TestCloudPackedTruthRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(12))
	dd := Cloud_Init(10, 18)
	for vars := int32(0); vars <= 9; vars++ {
		for trial := 0; trial < 5; trial++ {
			truth := make([]uint32, Kit_TruthWordNum(vars))
			for i := range truth {
				truth[i] = rng.Uint32()
			}
			if vars < 5 {
				mask := Kit_BitMask(1 << uint(vars))
				truth[0] &= mask
				for shift := uint(1) << uint(vars); shift < 32; shift *= 2 {
					truth[0] |= truth[0] << shift
				}
			}
			var nodes []int32
			var store [][]uint32
			if Kit_CreateCloudFromTruth(dd, truth, vars, &nodes) == 0 {
				t.Fatal("program creation failed")
			}
			out := Kit_CloudToTruth(nodes, vars, &store, 0)
			if Kit_TruthIsEqual(out, truth, vars) == 0 {
				t.Fatalf("vars=%d in=%x out=%x", vars, truth, out)
			}
		}
	}
	for _, truth := range []uint32{0, ^uint32(0)} {
		var nodes []int32
		var store [][]uint32
		Kit_CreateCloudFromTruth(dd, []uint32{truth}, 3, &nodes)
		if out := Kit_CloudToTruth(nodes, 3, &store, 0); out[0] != truth {
			t.Fatal("constant", out, truth)
		}
	}
}
func TestCloudTruthComposition(t *testing.T) {
	dd := Cloud_Init(3, 14)
	inputs := make([][]uint32, 3)
	for i := range inputs {
		inputs[i] = make([]uint32, 1)
		Kit_TruthIthVar(inputs[i], 3, int32(2-i))
	}
	truth := []uint32{0x96969696}
	var nodes []int32
	var store [][]uint32
	out := Kit_TruthCompose(dd, truth, 3, inputs, 3, &store, &nodes)
	if out[0] != truth[0] {
		t.Fatal("symmetric parity composition", out)
	}
	truth[0] = 0xd8d8d8d8
	out = Kit_TruthCompose(dd, truth, 3, inputs, 3, &store, &nodes)
	for x := 0; x < 8; x++ {
		y := (x&1)<<2 | (x & 2) | (x&4)>>2
		if (out[0] >> uint(x) & 1) != (truth[0] >> uint(y) & 1) {
			t.Fatal("composition order", x)
		}
	}
}
func TestBddFactoringVerification(t *testing.T) {
	for f := uint32(0); f < 256; f++ {
		truth := []uint32{f * 0x01010101}
		var cover, memory []uint32
		phase := Kit_TruthIsop(truth, 3, &cover, 0)
		if phase != 0 {
			t.Fatal(phase)
		}
		g := Kit_SopFactor(cover, 0, 3, &memory)
		if Kit_SopFactorVerify(cover, g, 3) != 1 {
			t.Fatal("BDD factoring mismatch", f)
		}
	}
	dd := Cloud_Init(3, 12)
	truth := []uint32{0xd8d8d8d8}
	f := Kit_TruthToBdd(dd, truth, 3, 0)
	for x := 0; x < 8; x++ {
		values := []bool{x&1 != 0, x&2 != 0, x&4 != 0}
		if Cloud_Evaluate(dd, f, values) != (truth[0]&(1<<uint(x)) != 0) {
			t.Fatal("BDD order", x)
		}
	}
}
