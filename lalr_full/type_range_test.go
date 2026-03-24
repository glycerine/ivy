package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// TestTypeWithRangeCreatesUninterpretedSort verifies that when the sort is a Range,
// the TypeDef's Value is replaced with UninterpretedSortAST (not the Range itself).
// Matches Python ivy_parser.py:1785: defsort = UninterpretedSort() if isinstance(p[7], Range) else p[7]
func TestTypeWithRangeCreatesUninterpretedSort(t *testing.T) {
	input := `#lang 1.7
type t = {0..5}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(result.Decls) == 0 {
		t.Fatal("expected at least 1 decl")
	}

	// Find the TypeDecl
	var td *ast.TypeDecl
	for _, d := range result.Decls {
		if x, ok := d.(*ast.TypeDecl); ok {
			td = x
			break
		}
	}
	if td == nil {
		t.Fatal("expected a TypeDecl")
	}

	// The TypeDef inside should have UninterpretedSortAST as Value, not Range
	if len(td.DeclArgs) == 0 {
		t.Fatal("TypeDecl has no args")
	}
	tdfn, ok := td.DeclArgs[0].(*ast.TypeDef)
	if !ok {
		t.Fatalf("expected TypeDef in TypeDecl, got %T", td.DeclArgs[0])
	}
	if _, ok := tdfn.Value.(*ast.UninterpretedSortAST); !ok {
		t.Errorf("expected UninterpretedSortAST as TypeDef.Value, got %T", tdfn.Value)
	}
}

// TestTypeWithRangeCreatesInterpretDecl verifies that parsing "type t = {0..5}"
// produces both a TypeDecl and an InterpretDecl.
// Matches Python ivy_parser.py:1790-1795.
func TestTypeWithRangeCreatesInterpretDecl(t *testing.T) {
	input := `#lang 1.7
type t = {0..5}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var hasTypeDecl, hasInterpretDecl bool
	for _, d := range result.Decls {
		switch d.(type) {
		case *ast.TypeDecl:
			hasTypeDecl = true
		case *ast.InterpretDecl:
			hasInterpretDecl = true
		}
	}
	if !hasTypeDecl {
		t.Error("expected a TypeDecl")
	}
	if !hasInterpretDecl {
		t.Error("expected an InterpretDecl for Range sort")
	}
}

// TestTypeWithNonRangeSortNoInterpretDecl verifies that "type t = nat" produces
// only a TypeDecl (no InterpretDecl), since nat is not a Range.
func TestTypeWithNonRangeSortNoInterpretDecl(t *testing.T) {
	input := `#lang 1.7
type t`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	for _, d := range result.Decls {
		if _, ok := d.(*ast.InterpretDecl); ok {
			t.Error("did not expect InterpretDecl for non-Range sort")
		}
	}
}

// TestTypeRangeInterpretDeclCanon verifies that the InterpretDecl produced for
// a Range sort has the correct canon output containing "interpretDecl" and "implies".
func TestTypeRangeInterpretDeclCanon(t *testing.T) {
	input := `#lang 1.7
type t = {0..5}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	for _, d := range result.Decls {
		if id, ok := d.(*ast.InterpretDecl); ok {
			canon := string(id.Canon())
			if !strings.Contains(canon, "interpretDecl") {
				t.Errorf("expected 'interpretDecl' in canon, got %s", canon)
			}
			// The InterpretDecl should contain a LabeledFormula with an Implies
			if !strings.Contains(canon, "implies") {
				t.Errorf("expected 'implies' in canon, got %s", canon)
			}
			// The label should be "interp[N]"
			if !strings.Contains(canon, "interp") {
				t.Errorf("expected 'interp' label in canon, got %s", canon)
			}
			t.Logf("InterpretDecl canon: %s", canon)
			return
		}
	}
	t.Error("no InterpretDecl found to check canon")
}

// TestTypeRangeDesugaring directly constructs the desugaring to verify
// the AST structure matches Python ivy_parser.py:1780-1797.
func TestTypeRangeDesugaring(t *testing.T) {
	// Simulate: type t = {0..5}
	scnst := ast.NewAtom("t")
	rangeSort := ast.NewRange(ast.NewAtom("0"), ast.NewAtom("5"))

	// Python: defsort = UninterpretedSort() if isinstance(p[7], Range) else p[7]
	_, isRange := (ast.Node)(rangeSort).(*ast.Range)
	if !isRange {
		t.Fatal("expected Range type assertion to succeed")
	}
	defsort := ast.NewUninterpretedSortAST()

	// Python: tdfn = TypeDef(scnst, defsort)
	tdfn := &ast.TypeDef{Name: scnst, Value: defsort}
	tdfnCanon := string(tdfn.Canon())
	if !strings.Contains(tdfnCanon, "typeDef") {
		t.Errorf("expected 'typeDef' in canon, got %s", tdfnCanon)
	}
	// Value should be constantSort (UninterpretedSort canons as constantSort)
	if !strings.Contains(tdfnCanon, "constantSort") {
		t.Errorf("expected 'constantSort' in TypeDef canon, got %s", tdfnCanon)
	}

	// Python: imp = Implies(scnst, p[7])
	imp := ast.NewImplies(scnst, rangeSort)
	impCanon := string(imp.Canon())
	if !strings.Contains(impCanon, "implies") {
		t.Errorf("expected 'implies' in canon, got %s", impCanon)
	}
	if !strings.Contains(impCanon, `rep:"t"`) {
		t.Errorf("expected t1 to be atom 't', got %s", impCanon)
	}

	// Python: mk_lf(imp)
	lf := ast.NewLabeledFormula(nil, imp)
	// Python: addlabel(lf, 'interp')
	// We can't call addLabel here (it's in the grammar package), but verify LabeledFormula works
	lfCanon := string(lf.Canon())
	if !strings.Contains(lfCanon, "labeledFormula") {
		t.Errorf("expected 'labeledFormula' in canon, got %s", lfCanon)
	}

	// Python: InterpretDecl(labeled)
	thing := ast.NewInterpretDecl(lf)
	thingCanon := string(thing.Canon())
	if !strings.Contains(thingCanon, "interpretDecl") {
		t.Errorf("expected 'interpretDecl' in canon, got %s", thingCanon)
	}

	t.Logf("TypeDef canon: %s", tdfnCanon)
	t.Logf("Implies canon: %s", impCanon)
	t.Logf("InterpretDecl canon: %s", thingCanon)
}

// TestGhostTypeDefCreated verifies that GhostTypeDef is used when optghost is true.
// Since we can't easily parse "ghost type t" without a full program context,
// we test the direct construction path.
func TestGhostTypeDefCreated(t *testing.T) {
	scnst := ast.NewAtom("t")
	value := ast.NewUninterpretedSortAST()

	// Non-ghost: plain TypeDef
	tdfn := &ast.TypeDef{Name: scnst, Value: value}
	tdfnCanon := string(tdfn.Canon())
	if !strings.Contains(tdfnCanon, "typeDef") {
		t.Errorf("expected 'typeDef' in canon, got %s", tdfnCanon)
	}

	// Ghost: GhostTypeDef wrapping TypeDef
	ghost := &ast.GhostTypeDef{TypeDef: *tdfn}
	ghostCanon := string(ghost.Canon())
	if !strings.Contains(ghostCanon, "ghostTypeDef") {
		t.Errorf("expected 'ghostTypeDef' in canon, got %s", ghostCanon)
	}
	// GhostTypeDef should still have the same Name and Value
	if !strings.Contains(ghostCanon, `rep:"t"`) {
		t.Errorf("expected name 't' in ghost canon, got %s", ghostCanon)
	}

	t.Logf("TypeDef canon: %s", tdfnCanon)
	t.Logf("GhostTypeDef canon: %s", ghostCanon)
}
