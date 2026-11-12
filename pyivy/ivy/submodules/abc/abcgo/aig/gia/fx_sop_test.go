package gia

import "testing"

func TestFxFactorSop(t *testing.T) {
	for _, tc := range []struct {
		sop  string
		want uint8
	}{
		{"11- 1\n-11 1\n", 0xc8},
		{"11- 0\n-11 0\n", 0x37},
		{"1-- 1\n", 0xaa},
		{"--- 1\n", 0xff},
	} {
		p := Gia_ManStart(40)
		leaves := Vec_IntAlloc(3)
		for i := 0; i < 3; i++ {
			Vec_IntPush(leaves, Gia_ManAppendCi(p))
		}
		Gia_ManHashStart(p)
		root := Gia_ManFactorNode(p, tc.sop, leaves)
		var truth uint8
		for assignment := 0; assignment < 8; assignment++ {
			values := make([]uint32, p.nObjs)
			for i := 0; i < 3; i++ {
				values[Vec_IntEntry(leaves, i)>>1] = uint32(assignment>>i) & 1
			}
			for _, obj := range p.pObjs {
				if Gia_ObjIsAnd(obj) != 0 {
					values[obj.id] = (values[Gia_ObjFaninId0(obj, obj.id)] ^ obj.fCompl0) & (values[Gia_ObjFaninId1(obj, obj.id)] ^ obj.fCompl1)
				}
			}
			truth |= uint8(values[root>>1]^uint32(root&1)) << assignment
		}
		if truth != tc.want {
			t.Fatalf("%q: got %02x want %02x", tc.sop, truth, tc.want)
		}
	}
}
