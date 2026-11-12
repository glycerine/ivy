package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfCutMemoryPages(t *testing.T) {
	reusable := vec.Vec_PtrAlloc(0)
	var m Lf_Mem_t
	Lf_MemAlloc(&m, 7, reusable, LF_CUT_WORDS)
	var handles []int32
	for i := int32(0); i < 40; i++ {
		c := Lf_Cut_t{nLeaves: 3, iFunc: 8}
		c.pLeaves[0], c.pLeaves[1], c.pLeaves[2] = 1, 7, 100+i
		handles = append(handles, Lf_MemSaveCut(&m, &c, 200+i))
	}
	if vec.Vec_PtrSize(&m.vPages) < 2 {
		t.Fatal("expected multiple pages")
	}
	for i, h := range handles {
		var c Lf_Cut_t
		Lf_MemLoadCut(&m, h, 200+int32(i), &c, 1, 1)
		if c.nLeaves != 3 || c.iFunc != 8 || c.pLeaves[2] != 100+int32(i) || c.Sign != Lf_CutGetSign(&c) {
			t.Fatalf("cut %d: %+v", i, c)
		}
	}
	if vec.Vec_PtrSize(reusable) == 0 {
		t.Fatal("old pages not recycled")
	}
	Lf_MemRecycle(&m)
	if m.iCur != 0 || vec.Vec_PtrSize(&m.vPages) != 0 {
		t.Fatal("recycle")
	}
	c := Lf_Cut_t{nLeaves: 1, iFunc: 2}
	c.pLeaves[0] = 4
	h := Lf_MemSaveCut(&m, &c, 8)
	var out Lf_Cut_t
	Lf_MemLoadCut(&m, h, 8, &out, 1, 0)
	if out.pLeaves[0] != 4 || out.Sign != 0 {
		t.Fatal("reused page", out)
	}
}
