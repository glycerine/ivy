package abci

import (
	"fmt"
	"slices"
	"testing"
)

func TestRecordedTruthStretchAllWidthsAndUntouchedSuffix(t *testing.T) {
	for source := int32(0); source <= LMS_VAR_MAX; source++ {
		for target := source; target <= LMS_VAR_MAX; target++ {
			// A full source workspace with distinct halves and poisoned tail.
			words := make([]word, LMS_MAX_WORD+1)
			for i := range words {
				words[i] = word(uint64(i+1)*0x123456789abcdef ^ 0xa5b6c7d8e9fa0123)
			}
			before := wordsToUint32(words)
			want := slices.Clone(before)
			step, limit := 1, 1
			if source > 5 {
				step = 1 << (source - 5)
			}
			if target > 5 {
				limit = 1 << (target - 5)
			}
			// Independent scalar 32-bit specification of the C helper.
			for i := step; i < limit; i++ {
				want[i] = before[i%step]
			}
			lmsStretch5(words, source, target)
			if got := wordsToUint32(words); !slices.Equal(got, want) {
				t.Fatalf("stretch%d->%d changed source repetition or workspace suffix", source, target)
			}
		}
	}
}

func TestRecordedTruthStretchNeedsNoTemporaryAllocation(t *testing.T) {
	words := make([]word, LMS_MAX_WORD)
	words[0] = 0x123456789abcdef
	for _, widths := range [][2]int32{{3, 6}, {6, 16}, {7, 16}, {16, 16}} {
		t.Run(fmt.Sprintf("%d_to_%d", widths[0], widths[1]), func(t *testing.T) {
			if allocations := testing.AllocsPerRun(20, func() { lmsStretch5(words, widths[0], widths[1]) }); allocations != 0 {
				t.Fatalf("source in-place repetition allocated%f temporary buffers", allocations)
			}
		})
	}
}

func BenchmarkRecordedTruthStretch(b *testing.B) {
	for _, widths := range [][2]int32{{3, 6}, {6, 16}, {7, 16}, {16, 16}} {
		b.Run(fmt.Sprintf("%d_to_%d", widths[0], widths[1]), func(b *testing.B) {
			words := make([]word, LMS_MAX_WORD)
			words[0] = 0x123456789abcdef
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lmsStretch5(words, widths[0], widths[1])
			}
		})
	}
}
