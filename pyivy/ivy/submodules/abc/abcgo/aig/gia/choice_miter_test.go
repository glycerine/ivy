package gia

import "testing"

func TestChoiceMiterSharedInputsAndOutputOrder(t *testing.T) {
	a, b := Gia_ManStart(4), Gia_ManStart(4)
	x, y := Gia_ManAppendCi(a), Gia_ManAppendCi(b)
	Gia_ManAppendCo(a, x)
	Gia_ManAppendCo(a, x^1)
	Gia_ManAppendCo(b, y^1)
	Gia_ManAppendCo(b, y)
	m := Gia_ManChoiceMiter([]*Gia_Man_t{a, b})
	if Gia_ManCiNum(m) != 1 || Gia_ManCoNum(m) != 4 {
		t.Fatal("incorrect miter interface")
	}
	for _, input := range []bool{false, true} {
		got, err := m.Evaluate([]bool{input})
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{input, !input, !input, input}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("output %d: %v", i, got)
			}
		}
	}
}
