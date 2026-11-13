package rwr

import "testing"

func TestSourceTraversalCounterReset(t *testing.T) {
	a, b := &Rwr_Node_t{TravId: 9}, &Rwr_Node_t{TravId: 4}
	p := &Rwr_Man_t{nTravIds: 7, vForest: []*Rwr_Node_t{a, b}}
	Rwr_ManIncTravId(p)
	if p.nTravIds != 8 || a.TravId != 9 || b.TravId != 4 {
		t.Fatal("ordinary increment modified marks")
	}
	p.nTravIds = -1879048193 // 0x8fffffff, the source's unsigned threshold.
	Rwr_ManIncTravId(p)
	if p.nTravIds != 1 || a.TravId != 0 || b.TravId != 0 {
		t.Fatal("threshold did not clear marks")
	}
}
