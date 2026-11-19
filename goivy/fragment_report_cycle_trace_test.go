package goivy

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// captureStdout returns what f writes to os.Stdout (xtracer's output).
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved, savedSup := os.Stdout, xtracer.Suppressed
	os.Stdout, xtracer.Suppressed = w, false
	defer func() { os.Stdout, xtracer.Suppressed = saved, savedSup }()
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	w.Close()
	return <-done
}

// Python report_cycle (ivy_fragment.py):
//
//	xtracer.trace("fragment/checker.reportCycle ENTER")
//	if cycle is not None:
//	    xtracer.trace("fragment/checker.reportCycle report cycle error")
//	    report_feu_error(...)          # raises
//	xtracer.trace("fragment/checker.reportCycle EXIT")
//
// The EXIT line is written only when there is no cycle: report_feu_error
// raises past it (and past check_feu's EXIT). goivy deferred EXIT, so it
// wrote it on the error path too.
func TestReportCycleTracesExitOnlyWithoutCycle(t *testing.T) {
	c := newChecker(nil, nil)
	n := NewUFNode()
	var err error
	out := captureStdout(t, func() {
		err = c.reportCycle([]arc{{from: n, to: n, fmla: True, argIdx: -1}})
	})
	if err == nil {
		t.Fatal("cycle: no error")
	}
	if !strings.Contains(out, "reportCycle report cycle error") || strings.Contains(out, "reportCycle EXIT") {
		t.Fatalf("cycle: trace = %q; want the error line and no EXIT", out)
	}
	out = captureStdout(t, func() { err = c.reportCycle(nil) })
	if err != nil || !strings.Contains(out, "reportCycle EXIT") {
		t.Fatalf("no cycle: err=%v trace=%q; want EXIT", err, out)
	}
}
