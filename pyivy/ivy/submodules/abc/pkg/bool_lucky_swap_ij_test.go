package abc

// Native branch-oracle tests for bool/lucky/luckySwapIJ.c.

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type luckySwapIJOracle struct {
	Cases  int       `json:"cases"`
	Hashes [4]string `json:"hashes"`
}

func luckySwapIJNext(state *uint64) uint64 {
	x := *state
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*state = x
	return x
}

func luckySwapIJCanonicalWord(word uint64, variables int) uint64 {
	if variables >= 6 {
		return word
	}
	period := 1 << uint(variables)
	word &= (uint64(1) << uint(period)) - 1
	for shift := period; shift < 64; shift *= 2 {
		word |= word << uint(shift)
	}
	return word
}

func luckySwapIJHash(hash, value uint64) uint64 {
	return (hash ^ value) * 1099511628211
}

func luckySwapIJCategory(first, second int) int {
	if first == second {
		return 0
	}
	if second <= 5 {
		return 1
	}
	if first <= 5 {
		return 2
	}
	return 3
}

func TestLuckySwapIJNativeBranches(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/lucky_swap_ij.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle luckySwapIJOracle
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	hashes := [4]uint64{1469598103934665603, 1469598103934665603, 1469598103934665603, 1469598103934665603}
	state := uint64(0x9e3779b97f4a7c15)
	cases := 0
	for variables := 1; variables <= 12; variables++ {
		count, err := TruthTableWordCount(variables)
		if err != nil {
			t.Fatal(err)
		}
		for first := 0; first < variables; first++ {
			for second := first; second < variables; second++ {
				words := make([]uint64, count)
				for i := range words {
					words[i] = luckySwapIJNext(&state)
				}
				if variables < 6 {
					words[0] = luckySwapIJCanonicalWord(words[0], variables)
				}
				table, err := TruthTableFromWords(variables, words, 0)
				if err != nil {
					t.Fatal(variables, err)
				}
				swapped, err := table.SwapVariables(first, second)
				if err != nil {
					t.Fatal(variables, first, second, err)
				}
				category := luckySwapIJCategory(first, second)
				hashes[category] = luckySwapIJHash(hashes[category], uint64(variables))
				hashes[category] = luckySwapIJHash(hashes[category], uint64(first))
				hashes[category] = luckySwapIJHash(hashes[category], uint64(second))
				for _, word := range swapped.Words() {
					hashes[category] = luckySwapIJHash(hashes[category], word)
				}
				cases++
			}
		}
	}
	if cases != oracle.Cases {
		t.Fatal(cases, oracle.Cases)
	}
	for category := range hashes {
		if got := fmt.Sprintf("%016x", hashes[category]); got != oracle.Hashes[category] {
			t.Fatal(category, got, oracle.Hashes[category])
		}
	}
}

func FuzzLuckySwapIJ(f *testing.F) {
	f.Add(uint16(0x6996), uint8(4), uint8(0), uint8(3))
	f.Add(uint16(0xa55a), uint8(9), uint8(2), uint8(8))
	f.Fuzz(func(t *testing.T, seed uint16, rawVariables, rawFirst, rawSecond uint8) {
		variables := 1 + int(rawVariables%10)
		first, second := int(rawFirst)%variables, int(rawSecond)%variables
		assignments := 1 << uint(variables)
		words := make([]uint64, (assignments+63)/64)
		for assignment := 0; assignment < assignments; assignment++ {
			bit := uint16(1) << uint(assignment%16)
			if seed&bit != 0 {
				words[assignment/64] |= uint64(1) << uint(assignment%64)
			}
		}
		if variables < 6 {
			words[0] = luckySwapIJCanonicalWord(words[0], variables)
		}
		table, err := TruthTableFromWords(variables, words, 0)
		if err != nil {
			t.Fatal(err)
		}
		swapped, err := table.SwapVariables(first, second)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < assignments; assignment++ {
			sourceAssignment := assignment
			if (assignment>>uint(first))&1 != (assignment>>uint(second))&1 {
				sourceAssignment ^= (1 << uint(first)) | (1 << uint(second))
			}
			got, err := swapped.Get(uint64(assignment))
			if err != nil || got != (seed&(uint16(1)<<uint(sourceAssignment%16)) != 0) {
				t.Fatal(variables, first, second, assignment, sourceAssignment, got, err)
			}
		}
	})
}
