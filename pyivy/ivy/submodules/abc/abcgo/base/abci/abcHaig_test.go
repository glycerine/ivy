package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
)

func haigChoiceNetwork(t *testing.T, reverse bool) *abc.Network {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	x, y, z := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	x.SetName("x")
	y.SetName("y")
	z.SetName("z")
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("f")
	aig := n.FunctionManager.(*abc.Abc_Aig_t)
	xy := abc.Abc_AigAnd(aig, x.Signal(), y.Signal())
	a := abc.Abc_AigAnd(aig, xy, z.Signal())
	yz := abc.Abc_AigAnd(aig, y.Signal(), z.Signal())
	b := abc.Abc_AigAnd(aig, x.Signal(), yz)
	root := b
	if reverse {
		// b depends on a only in this arrangement, which must be rejected as a
		// choice under a because that would make the choice class cyclic.
		b = abc.Abc_AigAnd(aig, x.Signal(), abc.Abc_AigAnd(aig, xy, z.Signal()))
		root = b
	}
	if err := n.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}
	if reverse {
		a.Object.Choice = b.Object
	} else {
		a.Object.Choice = b.Object
	}
	return n
}

func haigEval(n *abc.Network, assignment uint32) []bool {
	values := make(map[*abc.Object]bool)
	for _, object := range n.Objects() {
		switch object.Kind {
		case abc.ObjectConst1:
			values[object] = true
		case abc.ObjectPI:
			idx := 0
			for i, ci := range n.Inputs() {
				if ci == object {
					idx = i
					break
				}
			}
			values[object] = assignment&(1<<uint(idx)) != 0
		case abc.ObjectNode:
			f := object.Fanins()
			a, b := values[f[0].Object] != f[0].Complement, values[f[1].Object] != f[1].Complement
			values[object] = a && b
		}
	}
	result := make([]bool, len(n.Outputs()))
	for i, co := range n.Outputs() {
		f := co.Fanins()[0]
		result[i] = values[f.Object] != f.Complement
	}
	return result
}

func TestHopRemoveLoopsReconstructsSafeChoices(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		name := "independent equivalents"
		if reverse {
			name = "dependent equivalent"
		}
		t.Run(name, func(t *testing.T) {
			source := haigChoiceNetwork(t, reverse)
			defer abc.Abc_NtkDelete(source)
			manager := Abc_NtkPartStartHop(source)
			if len(hopObjectsForTest(manager)) == 0 {
				t.Fatal("HAIG manager did not retain its object order")
			}
			result := Abc_NtkHopRemoveLoops(source, manager)
			if result == nil {
				t.Fatal("loop removal failed to reconstruct the network")
			}
			defer abc.Abc_NtkDelete(result)
			if abc.Abc_NtkCheck(result) == 0 {
				t.Fatal("reconstructed network failed ABC check")
			}
			for assignment := uint32(0); assignment < 8; assignment++ {
				if got, want := haigEval(result, assignment), haigEval(source, assignment); got[0] != want[0] {
					t.Fatalf("assignment %03b: got %v, want %v", assignment, got, want)
				}
			}
			choices := 0
			for _, node := range abc.Abc_NtkNodes(result) {
				if node.Choice != nil {
					choices++
				}
			}
			if reverse && choices != 0 {
				t.Fatalf("dependent candidate produced %d cyclic choice links", choices)
			}
			if !reverse && choices == 0 {
				t.Fatal("independent equivalent candidate was not retained as a choice")
			}
		})
	}
}

func hopObjectsForTest(manager *Hop_Man_t) []*Hop_Obj_t { return hop.Hop_ManObjects(manager) }
