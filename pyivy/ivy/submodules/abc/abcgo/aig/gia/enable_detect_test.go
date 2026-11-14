package gia

import "testing"

func TestSequentialSignalDetection(t *testing.T) {
	p := Gia_ManStart(20)
	signal := Gia_ManAppendCi(p)
	r0 := Gia_ManAppendCi(p)
	r1 := Gia_ManAppendCi(p)
	a := Gia_ManAppendAnd(p, signal, r0)
	b := Gia_ManAppendAnd(p, signal, r1)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	Gia_ManSetRegNum(p, 2)
	found := Gia_ManDetectSeqSignalsWithFanout(p, 1, 0)
	if Vec_IntSize(found) != 1 || Vec_IntEntry(found, 0) != signal>>1 {
		t.Fatalf("signals: %v", Vec_IntArray(found))
	}
	Gia_ManDetectSeqSignals(p, 1, 0)
	if p.Refs != nil {
		t.Fatal("reference counts retained")
	}
}

func TestInitializedFrameUnroll(t *testing.T) {
	p := Gia_ManStart(10)
	input, reg := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, reg)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 1)
	q := Gia_ManUnrollInit(p, 3)
	if Gia_ManPiNum(q) != 3 || Gia_ManPoNum(q) != 3 || Gia_ManRegNum(q) != 0 {
		t.Fatal("incorrect unrolled interface")
	}
	values, err := giaEvaluateNodes(q, []bool{true, false, true})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{false, true, false} {
		co := Gia_ManCo(q, int32(i))
		got := values[Gia_ObjFaninId0(co, co.id)] != (Gia_ObjFaninC0(co) != 0)
		if got != want {
			t.Fatalf("frame %d = %v, want %v", i, got, want)
		}
	}
}

func TestRemoveEnableRetainsDataAndClass(t *testing.T) {
	p := Gia_ManStart(16)
	control, data, reg := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	next := Gia_ManAppendMux(p, control, data, reg)
	Gia_ManAppendCo(p, reg)
	Gia_ManAppendCo(p, next)
	Gia_ManSetRegNum(p, 1)
	q := Gia_ManRemoveEnables(p)
	if q.VFlopClasses == nil || Vec_IntSize(q.VFlopClasses) != 1 || Vec_IntEntry(q.VFlopClasses, 0) != 1 {
		t.Fatalf("flop classes: %v", q.VFlopClasses)
	}
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(inputs)
		if err != nil || out[0] != inputs[2] || out[1] != inputs[1] {
			t.Fatalf("inputs=%v outputs=%v error=%v", inputs, out, err)
		}
	}
}
