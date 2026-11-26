package abci

import (
	"bytes"
	"fmt"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	abcio "github.com/glycerine/abcgo/base/io"
	base_main "github.com/glycerine/abcgo/base/main"
)

var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

// runScript executes an ABC command script on a fresh global frame and
// returns its output without color codes. A panic fails the test.
func runScript(t *testing.T, script string) string {
	t.Helper()
	previous := base_main.Abc_FrameReadGlobalFrame()
	frame := base_main.NewFrame()
	base_main.Abc_FrameSetGlobalFrame(frame)
	defer base_main.Abc_FrameSetGlobalFrame(previous)
	defer frame.Close()
	// Commands print to the process stdout, as in C; capture it too.
	var out bytes.Buffer
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	copied := make(chan struct{})
	go func() { io.Copy(&out, r); close(copied) }()
	frame.Out, frame.Err = w, w
	base_main.Abc_FrameInit(frame)
	abcio.Io_Init(frame)
	Abc_Init(frame)
	// As in the CLI: If_Init installs the default 4-LUT library, which
	// commands such as lutpack read.
	ifpkg.If_Init(frame)
	status, panicked := 0, any(nil)
	func() {
		defer func() { panicked = recover() }()
		status = int(frame.Execute(script))
	}()
	w.Close()
	<-copied
	os.Stdout = stdout
	if panicked != nil {
		t.Fatalf("script %q panicked: %v\n%s", script, panicked, out.String())
	}
	if status != 0 {
		t.Fatalf("script %q failed:\n%s", script, out.String())
	}
	return ansiEscape.ReplaceAllString(out.String(), "")
}

// expectStats checks the final print_stats line against native ABC's.
func expectStats(t *testing.T, script, want string) {
	t.Helper()
	got := runScript(t, script)
	if !regexp.MustCompile(regexp.QuoteMeta(want)).MatchString(got) {
		t.Fatalf("%s\n got: %s\nwant: %s", script, got, want)
	}
}

// Native: abc -c "read i10.aig; if -K 6; print_stats".
func TestIfMappingMatchesNative(t *testing.T) {
	expectStats(t, "read ../../../i10.aig; if -K 6; print_stats",
		fmt.Sprintf("nd = %5d  edge = %6d  aig  = %5d  lev = %d", 611, 2674, 4096, 11))
}

// Native: abc -c "read i10.aig; if -K 4 -a; print_stats".
func TestIfAreaMappingMatchesNative(t *testing.T) {
	expectStats(t, "read ../../../i10.aig; if -K 4 -a; print_stats",
		fmt.Sprintf("nd = %5d  edge = %6d  aig  = %5d  lev = %d", 841, 2816, 3343, 25))
}

// Native: abc -c "read mul8.aig; if -K 6; print_stats -p" reports
// "power = 231.24" (Abc_NtkMfsTotalSwitching); mfs -p depends on the same
// strash/AIG mapping in Abc_NtkPowerEstimate.
func TestTotalSwitchingMatchesNative(t *testing.T) {
	expectStats(t, "read test_vectors/bench/mul8.aig; if -K 6; print_stats -p", "power = 231.24")
}

// Native: abc -c "read i10.aig; drf; print_stats" (it used to panic in Go).
func TestDarRefactorMatchesNative(t *testing.T) {
	expectStats(t, "read ../../../i10.aig; drf; print_stats", "and =   2301  lev = 49")
}

// Native: abc -c "read i10.aig; drw; print_stats".
func TestDarRewriteMatchesNative(t *testing.T) {
	expectStats(t, "read ../../../i10.aig; drw; print_stats", "and =   2124  lev = 48")
}

// Native: abc -c "read i10.aig; dc2; print_stats".
func TestDc2MatchesNative(t *testing.T) {
	expectStats(t, "read ../../../i10.aig; dc2; print_stats", "and =   1834  lev = 36")
}

// 'mfs' used to panic when a node simplified to a constant.
func TestMfsConstantNodeDoesNotPanic(t *testing.T) {
	if out := runScript(t, "read test_vectors/bench/sort16.aig; if -K 6; mfs; print_stats"); !strings.Contains(out, "nd =") {
		t.Fatalf("no statistics after mfs:\n%s", out)
	}
}

// Native: abc -c "read sort16.aig; &get; &if -K 6; &mfs; &ps" reports
// "lut = 103  edge = 437  lev = 6"; Go's &mfs used to build a corrupt
// mapping (LUTs whose root was one of their own leaves) and &put panicked.
func TestGiaMfsMappingMatchesNative(t *testing.T) {
	out := runScript(t, "read test_vectors/bench/sort16.aig; &get; &if -K 6; &mfs; &ps")
	if !strings.Contains(out, "lut =    103  edge =     437  lev =    6") {
		t.Fatalf("mapping after &mfs differs from native:\n%s", out)
	}
	runScript(t, "read test_vectors/bench/sort16.aig; &get; &if -K 6; &mfs; &put; print_stats")
}

// Native: abc -c "read add32.aig; dch; print_stats". With no choices found,
// Go used to fail the command because the choice AIG had no equivalence
// table (Dch_DeriveChoiceAig always allocates pEquivs).
func TestDchWithoutChoicesMatchesNative(t *testing.T) {
	const script = "read test_vectors/bench/add32.aig; dch; print_stats"
	if out := runScript(t, script); strings.Contains(out, "Command has failed") {
		t.Fatalf("dch failed:\n%s", out)
	}
	expectStats(t, script, "and =    220  lev = 64")
}

// Native: abc -c "read i10.aig; dch; print_stats". print_stats reports
// Abc_AigLevel (stored levels) for strashed networks, not Abc_NtkLevel.
func TestDchMatchesNative(t *testing.T) {
	expectStats(t, "read ../../../i10.aig; dch; print_stats", "and =   4206 (choice = 602)  lev = 33")
}
