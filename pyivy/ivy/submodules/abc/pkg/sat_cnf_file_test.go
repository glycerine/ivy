package abc

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCNFFileOutput(t *testing.T) {
	f, e := NewCNFFormula(3, [][]SatLit{{0, 3}, {4}, {}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	plain, e := f.DIMACS(CNFDIMACSOptions{})
	if e != nil {
		t.Fatal(e)
	}
	directory := t.TempDir()
	for _, name := range []string{"x", "formula.cnf", "formula.gz", "formula.GZ"} {
		path := filepath.Join(directory, name)
		if e = f.WriteDIMACSFile(path, CNFFileOptions{}); e != nil {
			t.Fatal(e)
		}
		got, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if name == "formula.gz" {
			got = cnfGunzip(t, got)
		}
		if !bytes.Equal(got, plain) {
			t.Fatal(name, string(got))
		}
	}
	path := filepath.Join(directory, "preserve.cnf")
	sentinel := []byte("existing data")
	if e = os.WriteFile(path, sentinel, 0600); e != nil {
		t.Fatal(e)
	}
	for _, options := range []CNFFileOptions{{DIMACS: CNFDIMACSOptions{MaxBytes: 1}}, {DIMACS: CNFDIMACSOptions{Limits: CNFTransformOptions{Stop: func() bool { return true }}}}} {
		if e = f.WriteDIMACSFile(path, options); !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(e)
		}
		got, e := os.ReadFile(path)
		if e != nil || !bytes.Equal(got, sentinel) {
			t.Fatal(got, e)
		}
	}
	if e = f.WriteDIMACSFile(directory, CNFFileOptions{}); e == nil {
		t.Fatal("directory destination accepted")
	}
	if e = f.WriteDIMACSFile(filepath.Join(directory, "missing", "file.cnf"), CNFFileOptions{}); e == nil {
		t.Fatal("missing parent accepted")
	}
	g := mappedCNFFixture(t, 0, 0)
	mapped, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if e = mapped.WriteDIMACSFile(filepath.Join(directory, "mapped.gz"), CNFFileOptions{}); e != nil {
		t.Fatal(e)
	}
	frames, e := mapped.UnrollFrames(g, 2, false, CNFFrameOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if e = frames.WriteDIMACSFile(filepath.Join(directory, "frames.cnf"), CNFFileOptions{}); e != nil {
		t.Fatal(e)
	}
}
