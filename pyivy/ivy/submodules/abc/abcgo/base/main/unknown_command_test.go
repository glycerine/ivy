package abcmain

import (
	"bytes"
	"testing"
)

// CmdCommandDispatch reports an unknown command with three lines on the
// error stream, pointing at a missing abc.rc.
func TestUnknownCommandMessageMatchesC(t *testing.T) {
	f := isolatedFrame(t)
	if status := f.Dispatch([]string{"bogus"}); status != 1 {
		t.Fatalf("status %d, want 1", status)
	}
	want := "** cmd error: unknown command 'bogus'\n" +
		"(this is likely caused by using an alias defined in \"abc.rc\"\n" +
		"without having this file in the current or parent directory)\n"
	if got := f.Err.(*bytes.Buffer).String(); got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

// A lone unknown word containing a dot is read as a file name, so
// "abc -c design.aig" reads the design.
func TestDottedWordIsRead(t *testing.T) {
	f := isolatedFrame(t)
	var got []string
	f.RegisterCommand(Command{Name: "read", Handler: func(_ *Frame, args []string) int32 {
		got = append([]string(nil), args...)
		return 0
	}})
	if status := f.Dispatch([]string{"design.aig"}); status != 0 || len(got) != 2 || got[0] != "read" || got[1] != "design.aig" {
		t.Fatalf("status %d, read args %q", status, got)
	}
}

// Abc_RealMain returns 0 after batch commands even when one fails.
func TestBatchFailureExitsZeroLikeC(t *testing.T) {
	isolatedFrame(t)
	if status := Abc_RealMain([]string{"abc", "-q", "bogus"}); status != 0 {
		t.Fatalf("exit status %d, want 0", status)
	}
}
