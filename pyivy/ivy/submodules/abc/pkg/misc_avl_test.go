package abc

// Standalone structural, model, lifecycle, and fuzz tests for misc/avl.

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

func TestAVLRotationsAndOrder(t *testing.T) {
	orders := [][]int{
		{3, 2, 1},
		{1, 2, 3},
		{3, 1, 2},
		{1, 3, 2},
		{8, 4, 12, 2, 6, 10, 14, 1, 3, 5, 7, 9, 11, 13, 15},
	}
	for _, order := range orders {
		tree, err := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range order {
			duplicate, insertErr := tree.Insert(key, key*10)
			if insertErr != nil || duplicate {
				t.Fatal(order, key, duplicate, insertErr)
			}
			if checkErr := tree.Check(); checkErr != nil {
				t.Fatal(order, key, checkErr)
			}
		}
		forward, err := tree.Items(AVLForward)
		if err != nil {
			t.Fatal(err)
		}
		backward, err := tree.Items(AVLBackward)
		if err != nil {
			t.Fatal(err)
		}
		for index := range forward {
			if forward[index].Key != index+1 || backward[index].Key != len(order)-index {
				t.Fatal(order, forward, backward)
			}
		}
		first, found, err := tree.First()
		if err != nil || !found || first.Key != 1 || first.Value != 10 {
			t.Fatal(first, found, err)
		}
		last, found, err := tree.Last()
		if err != nil || !found || last.Key != len(order) || last.Value != len(order)*10 {
			t.Fatal(last, found, err)
		}
	}
}

func TestAVLDuplicatesDeleteAndSlots(t *testing.T) {
	tree, err := NewAVLTree[int, string](AVLIntCompare, AVLTreeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for index, value := range []string{"first", "second", "third"} {
		duplicate, insertErr := tree.Insert(7, value)
		if insertErr != nil || duplicate != (index != 0) {
			t.Fatal(index, duplicate, insertErr)
		}
	}
	items, err := tree.Items(AVLForward)
	if err != nil || len(items) != 3 {
		t.Fatal(items, err)
	}
	values := []string{items[0].Value, items[1].Value, items[2].Value}
	slices.Sort(values)
	if !slices.Equal(values, []string{"first", "second", "third"}) {
		t.Fatal(values)
	}
	for remaining := 3; remaining > 0; remaining-- {
		stored, value, found, deleteErr := tree.Delete(7)
		if deleteErr != nil || !found || stored != 7 || !slices.Contains(values, value) {
			t.Fatal(stored, value, found, deleteErr)
		}
		if checkErr := tree.Check(); checkErr != nil {
			t.Fatal(checkErr)
		}
		count, _ := tree.Len()
		if count != remaining-1 {
			t.Fatal(count)
		}
	}
	if _, _, found, err := tree.Delete(7); err != nil || found {
		t.Fatal(found, err)
	}

	slot, existing, err := tree.FindOrAdd(4)
	if err != nil || existing {
		t.Fatal(existing, err)
	}
	if value, err := slot.Value(); err != nil || value != "" {
		t.Fatal(value, err)
	}
	if err = slot.Set("four"); err != nil {
		t.Fatal(err)
	}
	again, existing, err := tree.FindOrAdd(4)
	if err != nil || !existing {
		t.Fatal(existing, err)
	}
	if value, err := again.Value(); err != nil || value != "four" {
		t.Fatal(value, err)
	}
	if key, err := again.Key(); err != nil || key != 4 {
		t.Fatal(key, err)
	}
	if _, _, found, err := tree.Delete(4); err != nil || !found {
		t.Fatal(found, err)
	}
	if err := slot.Set("stale"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestAVLGeneratorTraversalAndClose(t *testing.T) {
	tree, _ := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{})
	for _, key := range []int{2, 1, 3} {
		_, _ = tree.Insert(key, key*10)
	}
	generator, err := tree.NewGenerator(AVLForward)
	if err != nil {
		t.Fatal(err)
	}
	// The generator owns its snapshot: later insertions are not generated and
	// deleting the next source node cannot create a dangling reference.
	_, _ = tree.Insert(4, 40)
	_, _, _, _ = tree.Delete(2)
	var generated []int
	for {
		item, found, nextErr := generator.Next()
		if nextErr != nil {
			t.Fatal(nextErr)
		}
		if !found {
			break
		}
		generated = append(generated, item.Key)
	}
	if !slices.Equal(generated, []int{1, 2, 3}) {
		t.Fatal(generated)
	}
	if err = generator.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err = generator.Next(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	var reverse []int
	err = tree.ForEach(AVLBackward, func(key, value int) error {
		reverse = append(reverse, key)
		if key == 3 {
			_, _ = tree.Insert(5, 50)
		}
		return nil
	})
	if err != nil || !slices.Equal(reverse, []int{4, 3, 1}) {
		t.Fatal(reverse, err)
	}
	if err = tree.ForEach(AVLForward, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = tree.Items(AVLDirection(9)); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	closeTree, _ := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{})
	var slots []*AVLSlot[int, int]
	for _, key := range []int{2, 1, 3} {
		slot, _, _ := closeTree.FindOrAdd(key)
		_ = slot.Set(key * 10)
		slots = append(slots, slot)
	}
	var keys, values []int
	if err = closeTree.Close(func(key int) { keys = append(keys, key) }, func(value int) { values = append(values, value) }); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(keys, []int{1, 3, 2}) || !slices.Equal(values, []int{10, 30, 20}) {
		t.Fatal(keys, values)
	}
	if _, err = slots[0].Value(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = closeTree.Len(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = closeTree.Close(nil, nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestAVLLimitsCancellationAndCorruption(t *testing.T) {
	if _, err := NewAVLTree[int, int](nil, AVLTreeOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{MaxEntries: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	limited, _ := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{MaxEntries: 2})
	_, _ = limited.Insert(1, 1)
	_, _ = limited.Insert(2, 2)
	if _, err := limited.Insert(3, 3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := limited.FindOrAdd(4); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	stop := false
	cancelled, _ := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{Stop: func() bool { return stop }})
	_, _ = cancelled.Insert(1, 1)
	stop = true
	if _, _, err := cancelled.Lookup(1); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := cancelled.Items(AVLForward); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := cancelled.Insert(2, 2); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}

	corrupt, _ := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{})
	_, _ = corrupt.Insert(1, 1)
	corrupt.root.height = 10
	if err := corrupt.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if AVLIntCompare(-1, 1) != -1 || AVLIntCompare(1, -1) != 1 || AVLIntCompare(1, 1) != 0 {
		t.Fatal("integer comparator")
	}
}

func TestAVLStringComparator(t *testing.T) {
	compare := func(a, b string) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
	tree, err := NewAVLTree[string, int](compare, AVLTreeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for index, key := range []string{"gamma", "alpha", "beta"} {
		_, _ = tree.Insert(key, index)
	}
	items, err := tree.Items(AVLForward)
	if err != nil || fmt.Sprint(items) != "[{alpha 1} {beta 2} {gamma 0}]" {
		t.Fatal(items, err)
	}
}

func FuzzAVLTreeAgainstMultiset(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {1, 2, 3}, {255, 7, 9, 2, 8}, {4, 4, 4, 4}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 96 {
			data = data[:96]
		}
		tree, err := NewAVLTree[int, int](AVLIntCompare, AVLTreeOptions{MaxEntries: 32})
		if err != nil {
			t.Fatal(err)
		}
		model := make(map[int][]int)
		total := 0
		removeValue := func(key, value int) bool {
			for index, candidate := range model[key] {
				if candidate == value {
					model[key] = append(model[key][:index], model[key][index+1:]...)
					if len(model[key]) == 0 {
						delete(model, key)
					}
					return true
				}
			}
			return false
		}
		for step, raw := range data {
			key := int(int8(raw)) % 9
			value := step*257 + int(raw)
			switch raw % 5 {
			case 0:
				duplicate, insertErr := tree.Insert(key, value)
				if total == 32 {
					if !errors.Is(insertErr, ErrLimit) {
						t.Fatal(step, insertErr)
					}
					break
				}
				if insertErr != nil || duplicate != (len(model[key]) != 0) {
					t.Fatal(step, duplicate, insertErr, model[key])
				}
				model[key] = append(model[key], value)
				total++
			case 1:
				got, found, lookupErr := tree.Lookup(key)
				if lookupErr != nil || found != (len(model[key]) != 0) || found && !slices.Contains(model[key], got) {
					t.Fatal(step, key, got, found, lookupErr, model[key])
				}
			case 2:
				stored, got, found, deleteErr := tree.Delete(key)
				if deleteErr != nil || found != (len(model[key]) != 0) || found && (stored != key || !removeValue(key, got)) {
					t.Fatal(step, key, stored, got, found, deleteErr, model[key])
				}
				if found {
					total--
				}
			case 3:
				slot, existing, slotErr := tree.FindOrAdd(key)
				if total == 32 && len(model[key]) == 0 {
					if !errors.Is(slotErr, ErrLimit) {
						t.Fatal(step, slotErr)
					}
					break
				}
				if slotErr != nil || existing != (len(model[key]) != 0) {
					t.Fatal(step, existing, slotErr)
				}
				old, valueErr := slot.Value()
				if valueErr != nil {
					t.Fatal(valueErr)
				}
				if existing {
					if !removeValue(key, old) {
						t.Fatal(step, old, model[key])
					}
				} else {
					total++
				}
				if setErr := slot.Set(value); setErr != nil {
					t.Fatal(setErr)
				}
				model[key] = append(model[key], value)
			case 4:
				contains, containsErr := tree.Contains(key)
				if containsErr != nil || contains != (len(model[key]) != 0) {
					t.Fatal(step, contains, containsErr)
				}
			}
			if err = tree.Check(); err != nil {
				t.Fatal(step, err)
			}
			if count, countErr := tree.Len(); countErr != nil || count != total {
				t.Fatal(step, count, total, countErr)
			}
		}
		items, err := tree.Items(AVLForward)
		if err != nil || len(items) != total {
			t.Fatal(len(items), total, err)
		}
		counts := make(map[[2]int]int)
		for key, values := range model {
			for _, value := range values {
				counts[[2]int{key, value}]++
			}
		}
		for index, item := range items {
			if index > 0 && items[index-1].Key > item.Key {
				t.Fatal("unordered", items)
			}
			pair := [2]int{item.Key, item.Value}
			counts[pair]--
			if counts[pair] < 0 {
				t.Fatal("unexpected item", item)
			}
		}
		for pair, count := range counts {
			if count != 0 {
				t.Fatal(pair, count)
			}
		}
	})
}
