package abci

import (
	"strings"
	"testing"
)

const declaNoMatch = "Cannot match the outputs of the arithmetic box."

// TestNativeDeclaWhenBoxMissesOutputs: on mul8 the adder box's root ranks
// cover fewer bits than the AIG has outputs, so Acec_RewriteTop's result is
// short and Acec_RewriteReplace read past it, crashing native ABC
// (CBUGS_FIXED.md #12). Now &decla reports it and keeps the AIG.
func TestNativeDeclaWhenBoxMissesOutputs(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/bench/mul8.aig; strash; &get; &decla; &ps")
	if !strings.Contains(out, declaNoMatch) || !strings.Contains(out, "and =     424") {
		t.Fatalf("native &decla on mul8:\n%s", out)
	}
}

func TestDeclaWhenBoxMissesOutputs(t *testing.T) {
	out := runScript(t, "read test_vectors/bench/mul8.aig; strash; &get; &decla; &ps")
	if !strings.Contains(out, declaNoMatch) || !strings.Contains(out, "and =     424") {
		t.Fatalf("&decla on mul8:\n%s", out)
	}
}

// On i10 the adder box has empty root ranks, and Acec_RewriteTop read the
// first entry of one (CBUGS_FIXED.md #12).
func TestNativeDeclaWithEmptyRootRanks(t *testing.T) {
	out := runNativeABC(t, "read ../../../i10.aig; strash; &get; &decla; &ps")
	if !strings.Contains(out, declaNoMatch) || !strings.Contains(out, "and =    2675") {
		t.Fatalf("native &decla on i10:\n%s", out)
	}
}

func TestDeclaWithEmptyRootRanks(t *testing.T) {
	out := runScript(t, "read ../../../i10.aig; strash; &get; &decla; &ps")
	if !strings.Contains(out, declaNoMatch) || !strings.Contains(out, "and =    2675") {
		t.Fatalf("&decla on i10:\n%s", out)
	}
}
