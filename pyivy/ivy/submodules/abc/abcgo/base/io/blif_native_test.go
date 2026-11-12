package io

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("test_vectors/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func TestBLIFHierarchyForwardBindingsAndRoundTrip(t *testing.T) {
	n, err := ParseBLIF("hierarchy.blif", readFixture(t, "hierarchy.blif"), false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Design.Modules) != 2 || n.Name != "top" {
		t.Fatal("lost models")
	}
	graph := abc.Abc_NtkDeriveFlatGia(n)
	if graph == nil {
		t.Fatal("flatten failed")
	}
	for bits := 0; bits < 4; bits++ {
		got, err := graph.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || got[0] != (bits == 1 || bits == 2) {
			t.Fatal(bits, got, err)
		}
	}
	var text bytes.Buffer
	if err = WriteBLIF(&text, n, false, true); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBLIF("roundtrip.blif", text.String(), false, true)
	if err != nil {
		t.Fatal(err, text.String())
	}
	if len(copy.Design.Modules) != 2 {
		t.Fatal("lost hierarchy")
	}
}
func TestBLIFMVSymbolicDefaultsSetsEquality(t *testing.T) {
	n, err := ParseBLIF("symbolic.mv", readFixture(t, "symbolic.mv"), true, true)
	if err != nil {
		t.Fatal(err)
	}
	node := n.ObjectsOfKind(abc.ObjectNode)[0]
	table := node.Data.(*abc.MVTable)
	for _, test := range []struct{ in, want []int }{{[]int{0, 2}, []int{1}}, {[]int{1, 0}, []int{1}}, {[]int{2, 0}, []int{2}}, {[]int{1, 2}, []int{2}}} {
		got, err := table.Evaluate(test.in)
		if err != nil || !reflect.DeepEqual(got, test.want) {
			t.Fatal(test.in, got, err)
		}
	}
	var text bytes.Buffer
	if err = WriteBLIF(&text, n, true, true); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBLIF("copy.mv", text.String(), true, true)
	if err != nil {
		t.Fatal(err, text.String())
	}
	if got := copy.ObjectsOfKind(abc.ObjectNode)[0].Data.(*abc.MVTable).String(); got != table.String() {
		t.Fatal(got, table.String())
	}
}
func TestBLIFMalformedLocations(t *testing.T) {
	for _, source := range []string{".model x\n.inputs a\n.outputs z\n.end", ".model x\n.inputs a a\n.end", ".model x\n.names z\n2\n.end", ".model x\n.mv a 3 red red blue\n.end", ".model x\n.subckt missing A=a\n.end"} {
		if _, err := ParseBLIF("bad.blif", source, true, true); err == nil {
			t.Error("accepted", source)
		} else if !strings.Contains(err.Error(), "bad.blif") {
			t.Error("missing filename", err)
		}
	}
}

func TestBLIFMVMultiOutputAndReset(t *testing.T) {
	const source = `.model state
.inputs d
.outputs q y
.mv d,q,y 3 red green blue
.latch d q
.reset q
red
.table d->y
.default blue
red green
blue red
.end
`
	n, err := ParseBLIF("reset.mv", source, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.ObjectsOfKind(abc.ObjectLatch)) != 2 {
		t.Fatal("shared reset-state latch missing")
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		for _, output := range node.Fanouts() {
			if output.Name() == "q" {
				table := node.Data.(*abc.MVTable)
				for _, test := range []struct{ in, want []int }{{[]int{0, 2, 0}, []int{0}}, {[]int{1, 2, 0}, []int{2}}} {
					got, err := table.Evaluate(test.in)
					if err != nil || !reflect.DeepEqual(got, test.want) {
						t.Fatal(got, err)
					}
				}
			}
		}
	}
	const multi = `.model multi
.inputs a
.outputs x y
.mv a,x,y 3
.names a -> x y
.default 0 1
2 1 2
.end
`
	m, err := ParseBLIF("multi.mv", multi, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ObjectsOfKind(abc.ObjectNode)) != 2 {
		t.Fatal("did not project output tables")
	}
}

func TestBLIFMVNumericSymbolsAndBinaryLabels(t *testing.T) {
	const source = `.model symbols
.inputs a
.outputs y
.mv a,y 2 1 0
.names a y
1 0
0 1
.end
`
	n, err := ParseBLIF("symbols.mv", source, true, true)
	if err != nil {
		t.Fatal(err)
	}
	table := n.ObjectsOfKind(abc.ObjectNode)[0].Data.(*abc.MVTable)
	got, err := table.Evaluate([]int{0})
	if err != nil || !reflect.DeepEqual(got, []int{1}) {
		t.Fatal("numeric symbol interpreted as index", got, err)
	}
	var text bytes.Buffer
	if err = WriteBLIF(&text, n, true, true); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBLIF("copy.mv", text.String(), true, true)
	if err != nil {
		t.Fatal(err, text.String())
	}
	got, err = copy.ObjectsOfKind(abc.ObjectNode)[0].Data.(*abc.MVTable).Evaluate([]int{0})
	if err != nil || !reflect.DeepEqual(got, []int{1}) {
		t.Fatal(got, err)
	}
	for _, net := range copy.ObjectsOfKind(abc.ObjectNet) {
		if variable := abc.Abc_ObjMvVar(net); variable == nil || !reflect.DeepEqual(variable.Names, []string{"1", "0"}) {
			t.Fatal("binary labels lost")
		}
	}
}
