package acec

import "testing"

func TestDeclARemovesCarryLogicWithoutChangingOutputs(t *testing.T) {
	p := reviewAdderChain(-1)
	dst := Acec_ManDecla(p, 0, 0)
	for i := 0; i < 32; i++ {
		in := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0, i&16 != 0}
		want, e1 := p.Evaluate(in)
		got, e2 := dst.Evaluate(in)
		if e1 != nil || e2 != nil || len(want) != len(got) {
			t.Fatalf("input %v outputs %v %v errors %v %v", in, want, got, e1, e2)
		}
		for j := range want {
			if want[j] != got[j] {
				t.Fatalf("input %v want %v got %v", in, want, got)
			}
		}
	}
}

func TestDeclAPreservesComplementedOutputs(t *testing.T) {
	for invert := 0; invert < 3; invert++ {
		p := reviewAdderChain(invert)
		dst := Acec_ManDecla(p, 0, 0)
		for i := 0; i < 32; i++ {
			in := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0, i&16 != 0}
			want, e1 := p.Evaluate(in)
			got, e2 := dst.Evaluate(in)
			if e1 != nil || e2 != nil || len(want) != len(got) {
				t.Fatalf("out %d input %v results %v %v errors %v %v", invert, in, want, got, e1, e2)
			}
			for j := range want {
				if want[j] != got[j] {
					t.Fatalf("out %d input %v want %v got %v", invert, in, want, got)
				}
			}
		}
	}
}
