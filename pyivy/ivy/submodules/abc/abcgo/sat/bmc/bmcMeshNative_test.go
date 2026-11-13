package bmc

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

func TestMeshOneHot(t *testing.T) {
	s := satoko.New()
	s.SetVariables(3)
	if s.SolveAssumptions([]int{0, 2, 4}) != satoko.SATOKO_SAT {
		t.Fatal("expected model")
	}
	if got := Bmc_MeshAddOneHotness(s, 0, 3); got != 3 {
		t.Fatalf("expected three exclusions, got %d", got)
	}
	if s.SolveAssumptions([]int{0, 2}) != satoko.SATOKO_UNSAT {
		t.Fatal("two selected cells must be impossible")
	}
}
func TestMeshAndPlacement(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	out := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, out)
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	Bmc_MeshTest(p, 3, 3, 2, 0)
	w.Close()
	os.Stdout = old
	data, _ := io.ReadAll(r)
	r.Close()
	if !strings.Contains(string(data), "Satisfying solution found") {
		t.Fatalf("unexpected result: %s", data)
	}
}
