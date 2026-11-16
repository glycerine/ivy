package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func isoSequentialReviewGraph(phase int32) *Gia_Man_t {
	graph := Gia_ManStart(24)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	qa, ra := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	qb, rb := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	x, y := Gia_ManAppendAnd(graph, qa, a), Gia_ManAppendAnd(graph, qb, b)
	nextA, nextB := Gia_ManAppendAnd(graph, a, ra^1), Gia_ManAppendAnd(graph, b, rb^1)
	for _, root := range []int32{x ^ phase, y ^ phase, x ^ (phase ^ 1), phase} {
		Gia_ManAppendCo(graph, root)
	}
	for _, root := range []int32{nextA, qa, nextB, qb} {
		Gia_ManAppendCo(graph, root)
	}
	Gia_ManSetRegNum(graph, 4)
	return graph
}

func TestIsoSequentialReductionAgainstIndependentTransitionSystem(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		for phase := int32(0); phase < 2; phase++ {
			for estimate := int32(0); estimate < 2; estimate++ {
				t.Run(fmt.Sprintf("mode%d/phase%d/estimate%d", mode, phase, estimate), func(t *testing.T) {
					source := isoSequentialReviewGraph(phase)
					var groups, permutations *Vec_Ptr_t
					var result *Gia_Man_t
					isoReviewCapture(t, func() {
						if mode == 0 {
							result = Gia_ManIsoReduce(source, &groups, &permutations, estimate, 0, 0, 0)
						} else {
							result = Gia_ManIsoReduce2(source, &groups, &permutations, estimate, int32(mode-1), 0, 0, 0)
						}
					})
					if result == nil {
						t.Fatal("sequential reduction failed")
					}
					defer Gia_ManStop(result)
					representatives := []int{0, 2, 3}
					registers := int32(2)
					if estimate != 0 {
						representatives = []int{0, 1, 2, 3}
						registers = 4
					}
					if Gia_ManPiNum(result) != 2 || Gia_ManRegNum(result) != registers || Gia_ManPoNum(result) != int32(len(representatives)) {
						t.Fatalf("wrong sequential interface: PIs%d regs%d POs%d", Gia_ManPiNum(result), Gia_ManRegNum(result), Gia_ManPoNum(result))
					}
					if estimate == 0 {
						want := [][]int32{{0, 1}, {2}, {3}}
						if groups == nil || vec.Vec_PtrSize(groups) != len(want) {
							t.Fatal("sequential equivalence class count differs")
						}
						for i, class := range want {
							if got := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, i).(*Vec_Int_t)); !slices.Equal(got, class) {
								t.Fatalf("class%d got%v want%v", i, got, class)
							}
						}
						if mode == 0 {
							for output := 0; output < 2; output++ {
								permutation, ok := vec.Vec_PtrEntry(permutations, output).(*Vec_Int_t)
								if !ok || permutation == nil || !slices.Equal(vec.Vec_IntArray(permutation), []int32{0}) {
									t.Fatalf("output%d did not return its cone-local PI permutation", output)
								}
							}
						} else if permutations != nil {
							t.Fatal("second-generation reduction invented PI permutations")
						}
					}
					for initial := 0; initial < 16; initial++ {
						for sequence := 0; sequence < 64; sequence++ {
							state := []bool{initial&1 != 0, initial&2 != 0, initial&4 != 0, initial&8 != 0}
							actualState := slices.Clone(state[:registers])
							for frame := 0; frame < 3; frame++ {
								a, b := sequence&(1<<(2*frame)) != 0, sequence&(1<<(2*frame+1)) != 0
								qa, ra, qb, rb := state[0], state[1], state[2], state[3]
								want := []bool{(qa && a) != (phase != 0), (qb && b) != (phase != 0), (qa && a) == (phase != 0), phase != 0}
								input := append([]bool{a, b}, actualState...)
								values, err := giaEvaluateNodes(result, input)
								if err != nil {
									t.Fatal(err)
								}
								for output, original := range representatives {
									if values[Gia_ManPo(result, int32(output)).id] != want[original] {
										t.Fatalf("initial%d sequence%d frame%d output%d changed", initial, sequence, frame, output)
									}
								}
								state = []bool{a && !ra, qa, b && !rb, qb}
								for register := range actualState {
									actualState[register] = values[Gia_ManRi(result, int32(register)).id]
									if actualState[register] != state[register] {
										t.Fatalf("initial%d sequence%d frame%d register%d transition changed", initial, sequence, frame, register)
									}
								}
							}
						}
					}
				})
			}
		}
	}
}

func TestIsoSequentialCanonicalKeysKeepRegisterCones(t *testing.T) {
	for phase := int32(0); phase < 2; phase++ {
		t.Run(fmt.Sprintf("phase%d", phase), func(t *testing.T) {
			source := isoSequentialReviewGraph(phase)
			var keys []string
			for output := int32(0); output < 4; output++ {
				var permutation *Vec_Int_t
				key := Gia_ManIsoFindString(source, output, 0, &permutation)
				if key == nil || permutation == nil {
					t.Fatalf("output%d has no canonical representation", output)
				}
				keys = append(keys, string(vec.Vec_StrArray(key)))
				expected := []int32{0}
				if output == 3 {
					expected = nil
				}
				if !slices.Equal(vec.Vec_IntArray(permutation), expected) {
					t.Fatalf("output%d PI permutation%v want%v", output, vec.Vec_IntArray(permutation), expected)
				}
			}
			if keys[0] != keys[1] || keys[0] == keys[2] || keys[0] == keys[3] || keys[2] == keys[3] {
				t.Fatal("sequential canonical keys lose register or output-phase distinctions")
			}
		})
	}
}
