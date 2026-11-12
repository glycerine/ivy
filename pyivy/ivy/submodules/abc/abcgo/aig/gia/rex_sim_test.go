package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRexTranspose64(t *testing.T) {
	var a [64]word
	a[3] = word(1) << (63 - 17)
	a[52] = word(1) << (63 - 7)
	Gia_ManAutomTranspose64(a[:])
	if a[17] != word(1)<<(63-3) || a[7] != word(1)<<(63-52) {
		t.Fatalf("unexpected transpose: %x %x", a[17], a[7])
	}
	Gia_ManAutomTranspose64(a[:])
	if a[3] != word(1)<<(63-17) || a[52] != word(1)<<(63-7) {
		t.Fatal("transpose is not involutive")
	}
}
func TestRexParallelStep(t *testing.T) {
	p := Gia_ManRex2Gia("(a|b)*c", 1, 0)
	temp := vec.Vec_WrdStart(int(p.nObjs))
	for _, state := range []word{0, word(1) << 63, word(1) << 62, (word(1) << 63) | (word(1) << 62)} {
		var next [64]word
		outputs := Gia_ManAutomStep(p, state, next[:], temp)
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			input := make([]bool, Gia_ManCiNum(p))
			input[i] = true
			for j := int32(0); j < Gia_ManRegNum(p); j++ {
				input[Gia_ManPiNum(p)+j] = state&(word(1)<<(63-j)) != 0
			}
			got, err := p.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			if (outputs&(word(1)<<(63-i)) != 0) != got[0] {
				t.Fatalf("state %x input %d output mismatch", state, i)
			}
			for j := int32(0); j < Gia_ManRegNum(p); j++ {
				if (next[i]&(word(1)<<(63-j)) != 0) != got[1+j] {
					t.Fatalf("state %x input %d register %d mismatch", state, i, j)
				}
			}
		}
	}
}
