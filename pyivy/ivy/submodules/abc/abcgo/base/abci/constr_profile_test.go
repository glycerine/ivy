package abci

import (
	"os"
	"strings"
	"testing"
)

// TestConstrProfileLikeNative: constr -v profiles constraints as
// Ssw_ManProfileConstraints does (Aig_ManRandom words from the all-zero
// state); the report equals native's (test_vectors golden).
func TestConstrProfileLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/native_s27_drw.blif; double; strash; constr -N 1; constr -v")
	var got []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Primary output") || strings.HasPrefix(line, "Constraint") || strings.HasPrefix(line, "TwoFrame") {
			got = append(got, line)
		}
	}
	want, err := os.ReadFile("test_vectors/constr_profile_s27_double.txt")
	if err != nil {
		t.Fatal(err)
	}
	if g := strings.Join(got, "\n") + "\n"; g != string(want) {
		t.Fatalf("got\n%s\nwant\n%s", g, want)
	}
}
