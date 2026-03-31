package gogen

import (
	"strings"
	"testing"

	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

func TestGoType_BooleanSort(t *testing.T) {
	s := &lg.BooleanSort{}
	if got := GoType(s); got != "bool" {
		t.Errorf("GoType(BooleanSort) = %q, want %q", got, "bool")
	}
}

func TestGoType_EnumeratedSort(t *testing.T) {
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	if got := GoType(s); got != "Color" {
		t.Errorf("GoType(EnumeratedSort) = %q, want %q", got, "Color")
	}
}

func TestGoType_EnumeratedSort_Empty(t *testing.T) {
	s := &lg.EnumeratedSort{Name: "", Extension: nil}
	if got := GoType(s); got != "Enum" {
		t.Errorf("GoType(EnumeratedSort{empty}) = %q, want %q", got, "Enum")
	}
}

func TestGoType_RangeSort(t *testing.T) {
	s := &lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}
	if got := GoType(s); got != "int" {
		t.Errorf("GoType(RangeSort) = %q, want %q", got, "int")
	}
}

func TestGoType_UninterpretedSort(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "node"}
	if got := GoType(s); got != "int" {
		t.Errorf("GoType(UninterpretedSort) = %q, want %q", got, "int")
	}
}

func TestGoType_FunctionSort_Unary(t *testing.T) {
	dom := &lg.UninterpretedSort{Name: "node"}
	rng := &lg.BooleanSort{}
	fs, err := lg.NewFunctionSort(dom, rng)
	if err != nil {
		t.Fatal(err)
	}
	got := GoType(fs)
	if got != "map[int]bool" {
		t.Errorf("GoType(node->bool) = %q, want %q", got, "map[int]bool")
	}
}

func TestGoType_FunctionSort_Binary_SameTypes(t *testing.T) {
	dom := &lg.UninterpretedSort{Name: "node"}
	rng := &lg.BooleanSort{}
	fs, err := lg.NewFunctionSort(dom, dom, rng)
	if err != nil {
		t.Fatal(err)
	}
	got := GoType(fs)
	if got != "map[[2]int]bool" {
		t.Errorf("GoType(node,node->bool) = %q, want %q", got, "map[[2]int]bool")
	}
}

func TestGoType_TopSort(t *testing.T) {
	s := &lg.TopSort{Name: "T"}
	if got := GoType(s); got != "interface{}" {
		t.Errorf("GoType(TopSort) = %q, want %q", got, "interface{}")
	}
}

func TestStateFieldType_Bool(t *testing.T) {
	s := &lg.BooleanSort{}
	got := StateFieldType("flag", s)
	if got != "bool" {
		t.Errorf("StateFieldType bool = %q, want %q", got, "bool")
	}
}

func TestStateFieldType_UnaryRelation(t *testing.T) {
	dom := &lg.UninterpretedSort{Name: "node"}
	rng := &lg.BooleanSort{}
	fs, _ := lg.NewFunctionSort(dom, rng)
	got := StateFieldType("visited", fs)
	if got != "map[int]bool" {
		t.Errorf("StateFieldType unary rel = %q, want %q", got, "map[int]bool")
	}
}

func TestStateFieldType_UnaryFunction(t *testing.T) {
	dom := &lg.UninterpretedSort{Name: "node"}
	rng := &lg.UninterpretedSort{Name: "value"}
	fs, _ := lg.NewFunctionSort(dom, rng)
	got := StateFieldType("data", fs)
	if got != "map[int]int" {
		t.Errorf("StateFieldType unary func = %q, want %q", got, "map[int]int")
	}
}

func TestStateFieldType_BinaryRelation(t *testing.T) {
	dom := &lg.UninterpretedSort{Name: "node"}
	rng := &lg.BooleanSort{}
	fs, _ := lg.NewFunctionSort(dom, dom, rng)
	got := StateFieldType("edge", fs)
	if got != "map[[2]int]bool" {
		t.Errorf("StateFieldType binary rel = %q, want %q", got, "map[[2]int]bool")
	}
}

func TestGoZeroValue(t *testing.T) {
	tests := []struct {
		sort lg.Sort
		want string
	}{
		{&lg.BooleanSort{}, "false"},
		{&lg.EnumeratedSort{Name: "color"}, "0"},
		{&lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "0"}}, "0"},
		{&lg.UninterpretedSort{Name: "node"}, "0"},
	}
	for _, tt := range tests {
		got := GoZeroValue(tt.sort)
		if got != tt.want {
			t.Errorf("GoZeroValue(%T) = %q, want %q", tt.sort, got, tt.want)
		}
	}
}

func TestGoZeroValue_FunctionSort(t *testing.T) {
	fs, _ := lg.NewFunctionSort(&lg.UninterpretedSort{Name: "n"}, &lg.BooleanSort{})
	if got := GoZeroValue(fs); got != "nil" {
		t.Errorf("GoZeroValue(FunctionSort) = %q, want %q", got, "nil")
	}
}

func TestGoSortValues_Boolean(t *testing.T) {
	vals := GoSortValues(&lg.BooleanSort{})
	if len(vals) != 2 || vals[0] != "false" || vals[1] != "true" {
		t.Errorf("GoSortValues(Boolean) = %v, want [false true]", vals)
	}
}

func TestGoSortValues_Enumerated(t *testing.T) {
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	vals := GoSortValues(s)
	if len(vals) != 3 {
		t.Fatalf("GoSortValues(color) len = %d, want 3", len(vals))
	}
	if vals[0] != "Red" || vals[1] != "Green" || vals[2] != "Blue" {
		t.Errorf("GoSortValues(color) = %v, want [Red Green Blue]", vals)
	}
}

func TestGoSortValues_Range_Nil(t *testing.T) {
	s := &lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "5"}}
	vals := GoSortValues(s)
	if vals != nil {
		t.Errorf("GoSortValues(RangeSort) = %v, want nil", vals)
	}
}

func TestGoSortValues_Uninterpreted_Nil(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "node"}
	vals := GoSortValues(s)
	if vals != nil {
		t.Errorf("GoSortValues(UninterpretedSort) = %v, want nil", vals)
	}
}

func TestEmitEnumDecl(t *testing.T) {
	w := NewCodeWriter()
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	EmitEnumDecl(w, s)
	out := w.String()

	if !strings.Contains(out, "type Color int") {
		t.Errorf("missing type declaration in:\n%s", out)
	}
	if !strings.Contains(out, "Red Color = iota") {
		t.Errorf("missing iota in:\n%s", out)
	}
	if !strings.Contains(out, "Green") {
		t.Errorf("missing Green in:\n%s", out)
	}
	if !strings.Contains(out, "Blue") {
		t.Errorf("missing Blue in:\n%s", out)
	}
	if !strings.Contains(out, "var allColor") {
		t.Errorf("missing allColor in:\n%s", out)
	}
}

func TestEmitEnumDecl_Empty(t *testing.T) {
	w := NewCodeWriter()
	s := &lg.EnumeratedSort{Name: "empty", Extension: nil}
	EmitEnumDecl(w, s)
	out := w.String()
	if !strings.Contains(out, "type Empty int") {
		t.Errorf("missing type declaration in:\n%s", out)
	}
	// No const block expected for empty extension.
	if strings.Contains(out, "const (") {
		t.Errorf("unexpected const block for empty enum in:\n%s", out)
	}
}

func TestEmitRangeHelpers(t *testing.T) {
	w := NewCodeWriter()
	s := &lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}
	EmitRangeHelpers(w, s)
	out := w.String()
	if !strings.Contains(out, "const IdxLo = 0") {
		t.Errorf("missing lo const in:\n%s", out)
	}
	if !strings.Contains(out, "const IdxHi = 10") {
		t.Errorf("missing hi const in:\n%s", out)
	}
}

func TestEmitSortDecls_WithModule(t *testing.T) {
	mod := module.New()
	colorSort := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	mod.SortOrder = []string{"color"}
	mod.Sig.Sorts["color"] = colorSort

	w := NewCodeWriter()
	EmitSortDecls(w, mod)
	out := w.String()
	if !strings.Contains(out, "type Color int") {
		t.Errorf("EmitSortDecls did not emit Color type:\n%s", out)
	}
}

func TestGoExportedName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"foo.bar", "Foo_bar"},
		{"", ""},
		{"a", "A"},
	}
	for _, tt := range tests {
		got := goExportedName(tt.in)
		if got != tt.want {
			t.Errorf("goExportedName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGoUnexportedName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Hello", "hello"},
		{"hello", "hello"},
		{"Foo.bar", "foo_bar"},
		{"", ""},
		{"A", "a"},
	}
	for _, tt := range tests {
		got := goUnexportedName(tt.in)
		if got != tt.want {
			t.Errorf("goUnexportedName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
