package fx

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestExtractSharedTerm(t *testing.T) {
	cubes := [][]int32{{3, 0, 2}, {4, 0, 2}, {5, 0, 2}}
	result, count, err := Extract(cubes, 6, Options{MaxNewNodes: 5})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int32{{3, 12}, {4, 12}, {5, 12}, {6, 0, 2}}
	if count != 1 || !reflect.DeepEqual(result, want) {
		t.Fatalf("result %v count %d", result, count)
	}
	for inputs := 0; inputs < 4; inputs++ {
		values := []bool{inputs&1 != 0, inputs&2 != 0, false, false, false, false, false}
		// The appended divisor is evaluated before the rewritten owners.
		values[6] = values[0] && values[1]
		for _, cube := range result[:3] {
			values[cube[0]] = values[cube[1]>>1]
			if values[cube[0]] != (values[0] && values[1]) {
				t.Fatalf("input %d cube %v", inputs, cube)
			}
		}
	}
}

func TestExtractNoChangesAndLimits(t *testing.T) {
	cubes := [][]int32{{2, 0, 1}}
	result, count, err := Extract(cubes, 3, Options{})
	if err != nil || count != 0 || !reflect.DeepEqual(result, cubes) {
		t.Fatalf("%v %d %v", result, count, err)
	}
	result[0][0] = 8
	if cubes[0][0] != 2 {
		t.Fatal("result aliases input")
	}
	if _, _, err = Extract(cubes, 3, Options{MaxNewNodes: -1}); err == nil {
		t.Fatal("negative limit accepted")
	}
}

func TestExtractPreservesRandomCovers(t *testing.T) {
	rng := rand.New(rand.NewSource(18))
	checked := 0
	for trial := 0; trial < 1000; trial++ {
		var cubes [][]int32
		for owner := int32(4); owner < 8; owner++ {
			seen := make(map[[4]int8]bool)
			for j := 0; j < 3; j++ {
				var pattern [4]int8
				cube := []int32{owner}
				for v := int32(0); v < 4; v++ {
					pattern[v] = int8(rng.Intn(3))
					if pattern[v] != 0 {
						cube = append(cube, 2*v+int32(pattern[v]-1))
					}
				}
				if seen[pattern] {
					continue
				}
				seen[pattern] = true
				cubes = append(cubes, cube)
			}
		}
		result, _, err := Extract(cubes, 8, Options{MaxNewNodes: 3})
		if err != nil {
			t.Fatalf("trial %d: %v; source=%v", trial, err, cubes)
		}
		checked++
		for input := 0; input < 16; input++ {
			before, err := evalFxOwners(cubes, input)
			if err != nil {
				t.Fatal(err)
			}
			after, err := evalFxOwners(result, input)
			if err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
			for owner := int32(4); owner < 8; owner++ {
				if before[owner] != after[owner] {
					t.Fatalf("trial=%d input=%d owner=%d before=%v after=%v source=%v result=%v", trial, input, owner, before[owner], after[owner], cubes, result)
				}
			}
		}
	}
	t.Logf("checked %d cases", checked)
	if checked < 1000 {
		t.Fatalf("only %d cases checked", checked)
	}
}

func evalFxOwners(cubes [][]int32, inputs int) (map[int32]bool, error) {
	covers := make(map[int32][][]int32)
	for _, cube := range cubes {
		covers[cube[0]] = append(covers[cube[0]], cube[1:])
	}
	values := make(map[int32]bool)
	for i := int32(0); i < 4; i++ {
		values[i] = inputs&(1<<i) != 0
	}
	active := make(map[int32]bool)
	var eval func(int32) (bool, error)
	eval = func(id int32) (bool, error) {
		if v, ok := values[id]; ok {
			return v, nil
		}
		if active[id] {
			return false, fmt.Errorf("cycle at node %d", id)
		}
		active[id] = true
		var result bool
		for _, cube := range covers[id] {
			product := true
			for _, lit := range cube {
				v, err := eval(lit >> 1)
				if err != nil {
					return false, err
				}
				product = product && (v != (lit&1 != 0))
			}
			result = result || product
		}
		delete(active, id)
		values[id] = result
		return result, nil
	}
	for id := int32(4); id < 8; id++ {
		if _, err := eval(id); err != nil {
			return nil, err
		}
	}
	return values, nil
}
