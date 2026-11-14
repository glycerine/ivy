package ver

import (
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"strings"
	"testing"
)

func TestParseBinaryConstantSourceValues(t *testing.T) {
	p := &Ver_Man_t{vNames: vec.Vec_PtrAlloc(0), Output: os.Stdout, fTopLevel: 1}
	if Ver_ParseConstant(p, "4'b10x1") != 1 {
		t.Fatal("valid constant rejected")
	}
	for i, want := range []any{int32(1), nil, nil, int32(1)} {
		if got := vec.Vec_PtrEntry(p.vNames, i); got != want {
			t.Fatalf("bit %d=%v want %v", i, got, want)
		}
	}
	for _, tc := range []struct{ word, message string }{{"3", "Cannot find symbol"}, {"3'h101", "only handle binary"}, {"3'b12x", "problem parsing"}} {
		if Ver_ParseConstant(p, tc.word) != 0 || !strings.Contains(string(p.sError[:]), tc.message) {
			t.Fatalf("%q: %q", tc.word, p.sError)
		}
	}
}
