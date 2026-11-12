package if_

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func decTestTruth(n int32, eval func(int) bool) []uint64 {
	t := make([]uint64, decWords(n))
	for m := 0; m < len(t)*64; m++ {
		if eval(m & ((1 << uint(n)) - 1)) {
			decSet(t, m)
		}
	}
	return t
}
func decTest32(t []uint64) []uint32 {
	out := make([]uint32, 2*len(t))
	for i, v := range t {
		out[2*i], out[2*i+1] = uint32(v), uint32(v>>32)
	}
	return out
}
func TestDec07GeneratedCascades(t *testing.T) {
	rng := rand.New(rand.NewSource(7719))
	for n := 5; n <= 7; n++ {
		for trial := 0; trial < 80; trial++ {
			order := rng.Perm(n)
			bound := order[:4]
			free := append([]int(nil), order[4:]...)
			for _, v := range bound[:7-n] {
				free = append(free, v)
			}
			config := uint64(uint16(rng.Uint32())) | uint64(uint16(rng.Uint32()))<<32
			for pin, v := range bound {
				config |= uint64(v) << uint(16+4*pin)
			}
			for pin, v := range free {
				config |= uint64(v) << uint(48+4*pin)
			}
			config |= uint64(7) << 60
			truth := decConfigTruth(config, int32(n))
			before := append([]uint64(nil), truth...)
			var derived uint64
			switch n {
			case 5:
				derived = If_Dec5Perform(truth[0], 1)
			case 6:
				derived = If_Dec6Perform(truth[0], 1)
			case 7:
				derived = If_Dec7Perform(truth, 1)
			}
			if derived == 0 {
				t.Fatal("lost generated cascade", n, trial)
			}
			actual := decConfigTruth(derived, int32(n))
			if !reflect.DeepEqual(actual, truth) {
				t.Fatal("configuration mismatch", n, trial)
			}
			if !reflect.DeepEqual(truth, before) {
				t.Fatal("input mutation")
			}
		}
	}
	If_Dec5PerformTest()
}
func TestDecSupportMinimumBaseAndMux(t *testing.T) {
	truth := decTestTruth(7, func(m int) bool { return m>>1&1 != 0 && m>>6&1 != 0 })
	support := []int32{10, 11, 12, 13, 14, 15, 16}
	var count int32
	If_Dec7MinimumBase(truth, support, 7, &count)
	if count != 2 || support[0] != 11 || support[1] != 16 || truth[0] != 0x8888888888888888 || truth[1] != truth[0] {
		t.Fatal("minimum base", support, count, truth)
	}
	mux := decTestTruth(7, func(m int) bool {
		if m&1 == 0 {
			return m&0xe == 0xe
		}
		return m&0x70 == 0x70
	})
	var zero, one [2]uint64
	selected := If_Dec7PickBestMux(mux, zero[:], one[:])
	if selected < 0 {
		t.Fatal("no mux decomposition")
	}
	for m := 0; m < 128; m++ {
		cofactor := zero[:]
		if m>>uint(selected)&1 != 0 {
			cofactor = one[:]
		}
		if decBit(cofactor, m) != decBit(mux, m) {
			t.Fatal("MUX cofactor mismatch", m)
		}
	}
	for variable := int32(0); variable < 7; variable++ {
		var a, b [2]uint64
		If_Dec08Cofactors(mux, 7, variable, a[:], b[:])
		for m := 0; m < 128; m++ {
			if decBit(a[:], m) != decBit(mux, m&^(1<<uint(variable))) || decBit(b[:], m) != decBit(mux, m|(1<<uint(variable))) {
				t.Fatal("cofactor", variable, m)
			}
		}
	}
}
func TestDecCompatibilityAndTruthAdapters(t *testing.T) {
	for n := int32(6); n <= 10; n++ {
		parity := decTestTruth(n, func(m int) bool {
			ones := 0
			for m != 0 {
				ones++
				m &= m - 1
			}
			return ones%2 == 1
		})
		before := append([]uint64(nil), parity...)
		if If_Dec10Perform(parity, n, 0) != 1 {
			t.Fatal("parity rejected", n)
		}
		if n <= 8 && If_Dec08Perform(parity, n, 0) != 1 {
			t.Fatal("Dec08 parity rejected", n)
		}
		if !reflect.DeepEqual(parity, before) {
			t.Fatal("compatibility check changed input")
		}
		if If_CutPerformCheck10(nil, decTest32(parity), n, n, "") != 1 {
			t.Fatal("truth adapter")
		}
	}
	irrelevant := decTestTruth(6, func(m int) bool { return m&1 != 0 })
	if If_CutPerformCheck07(nil, decTest32(irrelevant), 6, 6, "") != 0 || If_CutPerformCheck08(nil, decTest32(irrelevant), 6, 6, "") != 0 || If_CutPerformCheck10(nil, decTest32(irrelevant), 6, 6, "") != 0 {
		t.Fatal("accepted nonminimal full support")
	}
}
func TestDecSwapAndPermutation(t *testing.T) {
	rng := rand.New(rand.NewSource(491))
	for n := int32(6); n <= 10; n++ {
		truth := make([]uint64, decWords(n))
		for i := range truth {
			truth[i] = rng.Uint64()
		}
		for v := int32(0); v < n-1; v++ {
			out := make([]uint64, len(truth))
			If_Dec10SwapAdjacent(out, truth, v, n)
			for m := 0; m < 1<<uint(n); m++ {
				source := m
				if (m >> uint(v) & 1) != (m >> uint(v+1) & 1) {
					source ^= 3 << uint(v)
				}
				if decBit(out, m) != decBit(truth, source) {
					t.Fatal("variable swap", n, v, m)
				}
			}
			If_Dec10SwapAdjacent(out, out, v, n)
			if !reflect.DeepEqual(out, truth) {
				t.Fatal("aliased inverse swap")
			}
		}
	}
}
func TestDecConfigurationFourBitSelectors(t *testing.T) {
	truth := decTestTruth(10, func(m int) bool { return m>>9&1 != 0 })
	// One LUT projects input nine. The C debug verifier incorrectly masked it to one.
	config := uint32(0xaaaa) | uint32(9)<<16
	If_Dec10Verify(truth, 10, []uint32{config, 0})
}

func TestDecRecordedVectors(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("test_vectors", "decomposition_cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Variables int32    `json:"variables"`
		Truth     []string `json:"truth"`
		Cascade   bool     `json:"cascade"`
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		truth := make([]uint64, len(test.Truth))
		for i, hex := range test.Truth {
			truth[i], err = strconv.ParseUint(hex, 16, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		var config uint64
		switch test.Variables {
		case 5:
			config = If_Dec5Perform(truth[0], 1)
		case 6:
			config = If_Dec6Perform(truth[0], 1)
		case 7:
			config = If_Dec7Perform(truth, 1)
		}
		if (config != 0) != test.Cascade {
			t.Fatal(test)
		}
		if config != 0 && !reflect.DeepEqual(decConfigTruth(config, test.Variables), truth) {
			t.Fatal("recorded configuration reconstruction", test)
		}
	}
}
