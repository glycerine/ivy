package wlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWlcNtkDumpDotUsesDesignNames(t *testing.T) {
	ntk := Wlc_NtkAlloc("/tmp/design.rev1.blif", 8)
	ntk.pSpec = "/tmp/source.v"
	piID := Wlc_ObjAlloc(ntk, int32(WLC_OBJ_PI), 0, 0, 0)
	bufID := Wlc_ObjAlloc(ntk, int32(WLC_OBJ_BUF), 0, 0, 0)
	fanins := Vec_IntStart(1)
	Vec_IntWriteEntry(fanins, 0, piID)
	Wlc_ObjAddFanins(ntk, Wlc_NtkObj(ntk, bufID), fanins)
	Wlc_ObjSetCo(ntk, Wlc_NtkObj(ntk, bufID), 0)
	Vec_IntFree(fanins)
	defer Wlc_NtkFree(ntk)

	path := filepath.Join(t.TempDir(), "graph.dot")
	Wlc_NtkDumpDot(ntk, path, nil)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `Benchmark \"design\" from file \"source\"`) {
		t.Fatalf("DOT title lacks C design names: %s", text)
	}
}

func TestWlcShowDotNodeFormatting(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "dot-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	obj := &Wlc_Obj_t{Type: uint32(WLC_OBJ_BUF), End: 3, Beg: 0}
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
	words := [4]int32{-1985229329, 0x01234567}
	if got, want := dotConstHex(&words[0], 64), "0123456789abcdef"; got != want {
		t.Fatalf("constant hex = %q, want %q", got, want)
	}
	if got, want := dotConstHex(&words[0], 68), "0123456789abcdef..."; got != want {
		t.Fatalf("wide constant hex = %q, want %q", got, want)
	}
}
