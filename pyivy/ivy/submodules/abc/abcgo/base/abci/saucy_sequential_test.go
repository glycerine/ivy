package abci

import (
	"strings"

	"github.com/glycerine/abcgo/base/wlc"
	"testing"
)

const saucySequentialRefusal = "This command works only for combinational AIGs."

// TestNativeSaucyRefusesSequentialAig: saucy3 sized its arrays by the PIs
// and POs but computed supports over all CIs and COs, so on a sequential
// AIG native ABC crashed (CBUGS_FIXED.md #9).
func TestNativeSaucyRefusesSequentialAig(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/seq/s27.blif; strash; saucy3; print_stats")
	if !strings.Contains(out, saucySequentialRefusal) {
		t.Fatalf("native saucy3 on s27:\n%s", out)
	}
}

func TestSaucyRefusesSequentialAig(t *testing.T) {
	out, _ := runScriptStatus(t, "read test_vectors/seq/s27.blif; strash; saucy3")
	if !strings.Contains(out, saucySequentialRefusal) {
		t.Fatalf("saucy3 on s27:\n%s", out)
	}
}

const wlcTestNoDesign = "Abc_CommandTest(): There is no current design."

// TestNativeWlcTestWithoutDesign: %test explored the word-level design
// without checking that there is one (its check was commented out), so
// native ABC crashed when none had been read (CBUGS_FIXED.md #10).
func TestNativeWlcTestWithoutDesign(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/bench/mul8.aig; %test; print_stats")
	if !strings.Contains(out, wlcTestNoDesign) {
		t.Fatalf("native %%test without a design:\n%s", out)
	}
}

func TestWlcTestWithoutDesign(t *testing.T) {
	out, _ := runScriptStatusWith(t, "read test_vectors/bench/mul8.aig; %test", wlc.Wlc_Init)
	if !strings.Contains(out, wlcTestNoDesign) {
		t.Fatalf("%%test without a design:\n%s", out)
	}
}
