package gia

import "testing"

func BenchmarkAigerLoadI10(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		p, err := Gia_AigerRead("../../../i10.aig", 0, 0, 0)
		if err != nil {
			b.Fatal(err)
		}
		if Gia_ManAndNum(p) != 2675 || Gia_ManCiNum(p) != 257 || Gia_ManCoNum(p) != 224 {
			b.Fatal("unexpected i10 graph dimensions")
		}
		Gia_ManStop(p)
	}
}

func TestAigerVariableNumbering(t *testing.T) {
	for _, wire := range []string{
		"aig 3 2 0 1 1\n7\n\x01\x03",
		"aag 1024 2 0 1 1\n6\n100\n2049\n2048 101 6\n",
	} {
		p, err := parseAiger([]byte(wire), AigerReadOptions{SkipStrash: true}, 0)
		if err != nil {
			t.Fatal(err)
		}
		and, co := p.pObjs[3], Gia_ManCo(p, 0)
		if Gia_ObjFaninLit0(and, 3) != 2 || Gia_ObjFaninLit1(and, 3) != 5 || Gia_ObjFaninLit0(co, co.id) != 7 {
			t.Fatal("incorrect variable or polarity mapping")
		}
		Gia_ManStop(p)
	}
	for _, wire := range []string{"aig 1 1 0 1 0\n4\n", "aag 3 1 0 1 0\n6\n2\n"} {
		if _, err := parseAiger([]byte(wire), AigerReadOptions{}, 0); err == nil {
			t.Fatal("accepted undefined output variable")
		}
	}
}
