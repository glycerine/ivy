package goivy

import "testing"

func TestParseV16TopLevelInit(t *testing.T) {
	result, err := Parse("type t\nindividual x:t\ninit x = x", Version{1, 6}, WithFilename("init16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 init: %v", err)
	}
	if got := countDeclsOf[*InitDecl](result.Decls); got != 1 {
		t.Fatalf("InitDecl count = %d, want 1", got)
	}
}

func TestParseV16UnlabeledAxiomDoesNotSynthesizeLabel(t *testing.T) {
	result, err := Parse("type t\nrelation r(X:t)\naxiom forall X:t . r(X)", Version{1, 6}, WithFilename("axiom16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 axiom: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*AxiomDecl](t, result.Decls)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 unlabeled axiom synthesized label %q, want empty", lf.LabelName())
	}
}

func TestParseV16UnlabeledPropertyDoesNotSynthesizeLabel(t *testing.T) {
	result, err := Parse("type t\nrelation r(X:t)\nproperty forall X:t . r(X)", Version{1, 6}, WithFilename("prop16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 property: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*PropertyDecl](t, result.Decls)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 unlabeled property synthesized label %q, want empty", lf.LabelName())
	}
}

func TestParseV17UnlabeledAxiomStillSynthesizesLabel(t *testing.T) {
	result, err := Parse("type t\nrelation r(X:t)\naxiom forall X:t . r(X)", Version{1, 7}, WithFilename("axiom17.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.7 axiom: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*AxiomDecl](t, result.Decls)
	if lf.LabelName() == "" {
		t.Fatal("v1.7 unlabeled axiom should still synthesize a label")
	}
}

func TestParseV16VarDeclaresConstant(t *testing.T) {
	result, err := Parse("type t\nvar x:t", Version{1, 6}, WithFilename("var16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 var: %v", err)
	}
	decl := firstDeclOf[*ConstantDecl](t, result.Decls)
	arg := firstArgAs[*Atom](t, decl)
	if arg.Rep != "x" {
		t.Fatalf("var declared %q, want x", arg.Rep)
	}
	if got := arg.ASort.String(); got != "t" {
		t.Fatalf("var sort = %q, want t", got)
	}
}

func TestParseV16FunctionDeclaration(t *testing.T) {
	result, err := Parse("type t\nfunction f(X:t):t", Version{1, 6}, WithFilename("function16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 function: %v", err)
	}
	decl := firstDeclOf[*ConstantDecl](t, result.Decls)
	arg := firstArgAs[*Atom](t, decl)
	if arg.Rep != "f" {
		t.Fatalf("function declared %q, want f", arg.Rep)
	}
	if len(arg.Terms) != 1 {
		t.Fatalf("function argument count = %d, want 1", len(arg.Terms))
	}
	if got := arg.ASort.String(); got != "t" {
		t.Fatalf("function result sort = %q, want t", got)
	}
}

func TestParseV16FunctionDefinitionIsDerivedWithoutSyntheticLabel(t *testing.T) {
	result, err := Parse("type t\nfunction id(X:t):t = X", Version{1, 6}, WithFilename("function_def16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 function definition: %v", err)
	}
	decl := firstDeclOf[*DerivedDecl](t, result.Decls)
	lf := firstArgAs[*LabeledFormula](t, decl)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 function definition synthesized label %q, want empty", lf.LabelName())
	}
	if _, ok := lf.Formula.(*Definition); !ok {
		t.Fatalf("derived formula = %T, want *Definition", lf.Formula)
	}
}

func TestParseV16RelationDefinitionIsDerivedWithoutSyntheticLabel(t *testing.T) {
	result, err := Parse("type t\nrelation p(X:t) = true", Version{1, 6}, WithFilename("relation_def16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 relation definition: %v", err)
	}
	decl := firstDeclOf[*DerivedDecl](t, result.Decls)
	lf := firstArgAs[*LabeledFormula](t, decl)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 relation definition synthesized label %q, want empty", lf.LabelName())
	}
	if _, ok := lf.Formula.(*Definition); !ok {
		t.Fatalf("derived formula = %T, want *Definition", lf.Formula)
	}
}

func TestParseV16DerivedDeclWithoutSyntheticLabels(t *testing.T) {
	result, err := Parse("type t\nderived id(X:t):t = X, same(X:t) = X = X", Version{1, 6}, WithFilename("derived16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 derived: %v", err)
	}
	decl := firstDeclOf[*DerivedDecl](t, result.Decls)
	if got := len(decl.Args()); got != 2 {
		t.Fatalf("DerivedDecl args = %d, want 2", got)
	}
	for i, arg := range decl.Args() {
		lf, ok := arg.(*LabeledFormula)
		if !ok {
			t.Fatalf("DerivedDecl arg[%d] = %T, want *LabeledFormula", i, arg)
		}
		if lf.LabelName() != "" {
			t.Fatalf("v1.6 derived arg[%d] synthesized label %q, want empty", i, lf.LabelName())
		}
		if _, ok := lf.Formula.(*Definition); !ok {
			t.Fatalf("DerivedDecl arg[%d] formula = %T, want *Definition", i, lf.Formula)
		}
	}
}

func TestParseV16MacroDecl(t *testing.T) {
	result, err := Parse("macro step = { call ping }", Version{1, 6}, WithFilename("macro16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 macro: %v", err)
	}
	decl := firstDeclOf[*MacroDecl](t, result.Decls)
	def := firstArgAs[*Definition](t, decl)
	lhs, ok := def.Lhs.(*Atom)
	if !ok {
		t.Fatalf("MacroDecl lhs = %T, want *Atom", def.Lhs)
	}
	if lhs.Rep != "step" {
		t.Fatalf("MacroDecl lhs = %q, want step", lhs.Rep)
	}
	if _, ok := def.Rhs.(*CallAction); !ok {
		t.Fatalf("MacroDecl rhs = %T, want *CallAction", def.Rhs)
	}
}

func TestParseV16AliasDecl(t *testing.T) {
	result, err := Parse("alias short = target.step", Version{1, 6}, WithFilename("alias16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 alias: %v", err)
	}
	decl := firstDeclOf[*AliasDecl](t, result.Decls)
	def := firstArgAs[*Definition](t, decl)
	lhs, ok := def.Lhs.(*Atom)
	if !ok {
		t.Fatalf("AliasDecl lhs = %T, want *Atom", def.Lhs)
	}
	if lhs.Rep != "short" {
		t.Fatalf("AliasDecl lhs = %q, want short", lhs.Rep)
	}
	rhs, ok := def.Rhs.(*Atom)
	if !ok {
		t.Fatalf("AliasDecl rhs = %T, want *Atom", def.Rhs)
	}
	if rhs.Rep != "target.step" {
		t.Fatalf("AliasDecl rhs = %q, want target.step", rhs.Rep)
	}
}

func TestParseV16RelyDecls(t *testing.T) {
	result, err := Parse("rely p -> q\nrely stable", Version{1, 6}, WithFilename("rely16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 rely: %v", err)
	}
	if got := countDeclsOf[*RelyDecl](result.Decls); got != 2 {
		t.Fatalf("RelyDecl count = %d, want 2", got)
	}
	first := firstDeclOf[*RelyDecl](t, result.Decls)
	if _, ok := first.Args()[0].(*Implies); !ok {
		t.Fatalf("first RelyDecl arg = %T, want *Implies", first.Args()[0])
	}
}

func TestParseV16MixOrdDecl(t *testing.T) {
	result, err := Parse("mixord before -> after", Version{1, 6}, WithFilename("mixord16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 mixord: %v", err)
	}
	decl := firstDeclOf[*MixOrdDecl](t, result.Decls)
	if _, ok := decl.Args()[0].(*Implies); !ok {
		t.Fatalf("MixOrdDecl arg = %T, want *Implies", decl.Args()[0])
	}
}

func TestParseV16ProgressDecl(t *testing.T) {
	result, err := Parse("type t\nprogress done(X:t) = true", Version{1, 6}, WithFilename("progress16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 progress: %v", err)
	}
	decl := firstDeclOf[*ProgressDecl](t, result.Decls)
	def := firstArgAs[*Definition](t, decl)
	lhs, ok := def.Lhs.(*Atom)
	if !ok {
		t.Fatalf("ProgressDecl lhs = %T, want *Atom", def.Lhs)
	}
	if lhs.Rep != "done" {
		t.Fatalf("ProgressDecl lhs = %q, want done", lhs.Rep)
	}
}

func TestParseV16TopLevelAssert(t *testing.T) {
	result, err := Parse("relation p\nrelation q\nassert p -> q", Version{1, 6}, WithFilename("assert16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 assert: %v", err)
	}
	decl := firstDeclOf[*AssertDecl](t, result.Decls)
	if len(decl.Args()) != 1 {
		t.Fatalf("AssertDecl args = %d, want 1", len(decl.Args()))
	}
	if _, ok := decl.Args()[0].(*Implies); !ok {
		t.Fatalf("AssertDecl formula = %T, want *Implies", decl.Args()[0])
	}
}

func TestParseV16PrivateCallAtom(t *testing.T) {
	result, err := Parse("private hidden", Version{1, 6}, WithFilename("private16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 private: %v", err)
	}
	decl := firstDeclOf[*PrivateDecl](t, result.Decls)
	def := firstArgAs[*PrivateDef](t, decl)
	if got := def.Privatized(); got != "hidden" {
		t.Fatalf("private symbol = %q, want hidden", got)
	}
}

func TestParseV16DefinitionDeclWithoutSyntheticLabel(t *testing.T) {
	result, err := Parse("type t\ndefinition id(X:t):t = X", Version{1, 6}, WithFilename("definition16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 definition: %v", err)
	}
	decl := firstDeclOf[*DefinitionDecl](t, result.Decls)
	lf := firstArgAs[*LabeledFormula](t, decl)
	if lf.LabelName() != "" {
		t.Fatalf("v1.6 definition synthesized label %q, want empty", lf.LabelName())
	}
	if _, ok := lf.Formula.(*Definition); !ok {
		t.Fatalf("definition formula = %T, want *Definition", lf.Formula)
	}
}

func TestParseV16PropertyProofUsesBareSchemaInstantiation(t *testing.T) {
	result, err := Parse("property true proof intro", Version{1, 6}, WithFilename("proof16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 property proof: %v", err)
	}
	if got := countDeclsOf[*PropertyDecl](result.Decls); got != 1 {
		t.Fatalf("PropertyDecl count = %d, want 1", got)
	}
	proof := firstDeclOf[*ProofDecl](t, result.Decls)
	inst := firstArgAs[*SchemaInstantiation](t, proof)
	if inst.SchemaName.String() != "intro" {
		t.Fatalf("schema instantiation name = %q, want intro", inst.SchemaName.String())
	}
}

func TestParseV16PropertyProofWithMatches(t *testing.T) {
	result, err := Parse("property true proof intro with X = true", Version{1, 6}, WithFilename("proof_with16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 property proof with matches: %v", err)
	}
	proof := firstDeclOf[*ProofDecl](t, result.Decls)
	inst := firstArgAs[*SchemaInstantiation](t, proof)
	if len(inst.Matches) != 1 {
		t.Fatalf("schema instantiation matches = %d, want 1", len(inst.Matches))
	}
}

func firstLabeledFormulaInDecl[T Node](t *testing.T, decls []Node) *LabeledFormula {
	t.Helper()
	for _, decl := range decls {
		if typed, ok := decl.(T); ok {
			args := typed.Args()
			if len(args) == 0 {
				t.Fatalf("%T has no args", typed)
			}
			lf, ok := args[0].(*LabeledFormula)
			if !ok {
				t.Fatalf("%T arg[0] = %T, want *LabeledFormula", typed, args[0])
			}
			return lf
		}
	}
	t.Fatalf("no declaration of requested type found")
	return nil
}

func firstDeclOf[T Node](t *testing.T, decls []Node) T {
	t.Helper()
	for _, decl := range decls {
		if typed, ok := decl.(T); ok {
			return typed
		}
	}
	var zero T
	t.Fatalf("no declaration of requested type found")
	return zero
}

func firstArgAs[T Node](t *testing.T, node Node) T {
	t.Helper()
	args := node.Args()
	if len(args) == 0 {
		t.Fatalf("%T has no args", node)
	}
	typed, ok := args[0].(T)
	if !ok {
		t.Fatalf("%T arg[0] = %T, want requested type", node, args[0])
	}
	return typed
}
