package gia

import (
	"reflect"
	"testing"
)

func TestJfCutAccountingRestoresReferences(t *testing.T) {
	// The root references node 3 and input 1; node 3 references inputs 1 and 2.
	cuts := map[int32][]int32{1: {Jf_CutSetAll(0, 0, 1), 2}, 2: {Jf_CutSetAll(0, 0, 1), 4}, 3: {Jf_CutSetAll(0, 2, 2), 2, 4}}
	root := []int32{Jf_CutSetAll(0, 1, 2), 6, 2}
	refs := []int32{0, 1, 0, 0}
	ctx := jfCutAccounting{refs: refs, best: func(i int32) []int32 { return cuts[i] }}
	before := append([]int32(nil), refs...)
	if area := ctx.area(root, false); area != 3 {
		t.Fatalf("area=%d", area)
	}
	if area := ctx.area(root, true); area != (3<<4)+4 {
		t.Fatalf("edge area=%d", area)
	}
	if !reflect.DeepEqual(refs, before) {
		t.Fatalf("references not restored: %v", refs)
	}
	if n := ctx.ref(root, true, false, false); n != 3 {
		t.Fatalf("ref=%d", n)
	}
	if n := ctx.ref(root, false, false, false); n != 3 {
		t.Fatalf("deref=%d", n)
	}
	if !reflect.DeepEqual(refs, before) {
		t.Fatalf("round trip: %v", refs)
	}
}
func TestJfCutMffcLimitAndRestoration(t *testing.T) {
	cuts := map[int32][]int32{1: {1, 2}, 2: {1, 4}, 3: {2, 2, 4}}
	root := []int32{2, 6, 2}
	refs := []int32{0, 2, 1, 1}
	ctx := jfCutAccounting{refs: refs, best: func(i int32) []int32 { return cuts[i] }}
	before := append([]int32(nil), refs...)
	if ctx.checkMffc(root, 3) {
		t.Fatal("expected limit to stop traversal")
	}
	if !reflect.DeepEqual(refs, before) {
		t.Fatal("early stop did not restore refs", refs)
	}
	if !ctx.checkMffc(root, 5) {
		t.Fatal("unexpected limit")
	}
	if !reflect.DeepEqual(refs, before) {
		t.Fatal("refs not restored", refs)
	}
}
