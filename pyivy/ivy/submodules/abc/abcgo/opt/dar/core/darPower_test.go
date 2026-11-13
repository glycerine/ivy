package core
import (
	"context"
	"math"
	"sync"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func activityFixture() *aig.Graph {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	left, right := g.AND(a, b.Not()), g.AND(q, a)
	g.AddOutput(left)
	g.AddOutput(right)
	g.AddOutput(q.Not())
	aig.Aig_ManSetRegNum(g, 1)
	return g
}
func TestActivityDeterministicSequentialGolden(t *testing.T) {
	// Exact original giaSwitch.c kernel output: ten 32-bit words, 48 frames,
	// 16 settling frames, zero initial register with q'=!q.
	want := []uint32{1065353216, 1057023590, 1057002291, 1056964608, 1048641536, 1048602214, 1048641536, 1048602214, 1056964608}
	check := func() {
		g := activityFixture()
		got, err := SimulateActivity(nil, g, DefaultActivityOptions())
		if err != nil {
			t.Error(err)
			return
		}
		if len(got) != len(want) {
			t.Error(len(got))
			return
		}
		for i, p := range got {
			if math.Float32bits(p) != want[i] {
				t.Errorf("node %d got %08x want %08x", i, math.Float32bits(p), want[i])
			}
		}
	}
	check()
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() { defer workers.Done(); check() }()
	}
	workers.Wait()
}
func TestActivityStableIDsAndNativeXor(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	old := g.AND(a, b)
	g.AddOutput(old)
	xor := g.NewXor(c, d)
	aig.Aig_ObjReplace(g, old, xor.Not(), 0)
	probabilities, err := SimulateActivity(nil, g, DefaultActivityOptions())
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(probabilities[old.NodeID()]+probabilities[xor.NodeID()]-1)) > 1e-7 {
		t.Fatal("buffer phase lost", probabilities)
	}
	if probabilities[old.NodeID()] != probabilities[g.Outputs()[0].NodeID()] {
		t.Fatal("output probability lost")
	}
}
func TestPowerModeRejectsIncreasingCandidates(t *testing.T) {
	g := aig.NewGraph()
	a, q, r, s := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, q), g.AND(r, s))
	g.AddOutput(root)
	g.AddOutput(1)
	g.AddOutput(1)
	g.AddOutput(1)
	aig.Aig_ManSetRegNum(g, 3)
	params := DefaultRewriteParams()
	params.Power = true
	params.UseZeros = true
	m, err := Dar_ManStart(g, &params)
	if err != nil {
		t.Fatal(err)
	}
	defer Dar_ManStop(m)
	if err = m.Rewrite(); err != nil {
		t.Fatal(err)
	}
	if m.Stats.PowerRejected == 0 {
		t.Fatal("power mode did not reject a cost-increasing candidate", m.Stats)
	}
	for assignment := 0; assignment < 16; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		out, err := g.Evaluate(values)
		if err != nil || out[0] != (assignment == 15) {
			t.Fatal(assignment, out, err)
		}
	}
}
func TestPowerRewriteCanImproveLogic(t *testing.T) {
	g := aig.NewGraph()
	inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
	result := aig.Edge(1)
	for minterm := 0; minterm < 16; minterm++ {
		if bitsOnes(minterm)%2 == 0 {
			continue
		}
		term := aig.Edge(0)
		for variable, e := range inputs {
			if minterm&(1<<variable) == 0 {
				e = e.Not()
			}
			term = g.AND(term, e)
		}
		result = g.OR(result, term)
	}
	g.AddOutput(result)
	before := aig.Aig_ManDupDfs(g).NodeCount()
	p := DefaultRewriteParams()
	p.Power = true
	p.CutsMax = 128
	if err := Dar_ManRewrite(g, &p); err != nil {
		t.Fatal(err)
	}
	after := aig.Aig_ManDupDfs(g).NodeCount()
	if after >= before {
		t.Fatalf("power path did not optimize parity: %d -> %d", before, after)
	}
	for n := 0; n < 16; n++ {
		out, err := g.Evaluate([]bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0})
		if err != nil || out[0] != (bitsOnes(n)%2 == 1) {
			t.Fatal(n, out, err)
		}
	}
}
func bitsOnes(n int) int {
	count := 0
	for n != 0 {
		n &= n - 1
		count++
	}
	return count
}
func TestActivityBoundsAndCancellation(t *testing.T) {
	g := activityFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := SimulateActivity(ctx, g, DefaultActivityOptions()); err != context.Canceled {
		t.Fatal(err)
	}
	for _, options := range []ActivityOptions{{Words: 0, Frames: 48, Prefix: 16}, {Words: 1, Frames: 16, Prefix: 16}, {Words: math.MaxInt, Frames: 2}, {Words: math.MaxInt / 32, Frames: 1}, {Words: 1, Frames: 1, RandomFactor: -2}} {
		if _, err := SimulateActivity(nil, g, options); err == nil {
			t.Fatal("accepted bad activity dimensions", options)
		}
	}
}
