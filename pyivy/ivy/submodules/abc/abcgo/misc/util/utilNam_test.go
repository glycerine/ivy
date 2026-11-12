package util

import (
	"fmt"
	"testing"
)

func TestNameManagerGrowthAndLimits(t *testing.T) {
	p := Abc_NamStart(1, 1)
	defer Abc_NamStop(p)
	for i := 1; i <= 1000; i++ {
		name := fmt.Sprintf("name%d", i)
		if id := Abc_NamStrFindOrAdd(p, name, nil); id != int32(i) {
			t.Fatal(id, i)
		}
	}
	for i := 1; i <= 1000; i++ {
		name := fmt.Sprintf("name%d", i)
		var found int32
		if id := Abc_NamStrFindOrAdd(p, name, &found); id != int32(i) || found != 1 || Abc_NamStr(p, id) != name {
			t.Fatal(id, found, name)
		}
	}
	id := Abc_NamStrFindOrAddLim(p, "prefixSuffix", 6, nil)
	if Abc_NamStr(p, id) != "prefix" || Abc_NamStrFindLim(p, "prefixOther", 6) != id {
		t.Fatal(id)
	}
	if Abc_NamMemAlloc(p) < Abc_NamMemUsed(p) {
		t.Fatal("memory accounting")
	}
}
func TestNameManagerCommonMapping(t *testing.T) {
	a, b := Abc_NamStart(1, 1), Abc_NamStart(1, 1)
	x := Abc_NamStrFindOrAdd(a, "x", nil)
	y := Abc_NamStrFindOrAdd(a, "y", nil)
	Abc_NamStrFindOrAdd(b, "y", nil)
	m := Abc_NamComputeIdMap(a, b)
	if m[x] != 0 || m[y] != 1 || Abc_NamReportCommon([]int32{x, y}, a, b) != 1 || Abc_NamReportUnique([]int32{x, y}, a, b) != "x" {
		t.Fatal(m)
	}
	Abc_NamRef(a)
	Abc_NamDeref(a)
	if Abc_NamStr(a, x) != "x" {
		t.Fatal("early release")
	}
	Abc_NamDeref(a)
	if len(a.handles) != 0 {
		t.Fatal("not released")
	}
}
