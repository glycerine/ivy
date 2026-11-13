package fraig

import "testing"

func TestClauseSimulationPhases(t *testing.T) {
	for aPhase := 0; aPhase < 2; aPhase++ {
		for bPhase := 0; bPhase < 2; bPhase++ {
			for aComplement := 0; aComplement < 2; aComplement++ {
				for bComplement := 0; bComplement < 2; bComplement++ {
					a := &Fraig_Node_t{puSimR: []uint32{0x55555555}, fInv: uint32(aPhase)}
					b := &Fraig_Node_t{puSimR: []uint32{0xaaaaaaaa}, fInv: uint32(bPhase)}
					ae, be := Fraig_NotCond(a, aComplement != 0), Fraig_NotCond(b, bComplement != 0)
					x, y := a.puSimR[0], b.puSimR[0]
					if (aPhase != 0) != (aComplement != 0) {
						x = ^x
					}
					if (bPhase != 0) != (bComplement != 0) {
						y = ^y
					}
					want := x|y == ^uint32(0)
					if got := Fraig_ManCheckClauseUsingSimInfo(&Fraig_Man_t{nWordsRand: 1}, ae, be) != 0; got != want {
						t.Fatalf("phases %d %d, complements %d %d: got %t want %t", aPhase, bPhase, aComplement, bComplement, got, want)
					}
				}
			}
		}
	}
}
func TestSimulationInfoOwnsStorage(t *testing.T) {
	n := &Fraig_Node_t{Num: 0, puSimR: []uint32{1, 2}, puSimD: []uint32{3, 4}}
	m := &Fraig_Man_t{nWordsRand: 2, iWordStart: 1, vNodes: &Fraig_NodeVec_t{pArray: []*Fraig_Node_t{n}}}
	info := Fraig_ManGetSimInfo(m)
	if len(info) != 1 || len(info[0]) != 3 || info[0][2] != 3 {
		t.Fatalf("info = %v", info)
	}
	info[0][0] = 42
	if n.puSimR[0] != 1 {
		t.Fatal("simulation info aliases node")
	}
}
