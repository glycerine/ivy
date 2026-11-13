package fraig

import (
	"math/rand"
	"testing"
)

// A small arithmetic workload checks complemented edges, internal reconvergent
// cones and functional reduction against ordinary integer addition.
func TestEightBitAdderWorkload(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FDoSparse: 1, NBTLimit: -1})
	defer Fraig_ManFree(p)
	var a, b [8]*Fraig_Node_t
	for i := 0; i < 8; i++ {
		a[i] = Fraig_ManReadIthVar(p, int32(i))
		b[i] = Fraig_ManReadIthVar(p, int32(i+8))
	}
	carry := Fraig_Not(p.pConst1)
	var sum [8]*Fraig_Node_t
	for i := 0; i < 8; i++ {
		prop := Fraig_NodeExor(p, a[i], b[i])
		sum[i] = Fraig_NodeExor(p, prop, carry)
		carry = Fraig_NodeOr(p, Fraig_NodeAnd(p, a[i], b[i]), Fraig_NodeAnd(p, prop, carry))
	}
	rng := rand.New(rand.NewSource(20261113))
	for trial := 0; trial < 1024; trial++ {
		x, y := rng.Intn(256), rng.Intn(256)
		if trial < 4 {
			x = []int{0, 255, 255, 128}[trial]
			y = []int{0, 1, 255, 128}[trial]
		}
		model := make([]int32, 16)
		for i := 0; i < 8; i++ {
			model[i] = int32(x >> i & 1)
			model[8+i] = int32(y >> i & 1)
		}
		got := 0
		for i, edge := range sum {
			got |= int(fraigTestSimulateBitNode(p, edge, model)) << i
		}
		got |= int(fraigTestSimulateBitNode(p, carry, model)) << 8
		if got != x+y {
			t.Fatalf("%d + %d: got %d want %d", x, y, got, x+y)
		}
	}
}
