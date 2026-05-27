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

func TestParseV16PropertyNamedDecl(t *testing.T) {
	result, err := Parse("property true named witness", Version{1, 6}, WithFilename("named16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 named property: %v", err)
	}
	if got := countDeclsOf[*PropertyDecl](result.Decls); got != 1 {
		t.Fatalf("PropertyDecl count = %d, want 1", got)
	}
	decl := firstDeclOf[*NamedDecl](t, result.Decls)
	arg := firstArgAs[*Atom](t, decl)
	if arg.Rep != "witness" {
		t.Fatalf("NamedDecl arg = %q, want witness", arg.Rep)
	}
}

func TestParseV16SchemaDeclWithPropertyConclusion(t *testing.T) {
	src := `schema congruence = {
type d
type r
function f(X:d):r
property X = Y -> f(X) = f(Y)
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("schema16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 schema: %v", err)
	}
	decl := firstDeclOf[*SchemaDecl](t, result.Decls)
	schema := firstArgAs[*Schema](t, decl)
	def := firstArgAs[*Definition](t, schema)
	if lhs, ok := def.Lhs.(*Atom); !ok || lhs.Rep != "congruence" {
		t.Fatalf("schema lhs = %T %v, want Atom(congruence)", def.Lhs, def.Lhs)
	}
	body, ok := def.Rhs.(*SchemaBody)
	if !ok {
		t.Fatalf("schema rhs = %T, want *SchemaBody", def.Rhs)
	}
	if got := len(body.Args()); got != 4 {
		t.Fatalf("SchemaBody args = %d, want 4", got)
	}
	if _, ok := body.Conc().(*Implies); !ok {
		t.Fatalf("SchemaBody conclusion = %T, want *Implies", body.Conc())
	}
}

func TestParseV16SchemaPremisePropertyKeepsNoSyntheticLabel(t *testing.T) {
	src := `schema two_props = {
property p
property q
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("schema_props16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 schema with property premise: %v", err)
	}
	decl := firstDeclOf[*SchemaDecl](t, result.Decls)
	schema := firstArgAs[*Schema](t, decl)
	def := firstArgAs[*Definition](t, schema)
	body, ok := def.Rhs.(*SchemaBody)
	if !ok {
		t.Fatalf("schema rhs = %T, want *SchemaBody", def.Rhs)
	}
	if got := len(body.Args()); got != 2 {
		t.Fatalf("SchemaBody args = %d, want 2", got)
	}
	prem, ok := body.Args()[0].(*LabeledFormula)
	if !ok {
		t.Fatalf("SchemaBody premise = %T, want *LabeledFormula", body.Args()[0])
	}
	if prem.LabelName() != "" {
		t.Fatalf("v1.6 schema premise synthesized label %q, want empty", prem.LabelName())
	}
}

func TestParseV16ObjectDeclExpandsWithPrefix(t *testing.T) {
	src := `object obj = {
type t
relation p(X:t)
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("object16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 object: %v", err)
	}
	decl := firstDeclOf[*ObjectDecl](t, result.Decls)
	name := firstArgAs[*Atom](t, decl)
	if name.Rep != "obj" {
		t.Fatalf("ObjectDecl name = %q, want obj", name.Rep)
	}
	var foundPrefixedRelation bool
	for _, d := range result.Decls {
		cd, ok := d.(*ConstantDecl)
		if !ok || len(cd.Args()) == 0 {
			continue
		}
		if atom, ok := cd.Args()[0].(*Atom); ok && atom.Rep == "obj.p" {
			foundPrefixedRelation = true
			break
		}
	}
	if !foundPrefixedRelation {
		t.Fatalf("expanded object relation obj.p not found in declarations")
	}
}

func TestParseV16TypeDefinitions(t *testing.T) {
	src := `type color = {red, green}
type idx = {0 .. 3}
type pair = struct { first:color, second:idx }`
	result, err := Parse(src, Version{1, 6}, WithFilename("types16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 type definitions: %v", err)
	}
	if got := countDeclsOf[*TypeDecl](result.Decls); got != 3 {
		t.Fatalf("TypeDecl count = %d, want 3", got)
	}
	if got := countDeclsOf[*InterpretDecl](result.Decls); got != 1 {
		t.Fatalf("range type InterpretDecl count = %d, want 1", got)
	}
	var sawEnum, sawStruct bool
	for _, decl := range result.Decls {
		td, ok := decl.(*TypeDecl)
		if !ok || len(td.Args()) == 0 {
			continue
		}
		def, ok := td.Args()[0].(*TypeDef)
		if !ok {
			continue
		}
		switch def.Value.(type) {
		case *EnumeratedSort:
			sawEnum = true
		case *StructSort:
			sawStruct = true
		}
	}
	if !sawEnum {
		t.Fatalf("enumerated type definition not found")
	}
	if !sawStruct {
		t.Fatalf("struct type definition not found")
	}
}

func TestParseV16ActionReturnsAndCoreStatements(t *testing.T) {
	src := `type t
individual x:t
action read returns (out:t) = {
    assume true;
    out := x;
    assert out = x;
    ensures out = x
}
action havoc = {
    x := *
}
action branch = {
    if true {
        call read
    } else {
        call havoc
    }
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("actions16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 action syntax: %v", err)
	}
	if got := countDeclsOf[*ActionDecl](result.Decls); got != 3 {
		t.Fatalf("ActionDecl count = %d, want 3", got)
	}
	read := firstActionDefNamed(t, result.Decls, "read")
	if got := len(read.FormalReturns); got != 1 {
		t.Fatalf("read formal returns = %d, want 1", got)
	}
	seq, ok := read.Body.(*Sequence)
	if !ok {
		t.Fatalf("read body = %T, want *Sequence", read.Body)
	}
	if got := len(seq.Stmts); got != 4 {
		t.Fatalf("read body statement count = %d, want 4", got)
	}
	if _, ok := seq.Stmts[0].(*AssumeAction); !ok {
		t.Fatalf("read stmt[0] = %T, want *AssumeAction", seq.Stmts[0])
	}
	if _, ok := seq.Stmts[2].(*AssertAction); !ok {
		t.Fatalf("read stmt[2] = %T, want *AssertAction", seq.Stmts[2])
	}
	if _, ok := seq.Stmts[3].(*EnsuresAction); !ok {
		t.Fatalf("read stmt[3] = %T, want *EnsuresAction", seq.Stmts[3])
	}
	havoc := firstActionDefNamed(t, result.Decls, "havoc")
	if _, ok := havoc.Body.(*HavocAction); !ok {
		t.Fatalf("havoc body = %T, want *HavocAction", havoc.Body)
	}
	branch := firstActionDefNamed(t, result.Decls, "branch")
	if _, ok := branch.Body.(*IfAction); !ok {
		t.Fatalf("branch body = %T, want *IfAction", branch.Body)
	}
}

func TestParseV16AdvancedActionSyntax(t *testing.T) {
	src := `type t
individual a:t
relation p(X:t)
action helper(x:t)
action advanced = {
    var z:t := a;
    local w:t {
        w := z
    };
    if * {
        call helper
    } else {
        helper(a)
    };
    if some x:t . p(x) {
        instantiate witness
    };
    while p(z) invariant p(z) {
        z := a
    };
    let m = n {
        helper(a)
    };
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("advanced_actions16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 advanced action syntax: %v", err)
	}
	advanced := firstActionDefNamed(t, result.Decls, "advanced")
	seq, ok := advanced.Body.(*Sequence)
	if !ok {
		t.Fatalf("advanced body = %T, want *Sequence", advanced.Body)
	}
	want := []struct {
		name string
		ok   func(Node) bool
	}{
		{"local lowered from var", func(n Node) bool { _, ok := n.(*LocalAction); return ok }},
		{"choice", func(n Node) bool { _, ok := n.(*ChoiceAction); return ok }},
		{"if some", func(n Node) bool {
			ifa, ok := n.(*IfAction)
			if !ok {
				return false
			}
			_, isSome := ifa.Cond.(*Some)
			return isSome
		}},
		{"while", func(n Node) bool { _, ok := n.(*WhileAction); return ok }},
		{"let", func(n Node) bool { _, ok := n.(*LetAction); return ok }},
	}
	if got := len(seq.Stmts); got != len(want) {
		t.Fatalf("advanced body statement count = %d, want %d", got, len(want))
	}
	for i, w := range want {
		if !w.ok(seq.Stmts[i]) {
			t.Fatalf("advanced stmt[%d] = %T, want %s", i, seq.Stmts[i], w.name)
		}
	}
	ifa := seq.Stmts[2].(*IfAction)
	then, ok := ifa.Then.(*InstantiateAction)
	if !ok {
		t.Fatalf("if-some then = %T, want *InstantiateAction", ifa.Then)
	}
	if got := NodeRep(then.Args()[0]); got != "witness" {
		t.Fatalf("instantiate target = %q, want witness", got)
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

func firstActionDefNamed(t *testing.T, decls []Node, name string) *ActionDef {
	t.Helper()
	for _, decl := range decls {
		ad, ok := decl.(*ActionDecl)
		if !ok {
			continue
		}
		for _, arg := range ad.Args() {
			def, ok := arg.(*ActionDef)
			if !ok {
				continue
			}
			if atom, ok := def.Name.(*Atom); ok && atom.Rep == name {
				return def
			}
		}
	}
	t.Fatalf("no action %q found", name)
	return nil
}
