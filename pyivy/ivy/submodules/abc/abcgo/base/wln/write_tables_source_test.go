package wln

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTableOneSourceLayout(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), "table.v"))
	if err != nil {
		t.Fatal(err)
	}
	Wln_WriteTableOne(file, 2, 4, []uint64{0x3210}, 7)
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	const want = "module table7(ind, val);\n  input  [1:0] ind;\n  output [3:0] val;\n  reg    [3:0] val;\n  always @(ind)\n  begin\n    case (ind)\n      2'h0: val = 4'h0;\n      2'h1: val = 4'h1;\n      2'h2: val = 4'h2;\n      2'h3: val = 4'h3;\n    endcase\n  end\nendmodule\n\n"
	if string(got) != want {
		t.Fatalf("table output:\n%s\nwant:\n%s", got, want)
	}
}
