package tim

import "testing"

func timingFixture() *Tim_Man_t {
	p := &Tim_Man_t{nCis: 3, nCos: 3, pCis: make([]Tim_Obj_t, 3), pCos: make([]Tim_Obj_t, 3), fUseTravId: true, nTravIds: 1}
	for i := range p.pCis {
		p.pCis[i].iObj2Box = -1
		p.pCos[i].iObj2Box = -1
	}
	p.vDelayTables = [][]float32{{0, 2, 2, 1, 3, 4, 2}}
	Tim_ManCreateBox(p, 0, 2, 1, 2, 0, 0)
	return p
}
func TestTimingPropagation(t *testing.T) {
	p := timingFixture()
	Tim_ManSetCoArrival(p, 0, 10)
	Tim_ManSetCoArrival(p, 1, 20)
	if got := Tim_ManGetCiArrival(p, 1); got != 23 {
		t.Fatal(got)
	}
	if got := Tim_ManGetCiArrival(p, 2); got != 22 {
		t.Fatal(got)
	}
	Tim_ManIncrementTravId(p)
	Tim_ManSetCiRequired(p, 1, 50)
	Tim_ManSetCiRequired(p, 2, 40)
	if got := Tim_ManGetCoRequired(p, 0); got != 36 {
		t.Fatal(got)
	}
	if got := Tim_ManGetCoRequired(p, 1); got != 38 {
		t.Fatal(got)
	}
}
func TestDisconnectedTimingArc(t *testing.T) {
	p := timingFixture()
	p.vDelayTables[0][3] = -1000000000
	p.vDelayTables[0][6] = -1000000000
	Tim_ManSetCoArrival(p, 0, 10)
	Tim_ManSetCoArrival(p, 1, 20)
	if got := Tim_ManGetCiArrival(p, 1); got != 23 {
		t.Fatal(got)
	}
	if got := Tim_ManGetCiArrival(p, 2); got != 14 {
		t.Fatal(got)
	}
	Tim_ManIncrementTravId(p)
	Tim_ManSetCiRequired(p, 1, 50)
	Tim_ManSetCiRequired(p, 2, 40)
	if got := Tim_ManGetCoRequired(p, 0); got != 36 {
		t.Fatal(got)
	}
	if got := Tim_ManGetCoRequired(p, 1); got != 47 {
		t.Fatal(got)
	}
}
func TestTraversalAndConnectivity(t *testing.T) {
	p := timingFixture()
	if Tim_ManBoxForCi(p, 0) != -1 || Tim_ManBoxForCi(p, 2) != 0 || Tim_ManBoxForCo(p, 2) != -1 {
		t.Fatal("box mapping")
	}
	Tim_ManSetCurrentTravIdBoxInputs(p, 0)
	if !Tim_ManIsCoTravIdCurrent(p, 1) {
		t.Fatal("current input")
	}
	Tim_ManSetPreviousTravIdBoxInputs(p, 0)
	if Tim_ManIsCoTravIdCurrent(p, 1) {
		t.Fatal("previous input")
	}
	p.nTravIds = (1 << 30) - 1
	Tim_ManIncrementTravId(p)
	if p.nTravIds != 1 || p.pCos[1].TravId != 0 {
		t.Fatal("rollover")
	}
}

func TestFractionalArrivalPreserved(t *testing.T) {
	p := timingFixture()
	Tim_ManSetCoArrival(p, 0, 10.25)
	Tim_ManSetCoArrival(p, 1, 20.75)
	if got := Tim_ManGetCiArrival(p, 1); got != 23.75 {
		t.Fatal(got)
	}
}
