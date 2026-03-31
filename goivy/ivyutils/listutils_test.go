package ivyutils

import (
	"testing"
)

func TestConcat(t *testing.T) {
	result := Concat([][]int{{1, 2}, {3, 4}, {5}})
	want := []int{1, 2, 3, 4, 5}
	if len(result) != len(want) {
		t.Fatalf("Concat len = %d, want %d", len(result), len(want))
	}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("Concat[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestConcatEmpty(t *testing.T) {
	result := Concat[int](nil)
	if len(result) != 0 {
		t.Errorf("Concat(nil) len = %d, want 0", len(result))
	}
}

func TestListDiff(t *testing.T) {
	result := ListDiff([]string{"a", "b", "c", "d"}, []string{"b", "d"})
	want := []string{"a", "c"}
	if len(result) != len(want) {
		t.Fatalf("ListDiff len = %d, want %d", len(result), len(want))
	}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("ListDiff[%d] = %q, want %q", i, v, want[i])
		}
	}
}

func TestListUnion(t *testing.T) {
	result := ListUnion([]int{1, 2, 3}, []int{2, 3, 4, 5})
	want := []int{1, 2, 3, 4, 5}
	if len(result) != len(want) {
		t.Fatalf("ListUnion len = %d, want %d", len(result), len(want))
	}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("ListUnion[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestUnionToList(t *testing.T) {
	list := []string{"a", "b"}
	UnionToList(&list, []string{"b", "c", "d"})
	want := []string{"a", "b", "c", "d"}
	if len(list) != len(want) {
		t.Fatalf("UnionToList len = %d, want %d", len(list), len(want))
	}
	for i, v := range list {
		if v != want[i] {
			t.Errorf("UnionToList[%d] = %q, want %q", i, v, want[i])
		}
	}
}

func TestInverseMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	inv := InverseMap(m)
	if inv[1] != "a" || inv[2] != "b" {
		t.Errorf("InverseMap = %v, unexpected", inv)
	}
}

func TestComposeMaps(t *testing.T) {
	m1 := map[string]string{"a": "b", "c": "d"}
	m2 := map[string]string{"b": "x", "d": "y"}
	result := ComposeMaps(m1, m2)
	// a -> b -> x, c -> d -> y, b -> x (from m2), d -> y (from m2)
	if result["a"] != "x" {
		t.Errorf("ComposeMaps['a'] = %q, want 'x'", result["a"])
	}
	if result["c"] != "y" {
		t.Errorf("ComposeMaps['c'] = %q, want 'y'", result["c"])
	}
}

func TestPartition(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6}
	result := Partition(items, func(x int) string {
		if x%2 == 0 {
			return "even"
		}
		return "odd"
	})
	if len(result["even"]) != 3 || len(result["odd"]) != 3 {
		t.Errorf("Partition: even=%v odd=%v", result["even"], result["odd"])
	}
}

func TestSplitList(t *testing.T) {
	p := map[string]bool{"a": true, "c": true}
	yes, no := SplitList([]string{"a", "b", "c", "d"}, p)
	if len(yes) != 2 || len(no) != 2 {
		t.Errorf("SplitList: yes=%v no=%v", yes, no)
	}
}

func TestDistinctUnorderedPairs(t *testing.T) {
	result := DistinctUnorderedPairs([]int{1, 2, 3})
	// Should be: (1,2), (1,3), (2,3)
	if len(result) != 3 {
		t.Fatalf("DistinctUnorderedPairs len = %d, want 3", len(result))
	}
}

func TestUnzipAppend(t *testing.T) {
	tups := [][]int{{1, 2}, {3, 4}, {5, 6}}
	result := UnzipAppend(tups)
	if len(result) != 2 {
		t.Fatalf("UnzipAppend width = %d, want 2", len(result))
	}
	// First column: [1, 3, 5], Second column: [2, 4, 6]
	if len(result[0]) != 3 || result[0][0] != 1 || result[0][1] != 3 || result[0][2] != 5 {
		t.Errorf("UnzipAppend[0] = %v, want [1, 3, 5]", result[0])
	}
}
