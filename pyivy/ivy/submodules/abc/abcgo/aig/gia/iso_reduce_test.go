package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestIsoReduceEquivalentAndDistinctOutputs(t *testing.T) {
	g := Gia_ManStart(18)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x, y, z := Gia_ManAppendAnd(g, a, b^1), Gia_ManAppendAnd(g, c, d^1), Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, z)
	var groups, perms *Vec_Ptr_t
	result := Gia_ManIsoReduce(g, &groups, &perms, 0, 0, 0, 0)
	if result == nil || Gia_ManPoNum(result) != 2 {
		t.Fatalf("unexpected reduced output count")
	}
	if groups == nil || vec.Vec_PtrSize(groups) != 2 {
		t.Fatalf("groups %v", groups)
	}
	first := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 0).(*Vec_Int_t))
	second := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 1).(*Vec_Int_t))
	if !reflect.DeepEqual(first, []int32{0, 1}) || !reflect.DeepEqual(second, []int32{2}) {
		t.Fatalf("groups %v %v", first, second)
	}
	if vec.Vec_PtrEntry(perms, 0) == nil || vec.Vec_PtrEntry(perms, 1) == nil {
		t.Fatal("missing input permutations")
	}
	for mask := 0; mask < 16; mask++ {
		ins := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		before, err := giaEvaluateNodes(g, ins)
		if err != nil {
			t.Fatal(err)
		}
		after, err := giaEvaluateNodes(result, ins)
		if err != nil {
			t.Fatal(err)
		}
		if before[Gia_ManPo(g, 0).id] != after[Gia_ManPo(result, 0).id] || before[Gia_ManPo(g, 2).id] != after[Gia_ManPo(result, 1).id] {
			t.Fatalf("assignment %d changed representatives", mask)
		}
	}
}

func TestIsoReduceDualOutputs(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, b)
	Gia_ManAppendCo(g, c)
	Gia_ManAppendCo(g, d)
	var groups *Vec_Ptr_t
	result := Gia_ManIsoReduce(g, &groups, nil, 0, 1, 0, 0)
	if result == nil || Gia_ManPoNum(result) != 2 {
		t.Fatalf("dual reduction failed")
	}
	if groups == nil || vec.Vec_PtrSize(groups) != 1 {
		t.Fatalf("dual groups %v", groups)
	}
	group := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 0).(*Vec_Int_t))
	if !reflect.DeepEqual(group, []int32{0, 1}) {
		t.Fatalf("dual group %v", group)
	}
	for mask := 0; mask < 16; mask++ {
		ins := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		before, err := giaEvaluateNodes(g, ins)
		if err != nil {
			t.Fatal(err)
		}
		after, err := giaEvaluateNodes(result, ins)
		if err != nil {
			t.Fatal(err)
		}
		for i := int32(0); i < 2; i++ {
			if before[Gia_ManPo(g, i).id] != after[Gia_ManPo(result, i).id] {
				t.Fatalf("assignment %d output %d", mask, i)
			}
		}
	}
}
