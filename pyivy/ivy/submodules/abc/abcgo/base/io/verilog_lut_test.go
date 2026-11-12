package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestVerilogLutTruthAndPinOrder(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 0)
	a, b := n.NewObject(abc.ObjectNet), n.NewObject(abc.ObjectNet)
	a.SetName("a")
	b.SetName("b")
	node := n.NewObject(abc.ObjectNode)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	node.Data = "10 1\n"
	out := n.NewObject(abc.ObjectNet)
	out.SetName("y")
	abc.Abc_ObjAddFanin(out, node)
	var buf bytes.Buffer
	Io_WriteVerilogObjectsLut(&buf, n, 3)
	if got := buf.String(); !strings.Contains(got, "lut3 #(8'h22) lut_0 ( {1'b0, b, a}, y );") {
		t.Fatalf("unexpected LUT: %q", got)
	}
}
