package if_

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
)

// Enumerate cluster sets independently of the production boundary visitor.
// Candidates contain the root and zero, one or two of its direct LUT inputs.
// Smaller sets and earlier fanin positions win exact delay ties.
func iffReviewOracle(graph *Gia_Man_t, cuts map[int32][]int32, arrivals []float32, delays []float32, degree int) ([]float32, map[int32][]int32, []int32) {
	times := make([]float32, Gia_ManObjNum(graph))
	for i := range times {
		times[i] = math.Float32frombits(0xffffffff)
	}
	times[0] = 0
	visited := map[int32]bool{0: true}
	for i, arrival := range arrivals {
		id := gia.Gia_ObjId(graph, Gia_ManCi(graph, int32(i)))
		times[id], visited[id] = arrival, true
	}
	chosen := make(map[int32][]int32)
	for root := int32(1); root < Gia_ManObjNum(graph); root++ {
		leaves, ok := cuts[root]
		if !ok {
			continue
		}
		var eligible []int32
		for _, id := range leaves {
			if _, ok := cuts[id]; ok {
				eligible = append(eligible, id)
			}
		}
		consider := func(absorbed []int32) {
			members := append([]int32{root}, absorbed...)
			boundary := make(map[int32]bool)
			for _, member := range members {
				for _, input := range cuts[member] {
					if !slices.Contains(members, input) {
						boundary[input] = true
					}
				}
			}
			arrival := float32(0)
			for input := range boundary {
				arrival = max(arrival, times[input])
			}
			candidate := arrival + delays[len(boundary)]
			if math.IsNaN(float64(times[root])) || candidate < times[root] {
				times[root], chosen[root] = candidate, slices.Clone(absorbed)
			}
		}
		consider(nil)
		for _, a := range eligible {
			consider([]int32{a})
		}
		if degree == 3 {
			for i, a := range eligible {
				for _, b := range eligible[i+1:] {
					consider([]int32{a, b})
				}
			}
		}
	}
	packing := []int32{0}
	var collect func(int32)
	collect = func(root int32) {
		if visited[root] {
			return
		}
		visited[root] = true
		members := append([]int32{root}, chosen[root]...)
		for _, member := range members {
			for _, input := range cuts[member] {
				if !slices.Contains(members, input) {
					collect(input)
				}
			}
		}
		packing = append(packing, int32(len(members)))
		packing = append(packing, chosen[root]...)
		packing = append(packing, root)
		packing[0]++
	}
	for i := int32(0); i < Gia_ManCoNum(graph); i++ {
		co := Gia_ManCo(graph, i)
		root := gia.Gia_ObjFaninId0p(graph, co)
		times[gia.Gia_ObjId(graph, co)] = times[root]
		collect(root)
	}
	return times, chosen, packing
}

func iffReviewCircuit(t *testing.T, seed, phase int) (*Gia_Man_t, map[int32][]int32, []int32) {
	t.Helper()
	graph := Gia_ManStart(64)
	var available, roots []int32
	for range 4 + 2*(seed%2) {
		available = append(available, Gia_ManAppendCi(graph)>>1)
	}
	cuts := make(map[int32][]int32)
	state := uint32(seed + 1)
	for gate := 0; gate < 6; gate++ {
		count := 2
		if gate%3 == 0 {
			count = 3
		}
		var leaves []int32
		for len(leaves) < count {
			state = state*1664525 + 1013904223
			id := available[int(state>>8)%len(available)]
			if !slices.Contains(leaves, id) {
				leaves = append(leaves, id)
			}
		}
		literal := leaves[0]*2 ^ int32(phase&1)
		for pin, leaf := range leaves[1:] {
			literal = gia.Gia_ManAppendAnd(graph, literal, leaf*2^int32((phase>>(pin+1))&1))
		}
		root := literal >> 1
		cuts[root] = leaves
		available, roots = append(available, root), append(roots, root)
	}
	return graph, cuts, roots
}

func iffReviewMatrixCase(t *testing.T, seed, phase, degree, profile, library int, reverse bool) string {
	t.Helper()
	graph, cuts, roots := iffReviewCircuit(t, seed, phase)
	outputs := []int32{roots[5] * 2, roots[2]*2 ^ 1, roots[5]*2 ^ 1, 2, 1}
	if reverse {
		slices.Reverse(outputs)
	}
	for _, output := range outputs {
		Gia_ManAppendCo(graph, output)
	}
	iffReviewMapping(t, graph, cuts)
	mapping := graph.VMapping
	originalMapping := slices.Clone(Vec_IntArray(mapping))
	arrivals := make([]float32, Gia_ManCiNum(graph))
	for i := range arrivals {
		if profile == 1 {
			arrivals[i] = float32((3*i+seed)%5) / 2
		} else if profile == 2 {
			arrivals[i] = float32((i+seed)%4) * 1.25
		}
	}
	lib := &If_LibLut_t{LutMax: 10}
	delays := make([]float32, len(lib.pLutDelays))
	for size := range delays {
		delay := float32(1)
		if library == 1 {
			delay = float32(size+3) / 4
		} else if library == 2 {
			delay = float32((size*3)%7+1) / 4
		}
		delays[size], lib.pLutDelays[size][0] = delay, delay
	}
	wantTimes, wantChosen, wantPacking := iffReviewOracle(graph, cuts, arrivals, delays, degree)
	timing := tim.Tim_ManStart(len(arrivals), len(outputs))
	defer tim.Tim_ManStop(timing)
	for i, arrival := range arrivals {
		tim.Tim_ManInitPiArrival(timing, i, arrival)
	}
	p := Gia_ManIffPerform(graph, lib, timing, 4, int32(degree))
	defer Gia_ManIffStop(p)
	pairs, triples := 0, 0
	for id, want := range wantTimes {
		got := p.times[id]
		if math.Float32bits(got) != math.Float32bits(want) {
			t.Fatalf("arrival%d got%v want%v", id, got, want)
		}
	}
	for root := range cuts {
		want := [2]int32{-1, -1}
		copy(want[:], wantChosen[root])
		if got := [2]int32{p.matches[2][root], p.matches[3][root]}; got != want {
			t.Fatalf("root%d chosen%v want%v", root, got, want)
		}
		if len(wantChosen[root]) == 1 {
			pairs++
		} else if len(wantChosen[root]) == 2 {
			triples++
		}
	}
	if got := Vec_IntArray(Gia_ManIffSelect(p)); !slices.Equal(got, wantPacking) {
		t.Fatalf("packed postorder%v want%v", got, wantPacking)
	}
	maximum := float32(0)
	for i := range outputs {
		want := wantTimes[gia.Gia_ObjId(graph, Gia_ManCo(graph, int32(i)))]
		if got := tim.Tim_ManGetCoArrival(timing, i); got != want {
			t.Fatalf("CO%d got%v want%v", i, got, want)
		}
		maximum = max(maximum, want)
	}
	report := fmt.Sprintf("Max delay = %.2f.  Count1 = %d.  Count2 = %d.  Count3 = %d.\n", maximum, len(cuts)-pairs-triples, pairs, triples)
	if graph.VMapping != mapping || !slices.Equal(Vec_IntArray(mapping), originalMapping) || graph.VPacking != nil {
		t.Fatal("perform/select changed source mapping or installed packing")
	}
	return report
}

func TestIffClusteringAgainstIndependentSetOracle(t *testing.T) {
	var reports []string
	gotReports := dsdReportCapture(t, func() {
		for seed := 0; seed < 8; seed++ {
			for phase := 0; phase < 4; phase++ {
				for _, degree := range []int{2, 3} {
					for profile := 0; profile < 3; profile++ {
						for library := 0; library < 3; library++ {
							for _, reverse := range []bool{false, true} {
								name := fmt.Sprintf("seed%d/phase%d/degree%d/profile%d/library%d/reverse%t", seed, phase, degree, profile, library, reverse)
								t.Run(name, func(t *testing.T) {
									reports = append(reports, iffReviewMatrixCase(t, seed, phase, degree, profile, library, reverse))
								})
							}
						}
					}
				}
			}
		}
	})
	if want := strings.Join(reports, ""); gotReports != want {
		t.Fatal("one or more source summary rows differ from independently counted choices")
	}
}
