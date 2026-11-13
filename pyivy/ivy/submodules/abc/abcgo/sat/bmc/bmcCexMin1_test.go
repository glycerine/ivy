package bmc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/util"
)

func TestCexMinPerformWritesCubes(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 2
	if got := Saig_ManCexMinPerform(g, c); got != nil {
		t.Fatal("unexpected minimized trace")
	}
	model, err := ioa.Ioa_ReadAiger(filepath.Join(dir, "aigcube.aig"), 0)
	if err != nil || model == nil {
		t.Fatalf("missing cube graph: %v", err)
	}
}
