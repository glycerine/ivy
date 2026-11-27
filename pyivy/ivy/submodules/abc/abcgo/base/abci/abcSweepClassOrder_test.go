package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fraig_sweep used to collect its equivalence classes from a Go map, so the
// -w class listing and the order classes were merged (which numbers the
// inverters it adds) changed from run to run. Classes now come in the order
// their first node appears; each class lists its nodes as ABC's chain does
// (most recent first), and the merge keeps the node ABC keeps (n2, not n1).
// Native ABC orders classes by pointer hash, so only the members match it:
//
//	abc -c "read test_vectors/fraig_sweep_classes.blif; fraig_sweep -v -w"
func TestFraigSweepClassOrderIsDeterministic(t *testing.T) {
	blif := filepath.Join(t.TempDir(), "swept.blif")
	script := "read test_vectors/fraig_sweep_classes.blif; fraig_sweep -v -w; write_blif " + blif
	want := []string{
		"Class  0 : { n2(+)(1) n1(+)(1) }",
		"Class  1 : { m2(-)(1) m1(-)(1) }",
		"Class  2 : { k2(-)(1) k1(-)(1) }",
		"Class  3 : { p2(+)(1) p1(+)(1) }",
	}
	var first string
	for run := 0; run < 20; run++ {
		out := runScript(t, script)
		written, err := os.ReadFile(blif)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(written), "\n") {
			if !strings.HasPrefix(line, "#") { // the header carries a date
				out += line + "\n"
			}
		}
		var classes []string
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, "Class ") {
				classes = append(classes, line)
			}
		}
		if strings.Join(classes, "\n") != strings.Join(want, "\n") {
			t.Fatalf("run %d: classes\n%s\nwant\n%s", run, strings.Join(classes, "\n"), strings.Join(want, "\n"))
		}
		if run == 0 {
			first = out
			if !strings.Contains(out, ".names b a new_n2_") || strings.Contains(out, "new_n1_") {
				t.Fatalf("sweep kept n1 rather than ABC's n2:\n%s", out)
			}
		} else if out != first {
			t.Fatalf("run %d output differs:\n%s\nfirst:\n%s", run, out, first)
		}
	}
}
