package bmc

import (
	"os"
	"strings"
	"testing"
)

func TestExa2SynthesisDoesNotDumpBLIF(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	Exa_ManExactSynthesis2(&Bmc_EsPar_t{nVars: 2, nNodes: 1, pTtStr: "8"})
	w.Close()
	os.Stdout = original
	output := make([]byte, 8192)
	n, _ := r.Read(output)
	r.Close()
	if !strings.Contains(string(output[:n]), "Realization") {
		t.Fatalf("missing solution: %s", output[:n])
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("unexpected output files: %v", files)
	}
}
