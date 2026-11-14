package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestAbcNtkCascadeRejectsNonStrashNetwork(t *testing.T) {
	network := abc.NewNetwork("logic")
	network.Type = uint8(abc.ABC_NTK_LOGIC)
	if got := Abc_NtkCascade(network, 4, 0, 0); got != nil {
		t.Fatal("cascade should reject a non-strash network")
	}
}

func TestCascadeBDDComplementedEdge(t *testing.T) {
	dd := core.Cudd_Init(1, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		t.Fatal("CUDD manager initialization failed")
	}
	defer core.Cudd_Quit(dd)
	input := abc.NewNetwork("input").NewObject(abc.ObjectPI)
	variable := cudd.Cudd_bddIthVar(dd, 0)
	got := bddEdgeValue(map[*abc.Object]*cudd.DdNode{input: variable}, input.Signal().Not(), cudd.Cudd_ReadOne(dd))
	if got != core.Cudd_Not(variable) {
		t.Fatal("complemented AIG edge did not complement its BDD")
	}
}
