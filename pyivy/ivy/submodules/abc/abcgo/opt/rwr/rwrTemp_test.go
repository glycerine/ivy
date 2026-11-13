package rwr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceFiveVariableStatsUtility(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	out := filepath.Join(dir, "out")
	f, err := os.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	w := bufio.NewWriter(f)
	for i := 0; i < len(pTruths); i++ {
		value := uint32(1)
		if i == len(pTruths)-1 {
			value = 0x80000002
		}
		fmt.Fprintf(w, "0x%08x %d\n", value, i)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rwrTempFiles(in, out); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), ".names a b c d e 00\n00001 1\n11111 1\n") {
		t.Fatalf("first class not sorted/printed as C: %.300s", data)
	}
	if !strings.HasSuffix(string(data), ".end\n") {
		t.Fatal("missing BLIF end")
	}
}
