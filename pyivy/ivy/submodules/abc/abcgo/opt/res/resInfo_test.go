package res

import "testing"

func TestSourceInfoOperations(t *testing.T) {
	p := []uint32{0xffff0000, 0x12345678, 123}
	q := []uint32{0x0000ffff, 0xedcba987, 456}
	if Abc_InfoIsOne(p, 2) != 0 || !Abc_InfoIsOrOne(p, q, 2) {
		t.Fatal("two-word source tests failed")
	}
	r := []uint32{0, 0, 0}
	if !Abc_InfoIsOrOne3(p, q, r, 2) {
		t.Fatal("three-way source test failed")
	}
	Abc_InfoOr(p, q, 2)
	if Abc_InfoIsOne(p, 2) != 1 || p[2] != 123 {
		t.Fatalf("word-limited OR %x", p)
	}
	Abc_InfoClear(p, 2)
	if p[0] != 0 || p[1] != 0 || p[2] != 123 {
		t.Fatalf("word-limited clear %x", p)
	}
	Abc_InfoFill(p, 2)
	if p[0] != ^uint32(0) || p[1] != ^uint32(0) || p[2] != 123 {
		t.Fatalf("word-limited fill %x", p)
	}
}
