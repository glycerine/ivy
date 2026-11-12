package gia

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedCompiledSimulation(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x)
	var source bytes.Buffer
	giaWriteCompiledSimulation(&source, g)
	path := filepath.Join(t.TempDir(), "simulation.go")
	if err := os.WriteFile(path, source.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("go", "run", path).CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Res = ") {
		t.Fatalf("generated simulation failed: %s %v", output, err)
	}
}
