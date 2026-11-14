package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCec4ManagerLifecycleAndDiagnostics(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a, b := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, x, y)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, b>>1, a>>1)
	gia.Gia_ObjSetNext(g, a>>1, b>>1)
	gia.Gia_ObjSetNext(g, b>>1, -1)
	pars := &Cec_ParFra_t{JType: 2, FVerbose: 1}
	p := Cec4_ManCreate(g, pars)
	if p.pSat == nil || p.pSat.JustificationMode != 2 || vec.Vec_IntSize(p.vCexStamps) != g.ObjectCount() {
		t.Fatal("manager not initialized")
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = write
	Cec4_ManPrintClasses(g)
	Cec4_ManPrintClasses2(g)
	Cec4_ManPrintStats(g, pars, p, 1)
	Cec4_ManPrintStats(g, pars, p, 0)
	Cec4_ManDestroy(p)
	write.Close()
	os.Stdout = old
	data, err := io.ReadAll(read)
	read.Close()
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	for _, word := range []string{"Const0 class", "Class ", "Sim ", "SAT ", "TOTAL"} {
		if !strings.Contains(out, word) {
			t.Fatalf("missing %q in %q", word, out)
		}
	}
	if p.nItersSim != 1 || p.nItersSat != 1 || p.pSat != nil || p.vFails != nil {
		t.Fatal("diagnostics or cleanup failed")
	}
}
