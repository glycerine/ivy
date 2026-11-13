package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand/v2"
	"testing"
)

func TestSafetyRespectsConstraintViolations(t *testing.T) {
	// x=1 would make the bad latch true, but it violates the constraint.
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	for _, options := range []SafetyOptions{{}} {
		result := CheckSafety(g, 0, options)
		if result.Status != 1 {
			t.Fatalf("options=%+v got status %d error %v", options, result.Status, result.Error)
		}
	}
}

func TestSafetyConstrainedWitnessAndInvalidSelection(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x.Not())
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	for _, options := range []SafetyOptions{{}} {
		result := CheckSafety(g, 0, options)
		if result.Status != 0 || len(result.Inputs) != 2 || len(result.States) != 2 {
			t.Fatalf("options=%+v result=%+v", options, result)
		}
		for frame, inputs := range result.Inputs {
			if len(inputs) != 1 || !inputs[0] {
				t.Fatalf("illegal input at frame %d: %v", frame, inputs)
			}
		}
		if !result.States[1][0] {
			t.Fatal("missing reached state")
		}
		if bad := CheckSafety(g, 1, options); bad.Status != -1 || bad.Error == nil {
			t.Fatalf("constraint output accepted as property: %+v", bad)
		}
	}
}

func explicitConstrainedSafe(g *aig.Graph) bool {
	seen := map[int]bool{0: true}
	queue := []int{0}
	regs, pis := registers(g), trueInputs(g)
	for head := 0; head < len(queue); head++ {
		state := make([]bool, regs)
		for r := range state {
			state[r] = queue[head]&(1<<uint(r)) != 0
		}
		for mask := 0; mask < 1<<uint(pis); mask++ {
			inputs := make([]bool, pis)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			outs, next := evaluateStep(g, state, inputs)
			legal := true
			for i := trueOutputs(g) - int(aig.Aig_ManConstrNum(g)); i < trueOutputs(g); i++ {
				if outs[i] {
					legal = false
				}
			}
			if !legal {
				continue
			}
			if outs[0] {
				return false
			}
			key := 0
			for i, v := range next {
				if v {
					key |= 1 << uint(i)
				}
			}
			if !seen[key] {
				seen[key] = true
				queue = append(queue, key)
			}
		}
	}
	return true
}
func TestConstrainedSafetyAgainstReachability(t *testing.T) {
	random := rand.New(rand.NewPCG(817, 233))
	for trial := 0; trial < 100; trial++ {
		g := aig.NewGraph()
		signals := []aig.Edge{g.Const0(), g.Const1()}
		pis, regs := random.IntN(3), random.IntN(3)
		for i := 0; i < pis+regs; i++ {
			signals = append(signals, g.NewInput())
		}
		for i := 0; i < 6; i++ {
			a, b := signals[random.IntN(len(signals))], signals[random.IntN(len(signals))]
			if random.IntN(2) == 0 {
				a = a.Not()
			}
			if random.IntN(2) == 0 {
				b = b.Not()
			}
			if random.IntN(2) == 0 {
				signals = append(signals, g.AND(a, b))
			} else {
				signals = append(signals, g.XOR(a, b))
			}
		}
		g.AddOutput(signals[random.IntN(len(signals))])
		for i := 0; i < 2; i++ {
			g.AddOutput(signals[random.IntN(len(signals))])
		}
		for i := 0; i < regs; i++ {
			g.AddOutput(signals[random.IntN(len(signals))])
		}
		aig.Aig_ManSetRegNum(g, int32(regs))
		aig.Aig_ManSetConstrNum(g, 2)
		want := int32(0)
		if explicitConstrainedSafe(g) {
			want = 1
		}
		for _, options := range []SafetyOptions{{}} {
			got := CheckSafety(g, 0, options)
			if got.Status != want {
				t.Fatalf("trial %d opts=%+v got %d want %d err=%v", trial, options, got.Status, want, got.Error)
			}
			if got.Status == 0 {
				for f, in := range got.Inputs {
					outs, _ := evaluateStep(g, got.States[f], in)
					if outs[1] || outs[2] {
						t.Fatalf("trial %d invalid witness frame %d", trial, f)
					}
				}
			}
		}
	}
}

func TestLivenessReductionPreservesConstraints(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	network, err := NewNetwork(g, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	reduced, err := LivenessToSafety(network, TransformOptions{Mode: IGNORE_LIVENESS_KEEP_SAFETY_MODE, SafetyAssertions: []aig.Edge{state.Not()}})
	if err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ManConstrNum(reduced.Graph) != 1 {
		t.Fatal("constraint was dropped")
	}
	if got := CheckSafety(reduced.Graph, 0, SafetyOptions{}); got.Status != 1 {
		t.Fatalf("constrained reduction status %d error %v", got.Status, got.Error)
	}
}
