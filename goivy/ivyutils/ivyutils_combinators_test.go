package ivyutils

import "testing"

func TestUnique(t *testing.T) {
	result := Unique([]int{1, 2, 3, 2, 1, 4})
	want := []int{1, 2, 3, 4}
	if len(result) != len(want) {
		t.Fatalf("Unique len = %d, want %d", len(result), len(want))
	}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("Unique[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestGenList(t *testing.T) {
	// Double each element
	fn := func(x int) []int { return []int{x, x} }
	result := GenList(fn, []int{1, 2, 3})
	want := []int{1, 1, 2, 2, 3, 3}
	if len(result) != len(want) {
		t.Fatalf("GenList len = %d, want %d", len(result), len(want))
	}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("GenList[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestGenListList(t *testing.T) {
	fn := func(x int) []string {
		if x > 2 {
			return []string{"big"}
		}
		return []string{"small"}
	}
	result := GenListList(fn, [][]int{{1, 2}, {3, 4}})
	if len(result) != 4 {
		t.Fatalf("GenListList len = %d, want 4", len(result))
	}
}

func TestAnyIn(t *testing.T) {
	aSet := map[int]struct{}{3: {}, 7: {}}
	if !AnyIn(aSet, []int{1, 2, 3}) {
		t.Error("AnyIn should be true when set intersects")
	}
	if AnyIn(aSet, []int{1, 2, 4}) {
		t.Error("AnyIn should be false when no intersection")
	}
}

func TestGenToSet(t *testing.T) {
	result := GenToSet([]string{"a", "b", "a", "c"})
	if len(result) != 3 {
		t.Errorf("GenToSet size = %d, want 3", len(result))
	}
}

func TestFilter2(t *testing.T) {
	pairs := []Pair[string, int]{
		{"a", 1}, {"b", 2}, {"c", 1}, {"d", 3},
	}
	result := Filter2(pairs, 1)
	if len(result) != 2 {
		t.Errorf("Filter2 size = %d, want 2", len(result))
	}
	if _, ok := result["a"]; !ok {
		t.Error("Filter2 should contain 'a'")
	}
	if _, ok := result["c"]; !ok {
		t.Error("Filter2 should contain 'c'")
	}
}
