package fraig

import "testing"

func TestNodeSimulateAllPhases(t *testing.T) {
	for phaseA := 0; phaseA < 2; phaseA++ {
		for phaseB := 0; phaseB < 2; phaseB++ {
			for inv := 0; inv < 2; inv++ {
				a := &Fraig_Node_t{puSimR: []uint32{0x12345678}, fInv: uint32(phaseA)}
				b := &Fraig_Node_t{puSimR: []uint32{0x98765432}, fInv: uint32(phaseB)}
				n := &Fraig_Node_t{p1: a, p2: b, puSimR: []uint32{0}, fInv: uint32(inv)}
				Fraig_NodeSimulate(n, 0, 1, 1)
				x, y := a.puSimR[0], b.puSimR[0]
				if phaseA != 0 {
					x = ^x
				}
				if phaseB != 0 {
					y = ^y
				}
				want := x & y
				if inv != 0 {
					want = ^want
				}
				if n.puSimR[0] != want || n.uHashR != want*uint32(s_FraigPrimes[0]) {
					t.Fatalf("phase %d %d %d: sim %08x hash %08x, want %08x", phaseA, phaseB, inv, n.puSimR[0], n.uHashR, want)
				}
			}
		}
	}
}
