package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func checkBooleanNetwork(t *testing.T, n *abc.Network, want func([]bool) bool) {
	t.Helper()
	graph := abc.Abc_NtkDeriveFlatGia(n)
	if graph == nil {
		t.Fatal("network did not lower to AIG")
	}
	inputs := len(n.ObjectsOfKind(abc.ObjectPI))
	for bits := 0; bits < 1<<uint(inputs); bits++ {
		assignment := make([]bool, inputs)
		for i := range assignment {
			assignment[i] = bits>>uint(i)&1 != 0
		}
		got, err := graph.Evaluate(assignment)
		if err != nil || len(got) != 1 || got[0] != want(assignment) {
			t.Fatalf("%v: got%v err%v", assignment, got, err)
		}
	}
}
func TestEquationAndBenchRoundTrip(t *testing.T) {
	n, err := ParseEqn("mux.eqn", "INORDER = a b s; OUTORDER = y; y= a*s + b*!s;", true)
	if err != nil {
		t.Fatal(err)
	}
	want := func(a []bool) bool {
		if a[2] {
			return a[0]
		}
		return a[1]
	}
	checkBooleanNetwork(t, n, want)
	var text bytes.Buffer
	if err = WriteBench(&text, n); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBench("mux.bench", text.String(), true)
	if err != nil {
		t.Fatal(err, text.String())
	}
	checkBooleanNetwork(t, copy, want)
}
func TestDSDPrimeXorComplement(t *testing.T) {
	for _, test := range []struct {
		text string
		want func([]bool) bool
	}{{"!(a+b)", func(a []bool) bool { return a[0] == a[1] }}, {"8(a,b)", func(a []bool) bool { return a[0] && a[1] }}, {"(a*!b)", func(a []bool) bool { return a[0] && !a[1] }}} {
		n, err := ParseDSD(test.text)
		if err != nil {
			t.Fatal(err)
		}
		checkBooleanNetwork(t, n, test.want)
	}
	for _, bad := range []string{"(a*b+c)", "8(a,)", "!", "(a"} {
		if _, err := ParseDSD(bad); err == nil {
			t.Fatal("accepted", bad)
		}
	}
}
func TestBenchLUTAndReservedConstantNames(t *testing.T) {
	source := "INPUT(vdd)\nOUTPUT(z)\nz = LUT 0x6(vdd,gnd)\n"
	n, err := ParseBench("lut.bench", source, true)
	if err != nil {
		t.Fatal(err)
	}
	checkBooleanNetwork(t, n, func(a []bool) bool { return a[0] })
	var text bytes.Buffer
	if err = WriteBench(&text, n); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBench("copy.bench", text.String(), true)
	if err != nil {
		t.Fatal(err)
	}
	checkBooleanNetwork(t, copy, func(a []bool) bool { return a[0] })
}
func TestBenchArbitrarySOPAndConstant(t *testing.T) {
	n, err := ParseBLIF("sop.blif", ".model t\n.inputs a b\n.outputs y\n.names a b y\n10 1\n01 1\n.end\n", false, true)
	if err != nil {
		t.Fatal(err)
	}
	var text bytes.Buffer
	if err = WriteBench(&text, n); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseBench("sop.bench", text.String(), true)
	if err != nil {
		t.Fatal(err)
	}
	checkBooleanNetwork(t, copy, func(a []bool) bool { return a[0] != a[1] })
	if _, err = ParseEqn("bad.eqn", "INORDER=a;OUTORDER=y;y=missing;", true); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatal(err)
	}
}

func TestEquationWriterPreservesComplementedCover(t *testing.T) {
	n, err := ParseBLIF("nand.blif", ".model n\n.inputs a b\n.outputs y\n.names a b y\n11 0\n.end\n", false, true)
	if err != nil {
		t.Fatal(err)
	}
	var text bytes.Buffer
	if err = WriteEqn(&text, n); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseEqn("nand.eqn", text.String(), true)
	if err != nil {
		t.Fatal(err, text.String())
	}
	checkBooleanNetwork(t, copy, func(a []bool) bool { return !(a[0] && a[1]) })
}
