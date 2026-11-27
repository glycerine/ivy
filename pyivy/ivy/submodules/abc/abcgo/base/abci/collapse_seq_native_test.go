package abci

import (
	"strings"
	"testing"
)

// Abc_NtkFromGlobalBdds builds a node for every CO; the port built them for
// primary outputs only, leaving latch inputs undriven, so collapse failed
// on every sequential network ("The network check has failed.").
func TestCollapseSequential(t *testing.T) {
	for script, want := range map[string]string{
		"read test_vectors/native_s27_drw.blif; collapse; print_stats": "i/o =    4/    1  lat =    3  nd =     4  edge =     20  bdd  =    22  lev = 1",
		"read test_vectors/seq/s713.blif; collapse; print_stats":       "i/o =   35/   23  lat =   19  nd =    42  edge =    486  bdd  =  1130  lev = 1",
	} {
		if out := runScript(t, script); !strings.Contains(out, want) {
			t.Errorf("%s:\n%s", script, out)
		}
	}
}
