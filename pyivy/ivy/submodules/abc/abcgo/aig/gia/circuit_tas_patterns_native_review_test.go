package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestTasPatternsMatchOriginalABCAndScalarBits(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_tas_review.c")
	native, err := exec.Command(exe, "patterns").CombinedOutput()
	if err != nil {
		t.Fatalf("native TAS patterns: %v\n%s", err, native)
	}
	matrix := func() [][]uint32 {
		v := make([][]uint32, 5)
		for i := range v {
			v[i] = make([]uint32, 3)
		}
		return v
	}
	data, present := matrix(), matrix()
	var got strings.Builder
	for seed := 0; seed < 8; seed++ {
		random := uint32(seed + 103)
		var scalarData, scalarPresent [5][96]bool
		for i := 0; i < 5; i++ {
			for word := 0; word < 3; word++ {
				random = random*1664525 + 1013904223
				data[i][word] = random
				random = random*1664525 + 1013904223
				present[i][word] = 0
				if seed&1 != 0 {
					present[i][word] = random
				}
				for b := 0; b < 32; b++ {
					scalarData[i][32*word+b] = data[i][word]&(1<<uint(b)) != 0
					scalarPresent[i][32*word+b] = present[i][word]&(1<<uint(b)) != 0
				}
			}
		}
		boundaries := []int32{0, 1, 31, 32, 33, 63, 64, 95}
		for trial := 0; trial < 64; trial++ {
			count := trial % 6
			bit := boundaries[(trial+seed)%len(boundaries)]
			lits := make([]int32, count)
			for i := range lits {
				random = random*1664525 + 1013904223
				lits[i] = int32(2*((trial+2*i)%5)) + int32(random>>31)
			}
			fit := true
			for _, lit := range lits {
				if scalarPresent[lit>>1][bit] && scalarData[lit>>1][bit] != (lit&1 == 0) {
					fit = false
				}
			}
			if fit {
				for _, lit := range lits {
					scalarPresent[lit>>1][bit] = true
					scalarData[lit>>1][bit] = lit&1 == 0
				}
			}
			result := Tas_StorePatternTry(data, present, bit, lits, int32(count))
			if result != giaBool(fit) {
				t.Fatalf("seed%d trial%d fit%d expected%v", seed, trial, result, fit)
			}
			fmt.Fprintf(&got, "try %d %d %d %d %d", seed, trial, bit, count, result)
			for i := 0; i < 5; i++ {
				for word := 0; word < 3; word++ {
					var expectedData, expectedPresent uint32
					for b := 0; b < 32; b++ {
						if scalarData[i][32*word+b] {
							expectedData |= 1 << uint(b)
						}
						if scalarPresent[i][32*word+b] {
							expectedPresent |= 1 << uint(b)
						}
					}
					if data[i][word] != expectedData || present[i][word] != expectedPresent {
						t.Fatalf("scalar bitmap mismatch seed%d trial%d variable%d word%d", seed, trial, i, word)
					}
					fmt.Fprintf(&got, " %d:%d", data[i][word], present[i][word])
				}
			}
			got.WriteByte('\n')
		}
	}
	data, present = matrix(), matrix()
	var slots [32]int
	for bit := range slots {
		slots[bit] = -1
	}
	for trial := 0; trial < 64; trial++ {
		lits := make([]int32, 5)
		for i := range lits {
			lits[i] = int32(2*i + (((trial >> i) & 1) ^ 1))
		}
		expected := int32(0)
		for bit := 1; bit < 32; bit++ {
			if slots[bit] < 0 || slots[bit] == trial&31 {
				slots[bit] = trial & 31
				expected = 1
				break
			}
		}
		result := Tas_StorePattern(data, present, lits)
		if result != expected {
			t.Fatalf("pack trial%d result%d expected%d", trial, result, expected)
		}
		fmt.Fprintf(&got, "pack %d %d", trial, result)
		for i := 0; i < 5; i++ {
			var expectedData, expectedPresent uint32
			for bit := 1; bit < 32; bit++ {
				if slots[bit] >= 0 {
					expectedPresent |= 1 << uint(bit)
					if slots[bit]&(1<<uint(i)) != 0 {
						expectedData |= 1 << uint(bit)
					}
				}
			}
			if data[i][0] != expectedData || present[i][0] != expectedPresent || data[i][1] != 0 || present[i][1] != 0 {
				t.Fatalf("packed pattern mismatch trial%d variable%d", trial, i)
			}
			fmt.Fprintf(&got, " %d:%d", data[i][0], present[i][0])
		}
		got.WriteByte('\n')
	}
	if got.String() != string(native) {
		t.Fatal("native TAS bitmap snapshots differ from Go/scalar reference")
	}
	t.Log("512 exact native/scalar cross-word merges and 64 packing/saturation cases")
}
