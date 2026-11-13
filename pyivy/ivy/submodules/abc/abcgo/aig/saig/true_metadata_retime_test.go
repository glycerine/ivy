package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestHideExposePreservesStoredTerminalCounts(t *testing.T) {
	g := aig.NewGraph()
	x, q := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	li := g.AddOutput(q)
	aig.Aig_ManSetRegNum(g, 1)
	bad := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(bad, li)
	if got := Saig_ManHideBadRegs(g, bad); got != 1 {
		t.Fatalf("hidden %d", got)
	}
	if g.TruePiCount() != 2 || g.TruePoCount() != 2 || g.RegisterCount() != 0 {
		t.Fatalf("hidden counts %d/%d/%d", g.TruePiCount(), g.TruePoCount(), g.RegisterCount())
	}
	// Expose decrements the stored values, even if they are stale.
	g.SetTruePiCount(1)
	g.SetTruePoCount(1)
	Saig_ManExposeBadRegs(g, 1)
	if g.TruePiCount() != 0 || g.TruePoCount() != 0 || g.RegisterCount() != 1 {
		t.Fatalf("restored counts %d/%d/%d", g.TruePiCount(), g.TruePoCount(), g.RegisterCount())
	}
}

func TestRetimeDuplicatesCopyTrueOutputMetadata(t *testing.T) {
	g := aig.NewGraph()
	x, q := g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	g.Assertions = 1
	g.SetTruePoCount(0)
	cut := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(cut, q)
	for name, result := range map[string]*aig.Graph{
		"forward":  Saig_ManRetimeDupForward(g, cut),
		"backward": Saig_ManRetimeDupBackward(g, cut, nil),
	} {
		if result == nil {
			t.Fatalf("%s returned nil", name)
		}
		if result.InputCount() != 2 || result.OutputCount() != 1 || result.RegisterCount() != 1 {
			t.Fatalf("%s interface %d/%d/%d, C gives 2/1/1", name, result.InputCount(), result.OutputCount(), result.RegisterCount())
		}
		for _, edge := range append(result.Inputs(), result.Outputs()...) {
			if id := aig.Aig_ObjCioId(result, edge); id != 0 {
				t.Fatalf("%s terminal CioId=%d, C leaves zero", name, id)
			}
		}
		if aig.Aig_ManConstrNum(result) != 0 || result.Assertions != 0 {
			t.Fatalf("%s retained C-omitted constraint/assert metadata", name)
		}
		if result.TruePiCount() != g.TruePiCount() || result.TruePoCount() != 0 {
			t.Fatalf("%s copied counts %d/%d", name, result.TruePiCount(), result.TruePoCount())
		}
	}
}
