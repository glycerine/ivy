package abci

import (
	"strings"
	"testing"

	base_main "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mapper"
	"github.com/glycerine/abcgo/map/mio"
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

const mappedSeqPrep = "read test_vectors/seq/s27.blif; strash; if -K 4; "

func checkMappedSeqLikeNative(t *testing.T, cmd, want string) {
	t.Helper()
	script := mappedSeqPrep + cmd + "; print_stats"
	native := runNativeABC(t, script)
	out, _ := runScriptStatus(t, script)
	if want != "" && !strings.Contains(native, want) {
		t.Errorf("%s: native does not print %q:\n%s", cmd, want, native)
	}
	if printSymmNormalize(native) != printSymmNormalize(out) {
		t.Errorf("%s: output differs\nnative:\n%s\nport:\n%s", cmd, native, out)
	}
}

// C bug 61: print_cone (Abc_NtkDarPrintCone) ran Abc_NtkToDar on any
// network, reading a mapped network's LUTs as two-input ANDs, and printed
// cone statistics of that wrong AIG; the port panicked. As clockgate and
// synch (C bug 60), it now asks for an AIG.
func TestPrintConeNeedsAigLikeNative(t *testing.T) {
	checkMappedSeqLikeNative(t, "print_cone", "This command works only for AIGs")
}

// logicpush and mfse convert the network with Acb_NtkFromAbc, which asserts
// that every object is a CI, CO or node. Latch boxes are neither; release
// ABC compiles the assertion out and gives them no Acb object (Acb_NtkToAbc
// maps the result back by CI and CO order, so nothing reads them). The port
// kept the assertion and panicked on sequential networks.
func TestAcbCommandsOnSequentialNetworkLikeNative(t *testing.T) {
	for _, cmd := range []string{"logicpush", "mfse"} {
		checkMappedSeqLikeNative(t, cmd, "")
	}
}

// C bug 62: with a library that has no constant or buffer gates
// (amap_mux.genlib: AND2 INV XOR2 MUX2), phase_map ignored the NULL from
// Mio_CollectRootsNew2 (which reports "Error: Cannot find constant 0 gate in
// the library.") and read pCells[3]; mfs3 read the missing gates in
// Sfm_DecStart. Both segfaulted. They now stop after the message. The port
// ran mfs3 silently and printed its own phase_map error.
func TestLibraryWithoutConstantsLikeNative(t *testing.T) {
	for _, cmd := range []string{"phase_map", "mfs3"} {
		script := "read_genlib ../../map/amap/test_vectors/amap_mux.genlib; read test_vectors/bench/mul8.aig; strash; map; " + cmd + "; print_stats"
		native := runNativeABC(t, script)
		out, _ := runScriptStatusWith(t, script, func(f *base_main.Frame) { mapper.Map_Init(f); mio.Mio_Init(f) })
		if !strings.Contains(native, "Cannot find constant 0 gate in the library") {
			t.Errorf("%s: native does not report the missing gate:\n%s", cmd, native)
		}
		if printSymmNormalize(native) != printSymmNormalize(out) {
			t.Errorf("%s: output differs\nnative:\n%s\nport:\n%s", cmd, native, out)
		}
	}
}

// fretime on a mapped (SOP) network: fretInit.c converts the init-state
// network to BDDs (ABC_USE_CUDD) before Abc_NtkMiterSat; the port skipped
// the conversion and Abc_NtkMiterSatCreate panicked. The retimed network's
// statistics equal native's. (Native also prints a NetworkCheck message
// from fretMain.c's DEBUG_CHECK; that and fretime's verbose trace differ.)
func TestFretimeOnMappedNetworkLikeNative(t *testing.T) {
	script := mappedSeqPrep + "fretime; print_stats"
	native := runNativeABC(t, script)
	out, _ := runScriptStatus(t, script)
	last := func(s string) string {
		lines := strings.Split(strings.TrimSpace(s), "\n")
		return strings.TrimSpace(lines[len(lines)-1])
	}
	if last(native) != last(out) {
		t.Errorf("statistics differ\nnative: %s\nport:   %s", last(native), last(out))
	}
}
