package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestVerifyCounterExampleRequiresClearedSourceMarks(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	p := &Fra_Man_t{PManAig: g, PCla: Fra_ClassesStart(g)}
	cex := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(cex, 1)
	aig.Aig_ObjSetMarkB(g, x)
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts if an input simulation mark is already set")
		}
	}()
	Fra_FraigVerifyCounterEx(p, cex)
}

func TestSequentialSimulationChecksOutputOnlyWhenRequested(t *testing.T) {
	g := aig.NewGraph()
	pi, latch := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const1())
	g.AddOutput(pi)
	_ = latch
	aig.Aig_ManSetRegNum(g, 1)
	for _, tc := range []struct {
		check int32
		want  bool
	}{{0, false}, {1, true}} {
		s := Fra_SmlSimulateSeq(g, 1, 2, 1, tc.check)
		if s.NonconstantOutput != tc.want || s.PrefixFrames != 1 || s.Frames != 3 {
			t.Fatalf("check=%d: output=%v prefix=%d frames=%d", tc.check, s.NonconstantOutput, s.PrefixFrames, s.Frames)
		}
	}
}
