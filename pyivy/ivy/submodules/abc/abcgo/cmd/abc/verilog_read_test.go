package main

import (
	"os"
	"path/filepath"
	"testing"

	abcio "github.com/glycerine/abcgo/base/io"
)

// TestBinaryReadsVerilog: the abc binary must register the Verilog parser
// (base/ver), or "read x.v" panics where ABC reads the file.
func TestBinaryReadsVerilog(t *testing.T) {
	file := filepath.Join(t.TempDir(), "and.v")
	src := "module top ( a, b, y );\n  input a, b;\n  output y;\n  assign y = a & b;\nendmodule\n"
	if err := os.WriteFile(file, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	n := abcio.Io_ReadVerilog(file, 1)
	if n == nil {
		t.Fatal("Verilog file not read")
	}
}
