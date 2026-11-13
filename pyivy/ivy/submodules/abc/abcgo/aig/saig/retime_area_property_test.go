package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"testing"
)

func TestMinAreaRetimePreservesReachableOutputs(t *testing.T) {
	r := rand.New(rand.NewSource(5729))
	for trial := 0; trial < 32; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 8; i++ {
			a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 4; i++ {
			e := edges[r.Intn(len(edges))]
			if r.Intn(2) == 0 {
				e = e.Not()
			}
			g.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(g, 2)
		aig.Aig_ManCleanup(g)
		for _, mode := range []struct{ forward, backward int32 }{{1, 0}, {0, 1}, {0, 0}} {
			got := Saig_ManRetimeMinArea(g, 4, mode.forward, mode.backward, 1, 0)
			if got == nil {
				t.Fatalf("trial %d mode %+v returned nil", trial, mode)
			}
			for trace := 0; trace < 64; trace++ {
				oldState, newState := make([]bool, g.RegisterCount()), make([]bool, got.RegisterCount())
				for frame := 0; frame < 3; frame++ {
					ins := []bool{trace&(1<<(2*frame)) != 0, trace&(1<<(2*frame+1)) != 0}
					old, err := g.Evaluate(append(append([]bool(nil), ins...), oldState...))
					if err != nil {
						t.Fatal(err)
					}
					out, err := got.Evaluate(append(append([]bool(nil), ins...), newState...))
					if err != nil {
						t.Fatal(err)
					}
					if old[0] != out[0] || old[1] != out[1] {
						t.Fatalf("trial %d mode %+v trace %d frame %d old=%v new=%v", trial, mode, trace, frame, old, out)
					}
					copy(oldState, old[len(old)-len(oldState):])
					copy(newState, out[len(out)-len(newState):])
				}
			}
		}
	}
}

func TestMinAreaRetimeReducesSharedCuts(t *testing.T) {
	t.Run("backward", func(t *testing.T) {
		g := aig.NewGraph()
		x, y, a, b := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		gate := g.AND(x, y)
		g.AddOutput(a)
		g.AddOutput(b)
		g.AddOutput(gate)
		g.AddOutput(gate)
		aig.Aig_ManSetRegNum(g, 2)
		out := Saig_ManRetimeMinArea(g, 4, 0, 1, 1, 0)
		if out.RegisterCount() != 1 {
			t.Fatalf("registers = %d, want 1", out.RegisterCount())
		}
		checkRetimeTraces(t, g, out, 2, 2)
	})
	t.Run("forward", func(t *testing.T) {
		g := aig.NewGraph()
		x := g.NewInput()
		a, b := g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(a, b))
		g.AddOutput(x)
		g.AddOutput(x)
		aig.Aig_ManSetRegNum(g, 2)
		out := Saig_ManRetimeMinArea(g, 4, 1, 0, 1, 0)
		if out.RegisterCount() != 1 {
			t.Fatalf("registers = %d, want 1", out.RegisterCount())
		}
		checkRetimeTraces(t, g, out, 1, 1)
	})
}

func checkRetimeTraces(t *testing.T, source, result *aig.Graph, pis, pos int) {
	t.Helper()
	for trace := 0; trace < 1<<(pis*4); trace++ {
		oldState, newState := make([]bool, source.RegisterCount()), make([]bool, result.RegisterCount())
		for frame := 0; frame < 4; frame++ {
			inputs := make([]bool, pis)
			for i := range inputs {
				inputs[i] = trace&(1<<(frame*pis+i)) != 0
			}
			old, err := source.Evaluate(append(append([]bool(nil), inputs...), oldState...))
			if err != nil {
				t.Fatal(err)
			}
			got, err := result.Evaluate(append(append([]bool(nil), inputs...), newState...))
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < pos; i++ {
				if old[i] != got[i] {
					t.Fatalf("trace=%d frame=%d output=%d old=%v new=%v", trace, frame, i, old, got)
				}
			}
			copy(oldState, old[len(old)-len(oldState):])
			copy(newState, got[len(got)-len(newState):])
		}
	}
}

func TestBackwardRetimeRestoresExcludedRegisters(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	for _, e := range []aig.Edge{a, b, c, d, x, x.Not(), y, y} {
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 4)
	result := Saig_ManRetimeMinArea(g, 4, 0, 1, 1, 0)
	if result == nil {
		t.Fatal("retiming returned nil")
	}
	if result.RegisterCount() != 3 {
		t.Fatalf("registers=%d, want 3", result.RegisterCount())
	}
	checkRetimeTraces(t, g, result, 2, 4)
}

func TestBackwardRetimeRetriesUnsatisfiableInitialState(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	for _, e := range []aig.Edge{a, b, c, x, y, g.OR(x, y).Not()} {
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 3)
	edges := saigRetimeMinCut(g, false)
	cut := vec.Vec_PtrAlloc(len(edges))
	for _, edge := range edges {
		vec.Vec_PtrPush(cut, edge)
	}
	initial := Saig_ManRetimeDupInitState(g, cut)
	if len(edges) >= 3 || Saig_ManRetimeInitState(initial) != nil || Saig_ManRetimeUnsatCore(initial, 0) < 0 {
		t.Fatal("fixture does not exercise an unsatisfiable initial cut")
	}
	result := Saig_ManRetimeMinArea(g, 4, 0, 1, 1, 0)
	if result == nil {
		t.Fatal("retiming returned nil")
	}
	checkRetimeTraces(t, g, result, 2, 3)
}

func TestBackwardRetimeWithoutInitialState(t *testing.T) {
	g := aig.NewGraph()
	x, y, a, b := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	g.AddOutput(a)
	g.AddOutput(b)
	g.AddOutput(gate)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 2)
	result := Saig_ManRetimeMinArea(g, 1, 0, 1, 0, 0)
	if result == nil || result.RegisterCount() != 1 {
		t.Fatalf("backward-only no-initial retiming result: %v", result)
	}
}

func TestRetimePreservesOutputMetadata(t *testing.T) {
	g := aig.NewGraph()
	x, y, a, b := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	g.AddOutput(a)
	g.AddOutput(b)
	g.AddOutput(gate)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 2)
	aig.Aig_ManSetConstrNum(g, 1)
	g.Assertions = 1
	out := Saig_ManRetimeMinArea(g, 4, 0, 1, 1, 0)
	if out.RegisterCount() != 1 || aig.Aig_ManConstrNum(out) != 1 || out.Assertions != 1 {
		t.Fatalf("registers=%d constraints=%d assertions=%d", out.RegisterCount(), aig.Aig_ManConstrNum(out), out.Assertions)
	}
	checkRetimeTraces(t, g, out, 2, 2)
}

func TestBackwardRetimeSkipsConstrainedHiddenRegisters(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	for _, e := range []aig.Edge{a, b, c, d, x, x.Not(), y, y} {
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 4)
	aig.Aig_ManSetConstrNum(g, 1)
	out := Saig_ManRetimeMinArea(g, 4, 0, 1, 1, 0)
	if out.RegisterCount() != 4 || aig.Aig_ManConstrNum(out) != 1 {
		t.Fatalf("registers=%d constraints=%d", out.RegisterCount(), aig.Aig_ManConstrNum(out))
	}
	checkRetimeTraces(t, g, out, 2, 4)
}

func TestForwardRetimePreservesOutputMetadata(t *testing.T) {
	g := aig.NewGraph()
	x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(g.Const0())
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 2)
	aig.Aig_ManSetConstrNum(g, 1)
	g.Assertions = 1
	out := Saig_ManRetimeMinArea(g, 4, 1, 0, 1, 0)
	if out.RegisterCount() != 1 || aig.Aig_ManConstrNum(out) != 1 || out.Assertions != 1 {
		t.Fatalf("registers=%d constraints=%d assertions=%d", out.RegisterCount(), aig.Aig_ManConstrNum(out), out.Assertions)
	}
	checkRetimeTraces(t, g, out, 1, 2)
}
