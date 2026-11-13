package ver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCommentsAndEscapedNames(t *testing.T) {
	name := filepath.Join(t.TempDir(), "names.v")
	if err := os.WriteFile(name, []byte("// comment\n/* block * still block */ \\bus [7] , next;"), 0600); err != nil {
		t.Fatal(err)
	}
	p := Ver_StreamAlloc(name)
	if p == nil {
		t.Fatal("open failed")
	}
	defer Ver_StreamFree(p)
	manager := &Ver_Man_t{pReader: p}
	if got := Ver_ParseGetName(manager); got != "bus[7]" || manager.fNameLast != 1 {
		t.Fatalf("escaped=%q last=%d", got, manager.fNameLast)
	}
	if Ver_StreamPopChar(p) != ',' {
		t.Fatal("expected comma")
	}
	if got := Ver_ParseGetName(manager); got != "next" || manager.fNameLast != 0 {
		t.Fatalf("ordinary=%q last=%d", got, manager.fNameLast)
	}
}
