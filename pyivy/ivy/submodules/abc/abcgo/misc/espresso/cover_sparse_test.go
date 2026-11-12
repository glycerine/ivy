package espresso

import (
	_ "embed"
	"encoding/json"
	"slices"
	"testing"
)

//go:embed test_vectors/cover_sparse_oracle.json
var coverSparseOracle []byte

func TestCoverSparseOriginalC(t *testing.T) {
	savedCube, savedData := cube, cdata
	savedForce, savedRandom := force_irredundant, use_random_order
	savedTimes, savedCalls := total_time, total_calls
	defer func() {
		cube, cdata = savedCube, savedData
		force_irredundant, use_random_order = savedForce, savedRandom
		total_time, total_calls = savedTimes, savedCalls
	}()
	force_irredundant, use_random_order = false, false
	var cases []struct {
		On, DC, Off             []uint32
		Cost                    []int32
		Reduced, Sparse, Forced []uint32
	}
	if err := json.Unmarshal(coverSparseOracle, &cases); err != nil {
		t.Fatal(err)
	}
	family := func(masks []uint32) pcover {
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
	masks := func(f pcover) []uint32 {
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
	for trial, fixture := range cases {
		ConfigureCube(2, []int32{3, 2})
		cube.sparse[2], cube.sparse[3] = 1, 1
		f, d, r := family(fixture.On), family(fixture.DC), family(fixture.Off)
		forcedInput := sf_save(f)
		var cost cost_t
		cover_cost(f, &cost)
		gotCost := []int32{cost.cubes, cost.in, cost.out, cost.mv, cost.total, cost.primes}
		if !slices.Equal(gotCost, fixture.Cost) {
			t.Fatalf("trial %d cost %v want %v", trial, gotCost, fixture.Cost)
		}
		reduced := mv_reduce(sf_save(f), d)
		if got := masks(reduced); !slices.Equal(got, fixture.Reduced) {
			t.Fatalf("trial %d reduced %v want %v", trial, got, fixture.Reduced)
		}
		force_irredundant = false
		sparse := make_sparse(f, d, r)
		if got := masks(sparse); !slices.Equal(got, fixture.Sparse) {
			t.Fatalf("trial %d sparse %v want %v", trial, got, fixture.Sparse)
		}
		force_irredundant = true
		if got := masks(make_sparse(forcedInput, d, r)); !slices.Equal(got, fixture.Forced) {
			t.Fatalf("trial %d forced %v want %v", trial, got, fixture.Forced)
		}
	}
}
