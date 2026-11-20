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

func TestAigerHashBeforeDanglingCleanup(t *testing.T) {
	// x = a & b, y = x & 0, z = !a & b. The outputs are !y and !z.
	// Simplifying y makes x dangling even though x had a fanout on disk.
	wire := "aig 5 2 0 2 3\n9\n11\n\x02\x02\x02\x06\x06\x01"
	for _, tc := range []struct {
		opts             AigerReadOptions
		ands             int32
		driver0, driver1 int32
	}{
		{AigerReadOptions{}, 1, 1, 7},
		{AigerReadOptions{Simple: true}, 2, 1, 9},
		{AigerReadOptions{SkipStrash: true}, 3, 9, 11},
	} {
		p, err := ParseAiger([]byte(wire), tc.opts)
		if err != nil {
			t.Fatal(err)
		}
		a, b := Gia_ManCo(p, 0), Gia_ManCo(p, 1)
		if Gia_ManAndNum(p) != tc.ands || Gia_ObjFaninLit0(a, a.id) != tc.driver0 || Gia_ObjFaninLit0(b, b.id) != tc.driver1 {
			t.Fatalf("mode %+v got %d gates and drivers %d, %d", tc.opts, Gia_ManAndNum(p), Gia_ObjFaninLit0(a, a.id), Gia_ObjFaninLit0(b, b.id))
		}
	}
	indexed := []byte(wire + "c" + "u\x00\x00\x00\x08" + "12345678")
	if _, err := ParseAiger(indexed, AigerReadOptions{}); err == nil {
		t.Fatal("accepted stale object-indexed switching data")
	}
	p, err := ParseAiger(indexed, AigerReadOptions{SkipStrash: true})
	if err != nil || string(p.pSwitching) != "12345678" {
		t.Fatal("lost switching data with original object IDs", err)
	}
}
