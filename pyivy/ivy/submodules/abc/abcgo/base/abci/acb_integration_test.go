package abci

import base_main "github.com/glycerine/abcgo/base/main"

import (
	"github.com/glycerine/abcgo/base/acb"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Exercises the original command-based mapping flow, which requires ABCI.
func TestEcoRepairsSingleTarget(t *testing.T) {
	previous := base_main.Abc_FrameReadGlobalFrame()
	frame := base_main.NewFrame()
	base_main.Abc_FrameSetGlobalFrame(frame)
	defer base_main.Abc_FrameSetGlobalFrame(previous)
	defer frame.Close()
	base_main.Abc_FrameInit(frame)
	Abc_Init(frame)
	// The C synthesis scripts use aliases installed by the standard abc.rc.
	if frame.Execute("source ../../../abc.rc") != 0 {
		t.Fatal("loading standard ABC aliases failed")
	}
	// The source patch mapper requires a standard-cell library.
	library, err := mio.ReadLibrary("eco", `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0
GATE nand2 2 O=!(a*b); PIN * INV 1 999 1 0 1 0
GATE or2 2 O=a+b; PIN * NONINV 1 999 1 0 1 0
GATE nor2 2 O=!(a+b); PIN * INV 1 999 1 0 1 0
`, nil)
	if err != nil {
		t.Fatal(err)
	}
	mio.Mio_UpdateGenlib(library)
	ifpkg.Abc_FrameSetLibLut(ifpkg.If_LibLutSetSimple(6))
	previousPrint := util.PrintBridge
	var diagnostics strings.Builder
	util.PrintBridge = func(text string) { diagnostics.WriteString(text) }
	defer func() { util.PrintBridge = previousPrint }()
	t.Cleanup(func() {
		if strings.Contains(diagnostics.String(), "Error:") {
			t.Errorf("ECO emitted errors: %s", diagnostics.String())
		}
	})
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
