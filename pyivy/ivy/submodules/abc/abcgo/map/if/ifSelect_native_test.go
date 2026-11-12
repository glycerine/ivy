package if_

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestIfWordCountOnes(t *testing.T) {
	for _, test := range []struct {
		word uint32
		want int32
	}{{0, 0}, {1, 1}, {0xaaaaaaaa, 16}, {0xffffffff, 32}, {0x80010001, 3}} {
		if got := If_WordCountOnes(test.word); got != test.want {
			t.Fatalf("popcount(%#x)=%d, want %d", test.word, got, test.want)
		}
	}
}

func TestIfChoiceSATConstraintsSelectExactlyOne(t *testing.T) {
	s := bsat.NewSolver2()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	satSolverAddChoice(s, 0, []int32{1, 2})
	if !s.AddClause(bsat.Positive(0)) {
		t.Fatal("failed to assert the external choice")
	}
	if got := s.Solve(); got != bsat.Satisfiable {
		t.Fatalf("SAT status=%v", got)
	}
	selected := 0
	for _, v := range []int32{1, 2} {
		if value, ok := s.Value(v); ok && value {
			selected++
		}
	}
	if selected != 1 {
		t.Fatalf("selected %d implementations, want exactly one", selected)
	}
}

func TestIfNodeShapeMapCollectsChosenImplementation(t *testing.T) {
	p := If_ManStart(testManagerPars())
	defer If_ManStop(p)
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	cut := If_ObjCutBest(root)
	cut.nLeaves = 2
	cut.pLeaves = []int32{a.Id, b.Id}
	shape := vec.Vec_IntAlloc(4)
	defer vec.Vec_IntFree(shape)
	if If_ManNodeShapeMap(p, root, shape) == 0 {
		t.Fatal("shape traversal failed")
	}
	got := vec.Vec_IntArray(shape)
	if len(got) != 2 || got[0] != root.Id || got[1] != root.Id {
		t.Fatalf("shape=%v, want root pair [%d %d]", got, root.Id, root.Id)
	}
	if If_CutDataInt(If_ObjCutBest(a)) != 0 || If_CutDataInt(If_ObjCutBest(b)) != 0 || If_CutDataInt(cut) != 0 {
		t.Fatal("shape traversal leaked scratch marks")
	}
}
