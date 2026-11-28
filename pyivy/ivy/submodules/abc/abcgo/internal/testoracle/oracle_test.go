package testoracle

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRecordThenReplay records a command that prints, fails and writes a file
// into a temporary directory, and a file read, then replays them without the
// command or the file.
func TestRecordThenReplay(t *testing.T) {
	recordDir = t.TempDir()
	recordName = func(testing.TB) string { return "fixture" }
	defer func() {
		recordDir = filepath.Join("test_vectors", "oracle")
		recordName = func(t testing.TB) string { return t.Name() }
	}()
	session := func(t *testing.T) (string, int, string, []byte) {
		dir := t.TempDir()
		cmd := Command(t, "sh", "-c", `echo out "$0"; echo err >&2; printf 'written in %s' "$0" > "$0"/result.txt; exit 3`, dir)
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		err := cmd.Run()
		var x *ExitError
		if !errors.As(err, &x) || x.ExitCode() != 3 || cmd.ExitCode() != 3 {
			t.Fatalf("err = %v, exit %d", err, cmd.ExitCode())
		}
		written, err := os.ReadFile(filepath.Join(dir, "result.txt"))
		if err != nil {
			t.Fatal(err)
		}
		// recorded paths are this run's
		return strings.ReplaceAll(out.String()+"|"+errb.String()+string(written), dir, "DIR"), cmd.ExitCode(), "", nil
	}
	var recorded string
	t.Run("record", func(t *testing.T) {
		t.Setenv("ABCGO_ORACLE_RECORD", "1")
		o, _, w, d := session(t)
		recorded = o + w + string(d)
	})
	t.Run("replay", func(t *testing.T) {
		t.Setenv("ABCGO_ORACLE_RECORD", "")
		t.Setenv("PATH", "") // no sh: the run must come from the recording
		o, _, w, d := session(t)
		if got := o + w + string(d); got != recorded || got != "out DIR\n|err\nwritten in DIR" {
			t.Fatalf("replay %q, recorded %q", got, recorded)
		}
	})
}
