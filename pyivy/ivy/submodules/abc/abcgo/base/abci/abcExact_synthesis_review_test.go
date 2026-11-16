package abci

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Evaluate the source binary solution independently of SES's truth evaluator.
func exactSynthesisReviewModel(t *testing.T, model string, vars int, truth []word, arrivals ...[]int32) []int8 {
	t.Helper()
	if len(model) < 3 || int(model[0]) != vars || model[1] != 1 {
		t.Fatal("invalid synthesis header")
	}
	gates := int(model[2])
	if gates == 0 || len(model) != 3+4*gates+2+vars {
		t.Fatal("incomplete synthesized model")
	}
	distances := make([][]int, vars+gates)
	for node := range distances {
		distances[node] = make([]int, vars)
		for pin := range distances[node] {
			distances[node][pin] = -1
		}
		if node < vars {
			distances[node][node] = 0
		}
	}
	for gate := 0; gate < gates; gate++ {
		pos := 3 + 4*gate
		a, b := int(model[pos+2]), int(model[pos+3])
		if model[pos+1] != 2 || a >= b || b >= vars+gate {
			t.Fatal("invalid gate fanins")
		}
		for pin := 0; pin < vars; pin++ {
			depth := max(distances[a][pin], distances[b][pin])
			if depth >= 0 {
				distances[vars+gate][pin] = depth + 1
			}
		}
	}
	literal := int(model[3+4*gates])
	if literal>>1 >= gates {
		t.Fatal("invalid output gate")
	}
	root := vars + (literal >> 1)
	pins := make([]int8, vars)
	for i, depth := range distances[root] {
		if depth <= 0 || depth > 127 {
			t.Fatal("synthesized full-support function lost an input")
		}
		pins[i] = int8(depth)
	}
	trailer := 3 + 4*gates + 1
	wantDepth := int(slices.Max(pins))
	if len(arrivals) != 0 {
		wantDepth = 0
		for i, arrival := range arrivals[0] {
			wantDepth = max(wantDepth, int(arrival)+int(pins[i]))
		}
	}
	if int(model[trailer]) != wantDepth {
		t.Fatal("model depth disagrees with physical graph")
	}
	for i, pin := range pins {
		if int8(model[trailer+1+i]) != pin {
			t.Fatal("model pin profile disagrees with graph")
		}
	}
	values := make([]bool, vars+gates)
	for assignment := 0; assignment < 1<<vars; assignment++ {
		for i := 0; i < vars; i++ {
			values[i] = assignment&(1<<i) != 0
		}
		for gate := 0; gate < gates; gate++ {
			pos := 3 + 4*gate
			op := model[pos]
			a, b := values[model[pos+2]], values[model[pos+3]]
			values[vars+gate] = !a && b && op&1 != 0 || a && !b && op&2 != 0 || a && b && op&4 != 0
		}
		got := values[root] != (literal&1 != 0)
		want := truth[assignment>>6]>>(assignment&63)&1 != 0
		if got != want {
			t.Fatalf("model input%x: got%t want%t", assignment, got, want)
		}
	}
	return pins
}

func TestExactSynthesisAllFullSupportThreeInputFunctions(t *testing.T) {
	for _, aig := range []int32{0, 1} {
		for _, arrivals := range [][]int32{{0, 0, 0}, {0, 1, 2}} {
			functions := 0
			for table := 0; table < 256; table++ {
				fullSupport := true
				for pin := 0; pin < 3; pin++ {
					depends := false
					for assignment := 0; assignment < 8; assignment++ {
						if (table>>assignment)&1 != (table>>(assignment^(1<<pin)))&1 {
							depends = true
						}
					}
					fullSupport = fullSupport && depends
				}
				if !fullSupport {
					continue
				}
				functions++
				t.Run(fmt.Sprintf("aig%d/arrivals%v/truth%02x", aig, arrivals, table), func(t *testing.T) {
					exactMappingReviewStoreMode(t, aig)
					// Repeat the three-input table through a full Go word.
					truth := []word{uint64(table) * 0x0101010101010101}
					original := truth[0]
					times := slices.Clone(arrivals)
					pins := make([]int8, 3)
					var cost int32
					delay := Abc_ExactDelayCost(&truth[0], 3, &times[0], pins, &cost, -1)
					if delay == ABC_INFINITY || cost <= 0 {
						t.Fatalf("no synthesized model: delay%d cost%d undef%d", delay, cost, s_pSesStore.nUndefCalls)
					}
					if truth[0] != original || !slices.Equal(times, arrivals) {
						t.Fatal("synthesis changed caller truth or arrivals")
					}
					var model string
					if Ses_StoreGetEntry(s_pSesStore, &truth[0], 3, &times[0], &model) != 1 || model == "" {
						t.Fatal("synthesis did not store its model")
					}
					physicalPins := exactSynthesisReviewModel(t, model, 3, truth, arrivals)
					if !slices.Equal(pins, physicalPins) || int(cost) != int(model[2]) {
						t.Fatal("API cost/pin profile differs from physical model")
					}
					physicalDelay := int32(0)
					for pin, arrival := range arrivals {
						physicalDelay = max(physicalDelay, arrival+int32(pins[pin]))
					}
					if delay != physicalDelay || s_pSesStore.nSatCalls == 0 || s_pSesStore.nEntriesCount != 1 {
						t.Fatal("delay or actual synthesis accounting changed")
					}
					for gate := 0; gate < int(model[2]); gate++ {
						op := model[3+4*gate]
						if op != 1 && op != 2 && op != 4 && op != 7 && (aig != 0 || op != 3) {
							t.Fatalf("invalid operation%d in mode%d", op, aig)
						}
					}
				})
			}
			if functions != 218 {
				t.Fatalf("full-support three-input function inventory: got%d want218", functions)
			}
		}
	}
}

func TestExactSynthesisFullSupportAndBounds(t *testing.T) {
	for _, aig := range []int32{0, 1} {
		for vars := 3; vars <= 8; vars++ {
			for _, inverted := range []bool{false, true} {
				t.Run(fmt.Sprintf("aig%d/vars%d/inverted%t", aig, vars, inverted), func(t *testing.T) {
					exactMappingReviewStoreMode(t, aig)
					truth := exactCacheReviewAndTruth(vars)
					if inverted {
						for i := range truth {
							truth[i] = ^truth[i]
						}
					}
					original := slices.Clone(truth)
					arrivals := make([]int32, vars)
					pins := make([]int8, vars)
					var cost int32
					depth := int32(bits.Len(uint(vars - 1)))
					delay := Abc_ExactDelayCost(&truth[0], int32(vars), &arrivals[0], pins, &cost, depth)
					// A binary full-support function needs n-1 gates; AND/NAND
					// achieves the fanin-two depth lower bound ceil(log2(n)).
					if delay != depth || cost != int32(vars-1) {
						t.Fatalf("synthesis delay%d want%d cost%d want%d sat%d undef%d", delay, depth, cost, vars-1, s_pSesStore.nSatCalls, s_pSesStore.nUndefCalls)
					}
					if !slices.Equal(truth, original) {
						t.Fatal("synthesis changed caller truth")
					}
					var model string
					if Ses_StoreGetEntry(s_pSesStore, &truth[0], int32(vars), &arrivals[0], &model) != 1 || model == "" {
						t.Fatal("synthesis did not publish its model")
					}
					physicalPins := exactSynthesisReviewModel(t, model, vars, truth)
					if !slices.Equal(pins, physicalPins) {
						t.Fatal("API returned different pin depths")
					}
					sat := s_pSesStore.nSatCalls
					if sat == 0 || s_pSesStore.nEntriesCount != 1 || s_pSesStore.nCacheHits != 0 {
						t.Fatal("fresh synthesis/cache accounting changed")
					}
					for i := range arrivals {
						arrivals[i] = 7
					}
					if shifted := Abc_ExactDelayCost(&truth[0], int32(vars), &arrivals[0], pins, &cost, depth+7); shifted != depth+7 || s_pSesStore.nSatCalls != sat || s_pSesStore.nCacheHits != 1 {
						t.Fatal("uniform shift missed source cache or changed delay")
					}
					for _, arrival := range arrivals {
						if arrival != 7 {
							t.Fatal("cost query changed caller arrivals")
						}
					}
				})
			}
		}
	}
}

func TestExactSynthesisDepthHints(t *testing.T) {
	for _, bound := range []int32{0, 1, 2, -1, 100} {
		t.Run(fmt.Sprintf("bound%d", bound), func(t *testing.T) {
			exactMappingReviewStore(t)
			truth := exactCacheReviewAndTruth(3)
			arrivals := []int32{0, 0, 0}
			pins := []int8{-7, -7, -7}
			var cost int32
			delay := Abc_ExactDelayCost(&truth[0], 3, &arrivals[0], pins, &cost, bound)
			if bound == 0 || bound == 1 {
				if delay != ABC_INFINITY || cost != ABC_INFINITY || !slices.Equal(pins, []int8{-7, -7, -7}) || s_pSesStore.nSatCalls != 0 {
					t.Fatalf("impossible depth: delay%d cost%d pins%v calls%d", delay, cost, pins, s_pSesStore.nSatCalls)
				}
			} else {
				if delay != 2 || cost != 2 || s_pSesStore.nSatCalls == 0 {
					t.Fatalf("attainable depth: delay%d cost%d calls%d", delay, cost, s_pSesStore.nSatCalls)
				}
				var model string
				if Ses_StoreGetEntry(s_pSesStore, &truth[0], 3, &arrivals[0], &model) != 1 || model == "" {
					t.Fatal("missing synthesized model")
				}
				exactSynthesisReviewModel(t, model, 3, truth)
			}
			if s_pSesStore.nEntriesCount != 1 || s_pSesStore.nCacheHits != 0 {
				t.Fatal("cache-miss accounting changed")
			}
		})
	}
}

func TestExactSynthesisWideArrivalDepths(t *testing.T) {
	for _, gap := range []int32{29, 30, 31, 32} {
		for _, aig := range []int32{0, 1} {
			t.Run(fmt.Sprintf("aig%d/gap%d", aig, gap), func(t *testing.T) {
				exactMappingReviewStoreMode(t, aig)
				truth := exactCacheReviewAndTruth(2)
				arrivals := []int32{0, gap}
				pins := []int8{-7, -7}
				var cost int32
				delay := Abc_ExactDelayCost(&truth[0], 2, &arrivals[0], pins, &cost, gap+1)
				// A single AND gate achieves the full-support lower bound:
				// one unit after the later-arriving input.
				if delay != gap+1 || cost != 1 || !slices.Equal(pins, []int8{1, 1}) {
					t.Fatalf("wide arrival depth: delay%d cost%d pins%v", delay, cost, pins)
				}
				if !slices.Equal(arrivals, []int32{0, gap}) {
					t.Fatal("query changed caller arrivals")
				}
			})
		}
	}
}

func TestExactSynthesisVerboseMissHeader(t *testing.T) {
	exactMappingReviewStore(t)
	s_pSesStore.fVeryVerbose = 1
	truth := exactCacheReviewAndTruth(2)
	arrivals := []int32{0, 0}
	pins := []int8{-7, -7}
	var cost int32
	got := recordedReviewCapture(t, func() {
		Abc_ExactDelayCost(&truth[0], 2, &arrivals[0], pins, &cost, 0)
	})
	// A zero bound skips SAT entirely. Only the source cache-miss header
	// and its trailing padding should be printed.
	want := "\x1b[36m8\x1b[0m [0 0]@0:        \n"
	if got != want {
		t.Fatalf("cache-miss diagnostic: got%q want%q", got, want)
	}
}

func TestExactSynthesisResourceLimitDiagnostic(t *testing.T) {
	for _, bound := range []int32{4, 100} {
		t.Run(fmt.Sprintf("bound%d", bound), func(t *testing.T) {
			exactMappingReviewStore(t)
			s_pSesStore.fVeryVerbose = 1
			s_pSesStore.nBTLimit = 1
			// Four-input parity is full-support and needs multiple AIG gates.
			truth := []word{0x6996699669966996}
			arrivals := []int32{0, 0, 0, 0}
			pins := []int8{-7, -7, -7, -7}
			var cost int32
			got := recordedReviewCapture(t, func() {
				Abc_ExactDelayCost(&truth[0], 4, &arrivals[0], pins, &cost, bound)
			})
			if s_pSesStore.nUndefCalls == 0 {
				t.Fatalf("fixture did not hit the resource limit: cost%d report%q", cost, got)
			}
			depth, erase := "4", "\b"
			if bound == 100 {
				// Source caps a loose bound at twice (nVars+1).
				depth, erase = "10", "\b\b"
			}
			want := erase + ANSI_COLOR_RED + depth + ANSI_COLOR_RESET + "        \n"
			if !strings.HasSuffix(got, want) {
				t.Fatalf("resource limit must be red with both depth digits erased: got%q suffix%q", got, want)
			}
			if cost != ABC_INFINITY || !slices.Equal(pins, []int8{-7, -7, -7, -7}) || s_pSesStore.nUnsynthesizedRL != 1 {
				t.Fatalf("resource-limit result changed: cost%d pins%v failed%d", cost, pins, s_pSesStore.nUnsynthesizedRL)
			}
		})
	}
}

func TestExactSynthesisResourceLogUsesNormalizedTruth(t *testing.T) {
	exactMappingReviewStore(t)
	s_pSesStore.nBTLimit = 1
	filename := filepath.Join(t.TempDir(), "exact.log")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	// Store cleanup owns the debug file, as in Abc_ExactStart.
	s_pSesStore.pDebugEntries = file
	truth := []word{0x9669966996699669}
	arrivals := []int32{0, 0, 0, 0}
	pins := make([]int8, 4)
	var cost int32
	Abc_ExactDelayCost(&truth[0], 4, &arrivals[0], pins, &cost, 4)
	if s_pSesStore.nUndefCalls == 0 {
		t.Fatal("fixture did not hit the resource limit")
	}
	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	// C logs before SES cleanup restores inverted specifications. Reproduce
	// that state from the owned SES buffer, while leaving caller truth intact.
	want := "abc -c \"exact -v -C 1 -a -S 3 -D 4 -A 0,0,0,0 6996\" # no solution found before\n"
	if string(got) != want || truth[0] != 0x9669966996699669 {
		t.Fatalf("normalized failure log: got%q want%q caller%016x", got, want, truth[0])
	}
}
