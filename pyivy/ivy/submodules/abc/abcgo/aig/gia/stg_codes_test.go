package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestStateCodes(t *testing.T) {
	for hot := int32(1); hot <= 5; hot++ {
		var bits int32
		codes := Gia_ManAssignCodes(hot, 10, &bits)
		seen := map[string]bool{}
		for i := 0; i < 10; i++ {
			code := vec.Vec_IntArray(vec.Vec_VecEntryInt(codes, i))
			if len(code) != int(hot) {
				t.Fatal(code)
			}
			key := fmt.Sprint(code)
			if seen[key] {
				t.Fatal("duplicate code")
			}
			seen[key] = true
			for j, b := range code {
				if b >= bits || (j > 0 && code[j-1] >= b) {
					t.Fatal(code)
				}
			}
		}
	}
}
func TestStateOrGate(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashStart(p)
	lits := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(lits, a)
	vec.Vec_IntPush(lits, b)
	out := Gia_ManCreateOrGate(p, lits)
	Gia_ManHashStop(p)
	Gia_ManAppendCo(p, out)
	if byte(Gia_LutComputeTruth66_rec(p, Gia_ManCo(p, 0))) != 0xee {
		t.Fatal("OR truth mismatch")
	}
}
