package fraig

import (
	"math/rand"
	"testing"
)

// A 16-bit arithmetic graph drives the manager's simulation pool past the
// small fixtures and checks the full ripple carry cone on deterministic cases.
func TestSixteenBitAdderPoolBackedWorkload(t *testing.T) {
	const width = 16
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FDoSparse: 1, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := make([]*Fraig_Node_t, width), make([]*Fraig_Node_t, width)
	for i := 0; i < width; i++ {
		a[i] = Fraig_ManReadIthVar(p, int32(i))
		b[i] = Fraig_ManReadIthVar(p, int32(width+i))
	}
	carry := Fraig_Not(p.pConst1)
	sum := make([]*Fraig_Node_t, width)
	for i := 0; i < width; i++ {
		propagate := Fraig_NodeExor(p, a[i], b[i])
		sum[i] = Fraig_NodeExor(p, propagate, carry)
		carry = Fraig_NodeOr(p, Fraig_NodeAnd(p, a[i], b[i]), Fraig_NodeAnd(p, propagate, carry))
	}
	rng := rand.New(rand.NewSource(0x16add))
	for trial := 0; trial < 256; trial++ {
		x, y := rng.Uint32()&0xffff, rng.Uint32()&0xffff
		if trial < 4 {
			x = []uint32{0, 0xffff, 0xffff, 0x8000}[trial]
			y = []uint32{0, 1, 0xffff, 0x8000}[trial]
		}
		model := make([]int32, 2*width)
		for i := 0; i < width; i++ {
			model[i] = int32(x >> i & 1)
			model[width+i] = int32(y >> i & 1)
		}
		got := uint32(0)
		for i, node := range sum {
			got |= uint32(Fraig_ManSimulateBitNode(p, node, model)) << i
		}
		got |= uint32(Fraig_ManSimulateBitNode(p, carry, model)) << width
		if want := x + y; got != want {
			t.Fatalf("%d + %d = %d, want %d", x, y, got, want)
		}
	}
}
