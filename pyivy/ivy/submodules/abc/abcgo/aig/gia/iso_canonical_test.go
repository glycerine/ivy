package gia

import (
	"bytes"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestIsoCanonicalStringMatchesRenamedCone(t *testing.T) {
	g := Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x, y := Gia_ManAppendAnd(g, a, b^1), Gia_ManAppendAnd(g, c, d^1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	var perm0, perm1 *Vec_Int_t
	s0, s1 := Gia_ManIsoFindString(g, 0, 0, &perm0), Gia_ManIsoFindString(g, 1, 0, &perm1)
	if s0 == nil || s1 == nil {
		t.Fatal("canonical AIGER encoding failed")
	}
	if !bytes.Equal(vec.Vec_StrArray(s0), vec.Vec_StrArray(s1)) {
		t.Fatalf("canonical strings differ: %q %q", vec.Vec_StrArray(s0), vec.Vec_StrArray(s1))
	}
	if vec.Vec_IntSize(perm0) != 2 || vec.Vec_IntSize(perm1) != 2 {
		t.Fatalf("bad input permutations %v %v", vec.Vec_IntArray(perm0), vec.Vec_IntArray(perm1))
	}
}

func TestIsoCanonicalizationPreservesTruthAndConstant(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x^1)
	result := Gia_ManIsoCanonicize(g, 0)
	if Gia_ManCiNum(result) != 2 || Gia_ManPoNum(result) != 1 {
		t.Fatal("canonical interface changed")
	}
	var perm *Vec_Int_t
	encoded := Gia_ManIsoFindString(g, 0, 0, &perm)
	if encoded == nil || vec.Vec_IntSize(perm) != 2 {
		t.Fatal("missing canonical input permutation")
	}
	permutation := vec.Vec_IntArray(perm)
	for mask := 0; mask < 4; mask++ {
		original := []bool{mask&1 != 0, mask&2 != 0}
		reordered := []bool{original[permutation[0]], original[permutation[1]]}
		before, err := giaEvaluateNodes(g, original)
		if err != nil {
			t.Fatal(err)
		}
		after, err := giaEvaluateNodes(result, reordered)
		if err != nil {
			t.Fatal(err)
		}
		if before[Gia_ManPo(g, 0).id] != after[Gia_ManPo(result, 0).id] {
			t.Fatalf("assignment %d", mask)
		}
	}
	constant := Gia_ManStart(2)
	Gia_ManAppendCo(constant, 1)
	canon := Gia_ManIsoCanonicize(constant, 0)
	if Gia_ManCiNum(canon) != 0 || Gia_ObjFaninC0(Gia_ManPo(canon, 0)) != 1 {
		t.Fatal("constant one changed")
	}
	var empty *Vec_Int_t
	if Gia_ManIsoFindString(constant, 0, 0, &empty) == nil || vec.Vec_IntSize(empty) != 0 {
		t.Fatal("constant encoding failed")
	}
}
