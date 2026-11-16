package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func iso2ClassesReviewGraph(phase int32) *Gia_Man_t {
	g := Gia_ManStart(40)
	var in [6]int32
	for i := range in {
		in[i] = Gia_ManAppendCi(g)
	}
	a := Gia_ManAppendAnd(g, in[0], in[1])
	b := Gia_ManAppendAnd(g, in[2], in[3])
	c := Gia_ManAppendAnd(g, in[0], in[1]^1)
	d := Gia_ManAppendAnd(g, in[2], in[3]^1)
	e := Gia_ManAppendAnd(g, a, in[4])
	f := Gia_ManAppendAnd(g, b, in[5])
	for _, lit := range []int32{a, b, c, d, in[0], in[2], 0, 1, e, f, a, c} {
		Gia_ManAppendCo(g, lit^phase)
	}
	Gia_Iso2ManPrepare(g)
	Gia_Iso2ManPropagate(g)
	return g
}

func iso2ClassesReviewVector(groups [][]int32) *Vec_Wec_t {
	// Exact small capacity forces the full-quality worklist to move on append.
	v := vec.Vec_WecAllocExact(len(groups))
	for _, group := range groups {
		level := vec.Vec_WecPushLevel(v)
		for _, id := range group {
			vec.Vec_IntPush(level, id)
		}
	}
	return v
}

// Fixture cone kinds, independent of the production signature/pair routines.
var iso2ClassesReviewKinds = []int{0, 0, 1, 1, 2, 2, 3, 4, 5, 5, 0, 1}

func iso2ClassesReviewExpected(initial [][]int32, quality bool) (result, consumed [][]int32, transferred []bool) {
	queue := make([][]int32, len(initial))
	for i, group := range initial {
		queue[i] = slices.Clone(group)
	}
	for i := 0; i < len(queue); i++ {
		group := queue[i]
		var keep, reject []int32
		for _, id := range group {
			if iso2ClassesReviewKinds[id] == iso2ClassesReviewKinds[group[0]] {
				keep = append(keep, id)
			} else {
				reject = append(reject, id)
			}
		}
		if quality {
			result = append(result, keep)
			consumed = append(consumed, nil)
			if i < len(initial) {
				transferred = append(transferred, true)
			}
			if len(reject) > 0 {
				queue = append(queue, reject)
			}
		} else if len(reject) == 0 {
			result = append(result, slices.Clone(group))
			consumed = append(consumed, nil)
			transferred = append(transferred, true)
		} else {
			for _, id := range group {
				result = append(result, []int32{id})
			}
			consumed = append(consumed, slices.Clone(group))
			transferred = append(transferred, false)
		}
	}
	return
}

func TestIso2ClassRefinementConsumesSourceWorkspace(t *testing.T) {
	partitions := [][][]int32{
		{{0, 2, 4, 6, 8, 1, 3, 5, 7, 9, 10, 11}},
		{{0, 1, 10}, {2, 3, 11}, {4, 5}, {6}, {7}, {8, 9}},
		{{0, 2, 4, 8, 1, 3, 5, 9, 10, 11}, {6, 7}},
		{{}, {0, 2, 4, 6, 8, 1, 3, 5, 7, 9, 10, 11}, {}},
	}
	for partition, groups := range partitions {
		for _, quality := range []bool{false, true} {
			for phase := int32(0); phase < 2; phase++ {
				t.Run(fmt.Sprintf("partition%d/quality%t/phase%d", partition, quality, phase), func(t *testing.T) {
					graph := iso2ClassesReviewGraph(phase)
					classes := iso2ClassesReviewVector(groups)
					before := make([][]int32, len(groups))
					for i := range groups {
						before[i] = vec.Vec_IntArray(vec.Vec_WecEntry(classes, i))
					}
					want, consumed, transferred := iso2ClassesReviewExpected(groups, quality)
					var got *Vec_Wec_t
					isoReviewCapture(t, func() {
						if quality {
							got = Gia_Iso2ManCheckIsoClasses(graph, classes)
						} else {
							got = Gia_Iso2ManCheckIsoClassesSkip(graph, classes)
						}
					})
					if vec.Vec_WecSize(got) != len(want) {
						t.Fatalf("result classes%d want%d", vec.Vec_WecSize(got), len(want))
					}
					for i, group := range want {
						if actual := vec.Vec_IntArray(vec.Vec_WecEntry(got, i)); !slices.Equal(actual, group) {
							t.Errorf("result%d got%v want%v", i, actual, group)
						}
					}
					if vec.Vec_WecSize(classes) != len(consumed) {
						t.Errorf("source worklist length%d want%d", vec.Vec_WecSize(classes), len(consumed))
					}
					for i := 0; i < min(vec.Vec_WecSize(classes), len(consumed)); i++ {
						level := vec.Vec_WecEntry(classes, i)
						if !slices.Equal(vec.Vec_IntArray(level), consumed[i]) {
							t.Errorf("worklist%d got%v want%v", i, vec.Vec_IntArray(level), consumed[i])
						}
						if len(consumed[i]) == 0 && vec.Vec_IntCap(level) != 0 {
							t.Errorf("consumed class%d retains capacity%d", i, vec.Vec_IntCap(level))
						}
					}
					// Each accepted original class becomes its corresponding result buffer.
					output := 0
					for i, transfer := range transferred {
						if transfer {
							actual := vec.Vec_IntArray(vec.Vec_WecEntry(got, output))
							if len(actual) > 0 && &actual[0] != &before[i][0] {
								t.Errorf("class%d was copied instead of transferred", i)
							}
							output++
						} else {
							output += len(groups[i])
						}
					}
				})
			}
		}
	}
}

func TestIso2ClassRefinementReportsSourceProgress(t *testing.T) {
	for _, quality := range []bool{false, true} {
		t.Run(fmt.Sprintf("quality%t", quality), func(t *testing.T) {
			graph := Gia_ManStart(64)
			input := Gia_ManAppendCi(graph)
			groups := make([][]int32, 61)
			for i := range groups {
				Gia_ManAppendCo(graph, input)
				groups[i] = []int32{int32(i)}
			}
			classes := iso2ClassesReviewVector(groups)
			got := isoReviewCapture(t, func() {
				if quality {
					Gia_Iso2ManCheckIsoClasses(graph, classes)
				} else {
					Gia_Iso2ManCheckIsoClassesSkip(graph, classes)
				}
			})
			want := "Finished        0 outputs (out of       61)...\rFinished       50 outputs (out of       61)...\r"
			if got != want {
				t.Fatalf("progress got%q want%q", got, want)
			}
		})
	}
}
