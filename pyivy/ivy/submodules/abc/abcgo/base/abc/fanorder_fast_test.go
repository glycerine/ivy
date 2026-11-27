package abc

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// TestOrderFanColumnsFastMatchesGeneral: on random SOP nodes the byte-level
// fast path leaves the same SOP text and fanin order as the general path,
// for all three ordering modes.
func TestOrderFanColumnsFastMatchesGeneral(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for trial := 0; trial < 3000; trial++ {
		vars, cubes := 1+r.IntN(6), 1+r.IntN(8)
		var b strings.Builder
		out := "01"[r.IntN(2)]
		for c := 0; c < cubes; c++ {
			for v := 0; v < vars; v++ {
				b.WriteByte("-01"[r.IntN(3)])
			}
			b.WriteByte(' ')
			b.WriteByte(out)
			b.WriteByte('\n')
		}
		mode := r.IntN(3)
		make2 := func() *Object {
			n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
			node := Abc_NtkCreateNode(n)
			for v := 0; v < vars; v++ {
				Abc_ObjAddFanin(node, Abc_NtkCreatePi(n))
			}
			node.Data = b.String()
			return node
		}
		fast, general := make2(), make2()
		var sc fanOrderScratch
		if !orderFanColumnsFast(fast, mode, &sc) {
			t.Fatalf("fast path refused %q", b.String())
		}
		orderFanColumnsNode(general, mode)
		ids := func(o *Object) []int {
			var x []int
			for _, e := range o.fanins {
				x = append(x, e.Object.ID)
			}
			return x
		}
		if fast.Data != general.Data || !slices.Equal(ids(fast), ids(general)) {
			t.Fatalf("mode %d SOP %q: fast %q %v, general %q %v", mode, b.String(), fast.Data, ids(fast), general.Data, ids(general))
		}
	}
}
