package gia

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSimulationPatternFiles(t *testing.T) {
	p := Gia_ManStart(10)
	a := Gia_ManAppendCi(p)
	ro := Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, ro)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	dir := t.TempDir()
	input, output := filepath.Join(dir, "inputs"), filepath.Join(dir, "outputs")
	if err := os.WriteFile(input, []byte("1, 1\n0"), 0600); err != nil {
		t.Fatal(err)
	}
	Gia_ManSimSimulatePattern(p, input, output)
	got, err := os.ReadFile(output)
	if err != nil || string(got) != "0\n1\n0\n" {
		t.Fatalf("output %q, error %v", got, err)
	}
	for _, obj := range p.pObjs {
		if obj.fMark1 != 0 {
			t.Fatal("simulation left object marks")
		}
	}
}
