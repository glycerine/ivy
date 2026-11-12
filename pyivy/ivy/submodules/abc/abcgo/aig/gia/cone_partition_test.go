package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSequentialConePartitions(t *testing.T) {
	p := Gia_ManStart(30)
	a, b, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	n0 := Gia_ManAppendAnd(p, a, r0)
	n1 := Gia_ManAppendAnd(p, b, r1)
	Gia_ManAppendCo(p, n0)
	Gia_ManAppendCo(p, n0)
	Gia_ManAppendCo(p, n1)
	Gia_ManAppendCo(p, r1)
	Gia_ManAppendCo(p, b)
	Gia_ManSetRegNum(p, 2)
	if Gia_ManConeMark(p, 0, 100) != 2 || Gia_ManConeMark(p, 2, 100) != 1 {
		t.Fatal("wrong sequential cone size")
	}
	outs := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(outs, 0)
	vec.Vec_IntPush(outs, 2)
	if Gia_ManCountFlops(p, outs) != 2 {
		t.Fatal("flops counted twice")
	}
	pivots := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(pivots, int32(a>>1))
	vec.Vec_IntPush(pivots, int32(b>>1))
	signs := Gia_ManDeriveSigns(p, pivots, 0)
	if got := vec.Vec_WrdEntry(signs, int(Gia_ManPo(p, 0).id)); got != 3 {
		t.Fatalf("sequential signature %b", got)
	}
	if got := vec.Vec_WrdEntry(signs, int(Gia_ManPo(p, 2).id)); got != 2 {
		t.Fatalf("second signature %b", got)
	}
	bins := Gia_ManHashOutputs(p, signs, 0)
	if vec.Vec_PtrSize(bins) != 2 || vec.Vec_IntSize(vec.Vec_PtrEntry(bins, 0).(*Vec_Int_t)) != 2 {
		t.Fatal("incorrect output partitions")
	}
	reduced := giaDupPartitionCones(p, []int32{2})
	if Gia_ManPiNum(reduced) != 1 || Gia_ManRegNum(reduced) != 1 || Gia_ManPoNum(reduced) != 1 {
		t.Fatalf("incorrect interface: pi=%d reg=%d po=%d", Gia_ManPiNum(reduced), Gia_ManRegNum(reduced), Gia_ManPoNum(reduced))
	}
	for bits := 0; bits < 4; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		got, e := reduced.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		want, e := p.Evaluate([]bool{false, input[0], false, input[1]})
		if e != nil || got[0] != want[2] || got[1] != want[4] {
			t.Fatalf("bad reduced outputs %v %v %v", got, want, e)
		}
	}
}

func TestOutputPartitionMerge(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, a)
	m := Opa_ManStart(p)
	for i := 0; i < vec.Vec_IntSize(m.vFront); i++ {
		obj := Gia_ManObj(p, vec.Vec_IntEntry(m.vFront, i))
		for _, fanin := range giaObjectFanins(p, obj) {
			Opa_ManMoveOne(m, obj, Gia_ManObj(p, fanin))
		}
	}
	if m.nParts != 2 {
		t.Fatalf("groups = %d", m.nParts)
	}
}

func TestFindLargestOutputPartition(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, b)
	var bins *Vec_Ptr_t
	out := Gia_ManFindPoPartition(p, 0, 1, 1, 0, &bins)
	if out == nil || Gia_ManPoNum(out) != 2 || Gia_ManPiNum(out) != 1 || vec.Vec_PtrSize(bins) != 2 {
		t.Fatal("incorrect largest partition")
	}
	for _, v := range []bool{false, true} {
		got, err := out.Evaluate([]bool{v})
		if err != nil || got[0] != v || got[1] == v {
			t.Fatalf("incorrect outputs %v %v", got, err)
		}
	}
}
