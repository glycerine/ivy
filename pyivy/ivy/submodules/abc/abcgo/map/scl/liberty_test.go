package scl

import (
	"math"
	"strings"
	"testing"
)

const testLiberty = `/* comment */ library(test) {
 time_unit : "1ps"; capacitive_load_unit(1,ff);
 lu_table_template(t) {variable_1: input_net_transition;variable_2:total_output_net_capacitance;index_1("1,3");index_2("2,4");}
 cell(inv) {area:2;pin(A){direction:input;capacitance:1;}pin(Y){direction:output;function:"!A";
 timing(){related_pin:"A";timing_sense:negative_unate;
 cell_rise(t){values("3,5", "5,7");} cell_fall(t){values("3,5", "5,7");}
 rise_transition(t){values("3,5", "5,7");} fall_transition(t){values("3,5", "5,7");}
 }}}
 cell(skip){dont_use:true;pin(A){direction:input;}pin(Y){direction:output;function:"A";}}
 cell(seq){ff(IQ,IQN){next_state:D;}pin(D){direction:input;}pin(Q){direction:output;function:"IQ";}}
}`

func TestLibertyModelAndLookup(t *testing.T) {
	p, e := ParseLiberty("test.lib", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	if p.Name != "test" || len(p.Cells) != 1 || len(p.CellClasses) != 1 {
		t.Fatal(p)
	}
	c := p.Cells[0]
	if c.Inputs != 1 || c.Outputs != 1 || c.Area != 2 || SC_CellFunc(c)[0] != 0x5555555555555555 {
		t.Fatal(c)
	}
	surface := &Scl_CellPinTime(c, 0).CellRise
	for _, point := range [][3]float32{{2, 3, 5}, {0, 1, 1}, {4, 5, 9}} {
		if got := Scl_LibLookup(surface, point[0], point[1]); got != point[2] {
			t.Fatal(point, got)
		}
	}
	if Abc_SclFindInvertor(p, 0) != c || Abc_SclCellFind(p, "inv") != 0 || Abc_SclCellFind(p, "absent") != -1 {
		t.Fatal("lookup")
	}
}
func TestLibertyDiagnostics(t *testing.T) {
	for _, s := range []string{`library(x){`, `library(x){key:"unterminated}`, `library(x){/*`, `library(x){cell(x){pin(Y){function:"unknown";}}}`} {
		if _, e := ParseLiberty("bad.lib", s); e == nil {
			t.Fatal("accepted", s)
		}
	}
	if _, e := ParseLibertyTree("x", `library(x){ custom("http://a/*b*/") { value:"a;b"; } }`); e != nil {
		t.Fatal(e)
	}
}
func TestFormulaTruth(t *testing.T) {
	pins := []*SC_Pin{{Name: "A"}, {Name: "B"}, {Name: "C"}}
	for _, test := range []struct {
		formula string
		want    uint64
	}{{"A B + !C", 0x8f8f8f8f8f8f8f8f}, {"(A^B)'", 0x9999999999999999}, {"A * (B + C)", 0xa8a8a8a8a8a8a8a8}} {
		got, e := formulaTruth(test.formula, pins)
		if e != nil || got[0] != test.want {
			t.Fatal(test, got, e)
		}
	}
}
func TestNormalizeUnits(t *testing.T) {
	p, e := ParseLiberty("test", strings.ReplaceAll(strings.ReplaceAll(testLiberty, `"1ps"`, `"1ns"`), `(1,ff)`, `(1,pf)`))
	if e != nil {
		t.Fatal(e)
	}
	if p.UnitTime != 12 || p.UnitCapExponent != 15 || p.Cells[0].Pins[0].RiseCap != 1000 {
		t.Fatal(p)
	}
	s := &Scl_CellPinTime(p.Cells[0], 0).CellRise
	if math.Abs(float64(Scl_LibLookup(s, 2000, 3000)-5000)) > 1e-4 {
		t.Fatal(s)
	}
}
func TestFunctionClassRing(t *testing.T) {
	p, e := ParseLiberty("test", strings.Replace(testLiberty, "cell(inv)", "cell(inv2) {area:1;pin(A){direction:input;}pin(Y){direction:output;function:\"!A\";}} cell(inv)", 1))
	if e != nil {
		t.Fatal(e)
	}
	if len(p.CellClasses) != 1 || Abc_SclClassCellNum(p.CellClasses[0]) != 2 || p.CellClasses[0].Name != "inv2" {
		t.Fatal(p.CellClasses)
	}
}

func TestSCLBinaryRoundTripAndTruncation(t *testing.T) {
	p, e := ParseLiberty("test", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	data, e := MarshalSCL(p)
	if e != nil {
		t.Fatal(e)
	}
	q, e := Abc_SclReadFromStr(data)
	if e != nil {
		t.Fatal(e)
	}
	again, e := MarshalSCL(q)
	if e != nil {
		t.Fatal(e)
	}
	if string(again) != string(data) {
		t.Fatal("binary round trip changed bytes")
	}
	for i := 0; i < len(data); i++ {
		if _, e := Abc_SclReadFromStr(data[:i]); e == nil {
			t.Fatal("accepted truncated file", i)
		}
	}
	if _, e := Abc_SclReadFromStr(append(data, 0)); e == nil {
		t.Fatal("accepted trailing data")
	}
}

func TestQuotedPunctuationAndCommentTokens(t *testing.T) {
	tree, e := ParseLibertyTree("x", `library(x){custom(")",","){value:";";path:"http://foo/*bar*/";}}`)
	if e != nil {
		t.Fatal(e)
	}
	n := tree.Root.Children[0]
	if len(n.Arguments) != 2 || n.Arguments[0] != ")" || n.Arguments[1] != "," || n.value("value") != ";" || n.value("path") != "http://foo/*bar*/" {
		t.Fatal(n)
	}
}
func TestTimingFallbackAndWorstCase(t *testing.T) {
	text := strings.ReplaceAll(testLiberty, `cell_fall(t){values("3,5", "5,7");}`, "")
	p, e := ParseLiberty("test", text)
	if e != nil {
		t.Fatal(e)
	}
	if Scl_CellPinTime(p.Cells[0], 0).CellFall.Data[0][0] != 3 {
		t.Fatal("missing paired fallback")
	}
	tree, e := ParseLibertyTree("test", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	output := tree.Root.child("cell").Children[2]
	timing := output.child("timing")
	if timing == nil {
		t.Fatal("fixture")
	}
	duplicate := *timing
	duplicate.Children = append([]*LibertyItem(nil), timing.Children...)
	for i, c := range duplicate.Children {
		if c.Key == "cell_rise" {
			copy := *c
			copy.Children = []*LibertyItem{{Key: "values", Arguments: []string{"9,1", "2,3"}}}
			duplicate.Children[i] = &copy
		}
	}
	output.Children = append(output.Children, &duplicate)
	p, e = libraryFromTree(tree)
	if e != nil {
		t.Fatal(e)
	}
	group := p.Cells[0].Pins[1].Timings[0]
	if len(group.Timings) != 1 || group.Timings[0].CellRise.Data[0][0] != 9 || group.Timings[0].CellRise.Data[0][1] != 5 {
		t.Fatal(group)
	}
}

func TestLibertyWriterRoundTrip(t *testing.T) {
	p, e := ParseLiberty("test", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	text, e := LibertyText(p)
	if e != nil {
		t.Fatal(e)
	}
	q, e := ParseLiberty("roundtrip", text)
	if e != nil {
		t.Fatal(e, text)
	}
	a, e := MarshalSCL(p)
	if e != nil {
		t.Fatal(e)
	}
	b, e := MarshalSCL(q)
	if e != nil {
		t.Fatal(e)
	}
	if string(a) != string(b) {
		t.Fatal("Liberty roundtrip changed model")
	}
}
func TestMIOBridge(t *testing.T) {
	p, e := ParseLiberty("test", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Abc_SclDeriveGenlibSimple(p); e != nil {
		t.Fatal(e)
	}
	if _, e = Abc_SclDeriveGenlib(p, nil, 1, 1, 0, 0); e != nil {
		t.Fatal(e)
	}
	if id, ok := p.MioCellID("inv"); !ok || id != 0 {
		t.Fatal(id, ok)
	}
}
