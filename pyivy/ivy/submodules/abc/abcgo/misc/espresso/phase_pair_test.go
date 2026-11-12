package espresso

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func TestOutputPhaseSelectionIndependent(t *testing.T) {
	rng := rand.New(rand.NewSource(97))
	for trial := 0; trial < 100; trial++ {
		ConfigureCube(3, []int32{3})
		pla := &PLA_t{F: new_cover(8), R: new_cover(8), D: new_cover(0)}
		truth := make([]uint32, 8)
		for value := int32(0); value < 8; value++ {
			truth[value] = uint32(rng.Intn(8))
			f, r := binaryMinterm(value, 3), binaryMinterm(value, 3)
			for output := int32(0); output < 3; output++ {
				if truth[value]&(1<<output) != 0 {
					set_insert(f, 6+output)
				} else {
					set_insert(r, 6+output)
				}
			}
			pla.F = sf_addset(pla.F, f)
			pla.R = sf_addset(pla.R, r)
		}
		pla.phase = set_save(cube.fullset)
		flip := uint32(rng.Intn(8))
		for output := int32(0); output < 3; output++ {
			if flip&(1<<output) != 0 {
				set_remove(pla.phase, 6+output)
			}
		}
		set_phase(pla)
		for value := int32(0); value < 8; value++ {
			for output := int32(0); output < 3; output++ {
				want := (truth[value]^flip)&(1<<output) != 0
				for name, item := range map[string]struct {
					family pcover
					want   bool
				}{"on": {pla.F, want}, "off": {pla.R, !want}} {
					got := false
					for row := int32(0); row < item.family.count; row++ {
						p := GETSET(item.family, row)
						if matchesBinary(p, value, 3) && is_in_set(p, 6+output) {
							got = true
						}
					}
					if got != item.want {
						t.Fatalf("trial%d %s value%d out%d got%t want%t", trial, name, value, output, got, item.want)
					}
				}
			}
		}
	}
}
func TestOutputPhaseSetupAcrossWordBoundary(t *testing.T) {
	ConfigureCube(14, []int32{4})
	f, r, d := binaryMinterm(123, 14), binaryMinterm(456, 14), binaryMinterm(789, 14)
	for _, bit := range []int32{28, 31} {
		set_insert(f, bit)
	}
	set_insert(r, 30)
	set_insert(d, 29)
	pla := &PLA_t{F: sf_addset(new_cover(1), f), R: sf_addset(new_cover(1), r), D: sf_addset(new_cover(1), d)}
	output_phase_setup(pla, 0)
	if cube.size != 36 || pla.F.count != 2 || pla.R.count != 2 || pla.D.count != 1 {
		t.Fatalf("bad doubled geometry %d %d/%d/%d", cube.size, pla.F.count, pla.R.count, pla.D.count)
	}
	if !is_in_set(GETSET(pla.F, 0), 31) || !is_in_set(GETSET(pla.R, 0), 35) || !is_in_set(GETSET(pla.F, 1), 34) || !is_in_set(GETSET(pla.D, 0), 33) {
		t.Fatal("doubled phase bits lost")
	}
	if !matchesBinary(GETSET(pla.F, 0), 123, 14) {
		t.Fatal("binary fields changed")
	}
}
func TestPairingPreservesPreexistingMVParts(t *testing.T) {
	ConfigureCube(3, []int32{3, 2})
	old := new_cube()
	for _, bit := range []int32{0, 3, 4, 6, 8, 10} {
		set_insert(old, bit)
	}
	pla := &PLA_t{F: sf_addset(new_cover(1), old), D: new_cover(0), R: new_cover(0), pair: &pair_t{cnt: 1, var1: []int32{1}, var2: []int32{3}}}
	set_pair(pla)
	if cube.num_binary_vars != 1 || fmt.Sprint(cube.part_size) != "[2 4 3 2]" {
		t.Fatal(cube.part_size)
	}
	row := GETSET(pla.F, 0)
	want := maskFamily(cube.size, (1<<1)|(1<<5)|(1<<6)|(1<<8)|(1<<10))
	if !setp_equal(row, GETSET(want, 0)) {
		t.Fatal(familyMasks(pla.F), familyMasks(want))
	}
	if cube.sparse[1] != 0 {
		t.Fatal("paired variable must remain dense")
	}
}
func TestPairEnumerationAndNegativeCosts(t *testing.T) {
	for n, want := range map[int32]int{0: 1, 1: 1, 2: 1, 3: 3, 4: 3, 5: 15, 6: 15, 7: 105, 8: 105} {
		count := 0
		generate_all_pairs(pair_new(n), n, set_full(n), func(p ppair) {
			count++
			if p.cnt != n/2 {
				t.Fatal("incomplete matching")
			}
			used := map[int32]bool{}
			for i := int32(0); i < p.cnt; i++ {
				for _, v := range []int32{p.var1[i], p.var2[i]} {
					if used[v] {
						t.Fatal("duplicate matching endpoint")
					}
					used[v] = true
				}
			}
		})
		if count != want {
			t.Fatalf("n%d matchings%d want%d", n, count, want)
		}
	}
	ConfigureCube(4, []int32{1})
	costs := [][]int32{{0, -8, -2, -4}, {0, 0, -9, -3}, {0, 0, 0, -10}, {0, 0, 0, 0}}
	p := pair_best_cost(costs)
	if pairingBestCost != -5 || p == nil {
		t.Fatal("negative matching optimization", pairingBestCost)
	}
	var greedy ppair
	if cost := greedy_best_cost(costs, &greedy); cost != -5 {
		t.Fatal(cost)
	}
}
func TestMintermsMixedRadixAndGrayLayout(t *testing.T) {
	ConfigureCube(2, []int32{3, 2})
	f := maskFamily(cube.size, (1<<0)|(1<<3)|(1<<4)|(1<<6)|(1<<8))
	m := minterms(f)
	if set_ord(m) != 2 || !is_in_set(m, 14) || !is_in_set(m, 22) {
		t.Fatal("mixed radix minterms", m)
	}
	seen := map[int32]bool{}
	for r := int32(0); r < 16; r++ {
		for c := int32(0); c < 16; c++ {
			index := grayMapIndex(r, c)
			if seen[index] {
				t.Fatal("duplicate Gray map cell")
			}
			seen[index] = true
		}
	}
	if grayMapIndex(8, 0) != 160 || grayMapIndex(0, 8) != 80 || grayMapIndex(15, 15) != 202 {
		t.Fatal("Gray map differs from source")
	}
}

func TestPhasePairOriginalCFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/phase_pair_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Operation   int32
		On, DC, Off []uint32
		MappedOn    []uint32 `json:"mapped_on"`
		MappedDC    []uint32 `json:"mapped_dc"`
		MappedOff   []uint32 `json:"mapped_off"`
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for i, c := range cases {
		ConfigureCube(4, []int32{2})
		pla := &PLA_t{F: maskFamily(cube.size, c.On...), D: maskFamily(cube.size, c.DC...), R: maskFamily(cube.size, c.Off...)}
		switch c.Operation {
		case 0:
			pla.pair = &pair_t{cnt: 1, var1: []int32{1}, var2: []int32{3}}
			set_pair(pla)
		case 1:
			pla.phase = set_save(cube.fullset)
			set_remove(pla.phase, cube.last_part[cube.output])
			set_phase(pla)
		case 2:
			output_phase_setup(pla, 1)
		}
		for name, pair := range map[string]struct {
			family pcover
			want   []uint32
		}{"on": {pla.F, c.MappedOn}, "dc": {pla.D, c.MappedDC}, "off": {pla.R, c.MappedOff}} {
			if pair.family.count != int32(len(pair.want)) {
				t.Fatalf("case%d %s count %d want%d", i, name, pair.family.count, len(pair.want))
			}
			for row, want := range pair.want {
				p := GETSET(pair.family, int32(row))
				var mask uint32
				for bit := int32(0); bit < cube.size; bit++ {
					if is_in_set(p, bit) {
						mask |= 1 << bit
					}
				}
				if mask != want {
					t.Fatalf("case%d %s row%d mask%d want%d", i, name, row, mask, want)
				}
			}
		}
	}
}
