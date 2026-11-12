package avl

import (
	"math/rand"
	"slices"
	"testing"
)

func TestRandomDuplicateInsertDelete(t *testing.T) {
	tree := New(Avl_numcmp)
	counts := map[int]int{}
	r := rand.New(rand.NewSource(321))
	for step := 0; step < 5000; step++ {
		k := r.Intn(100)
		if r.Intn(3) != 0 {
			existed := tree.Insert(k, step)
			if existed != (counts[k] > 0) {
				t.Fatal("insert status", step)
			}
			counts[k]++
		} else {
			key, _, ok := tree.Delete(k)
			if ok != (counts[k] > 0) {
				t.Fatal("delete status", step)
			}
			if ok {
				if key != k {
					t.Fatal(key)
				}
				counts[k]--
			}
		}
		if tree.Check() != 0 {
			t.Fatalf("invalid AVL after operation %d", step)
		}
		total := 0
		for _, n := range counts {
			total += n
		}
		if tree.Count() != total {
			t.Fatal("count")
		}
	}
	var ordered []int
	tree.Foreach(func(k, v any) { ordered = append(ordered, k.(int)) }, AVL_FORWARD)
	if !slices.IsSorted(ordered) {
		t.Fatal("forward order")
	}
	var reverse []int
	g := tree.Generator(AVL_BACKWARD)
	for {
		key, _, ok := g.Next()
		if !ok {
			break
		}
		reverse = append(reverse, key.(int))
	}
	slices.Reverse(reverse)
	if !slices.Equal(ordered, reverse) {
		t.Fatal("backward order")
	}
	for key, n := range counts {
		for i := 0; i < n; i++ {
			if _, _, ok := tree.Delete(key); !ok {
				t.Fatal("missing key")
			}
			if tree.Check() != 0 {
				t.Fatal("delete invariant")
			}
		}
	}
	if tree.Count() != 0 {
		t.Fatal("not empty")
	}
}
func TestSlotsAndCompatibility(t *testing.T) {
	tree := Avl_init_table(Avl_numcmp)
	var slot *any
	if Avl_find_or_add(tree, 50, &slot) != 0 || *slot != nil {
		t.Fatal("initial slot")
	}
	*slot = "stored"
	for i := 0; i < 100; i++ {
		if i != 50 {
			Avl_insert(tree, i, i)
		}
	}
	var value any
	if Avl_lookup(tree, 50, &value) != 1 || value != "stored" {
		t.Fatal("slot lost during rotations")
	}
	if Avl_find_or_add(tree, 50, &slot) != 1 || *slot != "stored" {
		t.Fatal("existing slot")
	}
	key := any(49)
	if Avl_delete(tree, &key, nil) != 1 {
		t.Fatal("delete")
	}
	*slot = "updated"
	if v, _ := tree.Lookup(50); v != "updated" {
		t.Fatal("slot lost during deletion")
	}
	var first, last any
	Avl_first(tree, &first, nil)
	Avl_last(tree, &last, nil)
	if first != 0 || last != 99 {
		t.Fatal(first, last)
	}
	value = "unchanged"
	if Avl_lookup(tree, 101, &value) != 0 || value != "unchanged" {
		t.Fatal("failed lookup changed output")
	}
	keys, values := 0, 0
	Avl_free_table(tree, func(any) { keys++ }, func(any) { values++ })
	if keys != 99 || values != 99 || tree.Count() != 0 {
		t.Fatal(keys, values, tree.Count())
	}
	if _, _, ok := tree.First(); ok {
		t.Fatal("empty first")
	}
	if _, _, ok := tree.Last(); ok {
		t.Fatal("empty last")
	}
}
