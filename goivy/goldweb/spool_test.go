package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLineSpoolQueuesXTraceInMemory(t *testing.T) {
	spool, err := newLineSpool("test", filepath.Join(t.TempDir(), "spool.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer spool.close()

	if err := spool.addLine("XTRACE: first"); err != nil {
		t.Fatal(err)
	}

	line, ok, err := spool.nextLine()
	if err != nil {
		t.Fatalf("nextLine error: %v", err)
	}
	if !ok {
		t.Fatal("nextLine ended before queued XTRACE")
	}
	if line != "XTRACE: first\n" {
		t.Fatalf("nextLine = %q", line)
	}
}

func TestLineSpoolIgnoresNonXTraceForComparison(t *testing.T) {
	spool, err := newLineSpool("test", filepath.Join(t.TempDir(), "spool.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer spool.close()

	if err := spool.addLine("~br[after i=7]: progress\n"); err != nil {
		t.Fatal(err)
	}
	spool.markDone()
	line, ok, err := spool.nextLine()
	if err != nil {
		t.Fatalf("nextLine error: %v", err)
	}
	if ok {
		t.Fatalf("nextLine returned non-XTRACE comparison line %q", line)
	}
}

func TestLineSpoolPeriodicFlush(t *testing.T) {
	spool, err := newLineSpoolWithFlushInterval("test", filepath.Join(t.TempDir(), "spool.log"), 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer spool.close()

	if err := spool.addLine("XTRACE: periodic"); err != nil {
		t.Fatal(err)
	}

	got := make(chan string, 1)
	go func() {
		line, ok, err := spool.nextLine()
		if err != nil {
			t.Errorf("nextLine error: %v", err)
			return
		}
		if !ok {
			t.Error("nextLine ended before periodic flush")
			return
		}
		got <- line
	}()

	select {
	case line := <-got:
		if line != "XTRACE: periodic\n" {
			t.Fatalf("nextLine = %q", line)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for periodic flush")
	}
}

func TestJobMismatchStopsProducersWithXTraceIndex(t *testing.T) {
	client := &wsClient{
		send: make(chan []byte, 1),
		done: make(chan struct{}),
	}
	tmp := t.TempDir()
	j := &job{
		id:             "job-1",
		client:         client,
		status:         "running",
		xtraceCount:    7,
		browserLogPath: filepath.Join(tmp, "browser.xtrace.log"),
		pyLogPath:      filepath.Join(tmp, "py.xtrace.log"),
	}

	j.mismatch("xtrace divergence", "XTRACE: browser", "XTRACE: python")

	select {
	case data := <-client.send:
		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatal(err)
		}
		if env.Type != "cancel" {
			t.Fatalf("cancel Type = %q", env.Type)
		}
		if env.ID != "job-1" {
			t.Fatalf("cancel ID = %q", env.ID)
		}
		if env.XTraceIndex == nil || *env.XTraceIndex != 7 {
			t.Fatalf("XTraceIndex = %v, want 7", env.XTraceIndex)
		}
		if !strings.Contains(env.Message, "XTRACE index 7") {
			t.Fatalf("cancel message = %q", env.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for browser cancel")
	}

	if j.status != "mismatch" {
		t.Fatalf("status = %q, want mismatch", j.status)
	}
}

func TestJobAnnotatesNonXTraceWithSideIndex(t *testing.T) {
	var out strings.Builder
	old := sideOutput
	sideOutput = &out
	defer func() { sideOutput = old }()

	spool, err := newLineSpool("browser", filepath.Join(t.TempDir(), "browser.xtrace.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer spool.close()

	j := &job{status: "running"}
	if err := j.addObservedLine("browser", "warming up\n", spool); err != nil {
		t.Fatal(err)
	}
	if err := j.addObservedLine("browser", "XTRACE: first\n", spool); err != nil {
		t.Fatal(err)
	}
	if err := j.addObservedLine("browser", "still alive\n", spool); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "~br[after i=-1]: warming up\n") {
		t.Fatalf("missing pre-XTRACE annotation in %q", got)
	}
	if !strings.Contains(got, "~br[after i=0]: still alive\n") {
		t.Fatalf("missing post-XTRACE annotation in %q", got)
	}
	line, ok, err := spool.nextLine()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || line != "XTRACE: first\n" {
		t.Fatalf("queued comparison line = %q ok=%v, want first XTRACE", line, ok)
	}
}
