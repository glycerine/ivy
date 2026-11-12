package res

import (
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestResWndStrashBuildsExactRewrittenMiter(t *testing.T) {
	source := abc.NewNetwork("source")
	source.Function = uint8(abc.ABC_FUNC_AIG)
	source.FunctionManager = hop.Hop_ManStart()
	hopManager := source.FunctionManager.(*hop.Hop_Man_t)
	a := source.NewObject(abc.ObjectPI)
	b := source.NewObject(abc.ObjectPI)
	c := source.NewObject(abc.ObjectPI)
	center := source.NewObject(abc.ObjectNode)
	if err := source.AddFanin(center, a.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := source.AddFanin(center, b.Signal()); err != nil {
		t.Fatal(err)
	}
	center.Data = hop.Hop_And(hopManager, hop.Hop_IthVar(hopManager, 0), hop.Hop_IthVar(hopManager, 1))
	root := source.NewObject(abc.ObjectNode)
	if err := source.AddFanin(root, center.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := source.AddFanin(root, c.Signal()); err != nil {
		t.Fatal(err)
	}
	root.Data = hop.Hop_And(hopManager, hop.Hop_IthVar(hopManager, 0), hop.Hop_IthVar(hopManager, 1))

	oldCopy := abc.Edge{Object: c, Complement: true}
	center.CopyEdge = oldCopy
	root.CopyEdge = oldCopy
	p := &Res_Win_t{
		pNode:      center,
		nWinTfoMax: 1,
		vRoots:     vec.Vec_PtrAlloc(1),
		vLeaves:    vec.Vec_PtrAlloc(3),
		vBranches:  vec.Vec_PtrAlloc(0),
		vNodes:     vec.Vec_PtrAlloc(2),
		vDivs:      vec.Vec_PtrAlloc(0),
	}
	vec.Vec_PtrPush(p.vRoots, root)
	vec.Vec_PtrPush(p.vLeaves, a)
	vec.Vec_PtrPush(p.vLeaves, b)
	vec.Vec_PtrPush(p.vLeaves, c)
	vec.Vec_PtrPush(p.vNodes, center)
	vec.Vec_PtrPush(p.vNodes, root)

	got, err := resWndStrash(p)
	if err != nil {
		t.Fatal(err)
	}
	if center.CopyEdge != oldCopy || root.CopyEdge != oldCopy {
		t.Fatal("window translation left temporary object copies on the source network")
	}
	inputs := got.ObjectsOfKind(abc.ObjectPI)
	outputs := got.ObjectsOfKind(abc.ObjectPO)
	if len(inputs) != 3 || len(outputs) != 4 {
		t.Fatalf("got %d inputs and %d outputs, want 3 and 4", len(inputs), len(outputs))
	}
	for mask := 0; mask < 8; mask++ {
		values := map[*abc.Object]bool{
			inputs[0]: mask&1 != 0,
			inputs[1]: mask&2 != 0,
			inputs[2]: mask&4 != 0,
		}
		want := []bool{mask&4 != 0, !(mask&1 != 0 && mask&2 != 0), mask&1 != 0, mask&2 != 0}
		for i, output := range outputs {
			fanins := output.Fanins()
			if len(fanins) != 1 {
				t.Fatalf("output %d has %d fanins", i, len(fanins))
			}
			gotValue, err := got.Evaluate(fanins[0], values)
			if err != nil {
				t.Fatal(err)
			}
			if gotValue != want[i] {
				t.Fatalf("assignment %03b output %d: got %v want %v", mask, i, gotValue, want[i])
			}
		}
	}
}

func TestResWndStrashRejectsInvalidWindow(t *testing.T) {
	if _, err := resWndStrash(nil); err == nil {
		t.Fatal("expected nil window error")
	}
	source := abc.NewNetwork("source")
	center := source.NewObject(abc.ObjectNode)
	if _, err := resWndStrash(&Res_Win_t{pNode: center}); err == nil {
		t.Fatal("expected source Hop-manager error")
	}
}
