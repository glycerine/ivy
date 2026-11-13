package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
	"testing"
)

func TestClassManagerPreservesRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	aig.Aig_ManReprStart(g, 2)
	aig.Aig_ObjSetRepr(g, x, g.Const1())
	p := Ssw_ClassesStart(g)
	if Ssw_ClassesReadAig(p) != g || aig.Aig_ObjRepr(g, x) != g.Const1() {
		t.Fatal("existing representatives not preserved")
	}
	vec.Vec_PtrPush(Ssw_ClassesGetRefined(p), x)
	Ssw_ClassesClearRefined(p)
	if vec.Vec_PtrSize(Ssw_ClassesGetRefined(p)) != 0 {
		t.Fatal("refined set was not cleared")
	}
	Ssw_ClassesStop(p)
	if aig.Aig_ObjRepr(g, x) != g.Const1() {
		t.Fatal("stop changed graph representatives")
	}
}

// ABC's Aig_ManReprStart asserts Aig_ManBufNum(p) == 0. Check the
// correspondence entry point as well as the class manager so the source
// precondition does not become an accidental silent change of behavior.
func TestSignalCorrespondenceRejectsBufferWhenStartingReprs(t *testing.T) {
	for _, complemented := range []bool{false, true} {
		g := aig.NewGraph()
		input, state := g.NewInput(), g.NewInput()
		fanin := input
		if complemented {
			fanin = fanin.Not()
		}
		buffer := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{fanin, aig.InvalidEdge}})
		g.AddOutput(buffer)
		g.AddOutput(g.AND(input, state))
		aig.Aig_ManSetRegNum(g, 1)
		var pars Ssw_Pars_t
		Ssw_ManSetDefaultParams(&pars)
		func() {
			defer func() {
				v := recover()
				if v == nil || !strings.Contains(fmt.Sprint(v), "representatives require buffer-free graph") {
					t.Fatalf("complemented=%v: unexpected panic %v", complemented, v)
				}
			}()
			Ssw_SignalCorrespondence(g, &pars)
		}()
	}
}
func TestClassManagerStartsRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	p := Ssw_ClassesStart(g)
	if !aig.Aig_ManHasReprs(g) || Ssw_ClassesClassNum(p) != 0 || Ssw_ClassesCand1Num(p) != 0 || Ssw_ClassesLitNum(p) != 0 {
		t.Fatal("incorrect initial state")
	}
}

func TestClassManagerRejectsBuffersWithoutReprStorage(t *testing.T) {
	g := aig.NewGraph()
	in := g.NewInput()
	aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{in, aig.InvalidEdge}})
	defer func() {
		if recover() == nil {
			t.Fatal("expected representative-storage buffer precondition")
		}
	}()
	Ssw_ClassesStart(g)
}
