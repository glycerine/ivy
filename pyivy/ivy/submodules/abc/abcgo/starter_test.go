package abcgo

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStarterFiltersLinesWithoutTruncatingCommands(t *testing.T) {
	long := strings.Repeat("x", 4000)
	input := "\n\r\n\tignored\n ignored\n#ignored\nfirst\r\n" + long + "\nlast"
	var got []string
	var out, errOut bytes.Buffer
	execute := func(command string, _, _ io.Writer) error { got = append(got, command); return nil }
	if err := runStarterCommands(strings.NewReader(input), "commands", 1, &out, &errOut, execute); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"first", long, "last"}) {
		t.Fatalf("commands: %q", got)
	}
	if errOut.Len() != 0 {
		t.Fatal(errOut.String())
	}
}

func TestStarterConcurrencyBoundAndWorkerCompletion(t *testing.T) {
	var active, peak atomic.Int32
	started := make(chan struct{}, 6)
	release := make(chan struct{})
	var out bytes.Buffer
	done := make(chan error, 1)
	execute := func(_ string, out, errOut io.Writer) error {
		n := active.Add(1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-release
		io.WriteString(out, "child output\n")
		io.WriteString(errOut, "child error\n")
		active.Add(-1)
		return nil
	}
	go func() {
		done <- runStarterCommands(strings.NewReader(strings.Repeat("run\n", 6)), "commands", 3, &out, &out, execute)
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("workers did not start")
		}
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("workers did not finish")
	}
	if active.Load() != 0 || peak.Load() != 2 {
		t.Fatalf("active %d peak %d", active.Load(), peak.Load())
	}
	if strings.Count(out.String(), "child output") != 6 || strings.Count(out.String(), "child error") != 6 {
		t.Fatal(out.String())
	}
}

func TestStarterCommandFailuresContinueAndLimit(t *testing.T) {
	var calls atomic.Int32
	var out, errOut bytes.Buffer
	err := runStarterCommands(strings.NewReader(strings.Repeat("bad\n", MAX_COMM_NUM+1)), "commands", 4, &out, &errOut, func(string, io.Writer, io.Writer) error { calls.Add(1); return errors.New("failed") })
	if err != nil || calls.Load() != MAX_COMM_NUM {
		t.Fatalf("%v, calls %d", err, calls.Load())
	}
	if strings.Count(errOut.String(), "Sorry for the inconvenience.") != MAX_COMM_NUM {
		t.Fatal("missing failure diagnostics")
	}
	if !strings.Contains(out.String(), "Cannot execute more than 1000 commands") {
		t.Fatal("missing limit diagnostic")
	}
}

type starterReadError struct{ sent bool }

func (r *starterReadError) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		return copy(p, "run\n"), errors.New("read failed")
	}
	return 0, io.EOF
}

func TestStarterReadFailureStillJoinsWorkers(t *testing.T) {
	var completed atomic.Bool
	var out bytes.Buffer
	err := runStarterCommands(new(starterReadError), "commands", 2, &out, &out, func(string, io.Writer, io.Writer) error { completed.Store(true); return nil })
	if err == nil || !completed.Load() {
		t.Fatalf("err %v completed %v", err, completed.Load())
	}
	if err := RunStarter(strings.NewReader(""), "commands", 0, &out, &out); err == nil {
		t.Fatal("invalid CPU count accepted")
	}
}

func TestStarterMainShellWorkflow(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "commands")
	if err := os.WriteFile(filename, []byte("echo starter-ok\nexit 7\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := StarterMain([]string{"1", filename}, &out, &errOut); code != 0 {
		t.Fatal(code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "starter-ok") || !strings.Contains(out.String(), "Finished processing commands") {
		t.Fatal(out.String())
	}
	if !strings.Contains(errOut.String(), "\"exit 7\"") {
		t.Fatal(errOut.String())
	}
}

func TestStarterMainUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"1"}, {"0", "missing"}, {"abc", "missing"}, {"2", "missing"}} {
		var out, errOut bytes.Buffer
		if StarterMain(args, &out, &errOut) != 1 || !strings.Contains(out.String(), "usage:") {
			t.Fatalf("%q: %s %s", args, out.String(), errOut.String())
		}
	}
}
