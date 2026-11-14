package cba

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadBlifSimpleNetwork(t *testing.T) {
	const source = ".model demo\n.inputs a b\n.outputs y\n.names a b y\n11 1\n.end\n"
	path := filepath.Join(t.TempDir(), "demo.blif")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	networks := Prs_ManReadBlif(path)
	if networks == nil {
		t.Fatal("Prs_ManReadBlif returned nil")
	}
	defer Prs_ManVecFree(networks)
	if got := Vec_PtrSize(networks); got != 1 {
		t.Fatalf("network count = %d, want 1", got)
	}

	ntk := Prs_ManNtk(networks, 0)
	if got := Prs_NtkName(ntk); got != "demo" {
		t.Fatalf("network name = %q, want demo", got)
	}
	if got := Prs_NtkPiNum(ntk); got != 2 {
		t.Fatalf("input count = %d, want 2", got)
	}
	if got := Prs_NtkPoNum(ntk); got != 1 {
		t.Fatalf("output count = %d, want 1", got)
	}
	if got := Prs_NtkBoxNum(ntk); got != 1 {
		t.Fatalf("box count = %d, want 1", got)
	}
	if got := Prs_NtkSop(ntk, Prs_BoxNtk(ntk, 0)); got != "11 1\n" {
		t.Fatalf("node cover = %q, want %q", got, "11 1\n")
	}
}

func TestReadBlifRejectsEmptySignalList(t *testing.T) {
	const source = ".model demo\n.inputs\n.end\n"
	path := filepath.Join(t.TempDir(), "bad.blif")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if networks := Prs_ManReadBlif(path); networks != nil {
		Prs_ManVecFree(networks)
		t.Fatal("Prs_ManReadBlif returned networks for an empty input list")
	}
}
