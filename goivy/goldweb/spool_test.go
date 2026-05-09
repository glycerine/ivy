package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLineSpoolBuffersUntilDone(t *testing.T) {
	spool, err := newLineSpool("test", filepath.Join(t.TempDir(), "spool.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer spool.close()

	if err := spool.addLine("XTRACE: first"); err != nil {
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
			t.Error("nextLine ended before buffered line")
			return
		}
		got <- line
	}()

	select {
	case line := <-got:
		t.Fatalf("line became visible before flush/done: %q", line)
	case <-time.After(20 * time.Millisecond):
	}

	spool.markDone()
	select {
	case line := <-got:
		if line != "XTRACE: first\n" {
			t.Fatalf("nextLine = %q", line)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for done flush")
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
	j := &job{
		id:          "job-1",
		client:      client,
		status:      "running",
		xtraceCount: 7,
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
