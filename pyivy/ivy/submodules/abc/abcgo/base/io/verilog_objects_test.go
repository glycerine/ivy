package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
)

func TestVerilogAigObjects(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_AIG, 1)
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	a, b := n.NewObject(abc.ObjectNet), n.NewObject(abc.ObjectNet)
	a.SetName("a")
	b.SetName("b[0]")
	node := n.NewObject(abc.ObjectNode)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	node.Data = hop.Hop_And(manager, hop.Hop_IthVar(manager, 0), hop.Hop_Not(hop.Hop_IthVar(manager, 1)))
	out := n.NewObject(abc.ObjectNet)
	out.SetName("out")
	abc.Abc_ObjAddFanin(out, node)
	var buf bytes.Buffer
	Io_WriteVerilogObjects(&buf, n, 0)
	if got := buf.String(); got != "  assign out = (a & ~\\b[0] );\n" {
		t.Fatalf("unexpected assignment %q", got)
	}
	if hop.Hop_ObjData(hop.Hop_IthVar(manager, 0)) != nil {
		t.Fatal("temporary input name retained")
	}
}

func TestVerilogBoxEmission(t *testing.T) {
	model := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_AIG, 1)
	model.Name = "child"
	pi := model.NewObject(abc.ObjectPI)
	pin := model.NewObject(abc.ObjectNet)
	pin.SetName("in")
	abc.Abc_ObjAddFanin(pin, pi)
	po := model.NewObject(abc.ObjectPO)
	pout := model.NewObject(abc.ObjectNet)
	pout.SetName("out")
	abc.Abc_ObjAddFanin(po, pout)
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_AIG, 1)
	a := n.NewObject(abc.ObjectNet)
	a.SetName("source")
	bi := n.NewObject(abc.ObjectBI)
	abc.Abc_ObjAddFanin(bi, a)
	box := n.NewObject(abc.ObjectWhitebox)
	box.Data = model
	abc.Abc_ObjAddFanin(box, bi)
	bo := n.NewObject(abc.ObjectBO)
	abc.Abc_ObjAddFanin(bo, box)
	z := n.NewObject(abc.ObjectNet)
	z.SetName("dest")
	abc.Abc_ObjAddFanin(z, bo)
	var buf bytes.Buffer
	Io_WriteVerilogObjects(&buf, n, 0)
	if !strings.Contains(buf.String(), "  child box0(.in(source), .out(dest));\n") {
		t.Fatalf("unexpected box: %q", buf.String())
	}
}
