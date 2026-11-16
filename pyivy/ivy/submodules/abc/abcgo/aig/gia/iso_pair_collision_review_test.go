package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Positive AND gates with a level-seven Iso2 signature of zero. The fixture
// was derived using Go-source constants and scalar modular arithmetic; no C
// source or compiler is needed. Clone the DAG so every candidate signature is
// tied and uniquification adds no salts. Only the first copy drives outputs.
func isoPairReviewZeroSignatureGraph(t *testing.T) *Gia_Man_t {
	t.Helper()
	pairs := [][2]int32{
		{1, 2},
		{4, 3},
		{1, 2},
		{4, 6},
		{4, 3},
		{5, 8},
		{7, 4},
		{7, 5},
		{5, 8},
		{9, 12},
		{10, 4},
		{11, 4},
		{10, 5},
		{16, 10},
		{15, 9},
		{17, 13},
		{18, 14},
		{19, 20},
	}
	graph := Gia_ManStart(64)
	var inputs [2][3]int32
	for copy := range inputs {
		for i := range inputs[copy] {
			inputs[copy][i] = Gia_ManAppendCi(graph)
		}
	}
	var roots [2]int32
	for copy := range inputs {
		refs := make([]int32, 4, len(pairs)+4)
		for i, input := range inputs[copy] {
			refs[i+1] = input
		}
		for _, pair := range pairs {
			refs = append(refs, Gia_ManAppendAnd(graph, refs[pair[0]], refs[pair[1]]))
		}
		roots[copy] = refs[len(refs)-1]
	}
	Gia_ManAppendCo(graph, roots[0])
	Gia_ManAppendCo(graph, roots[0]^1)
	classes := Gia_Iso2ManPerform(graph, 0)
	if Gia_ManObj(graph, roots[0]>>1).Value != 0 || Gia_ManObj(graph, roots[1]>>1).Value != 0 || Gia_ManPo(graph, 0).Value != 0 || Gia_ManPo(graph, 1).Value != 0 {
		t.Fatal("fixture no longer produces the actual source signature collision")
	}
	if vec.Vec_WecSize(classes) != 1 || !slices.Equal(vec.Vec_IntArray(vec.Vec_WecEntry(classes, 0)), []int32{0, 1}) {
		t.Fatal("fixture did not enter one candidate class")
	}
	return graph
}

func TestIso2PairRejectsComplementedOutputSignatureCollision(t *testing.T) {
	graph := isoPairReviewZeroSignatureGraph(t)
	roots, o0, o1 := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	m0, m1 := vec.Vec_IntStart(int(Gia_ManObjNum(graph))), vec.Vec_IntStart(int(Gia_ManObjNum(graph)))
	Gia_Iso2ManCollectOrder(graph, []int32{0}, 1, roots, o0, m0)
	Gia_Iso2ManCollectOrder(graph, []int32{1}, 1, roots, o1, m1)
	if got := Gia_Iso2ManCheckIsoPair(graph, o0, o1, m0, m1); got != 0 {
		t.Fatal("structural comparison accepts an output and its complement after a real signature collision")
	}
}

func TestIsoPipelinesRetainOppositeOutputsAfterRealSignatureCollision(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		t.Run(fmt.Sprintf("mode%d", mode), func(t *testing.T) {
			graph := isoPairReviewZeroSignatureGraph(t)
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
				t.Fatal("reduction lost the complemented output after a real signature collision")
			}
			for i := 0; i < 2; i++ {
				if got := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, i).(*Vec_Int_t)); !slices.Equal(got, []int32{int32(i)}) {
					t.Fatalf("collision class%d got%v", i, got)
				}
			}
			// All gates are positive ANDs; the first root is the conjunction of its
			// three support inputs. Its complement has a different number of true
			// assignments, so no PI permutation can make these outputs isomorphic.
			for assignment := 0; assignment < 64; assignment++ {
				input := make([]bool, 6)
				for i := range input {
					input[i] = assignment&(1<<i) != 0
				}
				expected := input[0] && input[1] && input[2]
				values, err := giaEvaluateNodes(result, input)
				if err != nil {
					t.Fatal(err)
				}
				if values[Gia_ManPo(result, 0).id] != expected || values[Gia_ManPo(result, 1).id] == expected {
					t.Fatalf("assignment%d changed opposite output truth", assignment)
				}
			}
		})
	}
}

func TestIso2PairChecksOutputFaninMapping(t *testing.T) {
	graph := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	for _, driver := range []int32{a, b, a, a, b, b} {
		Gia_ManAppendCo(graph, driver)
	}
	Gia_Iso2ManPerform(graph, 0)
	for output := int32(1); output < 6; output++ {
		if Gia_ManPo(graph, output).Value != Gia_ManPo(graph, 0).Value {
			t.Fatal("fixture output signatures differ")
		}
	}
	roots, o0, o1 := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	m0, m1 := vec.Vec_IntStart(int(Gia_ManObjNum(graph))), vec.Vec_IntStart(int(Gia_ManObjNum(graph)))
	Gia_Iso2ManCollectOrder(graph, []int32{0, 1, 2}, 3, roots, o0, m0)
	Gia_Iso2ManCollectOrder(graph, []int32{3, 4, 5}, 3, roots, o1, m1)
	if vec.Vec_IntSize(o0) != vec.Vec_IntSize(o1) {
		t.Fatal("fixture traversal sizes differ")
	}
	// The ordered tuples (a,b,a) and (a,b,b) have equal signatures and
	// equal traversal sizes, but their output sharing cannot be permuted.
	if got := Gia_Iso2ManCheckIsoPair(graph, o0, o1, m0, m1); got != 0 {
		t.Fatal("structural comparison ignores inconsistent CO driver mappings")
	}
	// A consistent PI renaming must remain accepted: (a,b,a) -> (b,a,b).
	Gia_Iso2ManCollectOrder(graph, []int32{1, 0, 5}, 3, roots, o1, m1)
	if got := Gia_Iso2ManCheckIsoPair(graph, o0, o1, m0, m1); got != 1 {
		t.Fatal("structural comparison rejected a consistent PI renaming")
	}
}
