package abci

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestDeriveStatusArrays(t *testing.T) {
	if Abc_FrameDeriveStatusArray(nil) != nil || Abc_FrameDeriveStatusArray2(nil) != nil {
		t.Fatal("nil input")
	}
	cex := util.Abc_CexAlloc(0, 1, 1)
	entries := vec.Vec_PtrAllocArrayCopy([]any{nil, cex, uintptr(1)}, 3)
	if got := vec.Vec_IntArray(Abc_FrameDeriveStatusArray(entries)); !reflect.DeepEqual(got, []int32{-1, 0, 0}) {
		t.Fatal(got)
	}
	if got := vec.Vec_IntArray(Abc_FrameDeriveStatusArray2(entries)); !reflect.DeepEqual(got, []int32{-1, 0, 1}) {
		t.Fatal(got)
	}
	if vec.Vec_PtrEntry(entries, 2) != nil {
		t.Fatal("UNSAT sentinel was not consumed")
	}
}
