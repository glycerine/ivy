package util

import "testing"

func TestCexMakeTrivCombinational(t *testing.T) {
	p := Abc_CexMakeTriv(0, 2, 1, 0)
	if p.NRegs != 0 || p.NPis != 2 || p.NBits != 2 || p.IPo != 0 || p.IFrame != 0 {
		t.Fatalf("unexpected combinational witness: %+v", p)
	}
}

func TestCounterexampleBitsAndDup(t *testing.T) {
	p := Abc_CexCreate(2, 3, []int32{1, 0, 1, 0, 0, 0, 1, 1}, 1, 4, 0)
	if p.NBits != 8 || Abc_CexCountOnes(p) != 4 {
		t.Fatal(p)
	}
	q := Abc_CexDup(p, 4)
	if q.NBits != 10 || q.IPo != 4 || q.IFrame != 1 || Abc_InfoHasBit(q.PData, 0) || !Abc_InfoHasBit(q.PData, 4) || !Abc_InfoHasBit(q.PData, 8) {
		t.Fatal(q)
	}
	r := Abc_CexPermute(p, []int32{2, 1, 0})
	if !Abc_InfoHasBit(r.PData, 4) || !Abc_InfoHasBit(r.PData, 6) || !Abc_InfoHasBit(r.PData, 5) {
		t.Fatal(r)
	}
	Abc_CexFreeP(&q)
	if q != nil {
		t.Fatal("not freed")
	}
}
func TestCounterexampleUndc(t *testing.T) {
	p := Abc_CexCreate(0, 3, []int32{1, 0, 1, 0, 1, 0}, 1, 0, 0)
	q := Abc_CexTransformUndc(p, "1x")
	if q.NRegs != 2 || q.NPis != 2 || !Abc_InfoHasBit(q.PData, 0) || !Abc_InfoHasBit(q.PData, 1) || !Abc_InfoHasBit(q.PData, 2) {
		t.Fatal(q)
	}
}

func TestCounterexamplePhaseRegisterResize(t *testing.T) {
	p := Abc_CexCreate(2, 4, []int32{0, 0, 1, 0, 0, 1, 0, 1, 1, 0}, 1, 0, 0)
	q := Abc_CexTransformPhase(p, 2, 1, 4)
	if q.NBits != q.NRegs+(q.IFrame+1)*q.NPis {
		t.Fatalf("inconsistent phase dimensions: %+v", q)
	}
	if q.NBits != 12 || !Abc_InfoHasBit(q.PData, 4) || !Abc_InfoHasBit(q.PData, 7) {
		t.Fatal(q)
	}
}
func TestCounterexampleAllocationOverflow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("accepted wrapped dimensions")
		}
	}()
	_ = Abc_CexAlloc(0, 1<<30, 4)
}
