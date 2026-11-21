package gia

import "testing"

func TestAigerCleanupAfterOmittingInvariant(t *testing.T) {
	wire := []byte("aig 3 2 0 2 1\n2\n6\n\x02\x02o1 invariant:skip\n")
	for _, tc := range []struct{ simple, skip, ands int32 }{{0, 0, 0}, {1, 0, 1}, {0, 1, 1}} {
		p, err := Gia_AigerReadFromMemory(wire, int32(len(wire)), tc.simple, tc.skip, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got := Gia_ManAndNum(p); got != tc.ands || Gia_ManPoNum(p) != 1 {
			t.Errorf("simple %d skip %d gave %d ANDs, want %d", tc.simple, tc.skip, got, tc.ands)
		}
	}
}

func TestAigerUnusedUnknownRegisterNeedsNoMux(t *testing.T) {
	wire := []byte("aig 1 0 1 1 0\n0 2\n1\n")
	p, err := Gia_AigerReadFromMemory(wire, int32(len(wire)), 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if Gia_ManPiNum(p) != 1 || p.nRegs != 2 || Gia_ManAndNum(p) != 0 {
		t.Fatalf("unused X register yielded %d PIs, %d registers, %d ANDs", Gia_ManPiNum(p), p.nRegs, Gia_ManAndNum(p))
	}
}

func TestAigerTransformsKeepSkippedDuplicateGates(t *testing.T) {
	for _, wire := range []string{
		"aig 5 2 1 2 2\n0 1\n8\n10\n\x04\x02\x06\x02",
		"aig 4 2 0 2 2\n6\n8\n\x02\x02\x04\x02o1 constraint:enable\n",
	} {
		p, err := Gia_AigerReadFromMemory([]byte(wire), int32(len(wire)), 0, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		a, b := Gia_ManPo(p, 0), Gia_ManPo(p, 1)
		if Gia_ManAndNum(p) != 2 || Gia_ObjFanin0(a) == Gia_ObjFanin0(b) {
			t.Errorf("skip-strash transformation merged two ANDs, leaving %d", Gia_ManAndNum(p))
		}
	}
}

func TestDupWithConstraintsOmitsInvariantAndKeepsRIs(t *testing.T) {
	p := Gia_ManStart(8)
	a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, lit := range []int32{a, r, r, a} {
		Gia_ManAppendCo(p, lit)
	}
	Gia_ManSetRegNum(p, 1)
	q := Gia_ManDupWithConstraints(p, permutation(2, 1, 0))
	if q.nRegs != 1 || q.nConstrs != 1 || Gia_ManPoNum(q) != 2 {
		t.Fatal("port partition")
	}
	for i, want := range []int32{r, r ^ 1, a} {
		co := Gia_ManCo(q, int32(i))
		if Gia_ObjFaninLit0(co, co.id) != want {
			t.Fatalf("driver %d", i)
		}
	}
}

func TestDupZeroUndcClearsFanoutMarks(t *testing.T) {
	p := Gia_ManStart(8)
	r := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, r)
	Gia_ManAppendCo(p, 1)
	p.nRegs = 1
	for _, obj := range p.pObjs {
		obj.fMark0, obj.fMark1 = 1, 1
	}
	Gia_ManDupZeroUndc(p, "X", 0, 0, 0)
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 1 {
			t.Fatalf("marks on %d are %d/%d", obj.id, obj.fMark0, obj.fMark1)
		}
	}
}

func TestAigerMixedPropertiesKeepClassAndPhase(t *testing.T) {
	wire := []byte("aag 1 1 0 2 0 1 1\n2\n3\n2\n2\n3\no0 constraint:legacy\no1 normal\nb0 bad\nc0 assume\n")
	p, err := Gia_AigerReadFromMemory(wire, int32(len(wire)), 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if p.aigerBadOutputs != 1 || p.nConstrs != 2 {
		t.Fatalf("bad %d constraints %d", p.aigerBadOutputs, p.nConstrs)
	}
	for i, want := range []int32{2, 2, 2, 2} {
		co := Gia_ManPo(p, int32(i))
		if got := Gia_ObjFaninLit0(co, co.id); got != want {
			t.Fatalf("driver %d got %d want %d", i, got, want)
		}
	}
}

func BenchmarkDupZeroUndcChain(b *testing.B) {
	p := Gia_ManStart(4110)
	x := Gia_ManAppendCi(p)
	r := Gia_ManAppendCi(p)
	unused := Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	root := r
	for i := range 4096 {
		root = Gia_ManAppendAnd(p, x^int32(i&1), root)
	}
	for _, lit := range []int32{root, x, unused ^ 1, 0} {
		Gia_ManAppendCo(p, lit)
	}
	p.nRegs = 3
	b.ReportAllocs()
	for b.Loop() {
		q := Gia_ManDupZeroUndc(p, "X1X", 0, 0, 0)
		if Gia_ManAndNum(q) != 4099 || q.nRegs != 4 || Gia_ManPiNum(q) != 3 {
			b.Fatal("unexpected transformed circuit")
		}
		Gia_ManStop(q)
	}
}

func TestAigerSkippedOppositePinsSurviveResetConversion(t *testing.T) {
	wire := []byte("aig 3 1 1 1 1\n0 1\n6\n\x03\x01")
	q, err := Gia_AigerReadFromMemory(wire, int32(len(wire)), 0, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	driver := Gia_ObjFanin0(Gia_ManPo(q, 0))
	if Gia_ManAndNum(q) != 1 || Gia_ObjIsBuf(driver) != 0 || Gia_ObjFanin0(driver) != Gia_ObjFanin1(driver) || driver.fCompl0 == driver.fCompl1 {
		t.Fatal("opposite phases of the same PI should remain a raw AND")
	}
	for _, pi := range []bool{false, true} {
		if got := aigerOutputs(q, []bool{pi, false}); got[0] || !got[1] {
			t.Fatalf("transformed outputs %v", got)
		}
	}
}
