package ivyutils

import (
	"testing"
)

// --- UniqueRenamer tests ---

func TestUniqueRenamerBasic(t *testing.T) {
	rn := NewUniqueRenamer("", nil)
	name := rn.Rename("foo")
	if name != "foo" {
		t.Errorf("expected foo, got %s", name)
	}
	// Second call with same name should generate a different name
	name2 := rn.Rename("foo")
	if name2 == "foo" {
		t.Error("second rename of foo should be different")
	}
	if name2 == name {
		t.Error("should not duplicate")
	}
}

func TestUniqueRenamerWithPrefix(t *testing.T) {
	rn := NewUniqueRenamer("pre_", nil)
	name := rn.Rename("bar")
	if name != "pre_bar" {
		t.Errorf("expected pre_bar, got %s", name)
	}
}

func TestUniqueRenamerWithUsed(t *testing.T) {
	rn := NewUniqueRenamer("", []string{"foo"})
	name := rn.Rename("foo")
	if name == "foo" {
		t.Error("foo is already used, should generate different name")
	}
}

func TestUniqueRenamerEmptyName(t *testing.T) {
	rn := NewUniqueRenamer("x", nil)
	name := rn.Rename("")
	if name == "" {
		t.Error("should generate a non-empty name")
	}
}

// --- VariableGenerator tests ---

func TestVariableGenerator(t *testing.T) {
	gen := NewVariableGenerator()
	names := make([]string, 0)
	for i := 0; i < 30; i++ {
		n := gen.Generate("")
		names = append(names, n)
	}
	// Should get A, B, C, ..., N, P, Q, ..., Z, AA, BB, ...
	// (skipping O)
	if names[0] != "A" {
		t.Errorf("first name should be A, got %s", names[0])
	}
	// Check uniqueness
	seen := make(map[string]struct{})
	for _, n := range names {
		if _, ok := seen[n]; ok {
			t.Errorf("duplicate name: %s", n)
		}
		seen[n] = struct{}{}
	}
	// O should not appear
	if _, ok := seen["O"]; ok {
		t.Error("O should be skipped")
	}
}

func TestVariableGeneratorWithName(t *testing.T) {
	gen := NewVariableGenerator()
	name := gen.Generate("xyz")
	if name != "X" {
		t.Errorf("expected X, got %s", name)
	}
}

// --- ConstantNameGenerator tests ---

func TestConstantNameGenerator(t *testing.T) {
	gen := ConstantNameGenerator()
	// First 26 should be a-z
	for i := 0; i < 26; i++ {
		name := gen()
		expected := string(rune('a' + i))
		if name != expected {
			t.Errorf("expected %s, got %s", expected, name)
		}
	}
	// Next should be a0
	name := gen()
	if name != "a0" {
		t.Errorf("expected a0, got %s", name)
	}
}

func TestUnusedNameWithBase(t *testing.T) {
	used := map[string]struct{}{
		"x_a": {},
		"x_b": {},
	}
	name := UnusedNameWithBase("x", used)
	if name == "x_a" || name == "x_b" {
		t.Errorf("should not return used name: %s", name)
	}
	if name != "x_c" {
		t.Errorf("expected x_c, got %s", name)
	}
}

func TestDistinctRenaming(t *testing.T) {
	names1 := []string{"a", "b", "c"}
	names2 := []string{"a", "c"}
	result := DistinctRenaming(names1, names2)
	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
	// "b" should map to itself since it's not in names2
	if result["b"] != "b" {
		t.Errorf("b should map to b, got %s", result["b"])
	}
	// "a" should map to something different from "a"
	if result["a"] == "a" {
		t.Error("a should be renamed since it's in names2")
	}
}

// --- Parameter tests ---

func TestParameter(t *testing.T) {
	p := NewParameter("test_param_1", 42)
	if p.Get() != 42 {
		t.Errorf("expected 42, got %v", p.Get())
	}
	if err := p.Set("new_val"); err != nil {
		t.Fatal(err)
	}
	if p.Get() != "new_val" {
		t.Errorf("expected new_val, got %v", p.Get())
	}
}

func TestBooleanParameter(t *testing.T) {
	p := NewBooleanParameter("test_bool_1", false)
	if p.GetBool() != false {
		t.Error("expected false")
	}
	if err := p.Set("true"); err != nil {
		t.Fatal(err)
	}
	if p.GetBool() != true {
		t.Error("expected true after set")
	}
	if err := p.Set("invalid"); err == nil {
		t.Error("expected error for invalid boolean")
	}
}

func TestEnumeratedParameter(t *testing.T) {
	p := NewEnumeratedParameter("test_enum_1", []string{"a", "b", "c"}, "a")
	if p.GetString() != "a" {
		t.Errorf("expected a, got %s", p.GetString())
	}
	if err := p.Set("b"); err != nil {
		t.Fatal(err)
	}
	if p.GetString() != "b" {
		t.Errorf("expected b, got %s", p.GetString())
	}
	if err := p.Set("d"); err == nil {
		t.Error("expected error for invalid enum value")
	}
}

func TestParameterize(t *testing.T) {
	p := NewParameter("test_parameterize_1", "original")
	pz, err := NewParameterize(map[string]interface{}{
		"test_parameterize_1": "temporary",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Get() != "temporary" {
		t.Errorf("expected temporary, got %v", p.Get())
	}
	pz.Restore()
	if p.Get() != "original" {
		t.Errorf("expected original after restore, got %v", p.Get())
	}
}

// --- TopologicalSort tests ---

func TestTopologicalSort(t *testing.T) {
	items := []string{"a", "b", "c", "d"}
	order := [][2]string{
		{"a", "b"}, // a before b
		{"b", "c"}, // b before c
	}
	result := TopologicalSort(items, order, func(s string) string { return s })
	// a should come before b, b before c
	indexOf := make(map[string]int)
	for i, v := range result {
		indexOf[v] = i
	}
	if indexOf["a"] >= indexOf["b"] {
		t.Error("a should come before b")
	}
	if indexOf["b"] >= indexOf["c"] {
		t.Error("b should come before c")
	}
}

func TestTopologicalSortEmpty(t *testing.T) {
	result := TopologicalSort([]string{}, nil, func(s string) string { return s })
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

// --- FindCycle tests ---

func TestFindCycleNoCycle(t *testing.T) {
	arcs := []Arc[string]{
		{"a", "b"},
		{"b", "c"},
	}
	cycle := FindCycle(arcs)
	if cycle != nil {
		t.Errorf("expected no cycle, got %v", cycle)
	}
}

func TestFindCycleWithCycle(t *testing.T) {
	arcs := []Arc[string]{
		{"a", "b"},
		{"b", "c"},
		{"c", "a"},
	}
	cycle := FindCycle(arcs)
	if cycle == nil {
		t.Error("expected a cycle")
	}
}

func TestFindCycleSelfLoop(t *testing.T) {
	arcs := []Arc[string]{
		{"a", "a"},
	}
	cycle := FindCycle(arcs)
	if cycle == nil {
		t.Error("expected self-loop cycle")
	}
}

// --- Name utility tests ---

func TestComposeNames(t *testing.T) {
	if ComposeNames("a", "b", "c") != "a.b.c" {
		t.Errorf("got %s", ComposeNames("a", "b", "c"))
	}
	if ComposeNames("this", "b", "c") != "b.c" {
		t.Errorf("this should be skipped, got %s", ComposeNames("this", "b", "c"))
	}
}

func TestSplitName(t *testing.T) {
	parts := SplitName("a.b.c")
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b" || parts[2] != "c" {
		t.Errorf("expected [a b c], got %v", parts)
	}
}

func TestSplitNameWithSubscript(t *testing.T) {
	parts := SplitName("a.b[1].c")
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b[1]" || parts[2] != "c" {
		t.Errorf("expected [a b[1] c], got %v", parts)
	}
}

func TestSplitNameQuoted(t *testing.T) {
	parts := SplitName("\"foo.bar\"")
	if len(parts) != 1 || parts[0] != "\"foo.bar\"" {
		t.Errorf("quoted name should not split, got %v", parts)
	}
}

func TestBaseName(t *testing.T) {
	if BaseName("a.b.c") != "a" {
		t.Errorf("expected a, got %s", BaseName("a.b.c"))
	}
}

func TestParentChildName(t *testing.T) {
	pc := ParentChildName("a.b.c")
	if pc[0] != "a.b" || pc[1] != "c" {
		t.Errorf("expected [a.b c], got %v", pc)
	}
	pc2 := ParentChildName("solo")
	if pc2[0] != "this" || pc2[1] != "solo" {
		t.Errorf("expected [this solo], got %v", pc2)
	}
}

func TestExtractParametersName(t *testing.T) {
	name, parms := ExtractParametersName("f[1][2]")
	if name != "f" {
		t.Errorf("expected f, got %s", name)
	}
	if len(parms) != 2 || parms[0] != "1" || parms[1] != "2" {
		t.Errorf("expected [1 2], got %v", parms)
	}
}

func TestAddParamsName(t *testing.T) {
	result := AddParamsName("f", []string{"1", "2"})
	if result != "f[1][2]" {
		t.Errorf("expected f[1][2], got %s", result)
	}
}

func TestDistinct(t *testing.T) {
	if !Distinct([]int{1, 2, 3}) {
		t.Error("expected true for distinct elements")
	}
	if Distinct([]int{1, 2, 1}) {
		t.Error("expected false for duplicate elements")
	}
}

// --- Fuzz tests ---

func FuzzUniqueRenamer(f *testing.F) {
	f.Add("", "x")
	f.Add("pre_", "name")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, prefix, name string) {
		rn := NewUniqueRenamer(prefix, nil)
		// Should not panic, should return non-empty
		result := rn.Rename(name)
		if result == "" {
			t.Error("rename should never return empty string")
		}
		// Second call should return different name
		result2 := rn.Rename(name)
		if result == result2 {
			t.Error("second rename should be different")
		}
	})
}

func FuzzSplitName(f *testing.F) {
	f.Add("a.b.c")
	f.Add("simple")
	f.Add("a[1].b")
	f.Add("\"quoted\"")
	f.Fuzz(func(t *testing.T, name string) {
		// Should not panic
		_ = SplitName(name)
	})
}
