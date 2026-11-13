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

// cecSynth.c advances the traversal ID before Gia_ManCollectNodesCis, which
// itself advances it again. Unvisited marks from the old epoch are therefore
// older than the immediately previous epoch.
func TestRegisterPartitionAdvancesTraversalTwice(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b, r := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManSetRegNum(g, 1)
	_ = r
	gia.Gia_ManIncrementTravId(g)
	unused := gia.Gia_ManObj(g, b>>1)
	gia.Gia_ObjSetTravIdCurrent(g, unused)
	part := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(part, 0)
	Gia_ManRegCreatePart(g, part, nil, nil, nil)
	if gia.Gia_ObjIsTravIdPrevious(g, unused) != 0 {
		t.Fatal("unvisited mark remains in previous epoch")
	}
}

// cecSynth.c duplicates every Gia_ObjIsAnd node with Gia_ManAppendAnd,
// including the GIA encoding of native XOR. This inherits C's loss of XOR.
func TestRegisterPartitionNativeXorSourceBehavior(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	part := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(part, 0)
	out := Gia_ManRegCreatePart(g, part, nil, nil, nil)
	co := gia.Gia_ManCo(out, 1)
	driver := gia.Gia_ObjFanin0(co)
	if gia.Gia_ObjIsXor(driver) != 0 {
		t.Fatal("C duplicates native XOR as AND")
	}
	got, err := out.Evaluate([]bool{true, false, false})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1] {
		t.Fatalf("source AND output: %v", got)
	}
}
