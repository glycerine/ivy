package io

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestWriteVerilogLatches(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 0)
	for _, tc := range []struct {
		in, out string
		init    abc.Abc_InitType_t
	}{{"d[0]", "q[0]", abc.ABC_INIT_ZERO}, {"d1", "q1", abc.ABC_INIT_ONE}, {"d2", "q2", abc.ABC_INIT_DC}} {
		input := n.NewObject(abc.ObjectNet)
		input.SetName(tc.in)
		bi := n.NewObject(abc.ObjectBI)
		abc.Abc_ObjAddFanin(bi, input)
		latch := n.NewObject(abc.ObjectLatch)
		abc.Abc_ObjAddFanin(latch, bi)
		switch tc.init {
		case abc.ABC_INIT_ZERO:
			abc.Abc_LatchSetInit0(latch)
		case abc.ABC_INIT_ONE:
			abc.Abc_LatchSetInit1(latch)
		default:
			abc.Abc_LatchSetInitDc(latch)
		}
		bo := n.NewObject(abc.ObjectBO)
		abc.Abc_ObjAddFanin(bo, latch)
		output := n.NewObject(abc.ObjectNet)
		output.SetName(tc.out)
		abc.Abc_ObjAddFanin(output, bo)
	}
	var buf bytes.Buffer
	Io_WriteVerilogLatches(&buf, n)
	want := "  always @ (posedge clock) begin\n    \\q[0]  <= \\d[0] ;\n    q1 <= d1;\n    q2 <= d2;\n  end\n  initial begin\n    \\q[0]  <= 1'b0;\n    q1 <= 1'b1;\n  end\n"
	if got := buf.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
