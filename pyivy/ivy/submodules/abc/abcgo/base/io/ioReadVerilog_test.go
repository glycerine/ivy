package io

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestReadVerilogSelectsTop(t *testing.T) {
	verilogParser.RLock()
	old := verilogParser.parse
	verilogParser.RUnlock()
	defer RegisterVerilogParser(old)
	RegisterVerilogParser(func(filename string, check int32) *abc.Design {
		if filename != "top.v" || check != 1 {
			t.Fatal("wrong parser arguments")
		}
		design := abc.NewDesign("test")
		root := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
		root.Name = "top"
		if err := design.AddModule(root); err != nil {
			t.Fatal(err)
		}
		return design
	})
	root := Io_ReadVerilog("top.v", 1)
	if root == nil || root.Name != "top" || root.Spec != "top.v" || root.Design != nil {
		t.Fatalf("unexpected root: %+v", root)
	}
	abc.Abc_NtkDelete(root)
}
