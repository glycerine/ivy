package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"os"
	"testing"
)

func TestFaultPrepareFlip(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	options := &Bmc_ParFf_t{Algo: 3}
	mapping := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	tests := vec.Vec_IntAlloc(4)
	scratch := vec.Vec_IntAlloc(4)
	var m *gia.Gia_Man_t
	var data *cnf.Cnf_Dat_t
	var s *bsat.Solver
	if Gia_ManFaultPrepare(p, p, options, 2, mapping, tests, scratch, &m, &data, &s, 0) != 1 {
		t.Fatal("prepare failed")
	}
	if s.Solve() != bsat.Satisfiable {
		t.Fatal("expected detectable fault")
	}
	fault := data.VarNums[m.InputLiterals()[2]>>1]
	if value, ok := s.Value(fault); !ok || !value {
		t.Fatalf("fault enabled: %v %v", value, ok)
	}
}

func TestFaultOutputName(t *testing.T) {
	graph := gia.Gia_ManFromAig(func() *aig.Graph { g := aig.NewGraph(); g.Spec = "folder/design.aig"; return g }())
	if got := faultOutputName(graph, "_tests.txt", "tests.txt"); got != "folder/design_tests.txt" {
		t.Fatalf("output path %q", got)
	}
	if got := faultOutputName(gia.Gia_ManStart(1), "_tests.txt", "tests.txt"); got != "tests.txt" {
		t.Fatalf("default path %q", got)
	}
}

func TestFaultDriverDumpsExistingTestsWhenPreparationFinishes(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	graph := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendCo(graph, gia.Gia_ManAppendAnd(graph, a, b))
	if err = os.WriteFile("prior.txt", []byte("00\n01\n"), 0600); err != nil {
		t.Fatal(err)
	}
	Gia_ManFaultTest(graph, graph, &Bmc_ParFf_t{Algo: 3, PFileName: "prior.txt", FDump: 1})
	data, err := os.ReadFile("tests.txt")
	if err != nil {
		t.Fatalf("expected retained test suite: %v", err)
	}
	if string(data) != "00\n" {
		t.Fatalf("test suite = %q", data)
	}
}
