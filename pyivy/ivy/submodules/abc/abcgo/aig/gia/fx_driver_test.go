package gia

import "testing"

func TestPerformFxEndToEnd(t *testing.T) {
	for _, reverse := range []int32{0, 1} {
		p := Gia_ManStart(20)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		y := Gia_ManAppendAnd(p, x, c)
		z := Gia_ManAppendAnd(p, x, c^1)
		Gia_ManAppendCo(p, y)
		Gia_ManAppendCo(p, z^1)
		p.VMapping = Vec_IntStart(int(p.nObjs))
		for _, lut := range []struct {
			id  int32
			ins []int32
		}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}, {z >> 1, []int32{x >> 1, c >> 1}}} {
			Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(Vec_IntSize(p.VMapping)))
			Vec_IntPush(p.VMapping, int32(len(lut.ins)))
			for _, id := range lut.ins {
				Vec_IntPush(p.VMapping, id)
			}
			Vec_IntPush(p.VMapping, lut.id)
		}
		q := Gia_ManPerformFx(p, 5, 0, reverse, 0, 0)
		for input := 0; input < 8; input++ {
			values := make([]uint32, q.nObjs)
			for i := 0; i < 3; i++ {
				values[Gia_ManCi(q, int32(i)).id] = uint32(input>>i) & 1
			}
			for _, obj := range q.pObjs {
				if Gia_ObjIsAnd(obj) != 0 {
					values[obj.id] = (values[Gia_ObjFaninId0(obj, obj.id)] ^ obj.fCompl0) & (values[Gia_ObjFaninId1(obj, obj.id)] ^ obj.fCompl1)
				}
			}
			for i := int32(0); i < 2; i++ {
				obj := Gia_ManCo(q, i)
				got := values[Gia_ObjFaninId0(obj, obj.id)] ^ obj.fCompl0
				want := uint32(giaBool(input&3 == 3 && input&4 != 0))
				if i == 1 {
					want = uint32(giaBool(!(input&3 == 3 && input&4 == 0)))
				}
				if got != want {
					t.Fatalf("reverse=%d input=%d output=%d got=%d want=%d", reverse, input, i, got, want)
				}
			}
		}
	}
}
