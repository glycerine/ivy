package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
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
