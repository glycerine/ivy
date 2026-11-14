package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func truthNetwork(t *testing.T, variables int, truths ...uint64) *abc.Network {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	inputs := make([]*abc.Object, variables)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(n)
		inputs[i].SetName(fmt.Sprintf("x%d", i))
	}
	for i, truth := range truths {
		root := n.One().Not()
		for bits := 0; bits < 1<<variables; bits++ {
			if truth>>bits&1 == 0 {
				continue
			}
			term := n.One()
			for k, input := range inputs {
				edge := input.Signal()
				if bits>>k&1 == 0 {
					edge = edge.Not()
				}
				term = n.AndEdges(term, edge)
			}
			root = n.AndEdges(root.Not(), term.Not()).Not()
		}
		output := abc.Abc_NtkCreatePo(n)
		output.SetName(fmt.Sprintf("y%d", i))
		abc.Abc_ObjAddFanin(output, root)
	}
	return n
}

func evaluateNetwork(t *testing.T, n *abc.Network, bits int) []bool {
	t.Helper()
	inputs := map[*abc.Object]bool{}
	for i, input := range n.Inputs() {
		inputs[input] = bits>>i&1 != 0
	}
	result := make([]bool, len(n.Outputs()))
	for i, output := range n.Outputs() {
		var err error
		result[i], err = n.Evaluate(output.Signal(), inputs)
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func TestQuantificationExhaustiveThreeVariableFunctions(t *testing.T) {
	for truth := uint64(0); truth < 256; truth++ {
		for variable := 0; variable < 3; variable++ {
			for _, universal := range []int32{0, 1} {
				n := truthNetwork(t, 3, truth, ^truth&255)
				if Abc_NtkQuantify(n, universal, int32(variable), 0) != 1 {
					t.Fatal("quantify failed")
				}
				for bits := 0; bits < 8; bits++ {
					low, high := bits & ^(1<<variable), bits|1<<variable
					for i, got := range evaluateNetwork(t, n, bits) {
						f := truth
						if i == 1 {
							f = ^f
						}
						a, b := f>>low&1 != 0, f>>high&1 != 0
						want := a || b
						if universal != 0 {
							want = a && b
						}
						if got != want {
							t.Fatalf("truth=%02x variable=%d universal=%d bits=%d output=%d got=%v want=%v", truth, variable, universal, bits, i, got, want)
						}
					}
				}
				abc.Abc_NtkDelete(n)
			}
		}
	}
}

func TestMiterQuantificationAndCombination(t *testing.T) {
	for _, truth := range []uint64{0, 255, 0xaa, 0x96, 0xe8, 0x81, 0xfe} {
		for variable := int32(0); variable < 3; variable++ {
			for _, exist := range []int32{0, 1} {
				original := truthNetwork(t, 3, truth)
				result := Abc_NtkMiterQuantify(original, variable, exist)
				if result == nil {
					t.Fatal("nil quantified miter")
				}
				for bits := 0; bits < 8; bits++ {
					a, b := truth>>(bits & ^(1<<variable))&1 != 0, truth>>(bits|1<<variable)&1 != 0
					want := a && b
					if exist != 0 {
						want = a || b
					}
					if got := evaluateNetwork(t, result, bits)[0]; got != want {
						t.Fatalf("miter truth=%x variable=%d exist=%d bits=%d got=%v want=%v", truth, variable, exist, bits, got, want)
					}
				}
				abc.Abc_NtkDelete(result)
				abc.Abc_NtkDelete(original)
			}
		}
	}
	for _, mode := range []struct{ and, xor int32 }{{0, 0}, {1, 0}, {0, 1}} {
		n := truthNetwork(t, 3, 0xaa, 0xcc, 0xf0)
		if Abc_NtkCombinePos(n, mode.and, mode.xor) != 1 {
			t.Fatal("combine failed")
		}
		for bits := 0; bits < 8; bits++ {
			a, b, c := bits&1 != 0, bits&2 != 0, bits&4 != 0
			want := a || b || c
			if mode.and != 0 {
				want = a && b && c
			}
			if mode.xor != 0 {
				want = (a != b) != c
			}
			if got := evaluateNetwork(t, n, bits)[0]; got != want {
				t.Fatalf("mode=%v bits=%d got=%v want=%v", mode, bits, got, want)
			}
		}
		abc.Abc_NtkDelete(n)
	}
}

func TestReachabilityAllOneBitRelations(t *testing.T) {
	for relation := uint64(0); relation < 16; relation++ {
		n := truthNetwork(t, 2, relation)
		result := Abc_NtkReachability(n, 4, 0)
		if result == nil {
			t.Fatalf("relation=%x returned nil", relation)
		}
		reached := uint8(1)
		for step := 0; step < 4; step++ {
			next := reached
			for state := 0; state < 2; state++ {
				if reached>>state&1 == 0 {
					continue
				}
				for successor := 0; successor < 2; successor++ {
					if relation>>(state+2*successor)&1 != 0 {
						next |= 1 << successor
					}
				}
			}
			reached = next
		}
		for state := 0; state < 2; state++ {
			if got := evaluateNetwork(t, result, state)[0]; got != (reached>>state&1 == 0) {
				t.Fatalf("relation=%x state=%d unreachable=%v reached=%02b", relation, state, got, reached)
			}
		}
		abc.Abc_NtkDelete(result)
		abc.Abc_NtkDelete(n)
	}
}

func TestSequentialFramesAndTransitionRelation(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	defer abc.Abc_NtkDelete(n)
	input := abc.Abc_NtkCreatePi(n)
	input.SetName("input")
	state := abc.Abc_NtkAddLatch(n, nil, abc.ABC_INIT_ONE)
	latch := abc.Abc_ObjFanin0(state)
	li := abc.Abc_ObjFanin0(latch)
	state.SetName("state")
	latch.SetName("reg")
	li.SetName("next")
	next := n.AndEdges(input.Signal(), state.Signal().Not())
	abc.Abc_ObjAddFanin(li, next)
	output := abc.Abc_NtkCreatePo(n)
	output.SetName("out")
	abc.Abc_ObjAddFanin(output, state)
	abc.Abc_NtkOrderCisCos(n)
	for _, initial := range []int32{0, 1} {
		framed := Abc_NtkFrames(n, 3, initial, 0)
		if framed == nil {
			t.Fatal("framing failed")
		}
		for bits := 0; bits < 1<<len(framed.Inputs()); bits++ {
			got := evaluateNetwork(t, framed, bits)
			current := true
			if initial == 0 {
				current = bits>>3&1 != 0
			}
			for frame := 0; frame < 3; frame++ {
				if got[frame] != current {
					t.Fatalf("initial=%d bits=%d frame=%d got=%v want=%v", initial, bits, frame, got[frame], current)
				}
				current = bits>>frame&1 != 0 && !current
			}
			if initial == 0 && got[3] != current {
				t.Fatal("last latch input mismatch")
			}
		}
		abc.Abc_NtkDelete(framed)
	}
	relation := Abc_NtkTransRel(n, 0, 0)
	if relation == nil {
		t.Fatal("relation failed")
	}
	defer abc.Abc_NtkDelete(relation)
	for bits := 0; bits < 8; bits++ {
		current, next, input := bits&1 != 0, bits&2 != 0, bits&4 != 0
		want := next == (input && !current)
		if got := evaluateNetwork(t, relation, bits)[0]; got != want {
			t.Fatalf("bits=%d relation=%v want=%v", bits, got, want)
		}
	}
}
