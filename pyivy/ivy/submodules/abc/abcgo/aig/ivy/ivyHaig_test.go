package ivy

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHaigLifecycleChoiceRingAndSimulation(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	original := Ivy_And(p, Ivy_And(p, a, b), c)
	Ivy_ObjCreatePo(p, original)
	Ivy_ManHaigStart(p, 0)
	history := p.pHaig
	alternative := Ivy_And(p, a, Ivy_And(p, b, c))
	Ivy_ManHaigCreateChoice(p, original, alternative)
	var choices int32
	if Ivy_ManHaigCountChoices(p, &choices) != 1 || choices != 1 {
		t.Fatal("history choice count wrong")
	}
	root, member := Ivy_Regular(original.pEquiv), Ivy_Regular(alternative.pEquiv)
	if Ivy_HaigObjCountClass(root) != 2 || Ivy_Regular(Ivy_HaigObjRepr(member)) != root {
		t.Fatal("history ring representative wrong")
	}
	var out bytes.Buffer
	counts := ivyHaigSimulateSteps(p, 3, &out)
	if len(counts) != 3 || counts[0] != 0 || strings.Count(out.String(), "Non-determinate = 0") != 3 {
		t.Fatal("history combinational simulation wrong")
	}
	if Ivy_ManCheck(p) == 0 || Ivy_ManCheck(history) == 0 {
		t.Fatal("history simulation changed graph invariants")
	}
	Ivy_ManHaigStop(p)
	Ivy_ManHaigStop(p)
	if p.pHaig != nil || !history.memoryStopped {
		t.Fatal("history not released")
	}
	for _, n := range p.vObjs {
		if n != nil && n.pEquiv != nil {
			t.Fatal("history mapping retained after stop")
		}
	}
}
func TestHaigLatchResetUpdatesHashesAndCollisions(t *testing.T) {
	p := Ivy_ManStart()
	input := Ivy_ObjCreatePi(p)
	zero, one := Ivy_Latch(p, input, IVY_INIT_0), Ivy_Latch(p, input, IVY_INIT_1)
	Ivy_ObjCreatePo(p, zero)
	Ivy_ObjCreatePo(p, one)
	Ivy_ManHaigStart(p, 0)
	if Ivy_ManCheck(p.pHaig) == 0 {
		t.Fatal("DC latch reset left stale structural hashes")
	}
	if Ivy_ManLatchNum(p.pHaig) != 1 || Ivy_ManBufNum(p.pHaig) != 1 {
		t.Fatal("colliding reset latches did not retain canonical identity through a buffer")
	}
	if Ivy_ObjReal(zero.pEquiv) != Ivy_ObjReal(one.pEquiv) {
		t.Fatal("reset latch mappings do not share the canonical latch")
	}
	for _, count := range ivyHaigSimulateSteps(p, 4, nil) {
		if count != 0 {
			t.Fatal("original history latch not initialized determinately")
		}
	}
	if Ivy_ManCheck(p.pHaig) == 0 {
		t.Fatal("sequential diagnostic mutated latch state/levels")
	}
	Ivy_ManHaigStop(p)
}
func TestHaigTransferMovesOwnershipAndRecordsNewNodes(t *testing.T) {
	source := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(source), Ivy_ObjCreatePi(source)
	Ivy_ObjCreatePo(source, Ivy_And(source, a, b))
	Ivy_ManHaigStart(source, 0)
	destination := Ivy_ManStart()
	x, y := Ivy_ObjCreatePi(destination), Ivy_ObjCreatePi(destination)
	history := source.pHaig
	Ivy_ManHaigTrasfer(source, destination)
	if source.pHaig != nil || destination.pHaig != history {
		t.Fatal("history ownership not moved")
	}
	node := Ivy_And(destination, x, y)
	Ivy_ObjCreatePo(destination, node)
	if Ivy_Regular(node.pEquiv).owner != history || len(history.vPos) != 1 {
		t.Fatal("transferred history mapping/output boundary wrong")
	}
	added := Ivy_ObjCreatePi(destination)
	newNode := Ivy_CanonExor(destination, node, added)
	Ivy_ObjCreatePo(destination, newNode)
	if Ivy_Regular(newNode.pEquiv).owner != history || len(history.vPos) != 2 || len(history.vPis) != 3 {
		t.Fatal("new terminal/native-XOR history recording failed")
	}
	Ivy_ManHaigStop(destination)
	if !history.memoryStopped || Ivy_ManCheck(source) == 0 || Ivy_ManCheck(destination) == 0 {
		t.Fatal("history stop damaged working graphs")
	}
}
func TestHaigModifiedThreeValueRulesAndMalformedRing(t *testing.T) {
	for _, a := range []Ivy_Init_t{IVY_INIT_0, IVY_INIT_1, IVY_INIT_DC} {
		for _, b := range []Ivy_Init_t{IVY_INIT_0, IVY_INIT_1, IVY_INIT_DC} {
			got := Ivy_ManHaigSimulateAnd(a, b)
			want := IVY_INIT_0
			if a == IVY_INIT_DC || b == IVY_INIT_DC {
				want = IVY_INIT_DC
			} else if a == IVY_INIT_1 && b == IVY_INIT_1 {
				want = IVY_INIT_1
			}
			if got != want {
				t.Fatal("modified three-valued AND differs")
			}
		}
	}
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, a, b)
	Ivy_ObjCreatePo(p, root)
	root.pEquiv = root
	if Ivy_HaigObjCountClass(root) != 1 {
		t.Fatal("singleton ring count wrong")
	}
	root.pEquiv = a
	a.pEquiv = a
	defer func() {
		if recover() == nil {
			t.Error("malformed ring accepted")
		}
	}()
	Ivy_HaigObjCountClass(root)
}

func TestHaigNewLatchDiagnosticTraceAndPreservedState(t *testing.T) {
	p := Ivy_ManStart()
	input := Ivy_ObjCreatePi(p)
	Ivy_ObjCreatePo(p, input)
	Ivy_ManHaigStart(p, 0)
	latch := Ivy_Latch(p, input, IVY_INIT_DC)
	Ivy_ObjCreatePo(p, latch)
	var output bytes.Buffer
	counts := ivyHaigSimulateSteps(p, 3, &output)
	expected, err := os.ReadFile(filepath.Join("test_vectors", "haig_new_latch.trace"))
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != string(expected) || counts[0] != 1 || counts[1] != 0 {
		t.Fatalf("trace differs: %s", output.String())
	}
	mapped := Ivy_Regular(latch.pEquiv)
	if mapped.Init != IVY_INIT_DC || mapped.Level != 0 || Ivy_ManCheck(p.pHaig) == 0 {
		t.Fatal("diagnostic simulation changed structural latch data")
	}
	Ivy_ManHaigStop(p)
}
