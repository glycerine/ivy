package compiler

import (
	"slices"
	"testing"
)

// Helper to extract keys in their exact iteration order
func collectKeys[K comparable, V any](m *insMap[K, V]) []K {
	var keys []K
	for k, _ := range m.all() {
		keys = append(keys, k)
	}
	return keys
}

// Helper to extract values in their exact iteration order
func collectValues[K comparable, V any](m *insMap[K, V]) []V {
	var vals []V
	for _, v := range m.all() {
		vals = append(vals, v)
	}
	return vals
}

func TestInsMap_BasicInsertionOrder(t *testing.T) {
	m := newInsMap[string, int]()

	m.set("apple", 1)
	m.set("banana", 2)
	m.set("cherry", 3)

	if m.Len() != 3 {
		t.Fatalf("expected length 3, got %d", m.Len())
	}

	keys := collectKeys(m)
	expectedKeys := []string{"apple", "banana", "cherry"}
	if !slices.Equal(keys, expectedKeys) {
		t.Fatalf("expected order %v, got %v", expectedKeys, keys)
	}
}

func TestInsMap_UpdatePreservesOrder(t *testing.T) {
	m := newInsMap[string, int]()

	m.set("A", 1)
	m.set("B", 2)
	m.set("C", 3)

	// Updating "B" should NOT move it to the end.
	newlyAdded := m.set("B", 99)
	if newlyAdded {
		t.Fatal("expected newlyAdded to be false for an existing key")
	}

	keys := collectKeys(m)
	expectedKeys := []string{"A", "B", "C"}
	if !slices.Equal(keys, expectedKeys) {
		t.Fatalf("expected keys order to remain %v, got %v", expectedKeys, keys)
	}

	vals := collectValues(m)
	expectedVals := []int{1, 99, 3}
	if !slices.Equal(vals, expectedVals) {
		t.Fatalf("expected values order %v, got %v", expectedVals, vals)
	}
}

func TestInsMap_DeleteAndReinsertMovesToEnd(t *testing.T) {
	m := newInsMap[string, int]()

	m.set("X", 10)
	m.set("Y", 20)
	m.set("Z", 30)

	// Delete the middle element
	found, _ := m.delkey("Y")
	if !found {
		t.Fatal("expected to find 'Y' for deletion")
	}

	// Verify it's gone
	keysAfterDel := collectKeys(m)
	if !slices.Equal(keysAfterDel, []string{"X", "Z"}) {
		t.Fatalf("expected order after delete [X, Z], got %v", keysAfterDel)
	}

	// Re-insert the same key. It should now appear at the very end.
	m.set("Y", 99)

	keysAfterReinsert := collectKeys(m)
	expectedFinal := []string{"X", "Z", "Y"}
	if !slices.Equal(keysAfterReinsert, expectedFinal) {
		t.Fatalf("expected order after re-insertion %v, got %v", expectedFinal, keysAfterReinsert)
	}
}

func TestInsMap_DeleteAll(t *testing.T) {
	m := newInsMap[int, string]()

	m.set(1, "one")
	m.set(2, "two")

	m.deleteAll()

	if m.Len() != 0 {
		t.Fatalf("expected length 0 after deleteAll, got %d", m.Len())
	}

	keys := collectKeys(m)
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys yielded, got %d", len(keys))
	}
}

func TestInsMap_Get(t *testing.T) {
	m := newInsMap[string, string]()
	m.set("key1", "val1")

	v, found := m.get2("key1")
	if !found || v != "val1" {
		t.Fatalf("expected (val1, true), got (%v, %v)", v, found)
	}

	v2, found2 := m.get2("missing")
	if found2 {
		t.Fatalf("expected false for missing key, got true (val: %v)", v2)
	}
}

func TestInsMap_IterationCacheInvalidation(t *testing.T) {
	m := newInsMap[string, int]()
	m.set("A", 1)
	m.set("B", 2)
	m.set("C", 3)
	m.set("D", 4)

	// We will iterate, but when we hit "B", we will delete "C".
	// This forces the iterator to abandon `ordercache` and fall back
	// to the safe tree-walking path mid-iteration.
	var seen []string
	for k, _ := range m.all() {
		seen = append(seen, k)
		if k == "B" {
			m.delkey("C")
		}
	}

	expectedSeen := []string{"A", "B", "D"}
	if !slices.Equal(seen, expectedSeen) {
		t.Fatalf("expected to see %v, got %v", expectedSeen, seen)
	}
}

func TestInsMap_StructKeys(t *testing.T) {
	// Proving the generics `comparable` constraint works for structs
	type Point struct{ X, Y int }

	m := newInsMap[Point, string]()

	p1 := Point{0, 0}
	p2 := Point{10, 10}

	m.set(p1, "origin")
	m.set(p2, "far")

	if val := m.get(p2); val != "far" {
		t.Fatalf("expected 'far', got '%v'", val)
	}
}
