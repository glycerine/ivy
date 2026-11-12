package nm

import (
	"reflect"
	"testing"
)

func TestNamesakeRemoval(t *testing.T) {
	for _, order := range [][]int32{{1, 2, 3}, {2, 1, 3}, {3, 2, 1}} {
		p := Nm_ManCreate(1)
		for i := int32(1); i <= 3; i++ {
			Nm_ManStoreIdName(p, i, i, "shared", "")
		}
		for _, id := range order {
			if Nm_ManFindIdByName(p, "shared", id) != id {
				t.Fatal("missing type", id)
			}
			Nm_ManDeleteIdName(p, id)
			if Nm_ManFindIdByName(p, "shared", id) != -1 {
				t.Fatal("deleted type retained")
			}
			for _, e := range p.byID {
				if Nm_ManFindIdByName(p, "shared", int32(e.Type)) != int32(e.ObjId) {
					t.Fatal("broken namesake ring")
				}
			}
		}
		if Nm_ManNumEntries(p) != 0 || len(p.byName) != 0 {
			t.Fatal("not empty")
		}
	}
}
func TestManagerLifecycle(t *testing.T) {
	p := Nm_ManCreate(0)
	Nm_ManStoreIdName(p, 1, 2, "signal", "_out")
	if Nm_ManStoreIdName(p, 1, 3, "replacement", "") != "" || Nm_ManFindNameById(p, 1) != "signal_out" {
		t.Fatal("duplicate ID changed entry")
	}
	Nm_ManStoreIdName(p, 3, 4, "n2", "")
	Nm_ManStoreIdName(p, 4, 5, "n2_1", "")
	if Nm_ManCreateUniqueName(p, 2) != "n2_2" || Nm_ManCreateUniqueName(p, 1) != "signal_out" {
		t.Fatal("unique name")
	}
	if Nm_ManFindIdByNameTwoTypes(p, "signal_out", 9, 2) != 1 || Nm_ManFindIdByName(p, "missing", -1) != -1 {
		t.Fatal("lookup")
	}
	if !reflect.DeepEqual(Nm_ManReturnNameIds(p), []int32{1, 3, 4}) {
		t.Fatal("ids")
	}
	Nm_ManResize(p)
	if Nm_ManFindIdByName(p, "n2", -1) != 3 {
		t.Fatal("resize lost entry")
	}
	Nm_ManStoreIdName(p, 5, 1, "", "")
	if name, ok := p.Lookup(5); !ok || name != "" {
		t.Fatal("empty name")
	}
	Nm_ManDeleteIdName(p, 99)
	Nm_ManFree(p)
	if Nm_ManNumEntries(p) != 0 {
		t.Fatal("free")
	}
}
func TestManyEntries(t *testing.T) {
	p := Nm_ManCreate(1)
	for i := int32(0); i < 10000; i++ {
		Nm_ManStoreIdName(p, i, i%7, "group", "")
	}
	for i := int32(0); i < 10000; i += 2 {
		Nm_ManDeleteIdName(p, i)
	}
	if Nm_ManNumEntries(p) != 5000 {
		t.Fatal("count")
	}
	for i := int32(1); i < 10000; i += 2 {
		if e := Nm_ManTableLookupId(p, i); e == nil || e.Name != "group" {
			t.Fatal(i)
		}
	}
}
