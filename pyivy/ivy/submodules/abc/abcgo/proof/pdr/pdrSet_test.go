package pdr

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func testPdrSet(flops, inputs []int32) *Pdr_Set_t {
	return Pdr_SetCreate(vec.Vec_IntAllocArrayCopy(flops, len(flops)), vec.Vec_IntAllocArrayCopy(inputs, len(inputs)))
}

func TestPdrSetCreationAndOwnership(t *testing.T) {
	flops, inputs := []int32{6, 1, 4}, []int32{9, 8}
	set := testPdrSet(flops, inputs)
	if !reflect.DeepEqual(set.Lits, []int32{1, 4, 6, 9, 8}) || set.nLits != 3 || set.nTotal != 5 || set.nRefs != 1 {
		t.Fatalf("set = %+v", set)
	}
	flops[0], inputs[0] = 99, 99
	dup := Pdr_SetDup(set)
	dup.Lits[0] = 7
	if set.Lits[0] != 1 || set.Lits[3] != 9 {
		t.Fatal("cube data aliases its source")
	}
	removed := Pdr_SetCreateFrom(set, 1)
	if !reflect.DeepEqual(removed.Lits, []int32{1, 6, 9, 8}) || removed.nLits != 2 {
		t.Fatalf("removed = %+v", removed)
	}
	subset := Pdr_SetCreateSubset(set, []int32{6, 1}, 2)
	if !reflect.DeepEqual(subset.Lits, []int32{1, 6, 9, 8}) || subset.Sign != removed.Sign {
		t.Fatalf("subset = %+v", subset)
	}
	Pdr_SetRef(set)
	Pdr_SetDeref(set)
	if set.nRefs != 1 {
		t.Fatalf("references = %d", set.nRefs)
	}
}

func TestPdrSetContainmentAndInitialState(t *testing.T) {
	for mask := 1; mask < 256; mask++ {
		var oldLits []int32
		for i := 0; i < 8; i++ {
			if mask&(1<<i) != 0 {
				oldLits = append(oldLits, int32(i))
			}
		}
		old := testPdrSet(oldLits, nil)
		for sub := 1; sub < 256; sub++ {
			var subLits []int32
			for i := 0; i < 8; i++ {
				if sub&(1<<i) != 0 {
					subLits = append(subLits, int32(i))
				}
			}
			candidate := testPdrSet(subLits, nil)
			want := mask&sub == sub
			if (Pdr_SetContains(old, candidate) != 0) != want || (Pdr_SetContainsSimple(old, candidate) != 0) != want {
				t.Fatalf("old=%08b subset=%08b", mask, sub)
			}
		}
	}
	if Pdr_SetIsInit(testPdrSet([]int32{1, 3}, nil), -1) != 1 || Pdr_SetIsInit(testPdrSet([]int32{1, 2}, nil), -1) != 0 || Pdr_SetIsInit(testPdrSet([]int32{1, 2}, nil), 1) != 1 {
		t.Fatal("initial-state containment")
	}
}
