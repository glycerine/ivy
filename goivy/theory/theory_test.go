package theory

import (
	"strings"
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// --- Theory construction tests ---

func TestNewIntegerTheory(t *testing.T) {
	th := NewIntegerTheory()
	if th.Name != "int" {
		t.Errorf("expected name int, got %s", th.Name)
	}
	if th.Kind != IntegerKind {
		t.Error("expected IntegerKind")
	}
	if th.Finite {
		t.Error("int should not be finite")
	}
	if th.String() != "int" {
		t.Errorf("String() = %s, want int", th.String())
	}
}

func TestNewNaturalTheory(t *testing.T) {
	th := NewNaturalTheory()
	if th.Name != "nat" {
		t.Errorf("expected name nat, got %s", th.Name)
	}
	if th.Kind != NaturalKind {
		t.Error("expected NaturalKind")
	}
	if th.Finite {
		t.Error("nat should not be finite")
	}
}

func TestNewBitVectorTheory(t *testing.T) {
	th := NewBitVectorTheory(32)
	if th.Name != "bv[32]" {
		t.Errorf("expected name bv[32], got %s", th.Name)
	}
	if th.Kind != BitVectorKind {
		t.Error("expected BitVectorKind")
	}
	if !th.Finite {
		t.Error("bv should be finite")
	}
	if th.Bits() != 32 {
		t.Errorf("Bits() = %d, want 32", th.Bits())
	}
}

func TestBitsPanicOnNonBV(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Bits() on int theory")
		}
	}()
	th := NewIntegerTheory()
	th.Bits()
}

// --- ParseTheory tests ---

func TestParseTheoryInt(t *testing.T) {
	th, err := ParseTheory("int")
	if err != nil {
		t.Fatal(err)
	}
	if th.Kind != IntegerKind {
		t.Error("expected IntegerKind")
	}
}

func TestParseTheoryNat(t *testing.T) {
	th, err := ParseTheory("nat")
	if err != nil {
		t.Fatal(err)
	}
	if th.Kind != NaturalKind {
		t.Error("expected NaturalKind")
	}
}

func TestParseTheoryBV(t *testing.T) {
	th, err := ParseTheory("bv[64]")
	if err != nil {
		t.Fatal(err)
	}
	if th.Kind != BitVectorKind {
		t.Error("expected BitVectorKind")
	}
	if th.Bits() != 64 {
		t.Errorf("Bits() = %d, want 64", th.Bits())
	}
}

func TestParseTheoryErrors(t *testing.T) {
	cases := []struct {
		input   string
		wantErr string
	}{
		{"bv[32", "bad theory syntax"},
		{"unknown", "unknown theory"},
		{"int[1]", "wrong number of theory parameters"},
		{"bv", "wrong number of theory parameters"},
		{"bv[abc]", "bad theory syntax"},
	}
	for _, tc := range cases {
		_, err := ParseTheory(tc.input)
		if err == nil {
			t.Errorf("ParseTheory(%q): expected error containing %q", tc.input, tc.wantErr)
			continue
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("ParseTheory(%q): error = %q, want substring %q", tc.input, err.Error(), tc.wantErr)
		}
	}
}

// --- Theories / schema tests ---

func TestTheories17(t *testing.T) {
	schemas := Theories("1.7")
	s, ok := schemas["int"]
	if !ok {
		t.Fatal("missing int schema")
	}
	if !strings.Contains(s, "theorem [base]") {
		t.Error("1.7 schema should contain 'theorem [base]'")
	}
}

func TestTheories16(t *testing.T) {
	schemas := Theories("1.6")
	s, ok := schemas["int"]
	if !ok {
		t.Fatal("missing int schema")
	}
	if strings.Contains(s, "theorem") {
		t.Error("1.6 schema should not contain 'theorem'")
	}
}

func TestGetTheorySchemataInt(t *testing.T) {
	s := GetTheorySchemata("int", &lg.UninterpretedSort{Name: "int"}, "1.7")
	if s == "" {
		t.Error("expected non-empty schema for int")
	}
}

func TestGetTheorySchemataRange(t *testing.T) {
	rs := &lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}
	s := GetTheorySchemata("idx", rs, "1.7")
	if s == "" {
		t.Error("expected non-empty schema for RangeSort")
	}
}

func TestGetTheorySchemataBV(t *testing.T) {
	s := GetTheorySchemata("bv[32]", &lg.UninterpretedSort{Name: "bv32"}, "1.7")
	if s == "" {
		t.Error("expected non-empty schema for bv[32]")
	}
}

func TestGetTheorySchemataNat(t *testing.T) {
	s := GetTheorySchemata("nat", &lg.UninterpretedSort{Name: "nat"}, "1.7")
	if s == "" {
		t.Error("expected non-empty schema for nat")
	}
}

func TestGetTheorySchemataOldVersion(t *testing.T) {
	s := GetTheorySchemata("int", &lg.UninterpretedSort{Name: "int"}, "1.5")
	if s != "" {
		t.Error("expected empty schema for version < 1.6")
	}
}

func TestGetTheorySchemataUnknown(t *testing.T) {
	s := GetTheorySchemata("custom", &lg.UninterpretedSort{Name: "custom"}, "1.7")
	if s != "" {
		t.Errorf("expected empty schema for unknown theory, got %q", s)
	}
}

// --- GetSortTheory tests ---

func TestGetSortTheoryWithInterp(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "myint"}
	interp := map[string]interface{}{"myint": "int"}
	result := GetSortTheory(sort, interp)
	th, ok := result.(*Theory)
	if !ok {
		t.Fatalf("expected *Theory, got %T", result)
	}
	if th.Kind != IntegerKind {
		t.Error("expected IntegerKind")
	}
}

func TestGetSortTheoryWithTheoryInterp(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "myint"}
	existing := NewNaturalTheory()
	interp := map[string]interface{}{"myint": existing}
	result := GetSortTheory(sort, interp)
	th, ok := result.(*Theory)
	if !ok {
		t.Fatalf("expected *Theory, got %T", result)
	}
	if th != existing {
		t.Error("expected same theory object")
	}
}

func TestGetSortTheoryNoInterp(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "myint"}
	interp := map[string]interface{}{}
	result := GetSortTheory(sort, interp)
	if result != sort {
		t.Error("expected sort itself when no interpretation")
	}
}

// --- HasIntegerInterp tests ---

func TestHasIntegerInterpInt(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "myint"}
	interp := map[string]interface{}{"myint": "int"}
	if !HasIntegerInterp(sort, interp) {
		t.Error("expected true for int interp")
	}
}

func TestHasIntegerInterpNat(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "mynat"}
	interp := map[string]interface{}{"mynat": "nat"}
	if !HasIntegerInterp(sort, interp) {
		t.Error("expected true for nat interp")
	}
}

func TestHasIntegerInterpRange(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "idx"}
	rs := &lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "10"}}
	interp := map[string]interface{}{"idx": rs}
	if !HasIntegerInterp(sort, interp) {
		t.Error("expected true for RangeSort interp")
	}
}

func TestHasIntegerInterpFalse(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "foo"}
	interp := map[string]interface{}{"foo": "bv[8]"}
	if HasIntegerInterp(sort, interp) {
		t.Error("expected false for bv interp")
	}
}

func TestHasIntegerInterpMissing(t *testing.T) {
	sort := &lg.UninterpretedSort{Name: "foo"}
	interp := map[string]interface{}{}
	if HasIntegerInterp(sort, interp) {
		t.Error("expected false when no interp")
	}
}

// --- VersionLE tests ---

func TestVersionLE(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.6", "1.7", true},
		{"1.7", "1.7", true},
		{"1.8", "1.7", false},
		{"1", "1.0", true},
		{"2.0", "1.9", false},
		{"1.7.1", "1.7.2", true},
		{"1.7.2", "1.7.1", false},
	}
	for _, tc := range cases {
		got := TheoryVersionLE(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("TheoryVersionLE(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// --- Fuzz test ---

func FuzzParseTheory(f *testing.F) {
	f.Add("int")
	f.Add("nat")
	f.Add("bv[32]")
	f.Add("bv[0]")
	f.Add("unknown")
	f.Add("")
	f.Add("bv[")
	f.Add("bv[]")
	f.Add("int[1]")
	f.Add("bv[999999999999]")
	f.Add("[[[")

	f.Fuzz(func(t *testing.T, name string) {
		th, err := ParseTheory(name)
		if err != nil {
			// Errors are expected for most inputs.
			return
		}
		// If parsing succeeds, basic invariants must hold.
		if th == nil {
			t.Fatal("nil theory with no error")
		}
		if th.Name == "" {
			t.Fatal("empty theory name")
		}
		if th.Kind == BitVectorKind {
			if !th.Finite {
				t.Error("bv theory should be finite")
			}
			if len(th.Args) != 1 {
				t.Error("bv theory should have 1 arg")
			}
		} else {
			if th.Finite {
				t.Error("non-bv theory should not be finite")
			}
		}
	})
}
