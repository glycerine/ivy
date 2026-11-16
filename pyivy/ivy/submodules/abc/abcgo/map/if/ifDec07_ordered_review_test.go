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
