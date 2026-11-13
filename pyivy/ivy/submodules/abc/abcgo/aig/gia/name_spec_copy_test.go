package gia

import "testing"

func TestCopyNameSpec(t *testing.T) {
	src, dst := Gia_ManStart(1), Gia_ManStart(1)
	src.pName, src.pSpec = "source", "input.aig"
	Gia_ManCopyNameSpec(dst, src)
	if dst.pName != src.pName || dst.pSpec != src.pSpec {
		t.Fatalf("copied name=%q spec=%q", dst.pName, dst.pSpec)
	}
	src.pName = "changed"
	if dst.pName != "source" {
		t.Fatal("name changed with source")
	}
}
