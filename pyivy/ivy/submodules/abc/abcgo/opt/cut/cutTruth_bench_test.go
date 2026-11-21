package cut

import (
	"strconv"
	"testing"
)

// Keep every merged variable in at least one child, with overlap at every
// third variable. The computed functions exercise every word of the result.
func benchmarkTruthCuts(n int32) (c, a, b *Cut_Cut_t) {
	words := Cut_TruthWords(n)
	c = &Cut_Cut_t{NVarsMax: n, truth: make([]uint32, words)}
	a = &Cut_Cut_t{NVarsMax: n, truth: make([]uint32, words)}
	b = &Cut_Cut_t{NVarsMax: n, truth: make([]uint32, words)}
	for i := int32(0); i < n; i++ {
		c.leaves = append(c.leaves, i)
		if i%3 != 1 {
			a.leaves = append(a.leaves, i)
		}
		if i%3 != 0 {
			b.leaves = append(b.leaves, i)
		}
	}
	fill := func(child *Cut_Cut_t, seed uint32) {
		mask := 1<<len(child.leaves) - 1
		for bit := 0; bit < 32*len(child.truth); bit++ {
			index := bit & mask
			word := seed + 0x9e3779b9*uint32(index/32)
			child.truth[bit/32] |= (word >> uint(index%32) & 1) << uint(bit%32)
		}
	}
	fill(a, 0xf185349a)
	fill(b, 0x8b234ce7)
	return
}

func BenchmarkCutTruthMerge(b *testing.B) {
	for _, n := range []int32{4, 7, 12} {
		b.Run(strconv.Itoa(int(n)), func(b *testing.B) {
			c, left, right := benchmarkTruthCuts(n)
			b.ReportAllocs()
			for b.Loop() {
				cutTruthComputeOp(c, left, right, 1, 0, false)
			}
		})
	}
}

func TestCutTruthNativeBenchmarks(t *testing.T) {
	// Wordwise FNV hashes from test_vectors/perf_truth.c.
	for _, test := range []struct {
		width int32
		want  uint32
	}{{4, 0xe000cc70}, {7, 0x7b27cefa}, {12, 0x1f3692b1}} {
		c, left, right := benchmarkTruthCuts(test.width)
		cutTruthComputeOp(c, left, right, 1, 0, false)
		hash := uint32(2166136261)
		for _, word := range c.truth {
			hash = (hash ^ word) * 16777619
		}
		if hash != test.want {
			t.Fatalf("%d variables: truth hash %08x, C produced %08x", test.width, hash, test.want)
		}
	}
}

func TestCutTruthOutputMayBeRightInput(t *testing.T) {
	for _, width := range []int32{4, 7, 12} {
		params := DefaultParams(width)
		params.FTruth = 1
		p := Cut_ManStart(params)
		a := Cut_CutCreateTriv(p, 0)
		right := Cut_CutAlloc(p)
		for i := range width {
			right.leaves = append(right.leaves, i)
		}
		want := make([]uint32, len(right.truth))
		for i := range right.truth {
			right.truth[i] = uint32(0xcccccccc) ^ 0x2468acf*uint32(i)
			want[i] = 0xaaaaaaaa & right.truth[i]
		}
		Cut_TruthCompute(p, right, a, right, 0, 0)
		for i, word := range right.truth {
			if word != want[i] {
				t.Errorf("width %d word %d: in-place right input produced %08x, want %08x", width, i, word, want[i])
				break
			}
		}
		Cut_ManStop(p)
	}
}

func BenchmarkCutTruthPublic4(b *testing.B) {
	params := DefaultParams(4)
	params.FTruth = 1
	manager := Cut_ManStart(params)
	defer Cut_ManStop(manager)
	c, a, right := benchmarkTruthCuts(4)
	b.ReportAllocs()
	for b.Loop() {
		Cut_TruthCompute(manager, c, a, right, 1, 0)
	}
}
