package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	"os"
	"path/filepath"
	"testing"
)

func TestParserManagerLifecycleAndConversion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.v")
	if err := os.WriteFile(path, []byte("module"), 0600); err != nil {
		t.Fatal(err)
	}
	library := abc.Abc_DesCreate("library")
	defer abc.Abc_DesFree(library, nil)
	p := Ver_ParseStart(path, library)
	if p == nil || p.pGateLib != library || p.pDesign == nil || p.vNames == nil {
		t.Fatal("manager initialization")
	}
	n := Ver_ParseFindOrCreateNetwork(p, "a")
	if Ver_ParseConvertNetwork(p, n, 0) != 1 || n.Function != uint8(abc.ABC_FUNC_AIG) || n.FunctionManager != p.pDesign.FunctionManager {
		t.Fatal("AIG conversion")
	}
	m := Ver_ParseFindOrCreateNetwork(p, "b")
	sentinel := new(int)
	p.pGenlib = sentinel
	if Ver_ParseConvertNetwork(p, m, 1) != 1 || m.Function != uint8(abc.ABC_FUNC_MAP) || m.FunctionManager != sentinel {
		t.Fatal("mapped conversion")
	}
	// The test sentinel is not a library manager, so do not ask abc to release it.
	m.FunctionManager = nil
	Ver_ParseFreeData(p)
	Ver_ParseStop(p)
	if Ver_ParseStart(filepath.Join(t.TempDir(), "missing.v"), nil) != nil {
		t.Fatal("missing input opened")
	}
}
