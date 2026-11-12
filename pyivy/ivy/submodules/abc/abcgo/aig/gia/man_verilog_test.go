package gia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGiaDumpVerilog(t *testing.T) {
	g := Gia_ManStart(8)
	g.pName = "my-circuit"
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x)
	path := filepath.Join(t.TempDir(), "graph.v")
	Gia_ManDumpVerilog(g, path, nil)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"module my_circuit", "input x0x, x1x;", "not( i2i, x1x )", "and( n3n, n1n, i2i )", "buf( z0z, n3n )"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	if Gia_ManNameIsLegalInVerilog("good_1$") != 1 || Gia_ManNameIsLegalInVerilog("bad-name") != 0 {
		t.Fatal("name legality")
	}
}
