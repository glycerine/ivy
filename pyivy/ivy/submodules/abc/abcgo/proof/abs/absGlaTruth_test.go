package abs

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestGa2ISOPAllThreeInputFunctions(t *testing.T) {
	scratch := vec.Vec_IntAlloc(0)
	for truth := uint32(0); truth < 256; truth++ {
		cover := Ga2_ManCnfCompute(truth*0x01010101, 3, scratch)
		for assignment := 0; assignment < 8; assignment++ {
			got := false
			for _, raw := range vec.Vec_IntArray(cover) {
				cube := uint32(raw)
				matches := true
				for v := 0; v < 3; v++ {
					lit := (cube >> (2 * v)) & 3
					bit := assignment & (1 << v)
					if lit == 3 || lit == 1 && bit != 0 || lit == 2 && bit == 0 {
						matches = false
					}
				}
				if matches {
					got = true
					break
				}
			}
			want := truth&(1<<assignment) != 0
			if got != want {
				t.Fatalf("truth=%02x assignment=%d cover=%v", truth, assignment, vec.Vec_IntArray(cover))
			}
		}
		before := append([]int32(nil), vec.Vec_IntArray(cover)...)
		vec.Vec_IntClear(scratch)
		vec.Vec_IntPush(scratch, -1)
		if len(before) != vec.Vec_IntSize(cover) {
			t.Fatal("returned cover aliases scratch")
		}
		for i, value := range before {
			if vec.Vec_IntEntry(cover, i) != value {
				t.Fatal("returned cover aliases scratch")
			}
		}
	}
}

func TestGa2TruthDependenceAgainstAssignments(t *testing.T) {
	rng := rand.New(rand.NewSource(14876))
	for trial := 0; trial < 1000; trial++ {
		truth := rng.Uint32()
		for v := int32(0); v < 5; v++ {
			want := false
			for a := 0; a < 32; a++ {
				if ((truth >> a) & 1) != ((truth >> (a ^ (1 << uint(v)))) & 1) {
					want = true
					break
				}
			}
			if got := Ga2_ObjTruthDepends(truth, v) != 0; got != want {
				t.Fatalf("truth=%08x var=%d", truth, v)
			}
		}
	}
}
