package io

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestReadBlifModes(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "input.blif")
	if err := os.WriteFile(filename, []byte(".model top\n.inputs a b\n.outputs y\n.names a b y\n11 1\n.end\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, options := range [][]string{nil, {"-nm"}, {"-a"}} {
		frame := abcmain.NewFrame()
		var errors bytes.Buffer
		frame.Err = &errors
		args := append([]string{"read_blif"}, options...)
		args = append(args, filename)
		if IoCommandReadBlif(frame, args) != 0 {
			t.Fatalf("%v: %s", options, errors.String())
		}
		n, ok := frame.Network().(*abc.Network)
		if !ok || abc.Abc_NtkPiNum(n) != 2 {
			t.Fatalf("%v: bad network", options)
		}
		if len(options) > 0 && options[0] == "-a" && abc.Abc_NtkIsStrash(n) == 0 {
			t.Fatal("AIG mode did not strash")
		}
	}
}

func TestReadVerilogCommand(t *testing.T) {
	verilogParser.RLock()
	old := verilogParser.parse
	verilogParser.RUnlock()
	defer RegisterVerilogParser(old)
	RegisterVerilogParser(func(name string, check int32) *abc.Design {
		if name != "input.v" || check != 0 {
			t.Fatal("wrong parser options")
		}
		d := abc.NewDesign("top")
		n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
		n.Name = "top"
		if err := d.AddModule(n); err != nil {
			t.Fatal(err)
		}
		return d
	})
	frame := abcmain.NewFrame()
	if IoCommandReadVerilog(frame, []string{"read_verilog", "-c", "input.v"}) != 0 || frame.Network() == nil {
		t.Fatal("Verilog read failed")
	}
}
