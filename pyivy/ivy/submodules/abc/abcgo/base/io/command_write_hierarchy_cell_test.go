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

func TestWriteHierarchyAndCellNetCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
	pi := n.NewObject(abc.ObjectPI)
	pi.SetName("a")
	frame.ReplaceCurrentNetwork(n)
	dir := t.TempDir()
	cell := filepath.Join(dir, "out.cells")
	if IoCommandWriteCellNet(frame, []string{"write_cellnet", cell}) != 0 {
		t.Fatal("cellnet command failed")
	}
	if data, err := os.ReadFile(cell); err != nil || !strings.Contains(string(data), "CellNet") {
		t.Fatalf("cellnet %q: %v", data, err)
	}
	base := filepath.Join(dir, "base.blif")
	if err := os.WriteFile(base, []byte(".model base\n.inputs a\n.outputs y\n.names a y\n1 1\n.end\n"), 0600); err != nil {
		t.Fatal(err)
	}
	logic := abc.Abc_NtkCreateWithNode("1 1\n")
	frame.ReplaceCurrentNetwork(logic)
	output := filepath.Join(dir, "hierarchy.blif")
	if IoCommandWriteHie(frame, []string{"write_hie", base, output}) != 0 {
		t.Fatalf("hierarchy command failed: %s", errors.String())
	}
	data, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(data), ".model") {
		t.Fatalf("hierarchy output %q: %v; %s", data, err, errors.String())
	}
	if err := Io_WriteHie(logic, filepath.Join(dir, "missing.blif"), output); err == nil {
		t.Fatal("missing hierarchy base accepted")
	}
}
