package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaLutStructuresPreserveTruth(t *testing.T) {
	tests := []struct {
		name, structure          string
		n                        int32
		check75, check44, filter int32
	}{
		{"unstructured", "", 5, 0, 0, 0},
		{"single", "44", 4, 0, 0, 0},
		{"two", "44", 6, 0, 0, 0},
		{"three", "444", 8, 0, 0, 0},
		{"special44", "44", 6, 0, 1, 0},
		{"special75", "", 6, 1, 0, 0},
		{"filter", "", 6, 0, 0, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pars := &If_Par_t_{NLutSize: 8, NCutsMax: 4, FLut6Filter: tc.filter}
			p := If_ManStart(pars)
			defer If_ManStop(p)
			dst := Gia_ManStart(64)
			leaves, temp, mapping := Vec_IntAlloc(int(tc.n)), Vec_IntAlloc(16), Vec_IntAlloc(0)
			records, packing := vec.Vec_IntAllocArrayCopy([]int32{0}, 1), vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
			for i := int32(0); i < tc.n; i++ {
				Vec_IntPush(leaves, Gia_ManAppendCi(dst))
			}
			truth := make([]word, (1<<uint(tc.n)+63)/64)
			for x := 0; x < 1<<uint(tc.n); x++ {
				var value bool
				if tc.n <= 6 {
					value = (x&7 == 7) != (x>>3 == (1<<uint(tc.n-3))-1)
				} else {
					value = (x&7 == 7) != ((x>>3)&7 == 7) != (x>>6 == 3)
				}
				if value {
					truth[x>>6] |= 1 << uint(x&63)
				}
			}
			if tc.n < 6 {
				period := 1 << uint(tc.n)
				for i := period; i < 64; i++ {
					truth[0] |= (truth[0] >> uint(i%period) & 1) << uint(i)
				}
			}
			lit := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, tc.structure, nil, mapping, records, packing, tc.check75, tc.check44)
			if lit < 0 {
				t.Fatal("decomposition failed")
			}
			for mask := 0; mask < 1<<uint(tc.n); mask++ {
				values := make([]bool, Gia_ManObjNum(dst))
				for i, input := range Vec_IntArray(leaves) {
					values[input>>1] = mask>>uint(i)&1 != 0
				}
				for id := int32(1); id < Gia_ManObjNum(dst); id++ {
					obj := Gia_ManObj(dst, id)
					if gia.Gia_ObjIsAnd(obj) == 0 {
						continue
					}
					a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
					b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
					values[id] = a && b
				}
				got := values[lit>>1] != (lit&1 != 0)
				want := truth[mask>>6]>>uint(mask&63)&1 != 0
				if got != want {
					t.Fatalf("assignment %d: got %v want %v", mask, got, want)
				}
			}
		})
	}
}
