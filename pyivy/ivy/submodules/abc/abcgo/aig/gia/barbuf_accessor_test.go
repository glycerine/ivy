package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaManBarBufsAccessor(t *testing.T) {
	p := Gia_ManStart(4)
	v := Gia_ManBarBufs(p)
	vec.Vec_IntPush(v, 14)
	if Gia_ManBarBufs(p) != v || vec.Vec_IntSize(Gia_ManBarBufs(p)) != 1 || vec.Vec_IntEntry(v, 0) != 14 {
		t.Fatal("barrier-buffer accessor did not retain the manager vector")
	}
	Gia_ManSetNameSpec(p, "n", "s")
	if Gia_ManName(p) != "n" || Gia_ManSpec(p) != "s" {
		t.Fatal("manager metadata accessors did not round-trip")
	}
}
