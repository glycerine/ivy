package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRegisterPartitionPreservesConeAndMapping(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	first := gia.Gia_ManAppendAnd(g, a, c)
	second := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, first)
	gia.Gia_ManAppendCo(g, second)
	gia.Gia_ManSetRegNum(g, 2)
	part := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(part, 0)
	var pis, regs int32
	var mapping []int32
	out := Gia_ManRegCreatePart(g, part, &pis, &regs, &mapping)
	if pis != 1 || regs != 1 || gia.Gia_ManRegNum(out) != 1 {
		t.Fatalf("counts: PIs=%d other registers=%d result registers=%d", pis, regs, gia.Gia_ManRegNum(out))
	}
	if mapping[1] != a>>1 || mapping[2] != c>>1 || mapping[3] != b>>1 || mapping[5] != first>>1 {
		t.Fatalf("mapping %v", mapping)
	}
	for mask := 0; mask < 8; mask++ {
		a, c, b := mask&1 != 0, mask&2 != 0, mask&4 != 0
		got, err := out.Evaluate([]bool{a, c, b})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 || got[0] != b || got[1] != (a && c) {
			t.Fatalf("inputs %v %v %v got %v", a, c, b, got)
		}
	}
}
