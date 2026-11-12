package extrab

import "testing"

func stretchedTruth(table uint64, n int) []uint64 {
	t := make([]uint64, thresholdWords(n))
	for a := 0; a < len(t)*64; a++ {
		if table>>uint(a&((1<<uint(n))-1))&1 != 0 {
			thresholdSet(t, a, true)
		}
	}
	return t
}
func TestOriginalCThresholdFixture(t *testing.T) {
	var hash uint64 = 14695981039346656037
	number := func(v uint64) {
		for i := 0; i < 8; i++ {
			hash ^= v >> uint(i*8) & 255
			hash *= 1099511628211
		}
	}
	for n := 3; n <= 4; n++ {
		for table := uint64(0); table < uint64(1)<<uint(1<<uint(n)); table++ {
			truth := stretchedTruth(table, n)
			weights, chow := make([]int, n), make([]int, n)
			number(uint64(Extra_ThreshComputeChow(truth, n, chow)))
			for _, v := range chow {
				number(uint64(v))
			}
			threshold := Extra_ThreshCheck(truth, n, weights)
			number(uint64(threshold))
			number(truth[0])
			for _, v := range weights {
				number(uint64(v))
			}
			if threshold != 0 {
				for a := 0; a < 1<<uint(n); a++ {
					if (Extra_ThreshWeightedSum(weights, n, a) >= threshold) != thresholdBit(truth, a) {
						t.Fatal("invalid threshold certificate", n, table, a)
					}
				}
			}
		}
	}
	if hash != 0x23d92fb53fb7aab7 {
		t.Fatalf("original C threshold checksum: %016x", hash)
	}
}
func TestThresholdPrimeSOPsAndHeuristic(t *testing.T) {
	found := 0
	for table := uint64(1); table < 65535; table++ {
		truth := stretchedTruth(table, 4)
		monotone := true
		for v := 0; v < 4; v++ {
			pos, _ := thresholdPolarity(truth, 4, v)
			if !pos {
				monotone = false
				break
			}
		}
		if !monotone {
			continue
		}
		on, off := thresholdPrimeSOP(truth, 4, true), thresholdPrimeSOP(truth, 4, false)
		for _, part := range []struct {
			sop  string
			want bool
		}{{on, true}, {off, false}} {
			for a := 0; a < 16; a++ {
				covered := false
				for start := 0; start < len(part.sop); start += 7 {
					match := true
					for v := 0; v < 4; v++ {
						c := part.sop[start+v]
						if c != '-' && (c == '1') != (a&(1<<uint(v)) != 0) {
							match = false
						}
					}
					if match {
						covered = true
					}
				}
				if covered != (thresholdBit(truth, a) == part.want) {
					t.Fatal("prime cover truth", table, a, part.want)
				}
			}
		}
		weights := make([]int, 4)
		threshold := Extra_ThreshHeuristic(truth, 4, weights)
		if threshold == 0 {
			continue
		}
		found++
		// Source heuristic leaves the sorted positive-unate truth table complemented.
		for a := 0; a < 16; a++ {
			if (Extra_ThreshWeightedSum(weights, 4, a) >= threshold) == thresholdBit(truth, a) {
				t.Fatal("heuristic certificate", table, a, weights, threshold)
			}
		}
	}
	if found < 100 {
		t.Fatal("insufficient heuristic coverage", found)
	}
}
func TestThresholdHeuristicSixVariableBoundedBuffer(t *testing.T) {
	truth, n := thresholdExample()
	backing := []uint64{truth[0], 0x1122334455667788}
	weights := make([]int, n)
	threshold := Extra_ThreshHeuristic(backing[:1], n, weights)
	if backing[1] != 0x1122334455667788 {
		t.Fatal("overwrote adjacent truth word")
	}
	if threshold != 0 {
		for a := 0; a < 64; a++ {
			if (Extra_ThreshWeightedSum(weights, n, a) >= threshold) == thresholdBit(backing, a) {
				t.Fatal("six-var certificate")
			}
		}
	}
}
func TestThresholdInvalidAndBinate(t *testing.T) {
	truth := stretchedTruth(6, 2)
	original := truth[0]
	if Extra_ThreshCheck(truth, 2, make([]int, 2)) != 0 || truth[0] != original {
		t.Fatal("binate truth changed")
	}
	if Extra_ThreshComputeChow(nil, 4, make([]int, 4)) != 0 || Extra_ThreshCheck([]uint64{0}, 17, make([]int, 17)) != 0 {
		t.Fatal("invalid parameters accepted")
	}
	if Extra_ThreshAssignWeights(nil, "bad", "", 4, make([]int, 4), make([]int, 4), 1, 1) != 0 {
		t.Fatal("invalid SOP accepted")
	}
}

func TestOriginalCThresholdHeuristicFixture(t *testing.T) {
	var hash uint64 = 14695981039346656037
	number := func(v uint64) {
		for i := 0; i < 8; i++ {
			hash ^= v >> uint(i*8) & 255
			hash *= 1099511628211
		}
	}
	for n := 3; n <= 4; n++ {
		for table := uint64(1); table < (uint64(1)<<uint(1<<uint(n)))-1; table++ {
			truth := stretchedTruth(table, n)
			weights := make([]int, n)
			number(uint64(Extra_ThreshHeuristic(truth, n, weights)))
			number(truth[0])
			for _, v := range weights {
				number(uint64(v))
			}
		}
	}
	if hash != 0x14b0eda9522e02d5 {
		t.Fatalf("original inequality heuristic checksum: %016x", hash)
	}
}
