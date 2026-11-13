package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand/v2"
	"testing"
)

func evaluateStep(p *aig.Graph, state, inputs []bool) ([]bool, []bool) {
	values := make([]bool, len(p.Nodes()))
	values[0] = true
	for i, e := range p.Inputs() {
		if i < trueInputs(p) {
			values[e.NodeID()] = inputs[i]
		} else {
			values[e.NodeID()] = state[i-trueInputs(p)]
		}
	}
	var eval func(aig.Edge) bool
	seen := make([]bool, len(values))
	for _, e := range p.Inputs() {
		seen[e.NodeID()] = true
	}
	seen[0] = true
	eval = func(e aig.Edge) bool {
		id := e.NodeID()
		if !seen[id] {
			node := p.Node(aig.Aig_Regular(e))
			a := eval(node.Fanins[0])
			switch node.Kind {
			case aig.And:
				values[id] = a && eval(node.Fanins[1])
			case aig.ExclusiveOr:
				values[id] = a != eval(node.Fanins[1])
			case aig.Buffer, aig.Output:
				values[id] = a
			}
			seen[id] = true
		}
		return values[id] != e.Complemented()
	}
	outputs := make([]bool, trueOutputs(p))
	next := make([]bool, registers(p))
	for i := range outputs {
		outputs[i] = eval(outputDriver(p, i))
	}
	for r := range next {
		next[r] = eval(outputDriver(p, trueOutputs(p)+r))
	}
	return outputs, next
}
func explicitSafe(p *aig.Graph) bool {
	regs, pis := registers(p), trueInputs(p)
	visited := map[int]bool{0: true}
	queue := []int{0}
	for head := 0; head < len(queue); head++ {
		state := make([]bool, regs)
		for r := range state {
			state[r] = queue[head]&(1<<uint(r)) != 0
		}
		for pattern := 0; pattern < 1<<uint(pis); pattern++ {
			inputs := make([]bool, pis)
			for i := range inputs {
				inputs[i] = pattern&(1<<uint(i)) != 0
			}
			outputs, next := evaluateStep(p, state, inputs)
			if outputs[0] {
				return false
			}
			key := 0
			for r, value := range next {
				if value {
					key |= 1 << uint(r)
				}
			}
			if !visited[key] {
				visited[key] = true
				queue = append(queue, key)
			}
		}
	}
	return true
}
func TestSATReachableImagesAgainstEnumeration(t *testing.T) {
	random := rand.New(rand.NewPCG(383, 192))
	for trial := 0; trial < 250; trial++ {
		p := aig.Aig_ManStart(40)
		pis, regs := random.IntN(3), random.IntN(4)
		signals := []aig.Edge{0, 1}
		for i := 0; i < pis+regs; i++ {
			signals = append(signals, aig.Aig_ObjCreateCi(p))
		}
		for i := 0; i < 6; i++ {
			a, b := signals[random.IntN(len(signals))], signals[random.IntN(len(signals))]
			if random.IntN(2) == 0 {
				a = a.Not()
			}
			if random.IntN(2) == 0 {
				b = b.Not()
			}
			var gate aig.Edge
			if random.IntN(2) == 0 {
				gate = aig.Aig_And(p, a, b)
			} else {
				gate = aig.Aig_Exor(p, a, b)
			}
			signals = append(signals, gate)
		}
		aig.Aig_ObjCreateCo(p, signals[random.IntN(len(signals))])
		for r := 0; r < regs; r++ {
			aig.Aig_ObjCreateCo(p, signals[random.IntN(len(signals))])
		}
		aig.Aig_ManSetRegNum(p, int32(regs))
		result := CheckSafety(p, 0, SafetyOptions{})
		images := checkSafetyImages(p, 0, SafetyOptions{})
		if images.Status != result.Status {
			t.Fatalf("trial%d PDR%d reachable images%d", trial, result.Status, images.Status)
		}
		want := int32(0)
		if explicitSafe(p) {
			want = 1
		}
		if result.Status != want {
			t.Fatalf("trial%d SAT%d explicit%d", trial, result.Status, want)
		}
		if result.Status == 0 {
			state := make([]bool, regs)
			for frame, inputs := range result.Inputs {
				for r := range state {
					if state[r] != result.States[frame][r] {
						t.Fatalf("trial%d witness state mismatch frame%d", trial, frame)
					}
				}
				outputs, next := evaluateStep(p, state, inputs)
				if frame == len(result.Inputs)-1 && !outputs[0] {
					t.Fatalf("trial%d nonfailing witness", trial)
				}
				state = next
			}
		}
	}
}

func TestCheckSafetySelectsOutputAndReplaysLatchTrace(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state.Not())
	g.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(g, 1)
	for _, tc := range []struct{ output, frames int }{{0, 2}, {1, 1}} {
		result := CheckSafety(g, tc.output, SafetyOptions{})
		if result.Status != 0 || len(result.States) != tc.frames || len(result.Inputs) != tc.frames {
			t.Fatalf("output=%d status=%d frames=%d/%d", tc.output, result.Status, len(result.States), len(result.Inputs))
		}
		current := []bool{false}
		for i, inputs := range result.Inputs {
			if len(inputs) != 0 || len(result.States[i]) != 1 || result.States[i][0] != current[0] {
				t.Fatalf("output=%d frame=%d inputs=%v state=%v", tc.output, i, inputs, result.States[i])
			}
			outputs, next := evaluateStep(g, current, inputs)
			if (i == tc.frames-1) != outputs[tc.output] {
				t.Fatalf("output=%d frame=%d values=%v", tc.output, i, outputs)
			}
			current = next
		}
	}
}
