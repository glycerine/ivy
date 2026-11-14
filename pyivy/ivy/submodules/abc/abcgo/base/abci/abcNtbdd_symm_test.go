package abci

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestGlobalBddBuildPreservesMultiOutputFunctions(t *testing.T) {
	n, pis := multiTestAig(t)
	dd := Abc_NtkBuildGlobalBdds(n, 100000, 0, 0, 0, 0)
	if dd == nil {
		t.Fatal("global BDD build failed")
	}
	defer Abc_NtkFreeGlobalBdds(n, 1)
	pos := make([]*abc.Object, 0, 2)
	for _, o := range n.Objects() {
		if o.Kind == abc.ObjectPO {
			pos = append(pos, o)
		}
	}
	if len(pos) != 2 || Abc_ObjGlobalBdd(pos[0]) == nil || Abc_ObjGlobalBdd(pos[1]) == nil {
		t.Fatal("expected BDDs for both outputs")
	}
	for mask := 0; mask < 1<<len(pis); mask++ {
		assignment := make([]bool, len(pis))
		values := make(map[*abc.Object]bool, len(pis))
		for i, pi := range pis {
			assignment[i] = mask&(1<<i) != 0
			values[pi] = assignment[i]
		}
		for _, po := range pos {
			edge := po.Fanins()[0]
			expected := evalMultiNode(t, n, edge.Object, values, map[*abc.Object]bool{}) != edge.Complement
			got, err := dd.Evaluate(Abc_ObjGlobalBdd(po), assignment)
			if err != nil {
				t.Fatal(err)
			}
			if got != expected {
				t.Fatalf("output %s at assignment %04b: got %v, want %v", po.Name(), mask, got, expected)
			}
		}
	}
}

func TestSymmNpnCanonicalizationHandlesInputAndOutputPolarity(t *testing.T) {
	truth := []uint64{0x00000000f0f00f0f}
	flipped := append([]uint64(nil), truth...)
	util.Abc_TtFlip(flipped, 1, 1)
	util.Abc_TtNot(flipped, 1)
	schedule := extra.Extra_GreyCodeSchedule(3)
	Ntk_SymFunDeriveNpn(truth, 3, schedule)
	Ntk_SymFunDeriveNpn(flipped, 3, schedule)
	if truth[0] != flipped[0] {
		t.Fatalf("equivalent NPN variants canonicalized differently: %#x != %#x", truth[0], flipped[0])
	}
}

func TestBddImplicationUsesLibcRandSequence(t *testing.T) {
	rng := newGlibcRand(1)
	for i, want := range [][2]int{{183, 86}, {177, 115}, {193, 135}} {
		a, b := abcBddImplicationVars(rng, 200)
		if a != want[0] || b != want[1] {
			t.Fatalf("implication pair %d = (%d,%d), want (%d,%d)", i, a, b, want[0], want[1])
		}
	}
}

func TestSymmetryBDDTimingUsesAbcPrtFormat(t *testing.T) {
	old := util.PrintBridge
	defer func() { util.PrintBridge = old }()
	var output strings.Builder
	util.PrintBridge = func(s string) { output.WriteString(s) }
	abcNtkSymmPrintBddStats(1, 0, 1250_000_000, 500_000_000)
	want := "Constructing BDDs =     1.25 sec\nComputing symms   =     0.50 sec\nTOTAL             =     1.75 sec\n"
	if output.String() != want {
		t.Fatalf("ABC_PRT output:\n%q\nwant:\n%q", output.String(), want)
	}
}

func TestBddDeriveAndLocalMuxConversion(t *testing.T) {
	dd := core.Cudd_Init(2, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		t.Fatal("manager allocation failed")
	}
	defer core.Cudd_Quit(dd)
	f := core.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	if f == nil {
		t.Fatal("BDD construction failed")
	}
	core.Cudd_Ref(f)
	defer core.Cudd_RecursiveDeref(dd, f)
	names := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(names, "a")
	vec.Vec_PtrPush(names, "b")
	n := Abc_NtkDeriveFromBdd(dd, f, "both", names)
	vec.Vec_PtrFree(names)
	if n == nil || abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("BDD network derivation failed")
	}
	muxes := Abc_NtkBddToMuxes(n, 0, 1000)
	if muxes == nil || abc.Abc_NtkCheck(muxes) == 0 {
		t.Fatal("local BDD-to-MUX conversion failed")
	}
	if got := abc.Abc_NtkPoNum(muxes); got != 1 {
		t.Fatalf("outputs=%d, want 1", got)
	}
}
