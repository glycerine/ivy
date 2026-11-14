package wlc

import (
	"os"
	"testing"
)

func TestWlcShowDotNodeFormatting(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "dot-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	obj := &Wlc_Obj_t{Type: WLC_OBJ_BUF, End: 3, Beg: 0}
	writeDotNode(f, nil, 7, obj, false)
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "  Node7 [label = \"7: 4\", shape = triangle];\n"; got != want {
		t.Fatalf("DOT node = %q, want %q", got, want)
	}
}

func TestWlcShowConstantHexFormatting(t *testing.T) {
	words := [2]int32{0x89abcdef, 0x01234567}
	if got, want := dotConstHex(&words[0], 64), "0123456789abcdef"; got != want {
		t.Fatalf("constant hex = %q, want %q", got, want)
	}
	if got, want := dotConstHex(&words[0], 68), "0123456789abcdef..."; got != want {
		t.Fatalf("wide constant hex = %q, want %q", got, want)
	}
}
