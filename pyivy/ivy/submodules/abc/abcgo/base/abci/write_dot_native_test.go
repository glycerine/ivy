package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var dotTimeStamp = regexp.MustCompile(`Time was [^.]*\. `)

// TestWriteDotLikeNative: write_dot is Io_WriteDotNtk; s27's strashed DOT
// equals native's but for the time stamp.
func TestWriteDotLikeNative(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s27.dot")
	runScript(t, "read test_vectors/native_s27_drw.blif; strash; write_dot "+path)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("test_vectors/native_s27_strash.dot")
	if err != nil {
		t.Fatal(err)
	}
	if g, w := dotTimeStamp.ReplaceAllString(string(got), ""), dotTimeStamp.ReplaceAllString(string(want), ""); g != w {
		t.Fatalf("DOT differs from native:\n%s", g)
	}
}
