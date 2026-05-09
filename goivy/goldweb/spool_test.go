package main

import (
	"path/filepath"
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
