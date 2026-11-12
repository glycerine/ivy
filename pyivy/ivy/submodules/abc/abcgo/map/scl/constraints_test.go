package scl

import (
	"bytes"
	"github.com/glycerine/abcgo/misc/util"
	"strings"
	"testing"
)

func constraintNames(names ...string) *util.Abc_Nam_t {
	p := util.Abc_NamStart(16, 8)
	for _, s := range names {
		util.Abc_NamStrFindOrAdd(p, s, nil)
	}
	return p
}
func TestConstraintDefaultsOverridesAndWrite(t *testing.T) {
	p := Scl_ConAlloc("test.con", constraintNames("a", "b"), constraintNames("y"))
	defer Scl_ConFree(p)
	var diagnostics bytes.Buffer
	text := ".model test\n.default_input_cell INV\n.default_input_arrival 1.5\n.input_arrival b 2.5\n.output_load y 3.2\n.input_slew absent 9\n"
	if e := ParseConstraints(p, strings.NewReader(text), &diagnostics); e != nil {
		t.Fatal(e)
	}
	if p.ModelName != "test" || p.InputArrivals[0] != 1500 || p.InputArrivals[1] != 2500 || p.InputCells[1] != "INV" || p.OutputLoads[0] != 3200 || !strings.Contains(diagnostics.String(), "absent") {
		t.Fatal(p, diagnostics.String())
	}
	var out bytes.Buffer
	if e := WriteConstraints(&out, p); e != nil {
		t.Fatal(e)
	}
	q := Scl_ConAlloc("roundtrip", constraintNames("a", "b"), constraintNames("y"))
	defer Scl_ConFree(q)
	if e := ParseConstraints(q, &out, &diagnostics); e != nil {
		t.Fatal(e)
	}
	if q.InputArrivals[1] != 2500 || q.InputArrivals[0] != 1500 || q.OutputLoads[0] != 3200 {
		t.Fatal(q)
	}
}
func TestInvalidConstraintValue(t *testing.T) {
	p := Scl_ConAlloc("test.con", constraintNames("a"), constraintNames("y"))
	defer Scl_ConFree(p)
	if e := ParseConstraints(p, strings.NewReader(".input_arrival a nope"), &bytes.Buffer{}); e == nil {
		t.Fatal("accepted invalid number")
	}
}
