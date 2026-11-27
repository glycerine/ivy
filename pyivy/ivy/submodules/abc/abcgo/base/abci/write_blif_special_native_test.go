package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// write_blif -j/-S/-a (Io_WriteBlifSpecial) was not registered: the port
// printed "io: specialized BLIF writer is not registered".

const blifSpecialInput = "read test_vectors/native_mul8_if6.blif; strash; "

var blifSpecialMessage = regexp.MustCompile(`(?m)^(0x.*|Node ".*)$`)

// specialBLIF runs script with %s replaced by a file in a fresh directory
// and returns the file without its time-stamped first line.
func specialBLIF(t *testing.T, run func(string) string, script string) (file, messages string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out.blif")
	out := run(strings.ReplaceAll(script, "%s", path))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v\n%s", script, err, out)
	}
	text := string(data)
	if !strings.HasPrefix(text, "# Benchmark") {
		t.Fatalf("%s: no header:\n%s", script, text)
	}
	return text[strings.Index(text, "\n")+1:], strings.Join(blifSpecialMessage.FindAllString(out, -1), "\n")
}

func TestWriteBlifSpecialGoldens(t *testing.T) {
	for _, c := range []struct{ script, golden string }{
		{"if -K 6; write_blif -j %s", "native_mul8_if6_write_blif_j.blif"},
		{"if -S 44 -K 7; write_blif -S 44 -a %s", "native_mul8_if7_write_blif_s44a.blif"},
	} {
		got, _ := specialBLIF(t, func(s string) string { return runScript(t, s) }, blifSpecialInput+c.script)
		want, err := os.ReadFile(filepath.Join("test_vectors", c.golden))
		if err != nil {
			t.Fatal(err)
		}
		if got != string(want) {
			t.Fatalf("%s differs from %s:\n%s", c.script, c.golden, got)
		}
	}
}

// The -S 444 case crashed native (If_CluCheck3 read p->pPars with p NULL,
// CBUGS_FIXED.md #18).
func TestWriteBlifSpecialLikeNative(t *testing.T) {
	for _, script := range []string{
		"if -K 6; write_blif -j %s",
		"if -K 7; write_blif -j %s",
		"if -S 44 -K 7; write_blif -S 44 %s",
		"if -S 44 -K 7; write_blif -S 44 -a %s",
		"if -S 444 -K 9; write_blif -S 444 %s",
	} {
		native, nativeMsg := specialBLIF(t, func(s string) string { return runNativeABC(t, s) }, blifSpecialInput+script)
		port, portMsg := specialBLIF(t, func(s string) string { return runScript(t, s) }, blifSpecialInput+script)
		if native != port {
			t.Errorf("%s: file differs from native", script)
		}
		if nativeMsg != portMsg {
			t.Errorf("%s: messages differ:\nnative:\n%s\nport:\n%s", script, nativeMsg, portMsg)
		}
	}
}
