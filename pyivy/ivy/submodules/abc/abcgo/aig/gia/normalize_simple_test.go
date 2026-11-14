package gia

import (
	"fmt"
	"testing"
)

// Simple AIGER input can retain x & !x as an actual two-pin AND. Moving CIs
// and COs must not reject that node or reinterpret it as a one-pin buffer.
func TestNormalizeSimpleRepeatedFanin(t *testing.T) {
	for hash := int32(0); hash < 2; hash++ {
		for signs := int32(0); signs < 4; signs++ {
			t.Run(fmt.Sprintf("hash%d_signs%d", hash, signs), func(t *testing.T) {
				p := Gia_ManStart(1)
				p.fGiaSimple = 1
				a := Gia_ManAppendCi(p)
				Gia_ManAppendCo(p, a)
				b := Gia_ManAppendCi(p)
				root := Gia_ManAppendAnd(p, a^(signs&1), a^((signs>>1)&1))
				Gia_ManAppendCo(p, root)
				Gia_ManAppendCo(p, b^1)
				if Gia_ManIsNormalized(p) != 0 {
					t.Fatal("fixture should need input/output reordering")
				}
				q := Gia_ManDupNormalize(p, hash)
				if Gia_ManIsNormalized(q) == 0 {
					t.Fatal("duplicate still has interleaved terminals")
				}
				for mask := 0; mask < 4; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0}
					got, err := q.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					want := []bool{inputs[0], (inputs[0] != (signs&1 != 0)) && (inputs[0] != (signs&2 != 0)), !inputs[1]}
					for i := range want {
						if got[i] != want[i] {
							t.Fatalf("assignment %d: output %d = %t, want %t", mask, i, got[i], want[i])
						}
					}
				}
			})
		}
	}
}
