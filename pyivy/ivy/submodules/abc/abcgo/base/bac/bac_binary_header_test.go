package bac

import "testing"

func TestBacHeaderSourceCases(t *testing.T) {
	cases := []struct {
		line, name string
		nums       [4]int32
		ok         bool
	}{
		{"design 2 ", "design", [4]int32{2, 0, -1, -1}, true},
		{"net 1 2 4 3 ", "net", [4]int32{1, 2, 4, 3}, true},
		{"net 1", "net", [4]int32{1, -1, -1, -1}, false},
		{"net -1 2", "net", [4]int32{-1, 2, -1, -1}, true},
		{"net 0x 7", "net", [4]int32{0, 7, -1, -1}, true},
	}
	for _, tc := range cases {
		name, nums, ok := bacParseHeaderNumbers(tc.line)
		if name != tc.name || nums != tc.nums || ok != tc.ok {
			t.Fatalf("%q => %q %v %v", tc.line, name, nums, ok)
		}
	}
}
func TestBacHeaderLineBoundaries(t *testing.T) {
	data := []byte("#comment\ndesign 2 \n\x00\xff")
	first, pos, ok := bacReadHeaderLine(data, 0)
	if !ok || first != "#comment" {
		t.Fatal(first, pos, ok)
	}
	second, pos, ok := bacReadHeaderLine(data, pos)
	if !ok || second != "design 2 " || data[pos] != 0 {
		t.Fatal(second, pos, ok)
	}
	if _, _, ok := bacReadHeaderLine([]byte("unterminated"), 0); ok {
		t.Fatal("unterminated line accepted")
	}
}
