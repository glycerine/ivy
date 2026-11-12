package io

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestVerilogPortDeclarations(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 0)
	in := n.NewObject(abc.ObjectPI)
	inputNet := n.NewObject(abc.ObjectNet)
	inputNet.SetName("in[0]")
	abc.Abc_ObjAddFanin(inputNet, in)
	node := n.NewObject(abc.ObjectNode)
	abc.Abc_ObjAddFanin(node, inputNet)
	outNet := n.NewObject(abc.ObjectNet)
	outNet.SetName("out")
	abc.Abc_ObjAddFanin(outNet, node)
	out := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(out, outNet)
	feed := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(feed, inputNet)
	var buf bytes.Buffer
	Io_WriteVerilogPis(&buf, n, 3)
	if got := buf.String(); got != " \\in[0] " {
		t.Fatalf("inputs: %q", got)
	}
	buf.Reset()
	Io_WriteVerilogPos(&buf, n, 3)
	if got := buf.String(); got != " out" {
		t.Fatalf("outputs: %q", got)
	}
}

func TestVerilogNamesWrap(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 0)
	var nets []*abc.Object
	for _, name := range []string{"first_long_name", "second_long_name"} {
		net := n.NewObject(abc.ObjectNet)
		net.SetName(name)
		nets = append(nets, net)
	}
	var buf bytes.Buffer
	verilogNames(&buf, nets, 60)
	if got, want := buf.String(), " first_long_name,\n    second_long_name"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
