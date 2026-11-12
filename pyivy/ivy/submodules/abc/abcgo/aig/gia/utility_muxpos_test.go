package gia

import "testing"

func TestDupWithMuxOutputs(t *testing.T) {
	p := Gia_ManStart(40)
	a, b, c, d, e := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	inner := Gia_ManAppendMux(p, a, b, c)
	outer := Gia_ManAppendMux(p, d, inner, e)
	Gia_ManAppendCo(p, outer)
	q := Gia_ManDupWithMuxPos(p)
	if Gia_ManCoNum(q) <= 1 {
		t.Fatal("expected a shared MUX boundary output")
	}
	for mask := 0; mask < 32; mask++ {
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		before, e1 := p.Evaluate(inputs)
		after, e2 := q.Evaluate(inputs)
		if e1 != nil || e2 != nil || before[0] != after[0] {
			t.Fatalf("inputs=%v before=%v after=%v", inputs, before, after)
		}
	}
}
