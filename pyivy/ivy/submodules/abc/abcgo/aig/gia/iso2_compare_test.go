package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestIso2ExactConeComparison(t *testing.T) {
	g := Gia_ManStart(14)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x, y, z := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, c, d), Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, z)
	Gia_Iso2ManPrepare(g)
	Gia_Iso2ManPropagate(g)
	classes := vec.Vec_WecStart(1)
	for i := int32(0); i < 3; i++ {
		vec.Vec_WecPush(classes, 0, i)
	}
	result := Gia_Iso2ManCheckIsoClasses(g, classes)
	if vec.Vec_WecSize(result) != 2 {
		t.Fatalf("classes %d", vec.Vec_WecSize(result))
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(result, 0)); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatalf("matched class %v", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(result, 1)); !reflect.DeepEqual(got, []int32{2}) {
		t.Fatalf("distinct class %v", got)
	}
	skip := Gia_Iso2ManCheckIsoClassesSkip(g, classes)
	if vec.Vec_WecSize(skip) != 3 {
		t.Fatalf("skip mode should separate all candidates: %d", vec.Vec_WecSize(skip))
	}
}

func TestIsoReduce2GroupsAndKeepsRepresentatives(t *testing.T) {
	for _, quality := range []int32{0, 1} {
		g := Gia_ManStart(14)
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x, y, z := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, c, d), Gia_ManAppendAnd(g, a, b^1)
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, y)
		Gia_ManAppendCo(g, z)
		var groups *Vec_Ptr_t
		result := Gia_ManIsoReduce2(g, &groups, nil, 0, quality, 0, 0, 0)
		if result == nil || Gia_ManPoNum(result) != 2 {
			t.Fatalf("quality %d: reduction failed", quality)
		}
		if vec.Vec_PtrSize(groups) != 2 {
			t.Fatalf("quality %d: groups %d", quality, vec.Vec_PtrSize(groups))
		}
		if got := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 0).(*Vec_Int_t)); !reflect.DeepEqual(got, []int32{0, 1}) {
			t.Fatalf("quality %d: group %v", quality, got)
		}
		for mask := 0; mask < 16; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			before, err := giaEvaluateNodes(g, in)
			if err != nil {
				t.Fatal(err)
			}
			after, err := giaEvaluateNodes(result, in)
			if err != nil {
				t.Fatal(err)
			}
			if before[Gia_ManPo(g, 0).id] != after[Gia_ManPo(result, 0).id] || before[Gia_ManPo(g, 2).id] != after[Gia_ManPo(result, 1).id] {
				t.Fatalf("quality %d assignment %d", quality, mask)
			}
		}
	}
}
