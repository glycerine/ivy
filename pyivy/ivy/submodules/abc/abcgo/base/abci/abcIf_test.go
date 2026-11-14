package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

func TestIfTruthToBDDPreservesTruthTable(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		dd := core.NewBDDManager(3)
		// x0 XOR x2, with x1 unused.
		truth := []uint32{0x5a5a5a5a}
		root := ifTruthToBDD(dd, truth, 3, reverse)
		for assignment := 0; assignment < 8; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			got, err := dd.Evaluate(root, values)
			if err != nil {
				t.Fatal(err)
			}
			if want := values[0] != values[2]; got != want {
				t.Fatalf("reverse=%t assignment=%03b: got %t, want %t", reverse, assignment, got, want)
			}
		}
	}
}

func TestAbcNtkToIfCopiesAIGAndComplementedOutput(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	x, y := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := abc.Abc_AigAnd(n.FunctionManager.(*abc.Abc_Aig_t), x.Signal(), y.Signal().Not())
	out := abc.Abc_NtkCreatePo(n)
	if err := n.AddFanin(out, and.Not()); err != nil {
		t.Fatal(err)
	}

	pars := &ifpkg.If_Par_t{}
	ifpkg.If_ManSetDefaultPars(pars)
	pars.NLutSize = 6
	manager := Abc_NtkToIf(n, pars)
	if manager == nil {
		t.Fatal("Abc_NtkToIf returned nil")
	}
	defer ifpkg.If_ManStop(manager)
	if ifpkg.If_ManCiNum(manager) != 2 || ifpkg.If_ManCoNum(manager) != 1 || ifpkg.If_ManAndNum(manager) != 1 {
		t.Fatalf("mapped graph counts: CI=%d CO=%d AND=%d", ifpkg.If_ManCiNum(manager), ifpkg.If_ManCoNum(manager), ifpkg.If_ManAndNum(manager))
	}
	co := ifpkg.If_ManCo(manager, 0)
	if ifpkg.If_ObjFaninC0(co) != 1 {
		t.Fatal("complemented AIG output lost its output phase")
	}
	if _, ok := abc.Abc_ObjCopyData(and.Object).(*ifpkg.If_Obj_t); !ok {
		t.Fatal("AIG node's mapper copy field was not populated")
	}
}

func TestIFMappingPreservesTruthAcrossFunctionRepresentations(t *testing.T) {
	for _, representation := range []string{"Hop", "BDD", "SOP", "CNF", "MV"} {
		t.Run(representation, func(t *testing.T) {
			n := truthNetwork(t, 4, 0x6996, 0xa1f7, 0xaaaa, 0, 0xffff)
			defer abc.Abc_NtkDelete(n)
			var pars ifpkg.If_Par_t
			var bdds, sops, cnfs, mv int32
			switch representation {
			case "BDD":
				bdds = 1
			case "SOP":
				sops = 1
			case "CNF":
				cnfs = 1
			case "MV":
				mv = 1
			}
			pars.SetRenodeConfig(3, 8, 1, 1, 0, bdds, sops, cnfs, mv, 0, nil, nil)
			mapped := Abc_NtkIf(n, &pars)
			if mapped == nil {
				t.Fatal("mapping failed")
			}
			defer abc.Abc_NtkDelete(mapped)
			for bits := 0; bits < 16; bits++ {
				want, got := evaluateNetwork(t, n, bits), evaluateNetwork(t, mapped, bits)
				if len(got) != len(want) {
					t.Fatal("output count changed")
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("bits=%d output=%d got=%v want=%v", bits, i, got, want)
					}
				}
			}
		})
	}
}
