package espresso

import (
	_ "embed"
	"encoding/json"
	"math/rand/v2"
	"slices"
	"testing"
)

//go:embed test_vectors/gasp_oracle.json
var gaspOracle []byte

func gaspFamily(masks []uint32) pcover {
	f := sf_new(int32(len(masks)), cube.size)
	for _, mask := range masks {
		p := set_clear(GETSET(f, f.count), cube.size)
		f.count++
		for bit := int32(0); bit < cube.size; bit++ {
			if mask&(1<<bit) != 0 {
				set_insert(p, bit)
			}
		}
	}
	return f
}
func gaspMasks(f pcover) []uint32 {
	result := make([]uint32, 0, f.count)
	for row := int32(0); row < f.count; row++ {
		var mask uint32
		for bit := int32(0); bit < cube.size; bit++ {
			if is_in_set(GETSET(f, row), bit) {
				mask |= 1 << bit
			}
		}
		result = append(result, mask)
	}
	return result
}
func TestGaspAndReduceOriginalC(t *testing.T) {
	savedCube, savedData := cube, cdata
	savedToggle, savedRandom := reduceToggle, use_random_order
	savedTimes, savedCalls := total_time, total_calls
	defer func() {
		cube, cdata = savedCube, savedData
		reduceToggle, use_random_order = savedToggle, savedRandom
		total_time, total_calls = savedTimes, savedCalls
	}()
	reduceToggle, use_random_order = true, false
	var fixtures []struct {
		On, DC, Off, Base, Reduced, Last, Super []uint32
		PrimeSuper                              []uint32 `json:"prime_super"`
	}
	if err := json.Unmarshal(gaspOracle, &fixtures); err != nil {
		t.Fatal(err)
	}
	for trial, fixture := range fixtures {
		ConfigureCube(2, []int32{3, 2})
		cube.sparse[2], cube.sparse[3] = 1, 1
		base, d, r := gaspFamily(fixture.Base), gaspFamily(fixture.DC), gaspFamily(fixture.Off)
		// Gasp assumes its input cubes are already prime and irredundant.
		for row := int32(0); row < base.count; row++ {
			SET(GETSET(base, row), PRIME)
		}
		var cost cost_t
		reduced := reduce(sf_save(base), d)
		if got := gaspMasks(reduced); !slices.Equal(got, fixture.Reduced) {
			t.Fatalf("trial %d reduced %v want %v", trial, got, fixture.Reduced)
		}
		last := last_gasp(sf_save(base), d, r, &cost)
		if got := gaspMasks(last); !slices.Equal(got, fixture.Last) {
			t.Fatalf("trial %d last %v want %v", trial, got, fixture.Last)
		}
		super := super_gasp(sf_save(base), d, r, &cost)
		gotSuper := gaspMasks(super)
		if !slices.Equal(gotSuper, fixture.PrimeSuper) {
			t.Fatalf("trial %d corrected super %v want %v", trial, gotSuper, fixture.PrimeSuper)
		}
		// Raw C may return nonprimes because its all-primes base case drops
		// FREESET. Both reference versions must still preserve the function.
		for _, result := range [][]uint32{gaspMasks(last), gotSuper} {
			gaspCheckCover(t, result, fixture.On, fixture.DC, fixture.Off)
		}

	}
}

func TestSCCCPointEnumeration(t *testing.T) {
	savedCube, savedData := cube, cdata
	defer func() { cube, cdata = savedCube, savedData }()
	ConfigureCube(2, []int32{3, 2})
	points := []uint32{}
	for a := 0; a < 2; a++ {
		for b := 2; b < 4; b++ {
			for c := 4; c < 7; c++ {
				for d := 7; d < 9; d++ {
					points = append(points, 1<<a|1<<b|1<<c|1<<d)
				}
			}
		}
	}
	random := rand.New(rand.NewPCG(177, 29))
	for trial := 0; trial < 500; trial++ {
		masks := make([]uint32, random.IntN(12))
		for row := range masks {
			for _, parts := range [][2]int{{0, 2}, {2, 4}, {4, 7}, {7, 9}} {
				bits := 1 + random.IntN((1<<(parts[1]-parts[0]))-1)
				masks[row] |= uint32(bits) << parts[0]
			}
		}
		expected := uint32(0)
		for _, point := range points {
			covered := false
			for _, mask := range masks {
				covered = covered || point&mask == point
			}
			if !covered {
				expected |= point
			}
		}
		f := gaspFamily(masks)
		got := sccc(cube1list(f))
		gotMasks := gaspMasks(sf_addset(sf_new(1, cube.size), got))
		if gotMasks[0] != expected {
			t.Fatalf("trial %d sccc %d want %d", trial, gotMasks[0], expected)
		}
		if len(masks) > 0 {
			essential := uint32(0)
			for _, point := range points {
				if point&masks[0] == point && !gaspCovered(point, masks[1:]) {
					essential |= point
				}
			}
			under := reduce_cube(cube1list(f), GETSET(f, 0))
			reducedMask := gaspMasks(sf_addset(sf_new(1, cube.size), under))[0]
			if reducedMask != essential {
				t.Fatalf("trial %d maximal reduction %d want %d", trial, reducedMask, essential)
			}
		}

	}
}

func gaspPoints() []uint32 {
	var points []uint32
	for a := 0; a < 2; a++ {
		for b := 2; b < 4; b++ {
			for c := 4; c < 7; c++ {
				for d := 7; d < 9; d++ {
					points = append(points, 1<<a|1<<b|1<<c|1<<d)
				}
			}
		}
	}
	return points
}

func gaspCovered(point uint32, masks []uint32) bool {
	for _, mask := range masks {
		if point&mask == point {
			return true
		}
	}
	return false
}
func gaspCheckCover(t *testing.T, result, on, dc, off []uint32) {
	t.Helper()
	for _, point := range gaspPoints() {
		if gaspCovered(point, off) && gaspCovered(point, result) {
			t.Fatal("gasp covers off-set", point, result)
		}
		if gaspCovered(point, on) && !gaspCovered(point, dc) && !gaspCovered(point, result) {
			t.Fatal("gasp misses on-set", point, result)
		}
	}
	for row, mask := range result {
		for bit := uint(0); bit < 9; bit++ {
			if mask&(1<<bit) != 0 {
				continue
			}
			raised := mask | 1<<bit
			blocked := false
			for _, point := range off {
				blocked = blocked || point&raised == point
			}
			if !blocked {
				t.Fatal("gasp returned nonprime", mask, bit)
			}
		}
		essential := false
		for _, point := range gaspPoints() {
			if point&mask != point || gaspCovered(point, dc) {
				continue
			}
			other := false
			for index, candidate := range result {
				other = other || index != row && point&candidate == point
			}
			essential = essential || !other
		}
		if !essential {
			t.Fatal("gasp returned redundant cube", row, result)
		}
	}
}
