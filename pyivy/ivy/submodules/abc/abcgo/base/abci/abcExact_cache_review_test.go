package abci

import (
	"fmt"
	"slices"
	"testing"
)

// Single-output AND chain in the source packed format. Every variable occurs;
// later gates reuse input0. Input depths are g,g,g-1,...,g-n+2.
func exactCacheReviewChain(vars, gates int) string {
	data := []byte{byte(vars), 1, byte(gates)}
	for gate := 0; gate < gates; gate++ {
		a, b := 0, 1
		if gate != 0 {
			b = vars + gate - 1
			if gate < vars-1 {
				a = gate + 1
			}
		}
		data = append(data, 4, 2, byte(a), byte(b)) // AND operation, two operands.
	}
	data = append(data, byte(2*(gates-1)), byte(gates))
	for pin := 0; pin < vars; pin++ {
		depth := gates
		if pin > 1 {
			depth -= pin - 1
		}
		data = append(data, byte(depth))
	}
	return string(data)
}

func TestExactNetworkNormalizationHasNoHundredNodeLimit(t *testing.T) {
	for _, tc := range [][2]int{{2, 97}, {2, 98}, {2, 99}, {2, 100}, {2, 101}, {8, 92}, {8, 93}} {
		t.Run(fmt.Sprintf("vars%d/gates%d", tc[0], tc[1]), func(t *testing.T) {
			arrivals, want := make([]int32, tc[0]), make([]int32, tc[0])
			arrivals[1] = 4
			for pin := 2; pin < tc[0]; pin++ {
				want[pin] = int32(pin - 1)
			}
			model := exactCacheReviewChain(tc[0], tc[1])
			Abc_ExactNormalizeArrivalTimesForNetwork(int32(tc[0]), &arrivals[0], model)
			if !slices.Equal(arrivals, want) {
				t.Fatalf("network-normalized times%v want%v", arrivals, want)
			}
		})
	}
}

func exactCacheReviewAndTruth(vars int) []word {
	truth := make([]word, max(1, 1<<max(vars-6, 0)))
	mask := (1 << vars) - 1
	for assignment := 0; assignment < max(64, 1<<vars); assignment++ {
		if assignment&mask == mask {
			truth[assignment>>6] |= word(1) << (assignment & 63)
		}
	}
	return truth
}

func TestExactCostCachedChainsAndInputOwnership(t *testing.T) {
	for vars := 3; vars <= 8; vars++ {
		t.Run(fmt.Sprintf("vars%d", vars), func(t *testing.T) {
			exactMappingReviewStore(t)
			truth := exactCacheReviewAndTruth(vars)
			original := slices.Clone(truth)
			key, wantKey := make([]int32, vars), make([]int32, vars)
			for i := range key {
				key[i] = int32(i % 3)
				if i > 1 {
					wantKey[i] = int32(i - 1)
				}
			}
			model := exactCacheReviewChain(vars, vars-1)
			if Ses_StoreAddEntry(s_pSesStore, &truth[0], int32(vars), &key[0], model, 0) != 1 || !slices.Equal(key, wantKey) {
				t.Fatal("source network key normalization changed")
			}
			for _, offset := range []int32{-9, 0, 7} {
				for reverse := 0; reverse < 2; reverse++ {
					arrivals := make([]int32, vars)
					pins := make([]int8, vars+2)
					for i := range pins {
						pins[i] = -23
					}
					wantPins := make([]int8, vars)
					wantDelay := int32(0) // Source clamps a wholly negative arrival at zero.
					for i := range arrivals {
						value := i % 4
						if reverse != 0 {
							value = (vars - i) % 4
						}
						arrivals[i] = offset + int32(value)
						depth := vars - 1
						if i > 1 {
							depth -= i - 1
						}
						wantPins[i] = int8(depth)
						wantDelay = max(wantDelay, arrivals[i]+int32(depth))
					}
					before := slices.Clone(arrivals)
					var cost int32
					if delay := Abc_ExactDelayCost(&truth[0], int32(vars), &arrivals[0], pins, &cost, 0); delay != wantDelay || cost != int32(vars-1) || !slices.Equal(pins[:vars], wantPins) || pins[vars] != -23 || pins[vars+1] != -23 {
						t.Fatalf("cached cost: vars%d delay%d want%d cost%d pins%v", vars, delay, wantDelay, cost, pins)
					}
					if !slices.Equal(truth, original) || !slices.Equal(arrivals, before) {
						t.Fatal("cached cost changed caller truth/arrival data")
					}
				}
			}
			if s_pSesStore.nCacheHits != 6 || s_pSesStore.nSatCalls != 0 || s_pSesStore.nEntriesCount != 1 {
				t.Fatal("cache hit triggered synthesis or insertion")
			}
			changed := slices.Clone(truth)
			changed[len(changed)-1] ^= word(1) << 63
			if vars < 6 {
				for i := range changed {
					changed[i] = ^truth[i]
				}
			}
			if vars >= 6 && Ses_StoreTableHash(&changed[0], int32(vars)) != Ses_StoreTableHash(&truth[0], int32(vars)) {
				t.Fatal("expected high-word hash collision")
			}
			answer := "untouched"
			query := make([]int32, vars)
			if Ses_StoreGetEntry(s_pSesStore, &changed[0], int32(vars), &query[0], &answer) != 0 || answer != "untouched" {
				t.Fatal("cache ignored changed truth bits or modified a miss output")
			}
		})
	}
}

func TestExactCostCachedFailureKeepsPinsAndCounters(t *testing.T) {
	for _, limited := range []int32{0, 1} {
		t.Run(fmt.Sprintf("resource%d", limited), func(t *testing.T) {
			exactMappingReviewStore(t)
			truth := exactCacheReviewAndTruth(3)
			profile := []int32{0, 2, 1}
			if Ses_StoreAddEntry(s_pSesStore, &truth[0], 3, &profile[0], "", limited) != 1 {
				t.Fatal("could not seed source negative cache")
			}
			arrivals := []int32{7, 9, 8}
			pins := []int8{9, 10, 11}
			var cost int32
			if delay := Abc_ExactDelayCost(&truth[0], 3, &arrivals[0], pins, &cost, 20); delay != ABC_INFINITY || cost != ABC_INFINITY || !slices.Equal(pins, []int8{9, 10, 11}) || !slices.Equal(arrivals, []int32{7, 9, 8}) {
				t.Fatal("negative cache changed failure/pin semantics")
			}
			if s_pSesStore.nCacheHits != 1 || s_pSesStore.nSatCalls != 0 || s_pSesStore.nEntriesCount != 1 || s_pSesStore.nUnsynthesizedImp != uint64(1-limited) || s_pSesStore.nUnsynthesizedRL != uint64(limited) {
				t.Fatal("negative cache retried synthesis or changed insertion statistics")
			}
		})
	}
}

func TestExactCostRetainsSourceFirstCachedNetworkPolicy(t *testing.T) {
	exactMappingReviewStore(t)
	truth := exactCacheReviewAndTruth(3)
	first := exactCacheReviewChain(3, 2)                                     // depths[2,2,1]
	second := string([]byte{3, 1, 2, 4, 2, 1, 2, 4, 2, 0, 3, 2, 2, 1, 2, 2}) // depths[1,2,2]
	for _, model := range []string{first, second} {
		key := []int32{0, 0, 0}
		if Ses_StoreAddEntry(s_pSesStore, &truth[0], 3, &key[0], model, 0) != 1 {
			t.Fatal("distinct normalized network key was lost")
		}
	}
	arrivals, pins := []int32{0, 0, 9}, make([]int8, 3)
	var cost int32
	// Source takes the newest matching entry. Its delay11 is valid, although
	// the older network would give10. AigLevel only bounds cache-miss search.
	if delay := Abc_ExactDelayCost(&truth[0], 3, &arrivals[0], pins, &cost, 10); delay != 11 || cost != 2 || !slices.Equal(pins, []int8{1, 2, 2}) || s_pSesStore.nSatCalls != 0 {
		t.Fatalf("source cache policy: delay%d cost%d pins%v", delay, cost, pins)
	}
}

func TestExactCostLargeCachedNetworkUsesNormalizedKey(t *testing.T) {
	exactMappingReviewStore(t)
	truth := exactCacheReviewAndTruth(2)
	key := []int32{0, 4}
	model := exactCacheReviewChain(2, 99)
	if Ses_StoreAddEntry(s_pSesStore, &truth[0], 2, &key[0], model, 0) != 1 {
		t.Fatal("large cache entry was not stored")
	}
	arrivals, pins := []int32{0, 0}, make([]int8, 2)
	var cost int32
	if delay := Abc_ExactDelayCost(&truth[0], 2, &arrivals[0], pins, &cost, 0); delay != 99 || cost != 99 || !slices.Equal(pins, []int8{99, 99}) || s_pSesStore.nCacheHits != 1 || s_pSesStore.nSatCalls != 0 {
		t.Fatalf("large normalized cache: delay%d cost%d pins%v", delay, cost, pins)
	}
}
