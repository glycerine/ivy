package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func structureReviewVectors(n int) (*Gia_Man_t, *Vec_Int_t, *Vec_Int_t, *Vec_Int_t, *Vec_Int_t, *Vec_Int_t) {
	dst := Gia_ManStart(64)
	leaves, temp, mapping := Vec_IntAlloc(n), Vec_IntAlloc(n), Vec_IntAlloc(0)
	records, packing := vec.Vec_IntAllocArrayCopy([]int32{0}, 1), vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	for pin := 0; pin < n; pin++ {
		Vec_IntPush(leaves, Gia_ManAppendCi(dst)^int32(pin%2))
	}
	return dst, leaves, temp, mapping, records, packing
}

func TestStructureElementaryAndConstantRecords(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 15, NCutsMax: 4})
	defer If_ManStop(p)
	for n := 4; n <= 16; n++ {
		structure := "666"
		if n <= 5 {
			structure = "33"
		} else if n == 6 {
			structure = "44"
		}
		dst, leaves, temp, mapping, records, packing := structureReviewVectors(n)
		for variable := 0; variable < n; variable++ {
			for phase := 0; phase < 2; phase++ {
				t.Run(fmt.Sprintf("inputs%d/variable%d/phase%d", n, variable, phase), func(t *testing.T) {
					truth := make([]word, 1<<max(0, n-6))
					for x := 0; x < 64*len(truth); x++ {
						if (x>>variable&1)^phase != 0 {
							truth[x>>6] |= 1 << uint(x&63)
						}
					}
					before := slices.Clone(truth)
					got := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, structure, nil, mapping, records, packing, 0, 0)
					want := Vec_IntEntry(leaves, variable) ^ int32(phase)
					if got != want || Vec_IntSize(records) != 1 || Vec_IntSize(mapping) != 0 || !slices.Equal(Vec_IntArray(packing), []int32{0}) || Gia_ManObjNum(dst) != int32(n+1) || !slices.Equal(truth, before) {
						t.Fatalf("elementary return%d want%d records%v", got, want, Vec_IntArray(records))
					}
				})
			}
		}
		for phase := 0; phase < 4; phase++ {
			truth := make([]word, 1<<max(0, n-6))
			if phase&1 != 0 {
				for i := range truth {
					truth[i] = ^word(0)
				}
			}
			got := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, structure, nil, mapping, records, packing, 0, 0)
			if got != int32(phase&1) || !slices.Equal(Vec_IntArray(records), []int32{0, 0, 0}) || Vec_IntEntry(mapping, 0) != 1 || !slices.Equal(Vec_IntArray(packing), []int32{0}) {
				t.Fatalf("inputs%d constant%d got%d records%v", n, phase, got, Vec_IntArray(records))
			}
		}
		Gia_ManStop(dst)
		for _, v := range []*Vec_Int_t{leaves, temp, mapping, records, packing} {
			vec.Vec_IntFree(v)
		}
	}
}

func TestStructureUnequalSizesAndThreeLuts(t *testing.T) {
	for _, structure := range []string{"33", "34", "43", "44", "45", "54", "55", "66", "333", "343", "444", "456", "654", "666"} {
		leaf, root := int(structure[0]-'0'), int(structure[len(structure)-1]-'0')
		capacity, largest := leaf-1+root, max(leaf, root)
		if len(structure) == 3 {
			middle := int(structure[1] - '0')
			capacity += middle - 1
			largest = max(largest, middle)
		}
		for _, n := range []int{largest + 1, min(15, capacity)} {
			for family := 0; family < 2; family++ {
				for phase := 0; phase < 4; phase++ {
					t.Run(fmt.Sprintf("%s/inputs%d/family%d/phase%d", structure, n, family, phase), func(t *testing.T) {
						p := If_ManStart(&If_Par_t{NLutSize: int32(n), NCutsMax: 4})
						defer If_ManStop(p)
						dst, leaves, temp, mapping, records, packing := structureReviewVectors(n)
						defer Gia_ManStop(dst)
						for _, v := range []*Vec_Int_t{leaves, temp, mapping, records, packing} {
							defer vec.Vec_IntFree(v)
						}
						// XOR together conjunctions on disjoint groups fitting the requested
						// LUTs. At capacity a three-LUT structure must use all three groups.
						// Put the first group in the high variables so that the
						// source heuristic finds its initial bound partition.
						first := leaf
						second := 0
						if len(structure) == 3 && n > leaf+root-1 {
							second = int(structure[1] - '0')
						}
						value := func(x int) bool {
							if family == 0 {
								return x == (1<<n)-1
							}
							yes, result := true, false
							for rank := 0; rank < n; rank++ {
								if rank == first || second > 0 && rank == first+second {
									result = result != yes
									yes = true
								}
								yes = yes && (x>>(n-1-rank)&1 != 0)
							}
							return result != yes
						}
						truth := make([]word, 1<<max(0, n-6))
						for x := 0; x < 64*len(truth); x++ {
							if value(x&((1<<n)-1)) != (phase&1 != 0) {
								truth[x>>6] |= 1 << uint(x&63)
							}
						}
						if phase&2 != 0 {
							slices.Reverse(Vec_IntArray(leaves))
						}
						beforeTruth, beforeLeaves := slices.Clone(truth), slices.Clone(Vec_IntArray(leaves))
						lit := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, structure, nil, mapping, records, packing, 0, 0)
						if lit < 0 {
							t.Fatal("known realizable structure rejected")
						}
						if !slices.Equal(beforeTruth, truth) || !slices.Equal(beforeLeaves, Vec_IntArray(leaves)) {
							t.Fatal("source data changed")
						}
						packed := Vec_IntArray(packing)
						if len(packed) < 4 || packed[0] != 1 || packed[1] < 2 || packed[1] > int32(len(structure)) || packed[len(packed)-1] != lit>>1 {
							t.Fatalf("bad packing%v", packed)
						}
						for index, node := range packed[2:] {
							at := int(Vec_IntEntry(mapping, int(node)))
							data := Vec_IntArray(records)
							if at <= 0 || at >= len(data) {
								t.Fatalf("missing record for%d", node)
							}
							fans := int(data[at])
							limit := leaf
							if index == len(packed)-3 {
								limit = root
							} else if index == 1 {
								limit = int(structure[1] - '0')
							}
							if fans < 2 || fans > limit || at+fans+1 >= len(data) || data[at+fans+1] != node {
								t.Fatalf("bad LUT%d record%v", node, data[at:])
							}
						}
						gia.Gia_ManAppendCo(dst, lit)
						input := make([]bool, n)
						for x := 0; x < 1<<n; x++ {
							logical := 0
							for pin := range input {
								input[pin] = x>>pin&1 != 0
							}
							for pin, leafLit := range beforeLeaves {
								if (x>>uint((leafLit>>1)-1)&1)^int(leafLit&1) != 0 {
									logical |= 1 << pin
								}
							}
							want := value(logical) != (phase&1 != 0)
							got, err := dst.Evaluate(input)
							if err != nil || len(got) != 1 || got[0] != want {
								t.Fatalf("assignment%d got%v err%v want%t", x, got, err, want)
							}
						}
					})
				}
			}
		}
	}
}

func BenchmarkStructureElementaryLookup(b *testing.B) {
	p := If_ManStart(&If_Par_t{NLutSize: 15, NCutsMax: 4})
	defer If_ManStop(p)
	dst, leaves, temp, mapping, records, packing := structureReviewVectors(15)
	defer Gia_ManStop(dst)
	for _, v := range []*Vec_Int_t{leaves, temp, mapping, records, packing} {
		defer vec.Vec_IntFree(v)
	}
	truth := make([]word, 512)
	for x := 0; x < 1<<15; x++ {
		if x>>14&1 != 0 {
			truth[x>>6] |= 1 << uint(x&63)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, "666", nil, mapping, records, packing, 0, 0); got != Vec_IntEntry(leaves, 14) {
			b.Fatal(got)
		}
	}
}

// The C search is greedy, not an existence oracle. The following exact
// source-order trace rejects a realizable (abcd) XOR (efg) 4-4 decomposition.
func TestStructureGreedySearchOrdering(t *testing.T) {
	truth := make([]word, 2)
	for x := 0; x < 128; x++ {
		if (x&15 == 15) != (x>>4 == 7) {
			truth[x>>6] |= 1 << uint(x&63)
		}
	}
	work := slices.Clone(truth)
	v2p, p2v := cluIdentity(7)
	for round, want := range [][]int32{{2, 6, 5, 1, 0, 3, 4}, {0, 2, 6, 3, 4, 1, 5}} {
		if round != 0 {
			If_CluReverseOrder(work, 7, v2p, p2v, 0)
		}
		got := If_CluFindGroup(work, 7, 0, 0, v2p, p2v, 4, 1)
		if got.nVars != 0 || got.nMyu != 4 || !slices.Equal(p2v[:7], want) {
			t.Fatalf("search%d: group%+v order%v want%v", round, got, p2v[:7], want)
		}
		// Scalar evaluation proves the maintained permutation describes the
		// actual truth table, including swaps across the 64-bit boundary.
		for x := 0; x < 128; x++ {
			original := 0
			for pin, v := range want {
				original |= (x >> pin & 1) << v
			}
			expected := (original&15 == 15) != (original>>4 == 7)
			if got := work[x>>6]>>uint(x&63)&1 != 0; got != expected {
				t.Fatalf("search%d assignment%d", round, x)
			}
		}
	}
	if g := If_CluCheck(nil, truth, 7, 0, 0, 4, 4, nil, nil, nil, nil, 0); g.nVars != 0 {
		t.Fatalf("source greedy rejection changed: %+v", g)
	}
}
