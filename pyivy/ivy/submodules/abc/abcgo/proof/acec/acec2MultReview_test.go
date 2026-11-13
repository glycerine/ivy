package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestFindInputsRequiresValidFront(t *testing.T) {
	cuts := vec.Vec_WecStart(1)
	for _, v := range []int32{1, 2, 3, 4, 5} {
		vec.Vec_WecPush(cuts, 0, v)
	}
	for _, front := range []int32{-1, 100} {
		got := Sdb_StoFindInputs(cuts, front)
		if vec.Vec_IntSize(got) != 0 {
			t.Fatalf("front %d: unexpected inferred inputs %v", front, vec.Vec_IntArray(got))
		}
	}
}

// acec2Mult.c retains the front even when its count was cleared by the first
// group, then collects every zero-count index into the second group.
func TestFindInputsSourceZeroCountFront(t *testing.T) {
	cuts := vec.Vec_WecStart(1)
	vec.Vec_WecPush(cuts, 0, 1)
	vec.Vec_WecPush(cuts, 0, 2)
	got := vec.Vec_IntArray(Sdb_StoFindInputs(cuts, 1))
	want := []int32{1, 2, 1, 0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("inputs=%v, want source behavior %v", got, want)
	}
}
