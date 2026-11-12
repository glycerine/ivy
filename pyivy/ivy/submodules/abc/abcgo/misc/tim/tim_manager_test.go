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
