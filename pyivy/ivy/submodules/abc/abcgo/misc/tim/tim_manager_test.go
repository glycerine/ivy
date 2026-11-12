package tim

import (
	"bytes"
	"encoding/binary"
	"testing"
)

type testLibrary struct{}

func (testLibrary) BoxCount() int { return 1 }
func (testLibrary) BoxTiming(id int) (int, int, bool, []float32) {
	return 2, 2, true, []float32{1, 3, 4, 2}
}
func managedFixture() *Tim_Man_t {
	p := Tim_ManStart(5, 5)
	Tim_ManCreateBox(p, 0, 2, 1, 2, 0, 0)
	Tim_ManCreateBox(p, 2, 2, 3, 2, 0, 0)
	Tim_ManBoxSetCopy(p, 0, -1)
	Tim_ManBoxSetCopy(p, 1, -1)
	Tim_ManCreate(p, testLibrary{}, []float32{2.5}, []float32{90.5})
	return p
}
func TestManagerCreateCountsAndClone(t *testing.T) {
	p := managedFixture()
	if Tim_ManPiNum(p) != 1 || Tim_ManPoNum(p) != 1 || Tim_ManBoxNum(p) != 2 || Tim_ManBlackBoxNum(p) != 2 {
		t.Fatal("counts")
	}
	var ins, outs int
	Tim_ManBlackBoxIoNum(p, &ins, &outs)
	if ins != 4 || outs != 4 {
		t.Fatal(ins, outs)
	}
	if a := Tim_ManGetArrTimes(p); len(a) != 5 || a[0] != 2.5 {
		t.Fatal(a)
	}
	if r := Tim_ManGetReqTimes(p); len(r) != 5 || r[0] != 90.5 {
		t.Fatal(r)
	}
	q := Tim_ManDup(p, 0)
	if q.pCis[0].timeArr != 2.5 || q.pCos[4].timeReq != 90.5 {
		t.Fatal("terminal copy")
	}
	q.vDelayTables[0][3] = 99
	q.vBoxes[0].Inouts[0] = 1
	if p.vDelayTables[0][3] != 1 || p.vBoxes[0].Inouts[0] != 0 {
		t.Fatal("shallow clone")
	}
	unit := Tim_ManDup(p, 2)
	for _, delay := range unit.vDelayTables[0][3:] {
		if delay != 2 {
			t.Fatal(delay)
		}
	}
}
func TestTrimReduceAndAlign(t *testing.T) {
	p := managedFixture()
	q := Tim_ManTrim(p, []int{0, 1})
	if q.nCis != 3 || q.nCos != 3 || len(q.vBoxes) != 1 || q.vBoxes[0].iCopy != 1 {
		t.Fatal(q)
	}
	if q.pCos[2].timeReq != 90.5 || q.pCos[2].Id != 2 {
		t.Fatal(q.pCos)
	}
	mask := Tim_ManAlignTwo(p, q)
	if len(mask) != 2 || mask[0] != 0 || mask[1] != 1 {
		t.Fatal(mask)
	}
	reduced := Tim_ManReduce(p, []int{0}, 1)
	if reduced.nCis != 2 || reduced.nCos != 2 || Tim_ManPiNum(reduced) != 0 || Tim_ManPoNum(reduced) != 0 {
		t.Fatal(reduced)
	}
	Tim_ManStopP(&q)
	if q != nil {
		t.Fatal("stop pointer")
	}
}
func TestSaveLoadRoundTrip(t *testing.T) {
	p := managedFixture()
	data := Tim_ManSave(p, 0)
	if binary.BigEndian.Uint32(data[:4]) != 1 {
		t.Fatal("version encoding")
	}
	q := Tim_ManLoad(data, 0)
	if !bytes.Equal(Tim_ManSave(q, 0), data) {
		t.Fatal("non-round-tripping dump")
	}
	if q.pCis[0].timeArr != 2.5 || q.pCos[4].timeReq != 90.5 || len(q.vBoxes) != 2 {
		t.Fatal(q)
	}
	// Version 1 never stored black flags; restoring them requires the box library.
	if Tim_ManBlackBoxNum(q) != 0 {
		t.Fatal("unexpected serialized black flags")
	}
	hierarchy := Tim_ManSave(p, 1)
	q = Tim_ManLoad(hierarchy, 1)
	if len(q.vBoxes) != 2 || q.vDelayTables != nil {
		t.Fatal(q)
	}
}
func TestLoadRejectsCorruption(t *testing.T) {
	valid := Tim_ManSave(managedFixture(), 0)
	for _, data := range [][]byte{nil, valid[:3], valid[:len(valid)-1], append(append([]byte(nil), valid...), 0)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("accepted malformed stream")
				}
			}()
			Tim_ManLoad(data, 0)
		}()
	}
}
func TestDefaultDelayTablesAndPrimaryTimes(t *testing.T) {
	p := Tim_ManStart(2, 2)
	Tim_ManCreateBox(p, 0, 1, 1, 1, -1, 0)
	Tim_ManCreate(p, nil, []float32{7}, []float32{20})
	table := Tim_ManBoxDelayTable(p, 0)
	if len(table) != 4 || table[0] != 0 || table[3] != 1 {
		t.Fatal(table)
	}
	Tim_ManTravIdDisable(p)
	Tim_ManSetCoArrival(p, 0, 5)
	if got := Tim_ManGetCiArrival(p, 1); got != 6 {
		t.Fatal(got)
	}
	Tim_ManInitPiArrivalAll(p, 3)
	if p.pCis[0].timeArr != 3 || p.pCis[1].timeArr != 6 {
		t.Fatal(p.pCis)
	}
	Tim_ManInitPoRequiredAll(p, 25)
	if p.pCos[1].timeReq != 25 {
		t.Fatal(p.pCos)
	}
}

func TestDupUnitTimingDoesNotModifySource(t *testing.T) {
	p := managedFixture()
	q := Tim_ManDup(p, 1)
	if p.pCis[0].timeArr != 2.5 || p.pCos[4].timeReq != 90.5 {
		t.Fatal("duplication changed source constraints")
	}
	if q.pCis[0].timeArr != 0 || q.pCos[4].timeReq != TIM_ETERNITY {
		t.Fatal("unit clone kept old constraints")
	}
	fresh := Tim_ManDup(Tim_ManStart(1, 1), 0)
	Tim_ManCreate(fresh, nil, nil, nil)
}
func TestZeroPortBoxCloneAndSerialization(t *testing.T) {
	for _, ports := range [][2]int{{0, 1}, {1, 0}, {0, 0}} {
		ins, outs := ports[0], ports[1]
		p := Tim_ManStart(1+outs, 1+ins)
		Tim_ManCreateBox(p, 0, ins, 1, outs, -1, 0)
		Tim_ManCreate(p, nil, nil, nil)
		for _, q := range []*Tim_Man_t{Tim_ManDup(p, 0), Tim_ManLoad(Tim_ManSave(p, 0), 0)} {
			if Tim_ManPiNum(q) != 1 || Tim_ManPoNum(q) != 1 || Tim_ManBoxInputNum(q, 0) != ins || Tim_ManBoxOutputNum(q, 0) != outs {
				t.Fatal("zero-port box lost its boundary")
			}
		}
	}
}
func TestFailedTimingQueryDoesNotCacheFailure(t *testing.T) {
	p := Tim_ManStart(1, 1)
	Tim_ManCreateBox(p, 0, 1, 0, 1, -1, 0)
	Tim_ManIncrementTravId(p)
	for _, query := range []func(){func() { Tim_ManGetCiArrival(p, 0) }, func() { Tim_ManGetCoRequired(p, 0) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("missing table accepted")
				}
			}()
			query()
		}()
	}
	if p.pCis[0].TravId == p.nTravIds || p.pCos[0].TravId == p.nTravIds {
		t.Fatal("failed query cached as completed")
	}
	Tim_ManCreate(p, nil, nil, nil)
	Tim_ManSetCoArrival(p, 0, 8)
	if got := Tim_ManGetCiArrival(p, 0); got != 9 {
		t.Fatal(got)
	}
}
func TestLoadRejectsHugeInconsistentTopologyBeforeAllocation(t *testing.T) {
	var data []byte
	for _, value := range []uint32{1, 1 << 28, 1 << 28, 0, 0, 0} {
		data = binary.BigEndian.AppendUint32(data, value)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("inconsistent topology accepted")
		}
	}()
	Tim_ManLoad(data, 1)
}

func TestNegativeTerminalReductionRejected(t *testing.T) {
	p := managedFixture()
	defer func() {
		if recover() == nil {
			t.Error("negative terminal reduction accepted")
		}
		if p.nCis != 5 || p.nCos != 5 {
			t.Error("manager modified before rejection")
		}
	}()
	Tim_ManReduce(p, []int{0}, -1)
}

func TestManagerTransformsPreserveSourceTraversal(t *testing.T) {
	p := managedFixture()
	p.nTravIds = 17
	for i := range p.pCis {
		p.pCis[i].TravId = 17
	}
	for i := range p.pCos {
		p.pCos[i].TravId = 17
	}
	for _, q := range []*Tim_Man_t{Tim_ManDup(p, 1), Tim_ManTrim(p, []int{0, 1}), Tim_ManReduce(p, []int{0}, 0)} {
		for _, o := range q.pCis {
			if o.TravId != 0 {
				t.Fatal("clone retained source traversal")
			}
		}
		for _, o := range q.pCos {
			if o.TravId != 0 {
				t.Fatal("clone retained source traversal")
			}
		}
	}
	for _, o := range p.pCis {
		if o.TravId != 17 {
			t.Fatal("transform changed source traversal")
		}
	}
	for _, o := range p.pCos {
		if o.TravId != 17 {
			t.Fatal("transform changed source traversal")
		}
	}
}
