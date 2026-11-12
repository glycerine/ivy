package gia

import "testing"

func TestJfInitRefsMuxAndXor(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for _, coarsen := range []bool{false, true} {
			p := Gia_ManStart(16)
			a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			var root int32
			if xor {
				root = Gia_ManAppendXor(p, a, b)
			} else {
				root = Gia_ManAppendMux(p, a, b, c)
			}
			Gia_ManAppendCo(p, root)
			obj := Gia_ManObj(p, root>>1)
			var d1, d0 *Gia_Obj_t
			ctrl := Gia_ObjRecognizeMux(obj, &d1, &d0)
			expected := make([]int32, p.nObjs)
			for _, o := range p.pObjs {
				if Gia_ObjIsAnd(o) != 0 {
					expected[Gia_ObjFanin0(o).id]++
					expected[Gia_ObjFanin1(o).id]++
				}
			}
			expected[obj.id]++
			expected[Gia_Regular(ctrl).id]--
			if Gia_Regular(d0) == Gia_Regular(d1) {
				expected[Gia_Regular(d0).id]--
			}
			flows, n := jfInitRefs(p, coarsen)
			for i, r := range expected {
				if p.Refs[i] != r {
					t.Fatalf("xor=%v ref[%d]=%d want %d", xor, i, p.Refs[i], r)
				}
				if r < 1 {
					r = 1
				}
				if flows[i] != float32(r) {
					t.Fatalf("flow %d", i)
				}
			}
			if coarsen {
				if n != 2 || Jf_ObjIsUnit(Gia_ObjFanin0(obj)) || Jf_ObjIsUnit(Gia_ObjFanin1(obj)) {
					t.Fatalf("coarsening=%d", n)
				}
			} else if n != 0 {
				t.Fatalf("unexpected coarsening %d", n)
			}
		}
	}
}
