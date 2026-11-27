package abci

import (
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// TestRexWalkLikeNative: &rexwalk chooses moves with glibc's rand()
// compared in float, as giaRex.c does, so the walks and the state table
// equal native's (test_vectors golden).
func TestRexWalkLikeNative(t *testing.T) {
	util.C_Srand(1) // a fresh process: glibc's rand() starts at seed 1
	out := runScript(t, "&rex2gia (ab|c)*a(b|c); &rexwalk -S 20 -R 3 -v")
	want, err := os.ReadFile("test_vectors/rexwalk_native.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out, string(want)) {
		t.Fatalf("got\n%s\nwant suffix\n%s", out, want)
	}
}
