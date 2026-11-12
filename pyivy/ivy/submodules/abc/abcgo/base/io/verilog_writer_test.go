package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
)

func TestWriteVerilogAigModule(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_AIG, 1)
	n.Name = "top"
	pi := n.NewObject(abc.ObjectPI)
	pi.SetName("a")
	a := n.NewObject(abc.ObjectNet)
	a.SetName("a")
	abc.Abc_ObjAddFanin(a, pi)
	pi2 := n.NewObject(abc.ObjectPI)
	pi2.SetName("b")
	b := n.NewObject(abc.ObjectNet)
	b.SetName("b")
	abc.Abc_ObjAddFanin(b, pi2)
	node := n.NewObject(abc.ObjectNode)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	node.Data = hop.Hop_And(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1))
	y := n.NewObject(abc.ObjectNet)
	y.SetName("y")
	abc.Abc_ObjAddFanin(y, node)
	po := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po, y)
	var buf bytes.Buffer
	if err := WriteVerilog(&buf, n, 0); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"module top", "input  a, b;", "output y;", "assign y = (a & b);", "endmodule"} {
		if !strings.Contains(buf.String(), fragment) {
			t.Fatalf("missing %q in %q", fragment, buf.String())
		}
	}
}
