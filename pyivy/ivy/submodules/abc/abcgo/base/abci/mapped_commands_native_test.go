package abci

import (
	"strings"
	"testing"
)

// A crash sweep of every command on a mapped network ("read mul8; strash;
// if -K 6; <cmd>") found these. Each must print what (fixed) native ABC
// prints.
const mappedPrep = "read test_vectors/bench/mul8.aig; strash; if -K 6; "

func checkMappedLikeNative(t *testing.T, cmd, want string) {
	t.Helper()
	script := mappedPrep + cmd + "; print_stats"
	native := runNativeABC(t, script)
	out, _ := runScriptStatus(t, script)
	if want != "" && !strings.Contains(native, want) {
		t.Errorf("%s: native does not print %q:\n%s", cmd, want, native)
	}
	if printSymmNormalize(native) != printSymmNormalize(out) {
		t.Errorf("%s: output differs\nnative:\n%s\nport:\n%s", cmd, native, out)
	}
}

// C bug 59: faultclasses without a fault list: Abc_NtkDetectFinClasses
// prints "Current network does not have the required info." and returns
// NULL, which Abc_NtkDetectClassesTest passed to Abc_NtkFinCountPairs
// (segfault).
func TestFaultClassesWithoutFinsLikeNative(t *testing.T) {
	checkMappedLikeNative(t, "faultclasses", "does not have the required info")
}

// C bug 60: clockgate and synch ran Abc_NtkToDar on any network; on a
// mapped one it reads each LUT as a two-input AND ("Unexpected 43 dangling
// nodes", a 25-node result of the wrong logic). They now ask for an AIG, as
// the other Dar-based commands do. The port panicked.
func TestClockGateSynchNeedAigLikeNative(t *testing.T) {
	for _, cmd := range []string{"clockgate", "synch"} {
		checkMappedLikeNative(t, cmd, "This command works only for AIGs")
	}
}

// dress reads the specification (mul8.aig) back as a strashed network;
// Abc_NtkDressComputeEquivs asserts !Abc_NtkIsStrash on it, which release
// ABC compiles out (strashing an AIG again is harmless). The port kept the
// assertion and panicked.
func TestDressOnMappedNetworkLikeNative(t *testing.T) {
	checkMappedLikeNative(t, "dress", "names can be moved")
}

// debug (Abc_NtkAutoDebug with Abc_NtkRetimeDebug) converts the network to
// SOPs before retiming and compares with Abc_NtkSecFraig. The port skipped
// the conversion (min-area retiming panicked on the mapped network) and used
// its own equivalence check.
func TestDebugOnMappedNetworkLikeNative(t *testing.T) {
	checkMappedLikeNative(t, "debug", "does not cause the bug")
}
