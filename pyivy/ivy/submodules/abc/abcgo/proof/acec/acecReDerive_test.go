package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDeriveFullAndHalfAdderTruthPhases(t *testing.T) {
	sums := []int32{0x66, 0x99, 0x69, 0x96}
	carries := []int32{0xe8, 0xd4, 0xb2, 0x71, 0x8e, 0x4d, 0x2b, 0x17, 0x88, 0x44, 0x22, 0x11, 0x77, 0xbb, 0xdd, 0xee}
	for _, x := range sums {
		for j, y := range carries {
			h := vec.Hash_IntManStart(1)
			key := vec.Hsh_Int3ManInsert(h, 1, 2, 3)
			data := vec.Vec_IntAlloc(6)
			for _, v := range []int32{7, key, x, 8, key, y} {
				vec.Vec_IntPush(data, v)
			}
			got := Ree_ManDeriveAdds(h, data, 0)
			want := int32(j % 8)
			if x == 0x99 || x == 0x69 {
				want |= 8
			}
			if vec.Vec_IntSize(got) != 6 || vec.Vec_IntEntry(got, 5) != want {
				t.Fatalf("sum %x carry %x: %v", x, y, vec.Vec_IntArray(got))
			}
			for i, v := range []int32{1, 2, 3, 7, 8} {
				if vec.Vec_IntEntry(got, i) != v {
					t.Fatalf("record %v", vec.Vec_IntArray(got))
				}
			}
		}
	}
}
