package goivy

import "testing"

func TestParserNativeQuoteTextUsesNativeCode(t *testing.T) {
	typeSrc := "type idx = {0..3}\ntype vec\ninterpret vec -> <<< primitive std::vector<`idx`> >>>"
	typeResult, err := Parse(typeSrc, Version{1, 7})
	if err != nil {
		t.Fatalf("parse native type quote: %v", err)
	}
	nt := findParsedNativeType(t, typeResult)
	assertNativeQuoteHead(t, "NativeType", nt.Args())

	exprSrc := "type t\nindividual x:t\ndefinition value(a:t) = <<< `a` >>>"
	exprResult, err := Parse(exprSrc, Version{1, 7})
	if err != nil {
		t.Fatalf("parse native expr quote: %v", err)
	}
	ne := findParsedNativeExpr(t, exprResult)
	assertNativeQuoteHead(t, "NativeExpr", ne.Args())
}

func findParsedNativeType(t *testing.T, result *ParseResult) *NativeType {
	t.Helper()
	for _, decl := range result.Decls {
		id, ok := decl.(*InterpretDecl)
		if !ok || len(id.DeclArgs) == 0 {
			continue
		}
		lf, ok := id.DeclArgs[0].(*LabeledFormula)
		if !ok {
			continue
		}
		imp, ok := lf.Formula.(*Implies)
		if !ok {
			continue
		}
		if nt, ok := imp.T2.(*NativeType); ok {
			return nt
		}
	}
	t.Fatalf("native type quote not found in decls")
	return nil
}

func findParsedNativeExpr(t *testing.T, result *ParseResult) *NativeExpr {
	t.Helper()
	for _, decl := range result.Decls {
		dd, ok := decl.(*DefinitionDecl)
		if !ok || len(dd.DeclArgs) == 0 {
			continue
		}
		lf, ok := dd.DeclArgs[0].(*LabeledFormula)
		if !ok {
			continue
		}
		def, ok := lf.Formula.(*Definition)
		if !ok {
			continue
		}
		if ne, ok := def.Rhs.(*NativeExpr); ok {
			return ne
		}
	}
	t.Fatalf("native expr quote not found in decls")
	return nil
}

func assertNativeQuoteHead(t *testing.T, label string, args []Node) {
	t.Helper()
	if len(args) == 0 {
		t.Fatalf("%s has no args", label)
	}
	if _, ok := args[0].(*NativeCode); !ok {
		t.Fatalf("%s arg[0] = %T, want *NativeCode", label, args[0])
	}
}
