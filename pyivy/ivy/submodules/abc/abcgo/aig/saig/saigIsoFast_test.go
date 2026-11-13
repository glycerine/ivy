package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSaigIsoDetectFastGroupsStructurallyEqualOutputs(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	g.AddOutput(x)
	g.AddOutput(x)
	g.AddOutput(x.Not())
	aig.Aig_ManSetRegNum(g, 0)

	classes := Saig_IsoDetectFast(g)
	if got := vec.Vec_VecSize(classes); got != 2 {
		t.Fatalf("class count = %d, want 2", got)
	}
	foundDuplicateClass := false
	for i := 0; i < vec.Vec_VecSize(classes); i++ {
		class := vec.Vec_VecEntryInt(classes, i)
		if vec.Vec_IntSize(class) == 2 &&
			((vec.Vec_IntEntry(class, 0) == 0 && vec.Vec_IntEntry(class, 1) == 1) ||
				(vec.Vec_IntEntry(class, 0) == 1 && vec.Vec_IntEntry(class, 1) == 0)) {
			foundDuplicateClass = true
		}
	}
	if !foundDuplicateClass {
		t.Fatalf("equal output cones were not grouped: %v", classes)
	}
}

func TestIsoStoCollectInfoUsesCompactSourceBitWidths(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b.Not())
	output := g.AddOutput(root)
	aig.Aig_ManSetRegNum(g, 0)
	manager := Iso_StoStart(g)
	defer Iso_StoStop(manager)
	info := Iso_StoCollectInfo(manager, output)
	if vec.Vec_IntSize(info) != 3 {
		t.Fatalf("signature entries = %d, want root and both inputs", vec.Vec_IntSize(info))
	}
	for i := 0; i < vec.Vec_IntSize(info); i++ {
		if vec.Vec_IntEntry(info, i)>>16 != 1 {
			t.Fatalf("frequency field in entry %d = %d, want 1", i, vec.Vec_IntEntry(info, i)>>16)
		}
	}
}
