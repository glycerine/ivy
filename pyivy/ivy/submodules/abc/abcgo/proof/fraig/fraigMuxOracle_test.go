package fraig

import "testing"

// Recognition feeds the SAT encoding, so verify its returned decomposition
// against all inputs and all complemented control/data combinations.
func TestMuxRecognitionAgainstTruthTable(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	inputs := []*Fraig_Node_t{Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)}
	for c := 0; c < 3; c++ {
		for d := 0; d < 3; d++ {
			if d == c {
				continue
			}
			e := 3 - c - d
			for phase := 0; phase < 8; phase++ {
				args := []*Fraig_Node_t{inputs[c], inputs[d], inputs[e]}
				for i := range args {
					if phase&(1<<i) != 0 {
						args[i] = Fraig_Not(args[i])
					}
				}
				root := Fraig_Regular(Fraig_NodeMux(p, args[0], args[1], args[2]))
				if Fraig_NodeIsMuxType(root) == 0 {
					t.Fatalf("mux not recognized: %d %d %d", c, d, phase)
				}
				var yes, no *Fraig_Node_t
				control := Fraig_NodeRecognizeMux(root, &yes, &no)
				if Fraig_IsComplement(control) {
					t.Fatal("control is complemented")
				}
				for assignment := 0; assignment < 8; assignment++ {
					model := []int32{int32(assignment & 1), int32(assignment >> 1 & 1), int32(assignment >> 2 & 1)}
					eval := func(n *Fraig_Node_t) int32 { return Fraig_ManSimulateBitNode(p, n, model) }
					want := eval(no)
					if eval(control) != 0 {
						want = eval(yes)
					}
					if got := eval(root); got != want {
						t.Fatalf("control=%d data=%d phase=%d assignment=%d got=%d want=%d", c, d, phase, assignment, got, want)
					}
				}
			}
		}
	}
}

func TestXorRecognitionAgainstTruthTable(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	for phase := 0; phase < 4; phase++ {
		x, y := a, b
		if phase&1 != 0 {
			x = Fraig_Not(x)
		}
		if phase&2 != 0 {
			y = Fraig_Not(y)
		}
		root := Fraig_Regular(Fraig_NodeExor(p, x, y))
		if Fraig_NodeIsExorType(root) == 0 {
			t.Fatalf("xor not recognized phase %d", phase)
		}
		xor := Fraig_NodeIsExor(root) != 0
		for assignment := 0; assignment < 4; assignment++ {
			model := []int32{int32(assignment & 1), int32(assignment >> 1 & 1)}
			got := Fraig_ManSimulateBitNode(p, root, model) != 0
			want := (model[0] != model[1]) == xor
			if got != want {
				t.Fatalf("phase %d assignment %d xor=%t got=%t", phase, assignment, xor, got)
			}
		}
	}
}

// A supergate boundary is valid only if its leaves' conjunction retains the
// root function, including complemented and shared subgraphs.
func TestCollectSupergateAgainstTruthTable(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c, d := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2), Fraig_ManReadIthVar(p, 3)
	shared := Fraig_NodeAnd(p, a, b)
	mux := Fraig_NodeMux(p, c, shared, d)
	roots := []*Fraig_Node_t{shared, Fraig_Not(shared), Fraig_NodeAnd(p, shared, c), Fraig_NodeAnd(p, shared, Fraig_Not(d)), Fraig_Regular(mux), Fraig_NodeAnd(p, mux, shared)}
	for _, root := range roots {
		for _, stop := range []int32{0, 1} {
			leaves := Fraig_CollectSupergate(root, stop)
			if len(leaves.pArray) == 0 {
				t.Fatal("empty supergate")
			}
			for assignment := 0; assignment < 16; assignment++ {
				model := []int32{int32(assignment & 1), int32(assignment >> 1 & 1), int32(assignment >> 2 & 1), int32(assignment >> 3 & 1)}
				want := int32(1)
				for _, leaf := range leaves.pArray {
					want &= Fraig_ManSimulateBitNode(p, leaf, model)
				}
				if got := Fraig_ManSimulateBitNode(p, root, model); got != want {
					t.Fatalf("stop=%d assignment=%d got=%d want=%d", stop, assignment, got, want)
				}
			}
		}
	}
}
