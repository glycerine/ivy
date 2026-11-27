package abci

import (
	"strings"
	"testing"
)

// abcDsd.c kept Abc_Obj_t pointers in the DSD nodes' int Mark field
// ((int)(ABC_PTRINT_T)pNode): on 64-bit hosts the upper half of each
// pointer was lost and 'dsd' crashed in Abc_ObjAddFanin on every design
// (CBUGS_FIXED.md #43).
func TestNativeDsdDoesNotCrash(t *testing.T) {
	for _, design := range []string{"test_vectors/native_s27_drw.blif", "test_vectors/native_mul4_gen.aig"} {
		for _, cmd := range []string{"dsd", "dsd -r", "collapse; dsd -g"} {
			out := runNativeABC(t, "read "+design+"; strash; "+cmd+"; print_stats")
			if !strings.Contains(out, "nd =") {
				t.Errorf("%s %s: no statistics after dsd:\n%s", design, cmd, out)
			}
		}
	}
}
