// Tests for Z3 Enumerated Types as Native Sorts (todo.md item 8).
//
// Part 1: divergence fixes (getModelConstant + FunctionModelToClauses guards)
// Part 2: flag-gated enumerated_to_numeral implementation
package z3bridge

import (
	"strings"
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// ============================================================
// Part 1: Divergence fixes — getModelConstant / FunctionModelToClauses
// ============================================================

// TestGetModelConstant_NativeEnum verifies that with UseZ3Enums=true (default),
// getModelConstant takes the general constant_from_z3 path (not the iteration
// path). This tests the fix for the missing "not use_z3_enums" guard.
func TestGetModelConstant_NativeEnum(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	sig.AddSort(es)

	opts := module.DefaultSolverOptions()
	// UseZ3Enums defaults to true
	s := NewSolverFromSig(sig, opts)
	tr := s.Translator()
	ctx := s.Context()

	// Translate the enum sort so constructors are cached
	_, err := tr.TranslateSort(es)
	if err != nil {
		t.Fatalf("TranslateSort: %v", err)
	}

	// Create a variable x of sort color
	xConst := lg.NewConst("x", es)
	zx, err := tr.Translate(xConst)
	if err != nil {
		t.Fatalf("Translate x: %v", err)
	}

	// Create green constant
	greenConst := lg.NewConst("green", es)
	zgreen, err := tr.Translate(greenConst)
	if err != nil {
		t.Fatalf("Translate green: %v", err)
	}

	// Assert x == green
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(zx, zgreen))
	if slv.Check() != Sat {
		t.Fatal("x == green should be SAT")
	}
	model := slv.Model()

	// Build HerbrandModel and extract constant
	h := NewHerbrandModel(s, slv, model, nil)
	result := h.EvalConstant(xConst)
	if result.Name != "green" {
		t.Errorf("expected EvalConstant to return green, got %q", result.Name)
	}
}

// TestGetModelConstant_BinaryEncoding verifies that with UseZ3Enums=false,
// the iteration path (checking each enum value) works correctly.
func TestGetModelConstant_BinaryEncoding(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	sig.AddSort(es)

	opts := module.DefaultSolverOptions()
	s := NewSolverFromSig(sig, opts)
	s.SetUseNativeEnums(false) // force binary encoding path
	tr := s.Translator()
	ctx := s.Context()

	// Translate the sort to register constructors
	_, err := tr.TranslateSort(es)
	if err != nil {
		t.Fatalf("TranslateSort: %v", err)
	}

	// Create and translate constants
	xConst := lg.NewConst("x", es)
	zx, err := tr.Translate(xConst)
	if err != nil {
		t.Fatalf("Translate x: %v", err)
	}

	blueConst := lg.NewConst("blue", es)
	zblue, err := tr.Translate(blueConst)
	if err != nil {
		t.Fatalf("Translate blue: %v", err)
	}

	// Assert x == blue
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(zx, zblue))
	if slv.Check() != Sat {
		t.Fatal("x == blue should be SAT")
	}
	model := slv.Model()

	h := NewHerbrandModel(s, slv, model, nil)
	result := h.EvalConstant(xConst)
	if result.Name != "blue" {
		t.Errorf("expected EvalConstant to return blue, got %q", result.Name)
	}
}

// TestGetModelConstant_AllEnumValues verifies model extraction for each
// possible enum value with both UseZ3Enums=true and false.
func TestGetModelConstant_AllEnumValues(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}

	for _, useNative := range []bool{true, false} {
		for _, expected := range es.Extension {
			t.Run(expected+"_native="+boolStr(useNative), func(t *testing.T) {
				sig := il.NewSig()
				sig.Constructors["red"] = true
				sig.Constructors["green"] = true
				sig.Constructors["blue"] = true
				sig.AddSort(es)

				opts := module.DefaultSolverOptions()
				s := NewSolverFromSig(sig, opts)
				if !useNative {
					s.SetUseNativeEnums(false)
				}
				tr := s.Translator()
				ctx := s.Context()

				_, err := tr.TranslateSort(es)
				if err != nil {
					t.Fatalf("TranslateSort: %v", err)
				}

				xConst := lg.NewConst("x", es)
				zx, err := tr.Translate(xConst)
				if err != nil {
					t.Fatalf("Translate x: %v", err)
				}

				valConst := lg.NewConst(expected, es)
				zval, err := tr.Translate(valConst)
				if err != nil {
					t.Fatalf("Translate %s: %v", expected, err)
				}

				slv := ctx.NewZ3Solver()
				slv.Assert(ctx.Eq(zx, zval))
				if slv.Check() != Sat {
					t.Fatalf("x == %s should be SAT", expected)
				}
				model := slv.Model()

				h := NewHerbrandModel(s, slv, model, nil)
				result := h.EvalConstant(xConst)
				if result.Name != expected {
					t.Errorf("expected %q, got %q", expected, result.Name)
				}
			})
		}
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// ============================================================
// Part 2: Flag-gated enumerated_to_numeral
// ============================================================

// TestEnumeratedToNumeral_FlagOff verifies the error stub is preserved when
// EnableInterpretedEnums is false (the default). This matches Python's behavior.
func TestEnumeratedToNumeral_FlagOff(t *testing.T) {
	if EnableInterpretedEnums {
		t.Skip("EnableInterpretedEnums is true; this test checks the disabled path")
	}

	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Constructors["red"] = true
	sig.Constructors["green"] = true
	sig.Constructors["blue"] = true
	sig.AddSort(es)
	sig.Interp["color"] = "int" // interpret color as int

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()

	redConst := lg.NewConst("red", es)
	_, err := tr.Translate(redConst)
	if err == nil {
		t.Fatal("expected error for interpreted enum with flag off")
	}
	if !strings.Contains(err.Error(), "cannot interpret enumerated type") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestEnumeratedToNumeralZ3_Int directly tests the enumeratedToNumeralZ3 helper
// with int interpretation, bypassing the flag check.
func TestEnumeratedToNumeralZ3_Int(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Interp["color"] = "int"

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()
	ctx := s.Context()

	for i, name := range es.Extension {
		result, err := tr.enumeratedToNumeralZ3(i, es)
		if err != nil {
			t.Fatalf("enumeratedToNumeralZ3(%d/%s): %v", i, name, err)
		}
		expected := ctx.IntVal(int64(i))
		slv := ctx.NewZ3Solver()
		slv.Assert(ctx.Not(ctx.Eq(result, expected)))
		if slv.Check() != Unsat {
			t.Errorf("enum %q ordinal %d: expected IntVal(%d), got %s", name, i, i, result.String())
		}
	}
}

// TestEnumeratedToNumeralZ3_Nat tests enum interpreted as "nat".
func TestEnumeratedToNumeralZ3_Nat(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Interp["color"] = "nat"

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()
	ctx := s.Context()

	result, err := tr.enumeratedToNumeralZ3(2, es)
	if err != nil {
		t.Fatalf("enumeratedToNumeralZ3(2/blue): %v", err)
	}
	expected := ctx.IntVal(2)
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(result, expected)))
	if slv.Check() != Unsat {
		t.Errorf("blue ordinal 2: expected IntVal(2), got %s", result.String())
	}
}

// TestEnumeratedToNumeralZ3_Bv tests enum interpreted as bitvector.
func TestEnumeratedToNumeralZ3_Bv(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Interp["color"] = "bv[2]"

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()
	ctx := s.Context()

	// red=0, green=1, blue=2 should all fit in bv[2] (4 values)
	for i, name := range es.Extension {
		result, err := tr.enumeratedToNumeralZ3(i, es)
		if err != nil {
			t.Fatalf("enumeratedToNumeralZ3(%d/%s): %v", i, name, err)
		}
		expected := ctx.BvVal(int64(i), 2)
		slv := ctx.NewZ3Solver()
		slv.Assert(ctx.Not(ctx.Eq(result, expected)))
		if slv.Check() != Unsat {
			t.Errorf("enum %q: expected BvVal(%d,2), got %s", name, i, result.String())
		}
	}
}

// TestEnumeratedToNumeralZ3_BvOverflow verifies that an enum with too many
// values for the bitvector width produces an error.
func TestEnumeratedToNumeralZ3_BvOverflow(t *testing.T) {
	es := &lg.EnumeratedSort{
		Name:      "big",
		Extension: []string{"a", "b", "c", "d", "e"},
	}
	sig := il.NewSig()
	sig.Interp["big"] = "bv[2]" // bv[2] holds 0..3, enum has 5 values

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()

	// ordinal 4 (element "e") exceeds bv[2] capacity
	_, err := tr.enumeratedToNumeralZ3(4, es)
	if err == nil {
		t.Fatal("expected overflow error for ordinal 4 in bv[2]")
	}
	if !strings.Contains(err.Error(), "exceeds bv[2] capacity") {
		t.Errorf("unexpected error: %v", err)
	}

	// ordinal 3 (element "d") should still work (3 < 4)
	_, err = tr.enumeratedToNumeralZ3(3, es)
	if err != nil {
		t.Fatalf("ordinal 3 should fit in bv[2]: %v", err)
	}
}

// TestEnumeratedToNumeralZ3_RangeSort tests enum interpreted as a RangeSort.
func TestEnumeratedToNumeralZ3_RangeSort(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Interp["color"] = &lg.RangeSort{
		Name: "color_range",
		Lb:   lg.NumeralBound{Value: "0"},
		Ub:   lg.NumeralBound{Value: "5"},
	}

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()
	ctx := s.Context()

	result, err := tr.enumeratedToNumeralZ3(1, es)
	if err != nil {
		t.Fatalf("enumeratedToNumeralZ3(1/green): %v", err)
	}
	expected := ctx.IntVal(1)
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(result, expected)))
	if slv.Check() != Unsat {
		t.Errorf("green ordinal 1: expected IntVal(1), got %s", result.String())
	}
}

// TestEnumeratedToNumeralZ3_UnsupportedInterp verifies error for unsupported
// interpretation types.
func TestEnumeratedToNumeralZ3_UnsupportedInterp(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Interp["color"] = "float" // not a supported native sort

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()

	_, err := tr.enumeratedToNumeralZ3(0, es)
	if err == nil {
		t.Fatal("expected error for unsupported interpretation")
	}
	if !strings.Contains(err.Error(), "cannot interpret enum") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestEnumeratedToNumeralZ3_BadBvWidth verifies error for malformed bv width.
func TestEnumeratedToNumeralZ3_BadBvWidth(t *testing.T) {
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	sig := il.NewSig()
	sig.Interp["color"] = "bv[abc]" // bad width

	s := NewSolverFromSig(sig, nil)
	tr := s.Translator()

	_, err := tr.enumeratedToNumeralZ3(0, es)
	if err == nil {
		t.Fatal("expected error for bad bv width")
	}
	if !strings.Contains(err.Error(), "bad bv width") {
		t.Errorf("unexpected error: %v", err)
	}
}
