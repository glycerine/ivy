package abci

import (
	"github.com/glycerine/abcgo/base/acb"
	"os"
	"path/filepath"
	"testing"
)

// Exercises the original command-based mapping flow, which requires ABCI.
func TestEcoRepairsSingleTarget(t *testing.T) {
	Abc_Init(Abc_FrameGetGlobalFrame())
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.v")
	good := filepath.Join(dir, "good.v")
	out := filepath.Join(dir, "patched.v")
	for name, source := range map[string]string{
		bad:  "module fixture(a,b,y);\ninput a,b;\noutput y;\nwire t_fix;\nbuf(y,t_fix);\nendmodule\n",
		good: "module fixture(a,b,y);\ninput a,b;\noutput y;\nwire n;\nand(n,a,b);\nbuf(y,n);\nendmodule\n",
	} {
		if err := os.WriteFile(name, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p := acb.Acb_VerilogSimpleRead(bad, "")
	q := acb.Acb_VerilogSimpleRead(good, "")
	if p == nil || q == nil {
		t.Fatal("read failed")
	}
	defer acb.Acb_ManFree(acb.Acb_NtkMan(p))
	defer acb.Acb_ManFree(acb.Acb_NtkMan(q))
	if result := acb.Acb_NtkEcoPerform(p, q, []string{bad, good, "", out}, 0, 0, 1, 1, 0, 0); result != 1 {
		t.Fatalf("ECO result %d", result)
	}
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("missing repaired design: %v", err)
	}
}
