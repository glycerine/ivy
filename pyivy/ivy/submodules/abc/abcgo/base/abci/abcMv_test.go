package abci

import (
	"bufio"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestAbcMvReadCubeVariableOrderAndComplements(t *testing.T) {
	dd := core.Cudd_Init(32, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		t.Fatal("BDD manager creation failed")
	}
	defer core.Cudd_Quit(dd)

	want := core.Cudd_bddAnd(dd, core.Cudd_bddIthVar(dd, 29), core.Cudd_Not(core.Cudd_bddIthVar(dd, 28)))
	got := Abc_MvReadCube(dd, "10", 2)
	if got != want {
		t.Fatalf("cube 10 = %p, want var29 & !var28 = %p", got, want)
	}
	if got := Abc_MvReadCube(dd, "--", 2); got != core.Cudd_ReadOne(dd) {
		t.Fatalf("all-dash cube = %p, want one %p", got, core.Cudd_ReadOne(dd))
	}
}

func TestAbcMvReadScannerBuildsValueSetsAndEncodedFunctions(t *testing.T) {
	dd := core.Cudd_Init(32, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		t.Fatal("BDD manager creation failed")
	}
	defer core.Cudd_Quit(dd)
	p := &mvMan{dd: dd, nInputs: 9, nFuncs: 15}

	inputs := []byte(strings.Repeat("-", 18))
	inputs[0] = '1'
	line := string(inputs) + " " + "10" + strings.Repeat("--", 14) + "\n.e\n"
	if err := abcMvReadScanner(p, bufio.NewScanner(strings.NewReader(line))); err != nil {
		t.Fatal(err)
	}
	defer Abc_MvDeref(p)

	if p.bValues[0][1] != core.Cudd_ReadOne(dd) || p.bValues[0][0] != core.Cudd_ReadLogicZero(dd) {
		t.Fatal("PLA value sets do not add the complement of the specified domain")
	}
	for value := 0; value < 4; value++ {
		if p.bValues[1][value] != core.Cudd_ReadOne(dd) {
			t.Fatalf("don't-care value set output=1 value=%d is not unconstrained", value)
		}
	}
	want := core.Cudd_bddAnd(dd, core.Cudd_bddIthVar(dd, 30), core.Cudd_Not(core.Cudd_bddIthVar(dd, 31)))
	if p.bFuncs[0] != want {
		t.Fatal("multi-valued output was not encoded into the two output bits")
	}
}

func TestAbcMvReadCubeRejectsInvalidCharacter(t *testing.T) {
	dd := core.Cudd_Init(32, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	defer core.Cudd_Quit(dd)
	defer func() {
		if recover() == nil {
			t.Fatal("invalid cube character did not panic")
		}
	}()
	Abc_MvReadCube(dd, "x", 1)
}
