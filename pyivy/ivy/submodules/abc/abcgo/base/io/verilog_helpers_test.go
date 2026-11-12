package io

import (
	"bytes"
	"testing"
)

func TestVerilogNameEscaping(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"simple_123", "simple_123"}, {"0", "\\0 "}, {"1", "\\1 "}, {"net[3]", "\\net[3] "}, {"a.b", "\\a.b "}, {"27", "27"}} {
		if got := Io_WriteVerilogGetName(tc.in); got != tc.want {
			t.Errorf("name %q: got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWriteLutModule(t *testing.T) {
	var output bytes.Buffer
	Io_WriteLutModule(&output, 3)
	want := "module lut3 #( parameter TT = 8'h0 ) ( input [2:0] in, output out );\n    assign out = TT[in];\nendmodule\n\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}
