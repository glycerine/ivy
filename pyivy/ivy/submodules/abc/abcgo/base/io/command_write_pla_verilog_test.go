package io

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWritePlaAndVerilogCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	frame.ReplaceCurrentNetwork(abc.Abc_NtkCreateWithNode("11 1\n"))
	dir := t.TempDir()
	for _, tc := range []struct {
		command  func(*abcmain.Frame, []string) int32
		args     []string
		fragment string
	}{
		{IoCommandWritePla, []string{"write_pla", filepath.Join(dir, "out.pla")}, ".i 2"},
		{IoCommandWriteVerilog, []string{"write_verilog", "-K", "3", filepath.Join(dir, "lut.v")}, "lut3 #(8'h88)"},
		{IoCommandWriteVerilog, []string{"write_verilog", "-a", filepath.Join(dir, "aig.v")}, "assign"},
	} {
		if tc.command(frame, tc.args) != 0 {
			t.Fatalf("%v: %s", tc.args, errors.String())
		}
		data, err := os.ReadFile(tc.args[len(tc.args)-1])
		if err != nil || !strings.Contains(string(data), tc.fragment) {
			t.Fatalf("%v: %q, %v; errors %s", tc.args, data, err, errors.String())
		}
	}
	if IoCommandWriteVerilog(frame, []string{"write_verilog", "-K", "0", "bad.v"}) != 1 {
		t.Fatal("accepted invalid LUT size")
	}
}
