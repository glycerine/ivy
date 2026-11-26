package abc

import "testing"

// Abc_AigAndCreate orders the arguments so fanin 0 has the smaller ID.
// Every strashed AND creation path must do the same, because rewriting,
// cut enumeration and AIGER output all observe the fanin order.
func TestNetworkAndOrdersFaninsByID(t *testing.T) {
	n := NewNetwork("order")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	for _, e := range []Edge{n.AndEdges(b.Signal(), a.Signal().Not()), n.CreateAndEdges(b.Signal().Not(), a.Signal())} {
		f := e.Object.fanins
		if f[0].Object != a || f[1].Object != b {
			t.Fatalf("fanins are IDs %d,%d; want %d,%d", f[0].Object.ID, f[1].Object.ID, a.ID, b.ID)
		}
	}
	if f := n.AndEdges(b.Signal(), a.Signal().Not()).Object.fanins; !f[0].Complement || f[1].Complement {
		t.Fatal("fanin complements did not follow their objects")
	}
}
