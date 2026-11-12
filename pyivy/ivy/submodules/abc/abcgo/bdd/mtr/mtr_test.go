package mtr

import (
	"io"
	"strings"
	"testing"
)

func valid(t *testing.T, n *MtrNode) {
	t.Helper()
	if err := WriteGroups(io.Discard, n); err != nil {
		t.Fatal(err)
	}
}
func TestGroupNestingDissolve(t *testing.T) {
	r := Mtr_InitGroupTree(0, 20)
	a := Mtr_MakeGroup(r, 2, 3, MTR_FIXED)
	b := Mtr_MakeGroup(r, 8, 2, MTR_SOFT)
	outer := Mtr_MakeGroup(r, 1, 12, MTR_DEFAULT)
	if outer.Child != a || a.Younger != b || b.Parent != outer || r.Child != outer {
		t.Fatal("contained groups not reparented")
	}
	if Mtr_MakeGroup(r, 4, 5, 0) != nil {
		t.Fatal("accepted crossing group")
	}
	if Mtr_FindGroup(r, 8, 2) != b || Mtr_FindGroup(r, 15, 1) != nil {
		t.Fatal("find mismatch")
	}
	if Mtr_MakeGroup(r, 8, 2, MTR_FIXED) != b || b.Flags != MTR_FIXED {
		t.Fatal("existing flags not updated")
	}
	valid(t, r)
	if Mtr_DissolveGroup(outer) != r || r.Child != a || b.Parent != r || a.Elder != nil {
		t.Fatal("dissolve links")
	}
	valid(t, r)
	if Mtr_DissolveGroup(r) != nil || Mtr_DissolveGroup(a) != nil {
		t.Fatal("invalid dissolution accepted")
	}
}
func TestInsertionGapsAndCopy(t *testing.T) {
	r := Mtr_InitGroupTree(0, 30)
	for _, low := range []uint32{20, 0, 10, 5, 15} {
		if Mtr_MakeGroup(r, low, 2, 0) == nil {
			t.Fatal(low)
		}
	}
	valid(t, r)
	copy := Mtr_CopyTree(r, 2)
	if copy == nil || copy == r || copy.Size != 60 {
		t.Fatal("copy failed")
	}
	valid(t, copy)
	c := copy.Child
	for n := r.Child; n != nil; n = n.Younger {
		if c == n || c.Low != 2*n.Low || c.Parent != copy {
			t.Fatal("copy links")
		}
		c = c.Younger
	}
	if Mtr_CopyTree(r, 0) != nil {
		t.Fatal("zero expansion accepted")
	}
	Mtr_FreeTree(copy)
	if copy.Child != nil || copy.Size != 0 {
		t.Fatal("free links")
	}
	valid(t, r)
}
func TestSwapDescendantsAndFailureAtomicity(t *testing.T) {
	r := Mtr_InitGroupTree(0, 12)
	a := Mtr_MakeGroup(r, 0, 4, 0)
	b := Mtr_MakeGroup(r, 4, 8, 0)
	inner := Mtr_MakeGroup(a, 1, 2, 0)
	if Mtr_SwapGroups(b, a) != 1 || r.Child != b || b.Low != 0 || a.Low != 8 || inner.Low != 9 {
		t.Fatal("swap did not shift descendants")
	}
	valid(t, r)
	unrelated := Mtr_InitGroupTree(0, 2)
	if Mtr_SwapGroups(a, unrelated) != 0 {
		t.Fatal("unrelated swap")
	}
	if mtrShiftHL(a, -20) != 0 || a.Low != 8 || inner.Low != 9 {
		t.Fatal("failed shift mutated tree")
	}
}
func TestReadPrintAndErrors(t *testing.T) {
	r, err := ReadGroups(strings.NewReader("0 10 D\n2 3 F\n6 2 S"), 10)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := WriteGroups(&out, r); err != nil {
		t.Fatal(err)
	}
	if out.String() != "(0(2,4|F)(6,7|S)9)\n" {
		t.Fatal(out.String())
	}
	for _, input := range []string{"2 3", "0 1 Z", "-1 2 D", "9 2 F", "1 4 F 3 4 S"} {
		if _, err := ReadGroups(strings.NewReader(input), 10); err == nil {
			t.Fatal("accepted", input)
		}
	}
	if Mtr_InitGroupTree(-1, 2) != nil || Mtr_InitGroupTree(0, 0) != nil {
		t.Fatal("bad bounds")
	}
}
func TestBasicSiblingOperations(t *testing.T) {
	r := Mtr_InitTree()
	a := Mtr_CreateFirstChild(r)
	b := Mtr_CreateLastChild(r)
	c := Mtr_AllocNode()
	Mtr_MakeNextSibling(a, c)
	if r.Child != a || a.Younger != c || c.Younger != b || b.Elder != c || c.Parent != r {
		t.Fatal("sibling chain")
	}
	MTR_SET(a, MTR_FIXED|MTR_SOFT)
	MTR_RESET(a, MTR_SOFT)
	if MTR_TEST(a, MTR_FIXED) != MTR_FIXED {
		t.Fatal("flag handling")
	}
}
