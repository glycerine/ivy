package abci

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestOdcHashingAndTruthSimulation(t *testing.T) {
	p := Abc_NtkDontCareAlloc(5, 1, 0, 0)
	defer Abc_NtkDontCareFree(p)

	x, y := Odc_Var(p, 0), Odc_Var(p, 1)
	xy := Odc_And(p, x, y)
	if got := Odc_And(p, y, x); got != xy {
		t.Fatalf("structural hashing did not canonicalize fanins: %d != %d", got, xy)
	}
	if got := Odc_And(p, x, Odc_Not(x)); got != Odc_Const0() {
		t.Fatalf("x & !x = %d, want false literal %d", got, Odc_Const0())
	}

	Abc_NtkDontCareSimulateSetElem(p)
	p.iRoot = xy
	truth := make([]uint32, p.nWords)
	if got := Abc_NtkDontCareSimulate(p, truth); got != 8 {
		t.Fatalf("x & y truth has %d set minterms, want 8", got)
	}
	if truth[0] != 0x88888888 {
		t.Fatalf("x & y packed truth = %#08x, want %#08x", truth[0], uint32(0x88888888))
	}
}

func TestOdcClearResetsHashAndObjects(t *testing.T) {
	p := Abc_NtkDontCareAlloc(5, 1, 0, 0)
	defer Abc_NtkDontCareFree(p)

	_ = Odc_And(p, Odc_Var(p, 0), Odc_Var(p, 1))
	if Odc_NodeNum(p) != 1 || vec.Vec_IntSize(p.vUsedSpots) == 0 {
		t.Fatal("expected a node and a used hash bucket before clear")
	}
	Abc_NtkDontCareClear(p)
	if Odc_NodeNum(p) != 0 || p.iRoot != 0xffff || vec.Vec_IntSize(p.vUsedSpots) != 0 {
		t.Fatalf("clear left nodes/root/hash spots: nodes=%d root=%#x spots=%d", Odc_NodeNum(p), p.iRoot, vec.Vec_IntSize(p.vUsedSpots))
	}
	if got := Odc_And(p, Odc_Var(p, 0), Odc_Var(p, 1)); got == 0 {
		t.Fatal("hash table failed to create a node after clear")
	}
}

func TestOdcQuantifiesBranchExistentially(t *testing.T) {
	p := Abc_NtkDontCareAlloc(5, 1, 0, 0)
	defer Abc_NtkDontCareFree(p)
	branch := Odc_Var(p, p.nVarsMax)
	p.iRoot = Odc_Or(p, Odc_Var(p, 0), branch)
	vec.Vec_PtrPush(p.vBranches, nil)
	if got := Abc_NtkDontCareQuantify(p); got != 1 {
		t.Fatal("branch quantification reported an overflow")
	}
	if p.iRoot != Odc_Const1() {
		t.Fatalf("exists branch . (x | branch) = %#x, want true literal %#x", p.iRoot, Odc_Const1())
	}
}
