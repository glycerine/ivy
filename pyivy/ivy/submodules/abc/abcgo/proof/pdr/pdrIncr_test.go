package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestIncrementalSaveClausesCopiesCubes(t *testing.T) {
	p := &Pdr_Man_t{vClauses: vec.Vec_VecStart(3)}
	original := testPdrSet([]int32{2}, nil)
	vec.Vec_VecPush(p.vClauses, 1, original)
	vec.Vec_VecPush(p.vClauses, 2, testPdrSet([]int32{4}, nil))
	saved := IPdr_ManSaveClauses(p, 1)
	if vec.Vec_VecSize(saved) != 2 || vec.Vec_PtrSize(vec.Vec_VecEntry(saved, 1)) != 1 {
		t.Fatal("saved frames do not exclude the final frame")
	}
	copy := vec.Vec_PtrEntry(vec.Vec_VecEntry(saved, 1), 0).(*Pdr_Set_t)
	copy.Lits[0] = 9
	if original.Lits[0] != 2 {
		t.Fatal("saved cube aliases original")
	}
}

func TestIncrementalRestoreAbstractFlops(t *testing.T) {
	p := &Pdr_Man_t{vClauses: vec.Vec_VecStart(2), vAbsFlops: vec.Vec_IntStart(4)}
	vec.Vec_VecPush(p.vClauses, 1, testPdrSet([]int32{1, 6}, []int32{4}))
	IPdr_ManRestoreAbsFlops(p)
	for i, want := range []int32{1, 0, 0, 1} {
		if got := vec.Vec_IntEntry(p.vAbsFlops, i); got != want {
			t.Fatalf("flop %d: got %d, want %d", i, got, want)
		}
	}
}
