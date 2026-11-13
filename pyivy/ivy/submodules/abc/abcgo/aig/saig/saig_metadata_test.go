package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// saig.h reads nTruePis/nTruePos directly, even when those fields do not
// agree with the total terminal vectors. Aig_ManSetRegNum refreshes them.
func TestSAIGTerminalAccessorsUseStoredMetadata(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	ox, oy, oz := g.AddOutput(x), g.AddOutput(y), g.AddOutput(z)
	if Saig_ManPiNum(g) != 0 || Saig_ManPoNum(g) != 0 || Saig_ManCiNum(g) != 0 || Saig_ManCoNum(g) != 0 {
		t.Fatal("new graph should retain C's zero-initialized counts")
	}
	aig.Aig_ManSetRegNum(g, 1)
	if Saig_ManPiNum(g) != 2 || Saig_ManPoNum(g) != 2 || Saig_ManCiNum(g) != 3 || Saig_ManCoNum(g) != 3 || Saig_ManLo(g, 0) != z || Saig_ManLi(g, 0) != oz {
		t.Fatal("SetRegNum did not establish normal metadata")
	}
	g.SetTruePiCount(1)
	g.SetTruePoCount(1)
	if Saig_ManPiNum(g) != 1 || Saig_ManPoNum(g) != 1 || Saig_ManCiNum(g) != 2 || Saig_ManCoNum(g) != 2 {
		t.Fatal("stored metadata ignored")
	}
	if !Saig_ObjIsPi(g, x) || Saig_ObjIsPi(g, y) || !Saig_ObjIsLo(g, y) || !Saig_ObjIsPo(g, ox) || Saig_ObjIsPo(g, oy) || !Saig_ObjIsLi(g, oy) {
		t.Fatal("terminal classification ignored stored counts")
	}
	if Saig_ManLo(g, 0) != y || Saig_ManLi(g, 0) != oy || Saig_ObjLoToLi(g, y) != oy || Saig_ObjLiToLo(g, oy) != y || Saig_ObjRegId(g, y) != 0 || Saig_ObjRegId(g, oy) != 0 {
		t.Fatal("register mapping ignored stored counts")
	}
}

func TestSAIGRegisterIDRejectsPrimaryTerminal(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 0)
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts for a primary terminal")
		}
	}()
	Saig_ObjRegId(g, x)
}

// The C hide routine preserves the requested bad-list order and leaves each
// terminal's stored CioId unchanged until Aig_ManSetCioIds is called.
func TestHideExposePreservesCioIDsAndBadListOrder(t *testing.T) {
	g := aig.NewGraph()
	pi, r0, r1, r2 := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	po, l0, l1, l2 := g.AddOutput(pi), g.AddOutput(r0), g.AddOutput(r1), g.AddOutput(r2)
	aig.Aig_ManSetRegNum(g, 3)
	aig.Aig_ManSetConstrNum(g, 1)
	bad := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(bad, l2)
	vec.Vec_PtrPush(bad, l0)
	if got := Saig_ManHideBadRegs(g, bad); got != 2 {
		t.Fatalf("hidden %d", got)
	}
	for i, e := range []aig.Edge{pi, r2, r0, r1} {
		if g.Inputs()[i] != e {
			t.Fatalf("input order=%v", g.Inputs())
		}
	}
	for i, e := range []aig.Edge{po, l2, l0, l1} {
		if g.Outputs()[i] != e {
			t.Fatalf("output order=%v", g.Outputs())
		}
	}
	for _, tc := range []struct {
		e  aig.Edge
		id int32
	}{{pi, 0}, {r2, 3}, {r0, 1}, {r1, 2}, {po, 0}, {l2, 3}, {l0, 1}, {l1, 2}} {
		if got := aig.Aig_ObjCioId(g, tc.e); got != tc.id {
			t.Fatalf("edge %v id=%d want=%d", tc.e, got, tc.id)
		}
	}
	beforeExpose := append([]aig.Edge(nil), g.Outputs()...)
	Saig_ManExposeBadRegs(g, 2)
	for i, edge := range beforeExpose {
		if g.Outputs()[i] != edge {
			t.Fatal("C expose must not reorder constrained outputs")
		}
	}
	if g.TruePiCount() != 1 || g.TruePoCount() != 1 || g.RegisterCount() != 3 {
		t.Fatal("expose did not restore stored counts")
	}
	if aig.Aig_ObjCioId(g, r2) != 3 || aig.Aig_ObjCioId(g, l2) != 3 {
		t.Fatal("expose changed CioId")
	}
	aig.Aig_ManSetCioIds(g)
	if aig.Aig_ObjCioId(g, r2) != 1 || aig.Aig_ObjCioId(g, l2) != 1 {
		t.Fatal("explicit reset did not renumber")
	}
}

func TestHideBadRegistersRequiresSourceInterface(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		stale, useLO, duplicate bool
	}{
		{name: "stale counts", stale: true}, {name: "latch output", useLO: true}, {name: "duplicate latch input", duplicate: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			pi, lo := g.NewInput(), g.NewInput()
			g.AddOutput(pi)
			li := g.AddOutput(lo)
			aig.Aig_ManSetRegNum(g, 1)
			if tc.stale {
				g.SetTruePiCount(0)
			}
			bad := vec.Vec_PtrAlloc(2)
			if tc.useLO {
				vec.Vec_PtrPush(bad, lo)
			} else {
				vec.Vec_PtrPush(bad, li)
			}
			if tc.duplicate {
				vec.Vec_PtrPush(bad, li)
			}
			defer func() {
				if recover() == nil {
					t.Fatal("source-invalid interface was accepted")
				}
			}()
			Saig_ManHideBadRegs(g, bad)
		})
	}
}
