package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestStructuralReparameterizationOutputRange(t *testing.T) {
	for table := 0; table < 16; table++ {
		p := gia.Gia_ManStart(64)
		x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		gia.Gia_ManHashAlloc(p)
		result := int32(0)
		for mask := 0; mask < 4; mask++ {
			if table&(1<<mask) == 0 {
				continue
			}
			a, b := x, y
			if mask&1 == 0 {
				a ^= 1
			}
			if mask&2 == 0 {
				b ^= 1
			}
			result = gia.Gia_ManHashOr(p, result, gia.Gia_ManHashAnd(p, a, b))
		}
		gia.Gia_ManAppendCo(p, result)
		gia.Gia_ManHashStop(p)
		q := Abs_RpmPerform(p, 4, 0, 0)
		if q == nil || gia.Gia_ManPoNum(q) != 1 {
			t.Fatalf("table %04b lost output", table)
		}
		g := gia.Gia_ManToAigSimple(q)
		if len(g.Inputs()) > 10 {
			t.Fatalf("table %04b produced too many inputs", table)
		}
		seen := [2]bool{}
		for mask := 0; mask < 1<<len(g.Inputs()); mask++ {
			inputs := make([]bool, len(g.Inputs()))
			for i := range inputs {
				inputs[i] = mask&(1<<i) != 0
			}
			outputs, err := g.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if outputs[0] {
				seen[1] = true
			} else {
				seen[0] = true
			}
		}
		if seen[0] != (table != 15) || seen[1] != (table != 0) {
			t.Fatalf("table %04b output range %v", table, seen)
		}
	}
}

func TestSupportOrderingAndReferenceRestoration(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x, y)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManCreateRefs(p)
	gia.Gia_ObjSetMark1(gia.Gia_ManObj(p, x>>1), 1)
	gia.Gia_ObjSetMark1(gia.Gia_ManObj(p, y>>1), 1)
	root := gia.Gia_ManObj(p, gate>>1)
	if Abs_GiaObjDeref_rec(p, root) != 1 {
		t.Fatal("wrong dereferenced node count")
	}
	support := vec.Vec_IntAlloc(2)
	if Abs_ManSupport2(p, root, support) != 2 || Abs_GiaSortNodes(p, support) != 2 {
		t.Fatalf("wrong support %v", vec.Vec_IntArray(support))
	}
	if Abs_GiaObjRef_rec(p, root) != 1 || gia.Gia_ObjRefNum(p, gia.Gia_ManObj(p, x>>1)) != 1 || gia.Gia_ObjRefNum(p, gia.Gia_ManObj(p, y>>1)) != 1 {
		t.Fatal("reference restoration failed")
	}
}

func TestStructuralReparameterizationPreservesOutputPairs(t *testing.T) {
	for left := 0; left < 16; left++ {
		for right := 0; right < 16; right++ {
			p := gia.Gia_ManStart(64)
			x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
			gia.Gia_ManHashAlloc(p)
			for _, table := range []int{left, right} {
				out := int32(0)
				for mask := 0; mask < 4; mask++ {
					if table&(1<<mask) == 0 {
						continue
					}
					a, b := x, y
					if mask&1 == 0 {
						a ^= 1
					}
					if mask&2 == 0 {
						b ^= 1
					}
					out = gia.Gia_ManHashOr(p, out, gia.Gia_ManHashAnd(p, a, b))
				}
				gia.Gia_ManAppendCo(p, out)
			}
			gia.Gia_ManHashStop(p)
			q := Abs_RpmPerform(p, 4, 0, 0)
			g := gia.Gia_ManToAigSimple(q)
			if len(g.Inputs()) > 10 {
				t.Fatal("unexpectedly large input interface")
			}
			var want, got [4]bool
			for mask := 0; mask < 4; mask++ {
				index := (left >> mask & 1) | ((right >> mask & 1) << 1)
				want[index] = true
			}
			for mask := 0; mask < 1<<len(g.Inputs()); mask++ {
				inputs := make([]bool, len(g.Inputs()))
				for i := range inputs {
					inputs[i] = mask&(1<<i) != 0
				}
				outputs, err := g.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				index := 0
				if outputs[0] {
					index |= 1
				}
				if outputs[1] {
					index |= 2
				}
				got[index] = true
			}
			if got != want {
				t.Fatalf("outputs %04b,%04b: got %v want %v", left, right, got, want)
			}
		}
	}
}
