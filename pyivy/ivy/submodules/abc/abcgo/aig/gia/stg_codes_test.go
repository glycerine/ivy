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

func TestStateTransitionOneHot(t *testing.T) {
	lines := vec.Vec_IntAlloc(16)
	for _, v := range []int32{0, 1, 1, 0, 1, 1, 2, 1, 0, 2, 1, 0, 1, 2, 2, 1} {
		vec.Vec_IntPush(lines, v)
	}
	p := Gia_ManStgOneHot(lines, 1, 1, 2)
	if Gia_ManPiNum(p) != 1 || Gia_ManPoNum(p) != 1 || Gia_ManRegNum(p) != 2 {
		t.Fatal("bad state interface")
	}
	state := []bool{false, false}
	for _, input := range []bool{true, false} {
		values := []bool{input, state[0], state[1]}
		out, err := p.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != input {
			t.Fatalf("input %v output %v", input, out)
		}
		copy(state, out[1:])
	}
}
