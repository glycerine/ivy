package ver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestParseModulePortsHeaderAndDeclarations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "module.v")
	if err := os.WriteFile(path, []byte("module sample (a, y);\ninput a;\noutput y;\nassign y = a;\nendmodule\n"), 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_ParseStart(path, nil)
	if p == nil {
		t.Fatal("could not open Verilog source")
	}
	defer Ver_ParseStop(p)
	p.pProgress = nil
	if got := Ver_ParseGetName(p); got == nil || *got != "module" {
		t.Fatalf("first token = %v, want module", got)
	}
	if Ver_ParseModule(p) != 1 {
		t.Fatalf("module parse failed: %s", p.sError)
	}
	n := p.pDesign.Modules[0]
	if got := abc.Abc_NtkPiNum(n); got != 1 {
		t.Errorf("PI count = %d, want 1", got)
	}
	if got := abc.Abc_NtkPoNum(n); got != 1 {
		t.Errorf("PO count = %d, want 1", got)
	}
}
