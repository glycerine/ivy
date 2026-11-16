package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Reverse levels and scalar random arithmetic use only the abstract graph.
func (model isoSignatureReviewModel) reverseSignatures(values []uint64) (levels []int32, buckets [][]int32, next uint32) {
	levels = make([]int32, len(model.nodes))
	for id := len(model.nodes) - 1; id >= 0; id-- {
		node := model.nodes[id]
		if node.kind == 'a' {
			for _, edge := range node.edges {
				levels[edge.id] = max(levels[edge.id], levels[id]+1)
			}
		} else if node.kind == 'o' {
			levels[node.edges[0].id] = max(levels[node.edges[0].id], 1)
		}
	}
	var maximum int32
	for id, node := range model.nodes {
		if node.kind != 'a' && node.kind != 'o' {
			maximum = max(maximum, levels[id])
		}
	}
	buckets = make([][]int32, maximum+1)
	for id, node := range model.nodes {
		if id != 0 && (node.kind == 'o' || levels[id] != 0) {
			buckets[levels[id]] = append(buckets[levels[id]], int32(id))
		}
	}
	z, w := uint64(3716960521), uint64(2174103536)
	random := func() uint64 {
		z = (36969*(z%65536) + z/65536) & 0xffffffff
		w = (18000*(w%65536) + w/65536) & 0xffffffff
		return (z*65536 + w) & 0xffffffff
	}
	random() // Abc_Random(1) resets and consumes one sample.
	for level, ids := range buckets {
		constants := [2]uint64{random(), random()}
		for _, id := range ids {
			node := model.nodes[id]
			if level == 0 {
				values[id] = random()
				edge := node.edges[0]
				values[edge.id] = (values[edge.id] + values[id] + constants[edge.phase]) & 0xffffffff
			} else if node.kind == 'a' {
				for _, edge := range node.edges {
					values[edge.id] = (values[edge.id] + values[id] + constants[edge.phase]) & 0xffffffff
				}
			}
		}
	}
	return levels, buckets, uint32(random())
}

func TestIsoReverseSignaturesAgainstScalarLevelsAndRandom(t *testing.T) {
	for registers := 0; registers <= 2; registers++ {
		for _, phases := range []uint32{0, 0x3fff, 0x1555, 0x2aaa} {
			for seed := 0; seed < 2; seed++ {
				t.Run(fmt.Sprintf("registers%d/phases%04x/seed%d", registers, phases, seed), func(t *testing.T) {
					graph, model := isoSignatureReviewGraph(t, registers, phases)
					values := make([]uint64, len(model.nodes))
					fields := make([][8]uint32, len(model.nodes))
					for id := range values {
						if seed != 0 {
							values[id] = (0xfffffff0 + uint64(id)*101) & 0xffffffff
						}
						obj := Gia_ManObj(graph, int32(id))
						obj.Value = uint32(values[id])
						obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
						fields[id] = isoSignatureReviewFields(obj)
					}
					epoch := graph.nTravIds
					for repeat := 0; repeat < 2; repeat++ {
						levels, buckets, next := model.reverseSignatures(values)
						got := Gia_Iso4Gia(graph)
						checkIsoSignatureReview(t, graph, values, fields, "reverse signatures")
						if vec.Vec_WecSize(got) != len(buckets) {
							t.Fatal("reverse bucket count differs")
						}
						for i, ids := range buckets {
							if !slices.Equal(vec.Vec_IntArray(vec.Vec_WecEntry(got, i)), ids) {
								t.Fatalf("bucket%d got%v want%v", i, vec.Vec_IntArray(vec.Vec_WecEntry(got, i)), ids)
							}
						}
						if !slices.Equal(vec.Vec_IntArray(graph.vLevels), levels) || graph.nTravIds != epoch {
							t.Fatal("reverse levels or traversal state differ")
						}
						if sample := abcutil.Abc_Random(0); sample != next {
							t.Fatalf("next RNG sample%08x want%08x", sample, next)
						}
					}
				})
			}
		}
	}
}

func TestIsoReverseDriverReportsEverySourceLevel(t *testing.T) {
	graph, model := isoSignatureReviewGraph(t, 1, 0x1555)
	values := make([]uint64, len(model.nodes))
	_, buckets, _ := model.reverseSignatures(values)
	var want strings.Builder
	for level, ids := range buckets {
		fmt.Fprintf(&want, "Level %d\n", level)
		for _, id := range ids {
			fmt.Fprintf(&want, "Obj = %5d.  Value = %08x.\n", id, values[id])
		}
	}
	if got := isoReviewCapture(t, func() { Gia_Iso4Test(graph) }); got != want.String() {
		t.Fatalf("reverse driver got%q want%q", got, want.String())
	}
}

func TestIsoCompareVecsIncludesSourceStatisticsAndLevelEffects(t *testing.T) {
	a, modelA := isoSignatureReviewGraph(t, 0, 0)
	b, modelB := isoSignatureReviewGraph(t, 1, 0x3fff)
	a.pName, b.pName = "iso-left", "iso-right"
	for id, value := range []uint32{0, 0xffffffff, 7, 7, 0x80000000, 5} {
		Gia_ManObj(a, int32(id)).Value = value
	}
	for id, value := range []uint32{7, 7, 0x80000000, 0xffffffff, 9, 5} {
		Gia_ManObj(b, int32(id)).Value = value
	}
	// Lists may contain duplicate values and shared comparison uses signed order.
	left := iso2ClassesReviewVector([][]int32{{2, 1, 2, 3}, {0, 4}, {5}})
	right := iso2ClassesReviewVector([][]int32{{3, 1, 0, 3}, {2, 4, 2}})
	Gia_ManLevelRNum(a)
	Gia_ManLevelRNum(b)
	got := isoReviewCapture(t, func() { Gia_IsoCompareVecs(a, left, b, right) })
	if strings.Count(got, "i/o =") != 2 || !strings.Contains(got, "iso-left") || !strings.Contains(got, "iso-right") {
		t.Errorf("missing standard source graph statistics: %q", got)
	}
	want := "Printing 2 shared levels:\nLevel =   0. One =      1. Two =      1.  Common =      3.\nLevel =   1. One =      1. Two =      2.  Common =      1.\n"
	if !strings.HasSuffix(got, want) {
		t.Errorf("shared-level report got%q want suffix%q", got, want)
	}
	for _, check := range []struct {
		graph *Gia_Man_t
		model isoSignatureReviewModel
	}{{a, modelA}, {b, modelB}} {
		for id, node := range check.model.nodes {
			want := node.level
			if node.kind == 'o' {
				want = check.model.nodes[node.edges[0].id].level
			}
			if level := Gia_ObjLevelId(check.graph, int32(id)); level != want {
				t.Errorf("object%d stored level%d want source stats level%d", id, level, want)
			}
		}
	}
	if !slices.Equal(vec.Vec_IntArray(vec.Vec_WecEntry(left, 0)), []int32{2, 1, 2, 3}) || !slices.Equal(vec.Vec_IntArray(vec.Vec_WecEntry(right, 0)), []int32{3, 1, 0, 3}) {
		t.Error("comparison sorted input object lists")
	}
	data := Gia_IsoCollectData(a, vec.Vec_WecEntry(left, 0))
	if !slices.Equal(vec.Vec_IntArray(data), []int32{7, -1, 7, 7}) {
		t.Fatalf("collected data%v", vec.Vec_IntArray(data))
	}
	Gia_ManObj(a, 2).Value = 99
	if !slices.Equal(vec.Vec_IntArray(data), []int32{7, -1, 7, 7}) {
		t.Fatal("collected data aliases graph values")
	}
}

func TestIsoReverseTwoGraphDriverAgainstScalarSignatures(t *testing.T) {
	a, modelA := isoSignatureReviewGraph(t, 1, 0x1555)
	b, modelB := isoSignatureReviewGraph(t, 2, 0x2aaa)
	valuesA, valuesB := make([]uint64, len(modelA.nodes)), make([]uint64, len(modelB.nodes))
	_, bucketsA, _ := modelA.reverseSignatures(valuesA)
	_, bucketsB, _ := modelB.reverseSignatures(valuesB)
	count := min(len(bucketsA), len(bucketsB))
	var want strings.Builder
	fmt.Fprintf(&want, "Printing %d shared levels:\n", count)
	for level := 0; level < count; level++ {
		remaining := map[uint64]int{}
		for _, id := range bucketsA[level] {
			remaining[valuesA[id]]++
		}
		common := 0
		for _, id := range bucketsB[level] {
			if remaining[valuesB[id]] > 0 {
				common++
				remaining[valuesB[id]]--
			}
		}
		fmt.Fprintf(&want, "Level = %3d. One = %6d. Two = %6d.  Common = %6d.\n", level, len(bucketsA[level])-common, len(bucketsB[level])-common, common)
	}
	got := isoReviewCapture(t, func() { Gia_Iso4TestTwo(a, b) })
	if strings.Count(got, "i/o =") != 2 || !strings.HasSuffix(got, want.String()) {
		t.Fatalf("two-graph driver got%q want stats and suffix%q", got, want.String())
	}
	for _, check := range []struct {
		g      *Gia_Man_t
		values []uint64
	}{{a, valuesA}, {b, valuesB}} {
		for id, value := range check.values {
			if Gia_ManObj(check.g, int32(id)).Value != uint32(value) {
				t.Fatalf("two-graph object%d value differs", id)
			}
		}
	}
}

func TestIsoSynchronousDriverStopsOnSourceCountStability(t *testing.T) {
	for registers := 0; registers <= 2; registers++ {
		for _, phases := range []uint32{0, 0x3fff, 0x1555, 0x2aaa} {
			t.Run(fmt.Sprintf("registers%d/phases%04x", registers, phases), func(t *testing.T) {
				graph, model := isoSignatureReviewGraph(t, registers, phases)
				values := model.thirdInit()
				previous := -1
				var prefixes []string
				for iteration := 0; iteration < 500; iteration++ {
					var distinct []uint64
					for _, value := range values {
						if !slices.Contains(distinct, value) {
							distinct = append(distinct, value)
						}
					}
					prefixes = append(prefixes, fmt.Sprintf("Iter %3d : %6d  out of %6d  Time = ", iteration, len(distinct), len(values)))
					if len(distinct) == previous {
						break
					}
					previous = len(distinct)
					values = model.thirdPass(values)
				}
				got := isoReviewCapture(t, func() { Gia_Iso3Test(graph) })
				lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
				if len(lines) != len(prefixes) {
					t.Fatalf("driver iterations%d want%d", len(lines), len(prefixes))
				}
				for i, prefix := range prefixes {
					if !strings.HasPrefix(lines[i], prefix) || !strings.HasSuffix(lines[i], " sec") {
						t.Fatalf("iteration%d got%q want prefix%q", i, lines[i], prefix)
					}
				}
				for id, value := range values {
					if Gia_ManObj(graph, int32(id)).Value != uint32(value) {
						t.Fatalf("stopped object%d value differs", id)
					}
				}
			})
		}
	}
}

func TestIsoSnapshotUniqueCountKeepsInputOrder(t *testing.T) {
	for _, values := range [][]int32{nil, {0}, {-1, 0, -2147483648, 1, -1, 0, -2147483648}, {7, 7, 7}} {
		v := vec.Vec_IntAllocArrayCopy(values, len(values))
		var distinct []int32
		for _, value := range values {
			if !slices.Contains(distinct, value) {
				distinct = append(distinct, value)
			}
		}
		if got := Gia_Iso3Unique(v); got != int32(len(distinct)) {
			t.Fatalf("values%v unique%d want%d", values, got, len(distinct))
		}
		if !slices.Equal(vec.Vec_IntArray(v), values) {
			t.Fatal("uniqueness count reordered its input")
		}
	}
}
