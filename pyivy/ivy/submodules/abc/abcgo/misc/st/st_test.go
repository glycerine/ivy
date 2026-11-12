package st

import "testing"

func TestTableGrowthReplacementAndSlots(t *testing.T) {
	table := st__init_table_with_params(st__numcmp, st__numhash, 1, 1, 2, 1)
	for i := 0; i < 200; i++ {
		if table.Insert(i, i*i) {
			t.Fatal("new key reported old")
		}
	}
	if table.Count() != 200 || len(table.bins) == 1 {
		t.Fatal("growth/count")
	}
	for i := 0; i < 200; i++ {
		v, ok := table.Lookup(i)
		if !ok || v != i*i {
			t.Fatal(i, v)
		}
	}
	if !table.Insert(3, 99) {
		t.Fatal("replacement")
	}
	var slot *any
	if st__find_or_add(table, 3, &slot) != 1 {
		t.Fatal("existing slot")
	}
	*slot = 123
	if v, _ := table.Lookup(3); v != 123 {
		t.Fatal(v)
	}
	if st__find_or_add(table, 250, &slot) != 0 {
		t.Fatal("new slot")
	}
	*slot = 7
	copy := table.Copy()
	copy.Delete(3)
	if _, ok := table.Lookup(3); !ok {
		t.Fatal("copy shares entries")
	}
}
func TestIterationDeleteAndStop(t *testing.T) {
	table := New(st__numcmp, st__numhash)
	for i := 0; i < 50; i++ {
		table.Insert(i, i)
	}
	table.ForEach(func(k, v any) VisitResult {
		if k.(int)%2 == 0 {
			return Delete
		}
		return Continue
	})
	if table.Count() != 25 {
		t.Fatal(table.Count())
	}
	seen := map[int]bool{}
	g := table.Generator()
	for {
		k, _, ok := g.Next()
		if !ok {
			break
		}
		if seen[k.(int)] {
			t.Fatal("duplicate")
		}
		seen[k.(int)] = true
	}
	if len(seen) != 25 {
		t.Fatal(seen)
	}
	calls := 0
	if table.ForEach(func(k, v any) VisitResult { calls++; return Stop }) || calls != 1 {
		t.Fatal(calls)
	}
	table.Clear()
	if table.Count() != 0 {
		t.Fatal("clear")
	}
}
func TestStringsPointersAndMemoryManaged(t *testing.T) {
	table := New(st__strcmp, st__strhash)
	table.Insert("abc", 42)
	if v, ok := table.Lookup("abc"); !ok || v != 42 {
		t.Fatal(v)
	}
	p, q := new(int), new(int)
	pt := New(st__ptrcmp, st__ptrhash)
	pt.Insert(p, "p")
	pt.Insert(q, "q")
	if pt.Count() != 2 {
		t.Fatal("pointer identities")
	}
	mm := stmm_init_table(stmm_numcmp, stmm_numhash)
	stmm_insert(mm, 1, 2)
	stmm_clean(mm)
	if stmm_count(mm) != 0 {
		t.Fatal("memory-managed clear")
	}
}
