package abci

import (
	"strings"
	"testing"
)

// print_mffc names unnamed nodes as Abc_ObjName does ("n16"; the port
// printed empty names), print_cone prints Saig_ManPrintCones' legend, and
// &ps -m prints Gia_ManPrintMuxStats' fields ("xor =%8d %6.2f %%   ").
func TestPrintCommandsLikeABC(t *testing.T) {
	for script, want := range map[string]string{
		"read test_vectors/native_s27_drw.blif; strash; print_mffc":  "Node =    n16 : Supp =   2  Cone =   1  ( n16 )",
		"read test_vectors/native_s27_drw.blif; strash; print_cone":  "The format of this print-out: For each PO, x:a b=c+d+e, where \n- x is the time-frame counting back from the PO\n",
		"read test_vectors/native_s27_drw.blif; strash; &get; &ps -m": "XOR/MUX stats:  xor =       0   0.00 %   mux =       0   0.00 %   and =       7 100.00 %   obj =       7",
	} {
		if out := runScript(t, script); !strings.Contains(out, want) {
			t.Errorf("%s: missing %q:\n%s", script, want, out)
		}
	}
}

func TestPrintCommandsLikeNative(t *testing.T) {
	for _, design := range []string{"test_vectors/native_s27_drw.blif", "test_vectors/seq/s713.blif"} {
		for _, command := range []string{"print_mffc", "print_cone", "&get; &ps -m"} {
			script := "read " + design + "; strash; " + command
			if native, port := bmc3Normalize(runNativeABC(t, script)), bmc3Normalize(runScript(t, script)); native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
		}
	}
}
