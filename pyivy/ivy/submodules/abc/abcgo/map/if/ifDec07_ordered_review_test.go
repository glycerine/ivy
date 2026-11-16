package if_

import (
	"fmt"
	"slices"
	"testing"
)

func decOrderedReviewConjunction(n int, negative uint, complement bool) []uint64 {
	truth := make([]uint64, 1<<uint(max(0, n-6)))
	for assignment := 0; assignment < len(truth)*64; assignment++ {
		value := true
		for pin := 0; pin < n; pin++ {
			value = value && ((assignment >> pin & 1) != int(negative>>pin&1))
		}
		if value != complement {
			truth[assignment>>6] |= 1 << uint(assignment&63)
		}
	}
	return truth
}

func TestDec07OriginalConfigurationOrder(t *testing.T) {
	// Hand-derived from the source's first successful configuration: six-input
	// free variables0,1 (bound2..5), seven-input free0,1,2 (bound3..6).
	// The source five-input cofactor convention fills constant groups with1.
	for _, tc := range []struct {
		n          int
		negative   uint
		complement bool
		want       uint64
	}{
		{5, 0, false, 0x2170800021438fff},
		{6, 0, false, 0x0710808054328000},
		{6, 0, true, 0x07107f7f54328000},
		{6, 1, false, 0x0710404054328000},
		{6, 4, false, 0x0710808054324000},
		{6, 0x3c, false, 0x071008085432fffe},
		{7, 0, false, 0x7210800065438000},
		{7, 0, true, 0x72107fff65438000},
		{7, 1, false, 0x7210400065438000},
		{7, 8, false, 0x7210800065434000},
		{7, 0x78, false, 0x721000806543fffe},
	} {
		t.Run(fmt.Sprintf("n%d/negative%x/complement%t", tc.n, tc.negative, tc.complement), func(t *testing.T) {
			truth := decOrderedReviewConjunction(tc.n, tc.negative, tc.complement)
			before := slices.Clone(truth)
			var got, exists uint64
			switch tc.n {
			case 5:
				got = If_Dec5Perform(truth[0], 1)
				exists = If_Dec5Perform(truth[0], 0)
			case 6:
				got = If_Dec6Perform(truth[0], 1)
				exists = If_Dec6Perform(truth[0], 0)
			case 7:
				got = If_Dec7Perform(truth, 1)
				exists = If_Dec7Perform(truth, 0)
			}
			if got != tc.want {
				t.Errorf("configuration%016x, want original%016x", got, tc.want)
			}
			if exists != 1 || !slices.Equal(truth, before) {
				t.Fatal("existence result or input ownership changed")
			}
			if decoded := decConfigTruth(tc.want, int32(tc.n)); !slices.Equal(decoded, truth) {
				t.Fatal("source-derived expected configuration does not compose to the input truth")
			}
		})
	}
}

func TestDec6PrefersLaterDisjointConfiguration(t *testing.T) {
	var truth uint64
	for a := 0; a < 64; a++ {
		if (a&15 == 15) != (a>>4&3 == 3) {
			truth |= 1 << a
		}
	}
	// F=AND(a,b,c,d) XOR AND(e,f). Earlier mixed free pairs have shared
	// decompositions. Source continues to free(e,f); accumulated moves leave
	// bound variable order(d,c,b,a), with the first bound cofactor coded zero.
	const want uint64 = 0x0754787801238000
	if got := If_Dec6Perform(truth, 1); got != want {
		t.Fatalf("configuration%016x, want source disjoint%016x", got, want)
	}
	if decoded := If_Dec6Truth(want); decoded != truth {
		t.Fatal("source-derived expected configuration has wrong truth")
	}
}

func TestDec07OrderedPermutationScalarRemap(t *testing.T) {
	state := uint64(0x123456789abcdef)
	next := func() uint64 { state ^= state << 13; state ^= state >> 7; state ^= state << 17; return state }
	for _, n := range []int{6, 7} {
		for trial := 0; trial < 128; trial++ {
			original := [2]uint64{next(), next()}
			current := original
			order := make([]int, n)
			for i := range order {
				order[i] = i
			}
			for i := n - 1; i > 0; i-- {
				j := int(next() % uint64(i+1))
				order[i], order[j] = order[j], order[i]
			}
			p2v, v2p := dec07Permutations()
			for place, variable := range order {
				if n == 6 {
					current[0] = dec07Move6(current[0], variable, place, &p2v, &v2p)
				} else {
					dec07Move7(&current, variable, place, &p2v, &v2p)
				}
				for p := 0; p < n; p++ {
					if p2v[p] < 0 || p2v[p] >= n || v2p[p2v[p]] != p {
						t.Fatalf("n%d trial%d inverse maps broken", n, trial)
					}
				}
				for assignment := 0; assignment < 1<<n; assignment++ {
					source := 0
					for pin := 0; pin < n; pin++ {
						source |= (assignment >> pin & 1) << p2v[pin]
					}
					want := original[source>>6] >> uint(source&63) & 1
					got := current[assignment>>6] >> uint(assignment&63) & 1
					if got != want {
						t.Fatalf("n%d trial%d step%d assignment%d bit%d want%d", n, trial, place, assignment, got, want)
					}
				}
			}
		}
	}
}

func TestDec7LaterFreeTripleConfiguration(t *testing.T) {
	var truth [2]uint64
	for assignment := 0; assignment < 128; assignment++ {
		if (assignment&15 == 15) != (assignment>>4&7 == 7) {
			truth[assignment>>6] |= 1 << uint(assignment&63)
		}
	}
	// Only free(e,f,g) gives two cofactors. Earlier moves leave bound(d,c,b,a).
	const want uint64 = 0x76547f8001238000
	if got := If_Dec7Perform(truth[:], 1); got != want {
		t.Fatalf("configuration%016x, want source%016x", got, want)
	}
	if decoded := decConfigTruth(want, 7); !slices.Equal(decoded, truth[:]) {
		t.Fatal("source-derived expected config has incorrect truth")
	}
}

func TestDec6ConfigurationTruthScalarAgreement(t *testing.T) {
	state := uint64(0xcafebabe1234567)
	next := func() uint64 { state ^= state << 13; state ^= state >> 7; state ^= state << 17; return state }
	for trial := 0; trial < 512; trial++ {
		config := next()&0xffff | (next()&0xffff)<<32
		for pin := 0; pin < 4; pin++ {
			for _, shift := range []uint{16, 48} {
				source := next() % 8
				if shift == 16 && source == 7 {
					source = 6
				}
				config |= source << (shift + uint(4*pin))
				if source == 6 {
					// Make the associated truth independent of this unused formal pin.
					truthShift := uint(0)
					if shift == 48 {
						truthShift = 32
					}
					table := uint16(config >> truthShift)
					var projected uint16
					for m := 0; m < 16; m++ {
						projected |= (table >> uint(m&^(1<<pin)) & 1) << uint(m)
					}
					config = config&^(uint64(0xffff)<<truthShift) | uint64(projected)<<truthShift
				}
			}
		}
		want := decConfigTruth(config, 6)[0]
		if got := If_Dec6Truth(config); got != want {
			t.Fatalf("config%016x truth%016x, want scalar%016x", config, got, want)
		}
	}
}

var dec07BenchmarkTruth uint64

func BenchmarkDec6ConfigurationTruth(b *testing.B) {
	const config uint64 = 0x0754787801238000
	b.ReportAllocs()
	var sum uint64
	for i := 0; i < b.N; i++ {
		sum ^= If_Dec6Truth(config)
	}
	dec07BenchmarkTruth = sum
}

func TestDec07ReducedSupportGoCompatibility(t *testing.T) {
	// The source six-input entry asserts full support. Preserve the earlier Go
	// behavior outside that precondition without retaining the generic search.
	for n := 5; n <= 7; n++ {
		for _, value := range []uint64{0, ^uint64(0), 0xaaaaaaaaaaaaaaaa, 0x8888888888888888, 0x8080808080808080} {
			truth := make([]uint64, 1<<uint(max(0, n-6)))
			for i := range truth {
				truth[i] = value
			}
			var config uint64
			switch n {
			case 5:
				config = If_Dec5Perform(value, 1)
			case 6:
				config = If_Dec6Perform(value, 1)
			case 7:
				config = If_Dec7Perform(truth, 1)
			}
			if config == 0 || !slices.Equal(decConfigTruth(config, int32(n)), truth) {
				t.Fatalf("n%d reduced truth%016x config%016x", n, value, config)
			}
		}
	}
}
