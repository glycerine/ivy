package gia

import (
	"path/filepath"
	"testing"
)

func TestVerifyWithBoxesPlain(t *testing.T) {
	makeGraph := func(invert, seq bool) *Gia_Man_t {
		p := Gia_ManStart(8)
		a := Gia_ManAppendCi(p)
		if seq {
			r := Gia_ManAppendCi(p)
			lit := r
			if invert {
				lit ^= 1
			}
			Gia_ManAppendCo(p, lit)
			Gia_ManAppendCo(p, a)
			Gia_ManSetRegNum(p, 1)
		} else {
			if invert {
				a ^= 1
			}
			Gia_ManAppendCo(p, a)
		}
		return p
	}
	for _, seq := range []bool{false, true} {
		ref := makeGraph(false, seq)
		path := filepath.Join(t.TempDir(), "reference.aig")
		if err := Gia_AigerWrite(ref, path, 0, 0, 0); err != nil {
			t.Fatal(err)
		}
		mode := int32(0)
		if seq {
			mode = 1
		}
		if got := Gia_ManVerifyWithBoxes(makeGraph(false, seq), 1000, 1, mode, 0, 0, path); got != 1 {
			t.Fatalf("seq=%v equivalent %d", seq, got)
		}
		if got := Gia_ManVerifyWithBoxes(makeGraph(true, seq), 1000, 1, mode, 0, 0, path); got != 0 {
			t.Fatalf("seq=%v different %d", seq, got)
		}
	}
}
