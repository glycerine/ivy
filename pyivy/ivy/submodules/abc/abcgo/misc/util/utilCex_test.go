package util

import "testing"

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
