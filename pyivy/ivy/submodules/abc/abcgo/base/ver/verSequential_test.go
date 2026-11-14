package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	"os"
	"path/filepath"
	"testing"
)

func TestAlwaysAndInitialSequentialStatements(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "always.v")
	if err := os.WriteFile(path, []byte("begin q <= d; r <= ~d; end"), 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_ParseStart(path, nil)
	n := Ver_ParseFindOrCreateNetwork(p, "module")
	Ver_ParseConvertNetwork(p, n, 0)
	for _, name := range []string{"q", "r", "d"} {
		abc.Abc_NtkFindOrCreateNet(n, name)
	}
	if Ver_ParseAlways(p, n) != 1 {
		t.Fatal("always parse")
	}
	q := abc.Abc_NtkFindNet(n, "q")
	r := abc.Abc_NtkFindNet(n, "r")
	lq := abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(q))
	lr := abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(r))
	if abc.Abc_LatchIsInitDc(lq) != 1 || abc.Abc_LatchIsInitDc(lr) != 1 {
		t.Fatal("always latch init")
	}
	Ver_StreamFree(p.pReader)
	initial := filepath.Join(dir, "initial.v")
	if err := os.WriteFile(initial, []byte("begin q = 1'b1; r = 0; end"), 0600); err != nil {
		t.Fatal(err)
	}
	p.pReader = Ver_StreamAlloc(initial)
	if Ver_ParseInitial(p, n) != 1 || abc.Abc_LatchIsInit1(lq) != 1 || abc.Abc_LatchIsInit0(lr) != 1 {
		t.Fatal("initial latch values")
	}
	Ver_ParseFreeData(p)
	Ver_ParseStop(p)
}
