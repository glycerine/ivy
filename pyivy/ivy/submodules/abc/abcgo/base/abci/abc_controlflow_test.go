package abci

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"testing"
)

func TestAbcCommandLoadUsagePath(t *testing.T) {
	// The help switch exits through the command's usage path before the frame
	// is inspected, matching the C command's early goto usage behavior.
	if got := Abc_CommandAbcLoad(nil, 2, []string{"load", "-h"}); got != 1 {
		t.Fatalf("help command returned %d, want 1", got)
	}
}

func TestAbcCommandIfifUsagePath(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	if got := Abc_CommandIfif(frame, 2, []string{"ifif", "-h"}); got != 1 {
		t.Fatalf("help command returned %d, want 1", got)
	}
}
