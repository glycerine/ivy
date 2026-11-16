package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

var isoPipelineReviewGroups = [][]int32{{0, 1, 10}, {2, 3, 11}, {4, 5}, {6}, {7}, {8, 9}}
var isoPipelineReviewSupports = [][]int{{0, 1}, {2, 3}, {0, 1}, {2, 3}, {0}, {2}, nil, nil, {0, 1, 4}, {2, 3, 5}, {0, 1}, {0, 1}}

func isoPipelineReviewOutputs(input []bool, phase int32) []bool {
	a, b, c, d, e, f := input[0], input[1], input[2], input[3], input[4], input[5]
	outputs := []bool{a && b, c && d, a && !b, c && !d, a, c, false, true, a && b && e, c && d && f, a && b, a && !b}
	if phase != 0 {
		for i := range outputs {
			outputs[i] = !outputs[i]
		}
	}
	return outputs
}

func checkIsoPipelineReviewTruth(t *testing.T, graph *Gia_Man_t, representatives []int32, phase int32) {
	t.Helper()
	if Gia_ManCiNum(graph) != 6 || Gia_ManPoNum(graph) != int32(len(representatives)) || Gia_ManRegNum(graph) != 0 {
		t.Fatal("reduction changed the interface incorrectly")
	}
	for assignment := 0; assignment < 64; assignment++ {
		input := make([]bool, 6)
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		want := isoPipelineReviewOutputs(input, phase)
		actual, err := giaEvaluateNodes(graph, input)
		if err != nil {
			t.Fatal(err)
		}
		for output, representative := range representatives {
			if actual[Gia_ManPo(graph, int32(output)).id] != want[representative] {
				t.Fatalf("assignment%d output%d does not retain representative%d", assignment, output, representative)
			}
		}
	}
}

func TestIsoReductionPipelinesAgainstKnownCircuitClasses(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		for phase := int32(0); phase < 2; phase++ {
			for _, estimate := range []int32{0, 1} {
				for _, metadata := range []bool{false, true} {
					for verbosity := 0; verbosity < 3; verbosity++ {
						t.Run(fmt.Sprintf("mode%d/phase%d/estimate%d/metadata%t/verbosity%d", mode, phase, estimate, metadata, verbosity), func(t *testing.T) {
							graph := iso2ClassesReviewGraph(phase)
							graph.pName, graph.PSpec = "iso-pipeline", "pipeline-spec"
							oldObjects := Gia_ManObjNum(graph)
							groups, perms := vec.Vec_PtrAlloc(0), vec.Vec_PtrAlloc(0)
							oldGroups, oldPerms := groups, perms
							var groupArg, permArg **Vec_Ptr_t
							if metadata {
								groupArg, permArg = &groups, &perms
							}
							var result *Gia_Man_t
							verbose, veryVerbose := int32(0), int32(0)
							if verbosity == 1 {
								verbose = 1
							} else if verbosity == 2 {
								veryVerbose = 1
							}
							isoReviewCapture(t, func() {
								if mode == 0 {
									result = Gia_ManIsoReduce(graph, groupArg, permArg, estimate, 0, verbose, veryVerbose)
								} else {
									// The source's second variant ignores its dual flag.
									result = Gia_ManIsoReduce2(graph, groupArg, permArg, estimate, int32(mode-1), 1, verbose, veryVerbose)
								}
							})
							if result == nil || result == graph {
								t.Fatal("pipeline did not return an owned graph")
							}
							defer Gia_ManStop(result)
							if result.pName != graph.pName || result.PSpec != graph.PSpec || Gia_ManObjNum(graph) != oldObjects {
								t.Fatal("pipeline lost identity or changed source object count")
							}
							var representatives []int32
							if estimate != 0 {
								for i := int32(0); i < 12; i++ {
									representatives = append(representatives, i)
								}
							} else {
								for _, group := range isoPipelineReviewGroups {
									representatives = append(representatives, group[0])
								}
							}
							checkIsoPipelineReviewTruth(t, result, representatives, phase)
							if !metadata {
								return
							}
							if estimate != 0 {
								if mode != 0 {
									if groups != oldGroups || perms != oldPerms {
										t.Fatal("second-generation estimate changed output arguments")
									}
								} else {
									if groups != nil || perms == nil || vec.Vec_PtrSize(perms) != 12 {
										t.Fatal("first-generation estimate did not initialize metadata")
									}
									for _, entry := range vec.Vec_PtrArray(perms) {
										if entry != nil {
											t.Fatal("estimate computed a permutation")
										}
									}
								}
								return
							}
							if mode == 0 && (perms == nil || vec.Vec_PtrSize(perms) != 12) {
								t.Fatal("first-generation reduction lost permutation slots")
							}
							if groups == nil || vec.Vec_PtrSize(groups) != len(isoPipelineReviewGroups) {
								t.Fatal("wrong equivalence class count")
							}
							for i, want := range isoPipelineReviewGroups {
								if got := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, i).(*Vec_Int_t)); !slices.Equal(got, want) {
									t.Fatalf("class%d got%v want%v", i, got, want)
								}
								if mode != 0 || len(want) < 2 {
									continue
								}
								for _, output := range want {
									permutation, ok := vec.Vec_PtrEntry(perms, int(output)).(*Vec_Int_t)
									if !ok || permutation == nil {
										t.Fatalf("output%d missing input permutation", output)
									}
									ids := vec.Vec_IntArray(permutation)
									if len(ids) != len(isoPipelineReviewSupports[output]) {
										t.Fatalf("output%d permutation has wrong support size", output)
									}
									seen := make([]bool, len(ids))
									for _, id := range ids {
										if id < 0 || int(id) >= len(ids) || seen[id] {
											t.Fatalf("output%d permutation%v is not a bijection", output, ids)
										}
										seen[id] = true
									}
								}
							}
							if mode != 0 && perms != nil {
								t.Fatal("second-generation reduction invented PI permutations")
							}
							// Metadata must outlive both managers, as the C result ownership implies.
							Gia_ManStop(graph)
							Gia_ManStop(result)
							if !slices.Equal(vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 0).(*Vec_Int_t)), []int32{0, 1, 10}) {
								t.Fatal("result metadata aliases manager storage")
							}
						})
					}
				}
			}
		}
	}
}

func TestIsoCanonicalKeysAndConeLocalPermutations(t *testing.T) {
	for phase := int32(0); phase < 2; phase++ {
		t.Run(fmt.Sprintf("phase%d", phase), func(t *testing.T) {
			graph := iso2ClassesReviewGraph(phase)
			keys := make([]string, 12)
			normalized := make([]uint64, 12)
			for output, support := range isoPipelineReviewSupports {
				var permutation *Vec_Int_t
				encoded := Gia_ManIsoFindString(graph, int32(output), 0, &permutation)
				if encoded == nil || permutation == nil {
					t.Fatalf("output%d missing encoding or permutation", output)
				}
				keys[output] = string(vec.Vec_StrArray(encoded))
				if len(keys[output]) == 0 || vec.Vec_IntSize(permutation) != len(support) {
					t.Fatalf("output%d invalid encoding/support", output)
				}
				without := Gia_ManIsoFindString(graph, int32(output), 0, nil)
				if without == nil || string(vec.Vec_StrArray(without)) != keys[output] {
					t.Fatal("requesting permutations changes canonical encoding")
				}
				seen := make([]bool, len(support))
				for _, local := range vec.Vec_IntArray(permutation) {
					if local < 0 || int(local) >= len(support) || seen[local] {
						t.Fatalf("output%d invalid cone-local permutation", output)
					}
					seen[local] = true
				}
				for assignment := 0; assignment < 1<<len(support); assignment++ {
					input := make([]bool, 6)
					for canonical, local := range vec.Vec_IntArray(permutation) {
						input[support[local]] = assignment&(1<<canonical) != 0
					}
					if isoPipelineReviewOutputs(input, phase)[output] {
						normalized[output] |= 1 << assignment
					}
				}
			}
			for a := range keys {
				for b := range keys {
					same := iso2ClassesReviewKinds[a] == iso2ClassesReviewKinds[b]
					if (keys[a] == keys[b]) != same {
						t.Fatalf("outputs%d/%d canonical key equality does not match circuit kinds", a, b)
					}
					if same && normalized[a] != normalized[b] {
						t.Fatalf("outputs%d/%d permutations do not align Boolean functions", a, b)
					}
				}
			}
		})
	}
}

func TestIsoPipelinesDistinguishStructureFromBooleanEquivalence(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		for phase := int32(0); phase < 2; phase++ {
			t.Run(fmt.Sprintf("mode%d/phase%d", mode, phase), func(t *testing.T) {
				graph := Gia_ManStart(20)
				a, b, c := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
				ab, ac, bc := Gia_ManAppendAnd(graph, a, b), Gia_ManAppendAnd(graph, a, c), Gia_ManAppendAnd(graph, b, c)
				// The first graph shares a PI across two internal gates. All three
				// functions are a&&b&&c, but only the latter two graphs are isomorphic.
				reconvergent := Gia_ManAppendAnd(graph, ab, ac)
				chain0, chain1 := Gia_ManAppendAnd(graph, ab, c), Gia_ManAppendAnd(graph, a, bc)
				for _, root := range []int32{reconvergent, chain0, chain1} {
					Gia_ManAppendCo(graph, root^phase)
				}
				var groups *Vec_Ptr_t
				var result *Gia_Man_t
				isoReviewCapture(t, func() {
					if mode == 0 {
						result = Gia_ManIsoReduce(graph, &groups, nil, 0, 0, 0, 0)
					} else {
						result = Gia_ManIsoReduce2(graph, &groups, nil, 0, int32(mode-1), 0, 0, 0)
					}
				})
				if result == nil || Gia_ManPoNum(result) != 2 || groups == nil || vec.Vec_PtrSize(groups) != 2 {
					t.Fatal("structural grouping was replaced by Boolean equivalence")
				}
				for i, want := range [][]int32{{0}, {1, 2}} {
					if got := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, i).(*Vec_Int_t)); !slices.Equal(got, want) {
						t.Fatalf("structural class%d got%v want%v", i, got, want)
					}
				}
				for assignment := 0; assignment < 8; assignment++ {
					input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
					want := (input[0] && input[1] && input[2]) != (phase != 0)
					actual, err := giaEvaluateNodes(result, input)
					if err != nil {
						t.Fatal(err)
					}
					for output := int32(0); output < 2; output++ {
						if actual[Gia_ManPo(result, output).id] != want {
							t.Fatalf("assignment%d changed Boolean truth", assignment)
						}
					}
				}
				var keys []string
				for output := int32(0); output < 3; output++ {
					key := Gia_ManIsoFindString(graph, output, 0, nil)
					if key == nil {
						t.Fatal("canonical encoding failed")
					}
					keys = append(keys, string(vec.Vec_StrArray(key)))
				}
				if keys[0] == keys[1] || keys[1] != keys[2] {
					t.Fatal("canonical keys do not preserve graph structure")
				}
			})
		}
	}
}
