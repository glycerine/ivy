package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaCutConeChoiceAndCleanup(t *testing.T) {
	pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4}
	p := If_ManStart(pars)
	defer If_ManStop(p)
	a, b, c := If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, c)
	alternative := If_ManCreateAnd(p, a, If_Not(b))
	root.pEquiv = alternative
	root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}, dataInt: 97}
	a.CutBest.dataInt, b.CutBest.dataInt = 42, 43
	dst := Gia_ManStart(12)
	leaves := Vec_IntAlloc(2)
	Vec_IntPush(leaves, Gia_ManAppendCi(dst))
	Vec_IntPush(leaves, Gia_ManAppendCi(dst))
	for _, hash := range []int32{0, 1} {
		if hash != 0 {
			Gia_ManHashAlloc(dst)
		}
		lit := Gia_ManNodeIfToGia(dst, p, root, leaves, hash)
		if lit < 0 {
			t.Fatal("choice alternative was not found")
		}
		obj := Gia_ManObj(dst, lit>>1)
		for mask := 0; mask < 4; mask++ {
			values := make([]bool, Gia_ManObjNum(dst))
			for i, x := range Vec_IntArray(leaves) {
				values[x>>1] = mask>>uint(i)&1 != 0
			}
			for id := int32(1); id < Gia_ManObjNum(dst); id++ {
				node := Gia_ManObj(dst, id)
				if gia.Gia_ObjIsAnd(node) == 0 {
					continue
				}
				x := values[gia.Gia_ObjFaninId0(node, id)] != (gia.Gia_ObjFaninC0(node) != 0)
				y := values[gia.Gia_ObjFaninId1(node, id)] != (gia.Gia_ObjFaninC1(node) != 0)
				values[id] = x && y
			}
			got := values[gia.Gia_ObjId(dst, obj)] != (lit&1 != 0)
			want := mask&1 != 0 && mask&2 == 0
			if got != want {
				t.Fatalf("hash %d assignment %d: got %v want %v", hash, mask, got, want)
			}
		}
		if hash != 0 {
			Gia_ManHashStop(dst)
		}
	}
	if root.CutBest.dataInt != 97 || a.CutBest.dataInt != 42 || b.CutBest.dataInt != 43 {
		t.Fatal("cut scratch data changed")
	}
	root.pEquiv = nil
	if got := Gia_ManNodeIfToGia(dst, p, root, leaves, 0); got != -1 {
		t.Fatalf("missing cut leaf: got %d", got)
	}
	vec.Vec_IntWriteEntry(leaves, 0, 0)
	root.pEquiv = alternative
	if got := Gia_ManNodeIfToGia(dst, p, root, leaves, 0); got != 0 {
		t.Fatalf("constant leaf: got %d, want 0", got)
	}
}
