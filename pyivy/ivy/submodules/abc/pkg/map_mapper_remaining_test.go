package abc

// Standalone tests for portable mapper cut, truth, match, and timing semantics.

import (
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
)

func TestMapperCompiledSourceInventory(t *testing.T) {
	want := []string{"mapper.c", "mapperCanon.c", "mapperCore.c", "mapperCreate.c", "mapperCut.c", "mapperCutUtils.c", "mapperLib.c", "mapperMatch.c", "mapperRefs.c", "mapperSuper.c", "mapperSwitch.c", "mapperTable.c", "mapperTime.c", "mapperTree.c", "mapperTruth.c", "mapperUtils.c", "mapperVec.c"}
	got := MapperCompiledSources()
	if !slices.Equal(got, want) {
		t.Fatalf("inventory mismatch: %v", got)
	}
	got[0] = "changed"
	if MapperCompiledSources()[0] != "mapper.c" {
		t.Fatal("inventory aliases internal storage")
	}
}

func TestMapperMergeCutLeavesExhaustive(t *testing.T) {
	for maskA := 0; maskA < 1<<6; maskA++ {
		for maskB := 0; maskB < 1<<6; maskB++ {
			a, b, want := []int{}, []int{}, []int{}
			for i := 0; i < 6; i++ {
				if maskA&(1<<i) != 0 {
					a = append(a, i)
				}
				if maskB&(1<<i) != 0 {
					b = append(b, i)
				}
				if (maskA|maskB)&(1<<i) != 0 {
					want = append(want, i)
				}
			}
			got, err := MapperMergeCutLeaves(a, b, 6, MapperLimits{MaxWork: 100, MaxBytes: 48})
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("%x %x: %v %v want %v", maskA, maskB, got, err, want)
			}
			if len(want) > 3 {
				if got, err = MapperMergeCutLeaves(a, b, 3, MapperLimits{}); !errors.Is(err, ErrDimension) && !errors.Is(err, ErrLimit) || got != nil {
					t.Fatalf("capacity: %v %v", got, err)
				}
			}
		}
	}
	for _, tc := range []struct{ a, b []int }{{[]int{2, 1}, nil}, {[]int{1, 1}, nil}, {[]int{-1}, nil}} {
		if got, err := MapperMergeCutLeaves(tc.a, tc.b, 4, MapperLimits{}); !errors.Is(err, ErrDimension) || got != nil {
			t.Fatal(got, err)
		}
	}
	if got, err := MapperMergeCutLeaves(nil, nil, int(^uint(0)>>1), MapperLimits{}); !errors.Is(err, ErrDimension) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := MapperMergeCutLeaves([]int{1}, []int{2}, 2, MapperLimits{MaxBytes: 15}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := MapperMergeCutLeaves([]int{1}, []int{2}, 2, MapperLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || got != nil {
		t.Fatal(got, err)
	}
	polls := 0
	if got, err := MapperMergeCutLeaves([]int{0, 2, 4}, []int{1, 3, 5}, 6, MapperLimits{Stop: func() bool { polls++; return polls == 5 }}); !errors.Is(err, ErrCancelled) || got != nil {
		t.Fatal(got, err)
	}
}

func TestMapperTruthsAndPhaseExhaustive(t *testing.T) {
	for n := 0; n <= 6; n++ {
		truths, err := MapperElementaryTruths(n, MapperLimits{MaxWork: 1000, MaxBytes: 48})
		if err != nil {
			t.Fatal(err)
		}
		for v := 0; v < n; v++ {
			for m := 0; m < 64; m++ {
				if got := truths[v]&(uint64(1)<<m) != 0; got != (m&(1<<v) != 0) {
					t.Fatal(n, v, m)
				}
			}
		}
		seeds := []uint64{0, ^uint64(0), 0x6996966996696996, 0xaaaaaaaaaaaaaaaa}
		for _, truth := range seeds {
			for phase := 0; phase < 1<<n; phase++ {
				got, err := MapperTransformTruth(truth, n, uint32(phase), MapperLimits{MaxWork: 64})
				if err != nil {
					t.Fatal(err)
				}
				for m := 0; m < 64; m++ {
					want := truth&(uint64(1)<<uint(m^phase)) != 0
					if (got&(uint64(1)<<uint(m)) != 0) != want {
						t.Fatal(n, phase, m)
					}
				}
			}
		}
	}
	identity := uint64(0x0123456789abcdef)
	if got, err := MapperTransformTruth(identity, 0, 0, MapperLimits{MaxWork: 64}); err != nil || got != identity {
		t.Fatalf("zero-variable transform = %x, %v", got, err)
	}
	if _, err := MapperTransformTruth(0, 5, 1<<5, MapperLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if got, err := MapperElementaryTruths(6, MapperLimits{MaxWork: 1}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := MapperElementaryTruths(6, MapperLimits{MaxWork: 383}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := MapperElementaryTruths(6, MapperLimits{MaxBytes: 47}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	polls := 0
	if got, err := MapperElementaryTruths(6, MapperLimits{Stop: func() bool { polls++; return polls == 19 }}); !errors.Is(err, ErrCancelled) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := MapperTransformTruth(^uint64(0), 6, 0, MapperLimits{MaxWork: 63}); !errors.Is(err, ErrLimit) || got != 0 {
		t.Fatal(got, err)
	}
}

func TestMapperCutLeafPhaseExhaustive(t *testing.T) {
	for n := 1; n <= 8; n++ {
		for phase := uint32(0); phase < 1<<uint(n); phase++ {
			for leaf := 0; leaf < n; leaf++ {
				got, err := MapperCutLeafPhase(phase, leaf, n)
				want := phase&(1<<uint(leaf)) == 0
				if err != nil || got != want {
					t.Fatalf("n=%d phase=%x leaf=%d: %v %v", n, phase, leaf, got, err)
				}
			}
		}
	}
	for _, tc := range []struct {
		phase     uint32
		leaf      int
		leafCount int
	}{{0, -1, 1}, {0, 0, 0}, {0, 0, 33}, {2, 0, 1}} {
		if _, err := MapperCutLeafPhase(tc.phase, tc.leaf, tc.leafCount); !errors.Is(err, ErrDimension) {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
}

func mapperTestMatch(arrive, area float64, fanout, fanins int) MapperMatch {
	return MapperMatch{Arrival: MapperTime{Rise: arrive, Fall: arrive, Worst: arrive}, AreaFlow: area, Super: MapperSupergate{Area: area, FanoutLimit: fanout, Fanins: fanins}}
}

func TestMapperMatchOrdering(t *testing.T) {
	old := mapperTestMatch(10, 5, 4, 3)
	cases := []struct {
		candidate       MapperMatch
		areaFirst, want bool
	}{
		{mapperTestMatch(9, 100, 1, 6), false, true},
		{mapperTestMatch(10, 4, 1, 6), false, true},
		{mapperTestMatch(20, 4, 1, 6), true, true},
		{mapperTestMatch(10, 5, 5, 6), false, true},
		{mapperTestMatch(10, 5, 4, 2), false, true},
		{old, false, false},
	}
	for _, tc := range cases {
		got, err := MapperCompareMatches(old, tc.candidate, MapperMatchOrder{Epsilon: 1e-6, AreaFirst: tc.areaFirst})
		if err != nil || got != tc.want {
			t.Fatal(tc, got, err)
		}
	}
	bad := old
	bad.AreaFlow = math.NaN()
	if _, err := MapperCompareMatches(old, bad, MapperMatchOrder{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	bad = old
	bad.Arrival.Worst = bad.Arrival.Rise - 1
	if _, err := MapperCompareMatches(old, bad, MapperMatchOrder{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestMapperCutArrivalModel(t *testing.T) {
	arrivals := [][2]MapperTime{
		{{Rise: 1, Fall: 2, Worst: 2}, {Rise: 3, Fall: 4, Worst: 4}},
		{{Rise: 5, Fall: 6, Worst: 6}, {Rise: 7, Fall: 8, Worst: 8}},
	}
	gate := MapperSupergate{Area: 1, FanoutLimit: 10, Fanins: 2, PinDelays: []MapperPinDelay{
		{ToRise: MapperTime{Rise: 1, Fall: 2, Worst: 2}, ToFall: MapperTime{Rise: 3, Fall: 4, Worst: 4}},
		{ToRise: MapperTime{Rise: 2, Fall: 1, Worst: 2}, ToFall: MapperTime{Rise: 4, Fall: 3, Worst: 4}},
	}}
	for phase := uint32(0); phase < 4; phase++ {
		got, err := MapperCutArrival(arrivals, gate, phase, .5, 100, MapperLimits{MaxWork: 8, MaxBytes: 72})
		if err != nil {
			t.Fatal(err)
		}
		wantR, wantF := 0.0, 0.0
		for i := range arrivals {
			p := 0
			if phase&(1<<i) == 0 {
				p = 1
			}
			in, d := arrivals[i][p], gate.PinDelays[i]
			wantR = math.Max(wantR, math.Max(in.Rise+d.ToRise.Rise+.5, in.Fall+d.ToRise.Fall+.5))
			wantF = math.Max(wantF, math.Max(in.Rise+d.ToFall.Rise+.5, in.Fall+d.ToFall.Fall+.5))
		}
		if got.Rise != wantR || got.Fall != wantF || got.Worst != math.Max(wantR, wantF) {
			t.Fatal(phase, got, wantR, wantF)
		}
	}
	if got, err := MapperCutArrival(arrivals, gate, 0, .5, 1, MapperLimits{}); !errors.Is(err, ErrLimit) || got != (MapperTime{}) {
		t.Fatal(got, err)
	}
	if got, err := MapperCutArrival(arrivals, gate, 0, .5, 100, MapperLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || got != (MapperTime{}) {
		t.Fatal(got, err)
	}
	if got, err := MapperCutArrival(arrivals, gate, 0, .5, 100, MapperLimits{MaxWork: 7}); !errors.Is(err, ErrLimit) || got != (MapperTime{}) {
		t.Fatal(got, err)
	}
	if got, err := MapperCutArrival(arrivals, gate, 0, .5, 100, MapperLimits{MaxBytes: 71}); !errors.Is(err, ErrLimit) || got != (MapperTime{}) {
		t.Fatal(got, err)
	}
	// Native supergates use nonpositive delays for absent transitions.
	absent := gate
	absent.PinDelays = append([]MapperPinDelay(nil), gate.PinDelays...)
	absent.PinDelays[0] = MapperPinDelay{
		ToRise: MapperTime{Rise: -1, Fall: -1, Worst: -1},
		ToFall: MapperTime{Rise: -1, Fall: -1, Worst: -1},
	}
	got, err := MapperCutArrival(arrivals, absent, 0, .5, 100, MapperLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Rise != 9.5 || got.Fall != 11.5 || got.Worst != 11.5 {
		t.Fatalf("absent pin arcs contributed: %+v", got)
	}
}

func TestMapperConcurrentPureCalls(t *testing.T) {
	var wg sync.WaitGroup
	for k := 0; k < 8; k++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				got, err := MapperTransformTruth(uint64(seed+i)*0x9e3779b97f4a7c15, 6, uint32(i&63), MapperLimits{})
				if err != nil || got == 0 && seed+i != 0 {
					t.Errorf("%x %v", got, err)
					return
				}
			}
		}(k)
	}
	wg.Wait()
}

func FuzzMapperCutAndPhase(f *testing.F) {
	for _, seed := range []struct {
		a, b  uint8
		truth uint64
		phase uint8
	}{{0, 0, 0, 0}, {3, 5, 0x96, 3}, {63, 42, 0x6996966996696996, 63}} {
		f.Add(seed.a, seed.b, seed.truth, seed.phase)
	}
	f.Fuzz(func(t *testing.T, ma, mb uint8, truth uint64, rawPhase uint8) {
		a, b, want := []int{}, []int{}, []int{}
		for i := 0; i < 6; i++ {
			if ma&(1<<i) != 0 {
				a = append(a, i)
			}
			if mb&(1<<i) != 0 {
				b = append(b, i)
			}
			if (ma|mb)&(1<<i) != 0 {
				want = append(want, i)
			}
		}
		got, err := MapperMergeCutLeaves(a, b, 6, MapperLimits{MaxWork: 100, MaxBytes: 48})
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, err)
		}
		phase := uint32(rawPhase & 63)
		transformed, err := MapperTransformTruth(truth, 6, phase, MapperLimits{MaxWork: 64})
		if err != nil {
			t.Fatal(err)
		}
		back, err := MapperTransformTruth(transformed, 6, phase, MapperLimits{MaxWork: 64})
		if err != nil || back != truth {
			t.Fatal(back, truth, err)
		}
	})
}
