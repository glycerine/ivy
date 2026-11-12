package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestDupSupportAndOutputs(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, c)
	y := Gia_ManAppendAnd(p, x, b)
	Gia_ManAppendCo(p, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x)
	supports := Gia_ManCreateCoSupps(p, 0)
	want := [][]int32{{2}, {0, 1, 2}, {0, 2}}
	for i := range want {
		if got := vec.Vec_IntArray(vec.Vec_WecEntry(supports, i)); !reflect.DeepEqual(got, want[i]) {
			t.Fatalf("support %d: %v", i, got)
		}
	}
	if Gia_ManCoSuppSizeMax(p, supports) != 3 || Gia_ManCoLargestSupp(p, supports) != 1 || !reflect.DeepEqual(vec.Vec_IntArray(Gia_ManSortCoBySuppSize(p, supports)), []int32{1, 2, 0}) {
		t.Fatal("support ordering")
	}
	out := Gia_ManDupOuts(p)
	if Gia_ManPoNum(out) != 5 {
		t.Fatal("internal outputs missing")
	}
	got, e := out.Evaluate([]bool{true, true, false})
	if e != nil || got[0] || got[1] || got[2] || got[3] || got[4] {
		t.Fatalf("outputs %v %v", got, e)
	}
}
