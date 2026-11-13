package abs

import "testing"

func TestVtaObjectTableGrowthAndRollback(t *testing.T) {
	table := newVtaObjectTable(2)
	first := table.findOrAdd(3, -1)
	for i := int32(0); i < 1000; i++ {
		object := table.findOrAdd(i, i%7)
		if got := table.find(i, i%7); got != object || object.id != i+2 {
			t.Fatalf("object %d lost during growth", i)
		}
	}
	if table.find(3, -1) != first || first.id != 1 {
		t.Fatal("pointer changed during growth")
	}
	if table.findOrAdd(3, -1) != first || len(table.objects) != 1002 {
		t.Fatal("duplicate created")
	}
	table.rollback(501)
	if len(table.objects) != 501 || table.find(499, 499%7) != nil {
		t.Fatal("rollback retained objects")
	}
	if table.find(498, 498%7) == nil || table.find(3, -1) != first {
		t.Fatal("rollback lost prior objects")
	}
	if object := table.findOrAdd(499, 499%7); object.id != 501 {
		t.Fatalf("reused ID %d", object.id)
	}
}

func TestVtaObjectTableDeleteAndReinsert(t *testing.T) {
	table := newVtaObjectTable(2)
	old := table.findOrAdd(2, 5)
	table.delete(2, 5)
	if table.find(2, 5) != nil || old.iNext != -1 {
		t.Fatal("deleted object still reachable")
	}
	if replacement := table.findOrAdd(2, 5); replacement == old || replacement.id != 2 {
		t.Fatal("wrong replacement")
	}
	table.rollback(1)
	if table.find(2, 5) != nil {
		t.Fatal("rollback left replacement")
	}
}
