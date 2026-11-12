package espresso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
)

func familyMasks(f pset_family) []uint32 {
	r := make([]uint32, f.count)
	for i := int32(0); i < f.count; i++ {
		for j := int32(0); j < f.sf_size; j++ {
			if is_in_set(GETSET(f, i), j) {
				r[i] |= 1 << j
			}
		}
	}
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return r
}
func maskFamily(width int32, masks ...uint32) pset_family {
	f := sf_new(int32(len(masks)), width)
	for _, mask := range masks {
		p := set_new(width)
		for i := int32(0); i < width; i++ {
			if mask&(1<<i) != 0 {
				set_insert(p, i)
			}
		}
		f = sf_addset(f, p)
	}
	return f
}
func TestExactCoverIndependentEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(41))
	for trial := 0; trial < 600; trial++ {
		width := int32(1 + rng.Intn(6))
		rows := make([]uint32, 1+rng.Intn(8))
		for i := range rows {
			rows[i] = uint32(rng.Intn(1 << width))
		}
		f := maskFamily(width, rows...)
		var want []uint32
		for candidate := uint32(0); candidate < 1<<width; candidate++ {
			hits := func(c uint32) bool {
				for _, row := range rows {
					if c&row == 0 {
						return false
					}
				}
				return true
			}
			if !hits(candidate) {
				continue
			}
			minimal := true
			for bit := int32(0); bit < width; bit++ {
				if candidate&(1<<bit) != 0 && hits(candidate&^(1<<bit)) {
					minimal = false
					break
				}
			}
			if minimal {
				want = append(want, candidate)
			}
		}
		for name, result := range map[string]pset_family{"leveled": exact_minimum_cover(f), "recursive": unate_compl(sf_save(f))} {
			got := familyMasks(result)
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Fatalf("%s width=%d rows=%v got=%v want=%v", name, width, rows, got, want)
			}
		}
	}
}
func binaryMinterm(value, n int32) pcube {
	p := new_cube()
	for v := int32(0); v < n; v++ {
		set_insert(p, 2*v+((value>>v)&1))
	}
	return p
}
func matchesBinary(p pcube, value, n int32) bool {
	for v := int32(0); v < n; v++ {
		if !is_in_set(p, 2*v+((value>>v)&1)) {
			return false
		}
	}
	return true
}
func TestExpandIndependentTruthTables(t *testing.T) {
	rng := rand.New(rand.NewSource(47))
	for trial := 0; trial < 450; trial++ {
		n := int32(1 + rng.Intn(5))
		ConfigureCube(n, nil)
		on, off := new_cover(1<<n), new_cover(1<<n)
		truth := make([]bool, 1<<n)
		for value := int32(0); value < 1<<n; value++ {
			truth[value] = rng.Intn(2) == 1
			if truth[value] {
				on = sf_addset(on, binaryMinterm(value, n))
			} else {
				off = sf_addset(off, binaryMinterm(value, n))
			}
		}
		result := expand(on, off, false)
		for value, want := range truth {
			got := false
			for i := int32(0); i < result.count; i++ {
				got = got || matchesBinary(GETSET(result, i), int32(value), n)
			}
			if got != want {
				t.Fatalf("trial %d value %d got %t want %t", trial, value, got, want)
			}
		}
		for i := int32(0); i < result.count; i++ {
			p := GETSET(result, i)
			if !TESTP(p, PRIME) {
				t.Fatal("expanded row not prime-marked")
			}
			for bit := int32(0); bit < cube.size; bit++ {
				if is_in_set(p, bit) {
					continue
				}
				q := set_save(p)
				set_insert(q, bit)
				blocked := false
				for value, on := range truth {
					if !on && matchesBinary(q, int32(value), n) {
						blocked = true
						break
					}
				}
				if !blocked {
					t.Fatalf("trial %d nonprime result %s", trial, pc1(p))
				}
			}
		}
	}
}
func TestUnravelMixedValuedAndFormatting(t *testing.T) {
	ConfigureCube(1, []int32{3, 2})
	p := new_cube()
	for _, i := range []int32{0, 1, 2, 4, 5, 6} {
		set_insert(p, i)
	}
	f := sf_addset(new_cover(1), p)
	r := unravel_range(f, 1, 2)
	if r.count != 4 {
		t.Fatalf("unravel count=%d", r.count)
	}
	want := []string{"- 100 10", "- 100 01", "- 001 10", "- 001 01"}
	for i, s := range want {
		if got := pc1(GETSET(r, int32(i))); got != s {
			t.Fatalf("row %d=%q want %q", i, got, s)
		}
	}
	var out bytes.Buffer
	print_cube(&out, p, "~1")
	if out.String() != "- 101 11\n" {
		t.Fatal(out.String())
	}
	pla := new_PLA()
	pla.F = f
	pla.D = new_cover(0)
	pla.R = new_cover(0)
	makeup_labels(pla)
	out.Reset()
	fprint_pla(&out, pla, F_type)
	if !strings.Contains(out.String(), ".mv 3 1 3 2\n") || !strings.HasSuffix(out.String(), "- 101 11\n.e\n") {
		t.Fatal(out.String())
	}
}
func TestContainIndexedRowsAndDistanceMerge(t *testing.T) {
	ConfigureCube(2, nil)
	f := maskFamily(4, 1, 3, 3, 12)
	indices := []int32{10, 20, 30, 40}
	r := sf_ind_contain(f, indices)
	if fmt.Sprint(familyMasks(r)) != "[3 12]" {
		t.Fatal(familyMasks(r))
	}
	for i := int32(0); i < r.count; i++ {
		p := GETSET(r, i)
		if is_in_set(p, 0) {
			if indices[i] != 20 && indices[i] != 30 {
				t.Fatal(indices)
			}
		} else if indices[i] != 40 {
			t.Fatal(indices)
		}
	}
	f = maskFamily(4, 5, 6)
	r = d1merge(f, 0)
	if r.count != 1 || familyMasks(r)[0] != 7 {
		t.Fatal(familyMasks(r))
	}
}
func TestCubeSetupAndSavedContext(t *testing.T) {
	ConfigureCube(2, []int32{3, 2})
	if cube.size != 9 || cube.output != 3 || cube.inword != 1 {
		t.Fatalf("bad layout %+v", cube)
	}
	saved := set_save(cube.fullset)
	save_cube_struct()
	ConfigureCube(1, nil)
	restore_cube_struct()
	if cube.size != 9 || !setp_equal(cube.fullset, saved) {
		t.Fatal("context restore changed cube")
	}
}

func TestExpandOriginalCFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/expand_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ On, DC, Off, Expanded []uint32 }
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for i, c := range cases {
		ConfigureCube(2, []int32{3, 2})
		on := maskFamily(cube.size, c.On...)
		off := maskFamily(cube.size, c.Off...)
		got := expand(on, off, false)
		if got.count != int32(len(c.Expanded)) {
			t.Fatalf("case %d count got %d want %d", i, got.count, len(c.Expanded))
		}
		for j, want := range c.Expanded {
			p := GETSET(got, int32(j))
			var mask uint32
			for bit := int32(0); bit < cube.size; bit++ {
				if is_in_set(p, bit) {
					mask |= 1 << bit
				}
			}
			if mask != want {
				t.Fatalf("case %d row %d got %d want %d", i, j, mask, want)
			}
		}
	}
}

func TestAllPrimesIndependentEnumeration(t *testing.T) {
	for truth := uint32(0); truth < 256; truth++ {
		ConfigureCube(3, nil)
		on, off := new_cover(8), new_cover(8)
		for value := int32(0); value < 8; value++ {
			if truth&(1<<value) != 0 {
				on = sf_addset(on, binaryMinterm(value, 3))
			} else {
				off = sf_addset(off, binaryMinterm(value, 3))
			}
		}
		got := familyMasks(sf_dupl(all_primes(on, off)))
		consensusPrimes := familyMasks(sf_dupl(primes_consensus(cube1list(on))))
		var want []uint32
		for pattern := 0; pattern < 27; pattern++ {
			p := new_cube()
			n := pattern
			for v := int32(0); v < 3; v++ {
				digit := n % 3
				n /= 3
				if digit != 1 {
					set_insert(p, 2*v)
				}
				if digit != 0 {
					set_insert(p, 2*v+1)
				}
			}
			valid := func(q pcube) bool {
				for value := int32(0); value < 8; value++ {
					if matchesBinary(q, value, 3) && truth&(1<<value) == 0 {
						return false
					}
				}
				return true
			}
			if !valid(p) {
				continue
			}
			prime := true
			for bit := int32(0); bit < 6; bit++ {
				if !is_in_set(p, bit) {
					q := set_save(p)
					set_insert(q, bit)
					if valid(q) {
						prime = false
						break
					}
				}
			}
			if prime {
				mask := uint32(0)
				for bit := int32(0); bit < 6; bit++ {
					if is_in_set(p, bit) {
						mask |= 1 << bit
					}
				}
				want = append(want, mask)
			}
		}
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("truth=%02x got %v want %v", truth, got, want)
		}
		if fmt.Sprint(consensusPrimes) != fmt.Sprint(want) {
			t.Fatalf("consensus truth=%02x got %v want %v", truth, consensusPrimes, want)
		}
	}
}

func TestExpandNonsparsePreservesSparseParts(t *testing.T) {
	ConfigureCube(1, []int32{2})
	on := maskFamily(4, 5)
	off := new_cover(0)
	sparse := expand(sf_save(on), off, true)
	if fmt.Sprint(familyMasks(sparse)) != "[7]" {
		t.Fatal(familyMasks(sparse))
	}
	full := expand(on, off, false)
	if fmt.Sprint(familyMasks(full)) != "[15]" {
		t.Fatal(familyMasks(full))
	}
}
