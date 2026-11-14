package gia

import "testing"

func TestCopyNameSpec(t *testing.T) {
	src, dst := Gia_ManStart(1), Gia_ManStart(1)
	src.pName, src.PSpec = "source", "input.aig"
	Gia_ManCopyNameSpec(dst, src)
	if dst.pName != src.pName || dst.PSpec != src.PSpec {
		t.Fatalf("copied name=%q spec=%q", dst.pName, dst.PSpec)
	}
	src.pName = "changed"
	if dst.pName != "source" {
		t.Fatal("name changed with source")
	}
}
