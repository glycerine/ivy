package abci

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func capturePrintStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	got, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes.TrimRight(got, "\n"))
}

func TestAbcNtkShow6VarFuncRejectsBadLength(t *testing.T) {
	got := capturePrintStdout(t, func() { Abc_NtkShow6VarFunc("00", "0000000000000000") })
	if got != "Wrong length (2) of 6-var truth table." {
		t.Fatalf("unexpected diagnostic: %q", got)
	}
}

func TestAbcNtkShow6VarFuncRendersTruthTable(t *testing.T) {
	got := capturePrintStdout(t, func() { Abc_NtkShow6VarFunc("0000000000000000", "0000000000000000") })
	lines := bytes.Split([]byte(got), []byte("\n"))
	if len(lines) != 18 {
		t.Fatalf("expected header, separator, eight rows and eight rules; got %d lines", len(lines))
	}
	if !bytes.Contains(lines[0], []byte("000 100 110")) {
		t.Fatalf("unexpected column ordering: %q", lines[0])
	}
	if !bytes.Contains(lines[2], []byte(" |")) {
		t.Fatalf("unexpected first row: %q", lines[2])
	}
}

func TestAbcPrintBinUsesABCDecimalBuckets(t *testing.T) {
	bins := make([]int32, 70)
	for _, value := range []int32{9, 10, 99, 100, 999, 1000, 9_999_999} {
		abcPrintBin(bins, value)
	}
	for _, bucket := range []int{9, 11, 19, 21, 29, 31, 69} {
		if bins[bucket] != 1 {
			t.Fatalf("bucket %d = %d, want 1", bucket, bins[bucket])
		}
	}
}
