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

func TestParseV11StateDeclsAndCompilerPredicates(t *testing.T) {
	src := `state idle = entry
state visible = true | false
state guarded = {requires true modifies {} ensures false}`
	result, err := Parse(src, Version{1, 1}, WithFilename("state11.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.1 state declarations: %v", err)
	}
	if got := countDeclsOf[*StateDecl](result.Decls); got != 3 {
		t.Fatalf("StateDecl count = %d, want 3", got)
	}
	state := firstArgAs[*StateDef](t, firstDeclOf[*StateDecl](t, result.Decls))
	if state.Name != "idle" {
		t.Fatalf("first state name = %q, want idle", state.Name)
	}
	if _, ok := state.State.(*RME); !ok {
		t.Fatalf("entry state expression = %T, want *RME", state.State)
	}

	mod, err := IvyFromString("#lang ivy1.1\n" + src)
	if err != nil {
		t.Fatalf("Compile v1.1 state declarations: %v", err)
	}
	for _, name := range []string{"idle", "visible", "guarded"} {
		if mod.Predicates[name] == nil {
			t.Fatalf("compiled module missing predicate %q; predicates = %#v", name, mod.Predicates)
		}
	}
}

func TestParseV16StateKeywordMatchesPythonLexer(t *testing.T) {
	_, err := Parse("state idle = entry", Version{1, 6}, WithFilename("state16.ivy"))
	if err == nil {
		t.Fatal("Parse v1.6 accepted state declaration; Python v1.6 lexer treats state as a symbol")
	}
	if pe, ok := err.(*ParseError); !ok || pe.Token != "state" {
		t.Fatalf("v1.6 state error = %v, want token state", err)
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

func TestParseV16AndV17LabelsAcceptPythonSymbolSubscripts(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version Version
	}{
		{name: "v16", version: Version{1, 6}},
		{name: "v17", version: Version{1, 7}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse("type t\naxiom [lbl[this.part]] true", tc.version, WithFilename(tc.name+"_subscript_label.ivy"))
			if err != nil {
				t.Fatalf("Parse %s subscript label: %v", tc.name, err)
			}
			lf := firstLabeledFormulaInDecl[*AxiomDecl](t, result.Decls)
			if got := lf.LabelName(); got != "lbl[this.part]" {
				t.Fatalf("%s label = %q, want lbl[this.part]", tc.name, got)
			}
		})
	}
}

func TestParseV16UsingImportsWithPrefix(t *testing.T) {
	version := Version{1, 6}
	var calls []string
	var importer ImporterFunc
	importer = func(name string, parent *ivyAccum) (*ParseResult, error) {
		calls = append(calls, name)
		if name != "util" {
			t.Fatalf("unexpected import %q", name)
		}
		return Parse("type t\nvar x:t\naction ping = {}", version,
			WithImporter(importer),
			WithParentAccum(parent),
			WithNested(),
			WithAstConfig(parent.astCfg),
			WithFilename("util.ivy"),
		)
	}

	result, err := Parse("using util", version, WithImporter(importer), WithFilename("main_using16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 using: %v", err)
	}
	if len(calls) != 1 || calls[0] != "util" {
		t.Fatalf("using importer calls = %v, want [util]", calls)
	}
	var sawVar, sawAction bool
	for _, decl := range result.Decls {
		switch d := decl.(type) {
		case *ConstantDecl:
			if len(d.Args()) > 0 && NodeRep(d.Args()[0]) == "util.x" {
				sawVar = true
			}
		case *ActionDecl:
			if len(d.Args()) > 0 {
				if def, ok := d.Args()[0].(*ActionDef); ok && NodeRep(def.Name) == "util.ping" {
					sawAction = true
				}
			}
		}
	}
	if !sawVar {
		t.Fatalf("using did not declare prefixed util.x; decls = %v", result.Decls)
	}
	if !sawAction {
		t.Fatalf("using did not declare prefixed util.ping; decls = %v", result.Decls)
	}
}

func TestParseV16VarDeclaresConstant(t *testing.T) {
	result, err := Parse("type t\nvar x:t", Version{1, 6}, WithFilename("var16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 var: %v", err)
	}
	decl := firstDeclOf[*ConstantDecl](t, result.Decls)
	arg := firstArg(t, decl)
	if got := NodeRep(arg); got != "x" {
		t.Fatalf("var declared %q, want x", got)
	}
	if got := nodeSortString(arg); got != "t" {
		t.Fatalf("var sort = %q, want t", got)
	}
}

func TestParseV16FunctionDeclaration(t *testing.T) {
	result, err := Parse("type t\nfunction f(X:t):t", Version{1, 6}, WithFilename("function16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 function: %v", err)
	}
	decl := firstDeclOf[*ConstantDecl](t, result.Decls)
	arg := firstArg(t, decl)
	if got := NodeRep(arg); got != "f" {
		t.Fatalf("function declared %q, want f", got)
	}
	if got := len(arg.Args()); got != 1 {
		t.Fatalf("function argument count = %d, want 1", got)
	}
	if got := nodeSortString(arg); got != "t" {
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
	result, err := Parse("mixord pre -> post", Version{1, 6}, WithFilename("mixord16.ivy"))
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

func TestParseV16PropertyNamedOperatorAndLabeledProof(t *testing.T) {
	src := "property true named (X + Y) proof [pf] intro"
	result, err := Parse(src, Version{1, 6}, WithFilename("named_operator_proof16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 named operator/labeled proof: %v", err)
	}
	named := firstDeclOf[*NamedDecl](t, result.Decls)
	if got := NodeRep(firstArg(t, named)); got != "+" {
		t.Fatalf("NamedDecl operator = %q, want +", got)
	}
	proof := firstDeclOf[*ProofDecl](t, result.Decls)
	lf := firstArgAs[*LabeledFormula](t, proof)
	if lf.LabelName() != "pf" {
		t.Fatalf("proof label = %q, want pf", lf.LabelName())
	}
	if _, ok := lf.Formula.(*SchemaInstantiation); !ok {
		t.Fatalf("proof formula = %T, want *SchemaInstantiation", lf.Formula)
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
	local, ok := advanced.Body.(*LocalAction)
	if !ok {
		t.Fatalf("advanced body = %T, want *LocalAction lowered from leading var", advanced.Body)
	}
	if got := len(local.Args()); got != 2 {
		t.Fatalf("lowered var LocalAction args = %d, want 2", got)
	}
	seq, ok := local.Args()[1].(*Sequence)
	if !ok {
		t.Fatalf("lowered var body = %T, want *Sequence", local.Args()[1])
	}
	want := []struct {
		name string
		ok   func(Node) bool
	}{
		{"explicit local", func(n Node) bool { _, ok := n.(*LocalAction); return ok }},
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

func TestParseV16SharedTopLevelDeclarationSyntax(t *testing.T) {
	src := `temporal axiom true
method meth = {}
action step = {}
before step { call step }
after step { call step }
implement step { call step }
mixin step before step
trusted isolate iso = step with step
delegate step -> target
interpret sort_a -> sort_b
attribute step = true
variant child of sort_a = {left, right}`
	result, err := Parse(src, Version{1, 6}, WithFilename("shared_top16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 shared top-level syntax: %v", err)
	}
	lf := firstLabeledFormulaInDecl[*AxiomDecl](t, result.Decls)
	if !lf.IsTemporal() {
		t.Fatalf("temporal axiom was not marked temporal")
	}
	if got := countDeclsOf[*ActionDecl](result.Decls); got != 5 {
		t.Fatalf("ActionDecl count = %d, want 5", got)
	}
	if got := countDeclsOf[*MixinDecl](result.Decls); got != 4 {
		t.Fatalf("MixinDecl count = %d, want 4", got)
	}
	if got := countDeclsOf[*IsolateDecl](result.Decls); got != 1 {
		t.Fatalf("IsolateDecl count = %d, want 1", got)
	}
	if got := countDeclsOf[*DelegateDecl](result.Decls); got != 1 {
		t.Fatalf("DelegateDecl count = %d, want 1", got)
	}
	if got := countDeclsOf[*InterpretDecl](result.Decls); got != 1 {
		t.Fatalf("InterpretDecl count = %d, want 1", got)
	}
	if got := countDeclsOf[*AttributeDecl](result.Decls); got != 1 {
		t.Fatalf("AttributeDecl count = %d, want 1", got)
	}
	if got := countDeclsOf[*VariantDecl](result.Decls); got != 1 {
		t.Fatalf("VariantDecl count = %d, want 1", got)
	}
	for _, decl := range result.Decls {
		md, ok := decl.(*MixinDecl)
		if !ok || len(md.Args()) == 0 {
			continue
		}
		if def, ok := md.Args()[0].(*MixinBeforeDef); ok {
			if def.Mixer() != "step[before]" {
				t.Fatalf("v1.6 before mixin name = %q, want step[before]", def.Mixer())
			}
			return
		}
	}
	t.Fatalf("before MixinDecl not found")
}

func TestParseV16DelegateGeneratedMixinNameSubscripts(t *testing.T) {
	src := `action step = {}
before step { call step }
after step { call step }
delegate step[before], step[after] -> target`
	result, err := Parse(src, Version{1, 6}, WithFilename("delegate_mixin_subscript16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 delegate generated mixin names: %v", err)
	}
	decl := firstDeclOf[*DelegateDecl](t, result.Decls)
	if got := len(decl.Args()); got != 2 {
		t.Fatalf("DelegateDecl args = %d, want 2", got)
	}
	for i, want := range []string{"step[before]", "step[after]"} {
		def, ok := decl.Args()[i].(*DelegateDef)
		if !ok {
			t.Fatalf("DelegateDecl arg[%d] = %T, want *DelegateDef", i, decl.Args()[i])
		}
		if got := def.Delegated(); got != want {
			t.Fatalf("DelegateDef[%d].Delegated() = %q, want %q", i, got, want)
		}
		if got := def.Delegee(); got != "target" {
			t.Fatalf("DelegateDef[%d].Delegee() = %q, want target", i, got)
		}
	}
}

func TestParseV16GeneratedClassNameSubscript(t *testing.T) {
	src := `object marcelo[class](self:marcelo[class].t) = {
type t
}
alias marcelo_alias = marcelo[class].t`
	if _, err := Parse(src, Version{1, 6}, WithFilename("class_subscript16.ivy")); err != nil {
		t.Fatalf("Parse v1.6 generated class name subscript: %v", err)
	}
}

func TestParseV16ObjectBodyDoesNotRedefineObjectName(t *testing.T) {
	src := `type money
object account = {
individual balance : money
init balance = 0
}`
	if _, err := Parse(src, Version{1, 6}, WithFilename("object_redef16.ivy")); err != nil {
		t.Fatalf("Parse v1.6 object with body declarations: %v", err)
	}
}

func TestParseV16ObjectInterpretDoesNotRedefineObjectName(t *testing.T) {
	src := `object packet = {
type t
interpret t -> bv[1]
}`
	if _, err := Parse(src, Version{1, 6}, WithFilename("object_interpret16.ivy")); err != nil {
		t.Fatalf("Parse v1.6 object with interpret declaration: %v", err)
	}
}

func TestParseV16LabeledDeclLabelsDoNotDefineNames(t *testing.T) {
	src := `axiom [same] true
property [same] true
conjecture [same] true`
	if _, err := Parse(src, Version{1, 6}, WithFilename("label_defs16.ivy")); err != nil {
		t.Fatalf("Parse v1.6 repeated labeled declarations: %v", err)
	}
}

func TestParseV17LabeledDeclLabelsStillDefineNames(t *testing.T) {
	src := `axiom [same] true
property [same] true`
	if _, err := Parse(src, Version{1, 7}, WithFilename("label_defs17.ivy")); err == nil {
		t.Fatal("Parse v1.7 repeated labeled declarations succeeded, want redefinition error")
	}
}

func TestParseV16ModuleClassAndRMEAssertSyntax(t *testing.T) {
	src := `module m = {
type inner
}
class c = {
relation p(X:this)
}
relation req
assert req -> { requires true modifies * ensures true }`
	result, err := Parse(src, Version{1, 6}, WithFilename("module_class_rme16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 module/class/RME syntax: %v", err)
	}
	if got := countDeclsOf[*ModuleDecl](result.Decls); got != 1 {
		t.Fatalf("ModuleDecl count = %d, want 1", got)
	}
	if got := countDeclsOf[*ObjectDecl](result.Decls); got != 1 {
		t.Fatalf("class ObjectDecl count = %d, want 1", got)
	}
	decl := firstDeclOf[*AssertDecl](t, result.Decls)
	imp, ok := decl.Args()[0].(*Implies)
	if !ok {
		t.Fatalf("AssertDecl arg = %T, want *Implies", decl.Args()[0])
	}
	if _, ok := imp.T2.(*RME); !ok {
		t.Fatalf("AssertDecl RHS = %T, want *RME", imp.T2)
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

func TestParseV16ExportMethodCallAtom(t *testing.T) {
	result, err := Parse("export method", Version{1, 6}, WithFilename("export_method16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 export method: %v", err)
	}
	decl := firstDeclOf[*ExportDecl](t, result.Decls)
	def := firstArgAs[*ExportDef](t, decl)
	if got := NodeRep(def.ExportedNode); got != "method" {
		t.Fatalf("exported callatom = %q, want method", got)
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

func TestParseV16PropertyProofGroupSequence(t *testing.T) {
	result, err := Parse("property true proof { first; second }", Version{1, 6}, WithFilename("proof_group_sequence16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 property proof group sequence: %v", err)
	}
	proof := firstDeclOf[*ProofDecl](t, result.Decls)
	seq := firstArgAs[*ComposeTactics](t, proof)
	if got := len(seq.Tactics); got != 2 {
		t.Fatalf("ComposeTactics length = %d, want 2", got)
	}
	for i, want := range []string{"first", "second"} {
		inst, ok := seq.Tactics[i].(*SchemaInstantiation)
		if !ok {
			t.Fatalf("ComposeTactics[%d] = %T, want *SchemaInstantiation", i, seq.Tactics[i])
		}
		if got := inst.SchemaName.String(); got != want {
			t.Fatalf("ComposeTactics[%d] schema = %q, want %q", i, got, want)
		}
	}
}

func TestParseV16NamedBinderTerms(t *testing.T) {
	src := `type t
individual a:t
relation p(X:t)
conjecture ($snap X. p(X))(a)
conjecture $saved. p(a)
conjecture $saved $ p(a)`
	if _, err := Parse(src, Version{1, 6}, WithFilename("named_binder16.ivy")); err != nil {
		t.Fatalf("Parse v1.6 named binders: %v", err)
	}
}

func TestParseV16FreshSchemaDeclsAndDestructor(t *testing.T) {
	src := `type t
destructor get(X:t):t
schema fresh_bits = {
fresh individual a:t
fresh relation p(X:t)
fresh function f(X:t):t
property true
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("fresh_destructor16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 fresh schema/destructor: %v", err)
	}
	if got := countDeclsOf[*DestructorDecl](result.Decls); got != 1 {
		t.Fatalf("DestructorDecl count = %d, want 1", got)
	}
	schema := firstArgAs[*Schema](t, firstDeclOf[*SchemaDecl](t, result.Decls))
	def := firstArgAs[*Definition](t, schema)
	body, ok := def.Rhs.(*SchemaBody)
	if !ok {
		t.Fatalf("schema rhs = %T, want *SchemaBody", def.Rhs)
	}
	var fresh int
	for _, arg := range body.Args() {
		if _, ok := arg.(*FreshConstantDecl); ok {
			fresh++
		}
	}
	if fresh != 3 {
		t.Fatalf("fresh schema declarations = %d, want 3", fresh)
	}
}

func TestParseV16NativeQuoteDeclarationsAndDefinitions(t *testing.T) {
	src := "type idx = {0..3}\ntype vec\ninterpret vec -> <<< primitive `idx` >>>\ndefinition value(A:idx) = <<< `A` >>>\n<<< native `value` >>>"
	result, err := Parse(src, Version{1, 6}, WithFilename("native16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 nativequote forms: %v", err)
	}
	if got := countDeclsOf[*NativeDecl](result.Decls); got != 1 {
		t.Fatalf("NativeDecl count = %d, want 1", got)
	}
	if nt := findParsedNativeType(t, result); nt == nil {
		t.Fatalf("native type quote not found")
	}
	if ne := findParsedNativeExpr(t, result); ne == nil {
		t.Fatalf("native expr quote not found")
	}
}

func TestParseV16ScenarioUsesDeterministicMixinNames(t *testing.T) {
	src := `action step = {}
scenario { -> s0; s0 -> s1 : before step { call step } s1 : after step { call step } }`
	result, err := Parse(src, Version{1, 6}, WithFilename("scenario16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 scenario: %v", err)
	}
	decl := firstDeclOf[*ScenarioDecl](t, result.Decls)
	sdef := firstArgAs[*ScenarioDef](t, decl)
	transitions := sdef.Transitions()
	if got := len(transitions); got != 2 {
		t.Fatalf("scenario transition count = %d, want 2", got)
	}
	before, ok := transitions[0].Action.(*ScenarioBeforeMixin)
	if !ok {
		t.Fatalf("transition[0] action = %T, want *ScenarioBeforeMixin", transitions[0].Action)
	}
	if got := NodeRep(before.Mixer); got != "step[before]" {
		t.Fatalf("before scenario mixer = %q, want step[before]", got)
	}
	after, ok := transitions[1].Action.(*ScenarioAfterMixin)
	if !ok {
		t.Fatalf("transition[1] action = %T, want *ScenarioAfterMixin", transitions[1].Action)
	}
	if got := NodeRep(after.Mixer); got != "step[after]" {
		t.Fatalf("after scenario mixer = %q, want step[after]", got)
	}
}

func TestParseV16TopLevelInstantiateConceptAndUpdate(t *testing.T) {
	src := `type t
relation r(X:t)
relation s(X:t)
action step = {}
instantiate missing
concept c = { true }, d = r
update r from s params x:t in { call step } -> requires true ensures true`
	result, err := Parse(src, Version{1, 6}, WithFilename("instantiate_concept_update16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 instantiate/concept/update: %v", err)
	}
	instantiate := firstDeclOf[*InstantiateDecl](t, result.Decls)
	inst := firstArgAs[*Instantiation](t, instantiate)
	if got := NodeRep(inst.Sort); got != "missing" {
		t.Fatalf("instantiate target = %q, want missing", got)
	}
	concept := firstDeclOf[*ConceptDecl](t, result.Decls)
	if got := len(concept.Args()); got != 2 {
		t.Fatalf("ConceptDecl args = %d, want 2", got)
	}
	update := firstDeclOf[*UpdateDecl](t, result.Decls)
	if _, ok := update.Args()[0].(*PatternBasedUpdate); !ok {
		t.Fatalf("UpdateDecl arg = %T, want *PatternBasedUpdate", update.Args()[0])
	}
}

func TestParseV16GhostTypesAndModuleVariants(t *testing.T) {
	src := `ghost type spec_t
ghost type spec_idx = {0..1}
module object mo = {
type inner
}
module isolate mi with api = {
action a = {}
}`
	result, err := Parse(src, Version{1, 6}, WithFilename("ghost_module16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 ghost/module variants: %v", err)
	}
	var ghosts int
	for _, decl := range result.Decls {
		td, ok := decl.(*TypeDecl)
		if !ok || len(td.Args()) == 0 {
			continue
		}
		if _, ok := td.Args()[0].(*GhostTypeDef); ok {
			ghosts++
		}
	}
	if ghosts != 2 {
		t.Fatalf("GhostTypeDef count = %d, want 2", ghosts)
	}
	if got := countDeclsOf[*ModuleDecl](result.Decls); got != 2 {
		t.Fatalf("ModuleDecl count = %d, want 2", got)
	}
	var sawInnerIsolate bool
	for _, decl := range result.Decls {
		md, ok := decl.(*ModuleDecl)
		if !ok || len(md.Args()) == 0 {
			continue
		}
		def, ok := md.Args()[0].(*Definition)
		if !ok || NodeRep(def.Lhs) != "mi" {
			continue
		}
		body, ok := def.Rhs.(*ivyAccum)
		if !ok {
			t.Fatalf("module isolate body = %T, want *ivyAccum", def.Rhs)
		}
		if countDeclsOf[*IsolateDecl](body.decls) == 1 {
			sawInnerIsolate = true
		}
	}
	if !sawInnerIsolate {
		t.Fatalf("module isolate did not synthesize inner isolate declaration")
	}
}

func TestParseV16TopLevelProofAndImplementType(t *testing.T) {
	src := `proof [pf] intro
implement type spec_t with impl_t`
	result, err := Parse(src, Version{1, 6}, WithFilename("proof_implement_type16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 proof/implement type: %v", err)
	}
	proof := firstDeclOf[*ProofDecl](t, result.Decls)
	lf := firstArgAs[*LabeledFormula](t, proof)
	if lf.LabelName() != "pf" {
		t.Fatalf("proof label = %q, want pf", lf.LabelName())
	}
	if _, ok := lf.Formula.(*SchemaInstantiation); !ok {
		t.Fatalf("proof formula = %T, want *SchemaInstantiation", lf.Formula)
	}
	if got := countDeclsOf[*ImplementTypeDecl](result.Decls); got != 1 {
		t.Fatalf("ImplementTypeDecl count = %d, want 1", got)
	}
}

func TestParseV16IsolateAndExtractObjectForms(t *testing.T) {
	src := `isolate objiso = {
action a = {}
} with exported
extract worker = {
action b = {}
} with exported
extract simple = a, b`
	result, err := Parse(src, Version{1, 6}, WithFilename("isolate_extract16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 isolate/extract object forms: %v", err)
	}
	if got := countDeclsOf[*ObjectDecl](result.Decls); got != 2 {
		t.Fatalf("ObjectDecl count = %d, want 2", got)
	}
	if got := countDeclsOf[*IsolateObjectDecl](result.Decls); got != 2 {
		t.Fatalf("IsolateObjectDecl count = %d, want 2", got)
	}
	var sawExtract bool
	for _, decl := range result.Decls {
		id, ok := decl.(*IsolateDecl)
		if !ok || len(id.Args()) == 0 {
			continue
		}
		if ed, ok := id.Args()[0].(*ExtractDef); ok && ed.Kind == "extract" {
			sawExtract = true
		}
	}
	if !sawExtract {
		t.Fatalf("plain extract declaration not found")
	}
}

func TestParseV16DefinitionSomeExprAndOperatorLHS(t *testing.T) {
	src := `type t
relation p(X:t)
definition choice(X:t):t = some Y:t . p(Y) in Y else X,
           (X + Y):t = X`
	result, err := Parse(src, Version{1, 6}, WithFilename("definition_shapes16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 definition shapes: %v", err)
	}
	decl := firstDeclOf[*DefinitionDecl](t, result.Decls)
	if got := len(decl.Args()); got != 2 {
		t.Fatalf("DefinitionDecl args = %d, want 2", got)
	}
	first := decl.Args()[0].(*LabeledFormula).Formula.(*Definition)
	if _, ok := first.Rhs.(*SomeExpr); !ok {
		t.Fatalf("choice rhs = %T, want *SomeExpr", first.Rhs)
	}
	second := decl.Args()[1].(*LabeledFormula).Formula.(*Definition)
	if got := NodeRep(second.Lhs); got != "+" {
		t.Fatalf("operator definition lhs = %q, want +", got)
	}
}

func TestParseV16ActionNativeCrashAndOperatorTterm(t *testing.T) {
	src := `type t
var (X + Y):t
action native = { <<< impure
do_native(` + "`X`" + `)
>>> }
action crash = *`
	result, err := Parse(src, Version{1, 6}, WithFilename("native_crash_tterm16.ivy"))
	if err != nil {
		t.Fatalf("Parse v1.6 native/crash/operator tterm: %v", err)
	}
	constant := firstDeclOf[*ConstantDecl](t, result.Decls)
	if got := NodeRep(firstArg(t, constant)); got != "+" {
		t.Fatalf("operator tterm rep = %q, want +", got)
	}
	native := firstActionDefNamed(t, result.Decls, "native")
	if _, ok := native.Body.(*NativeAction); !ok {
		t.Fatalf("native action body = %T, want *NativeAction", native.Body)
	}
	crash := firstActionDefNamed(t, result.Decls, "crash")
	if _, ok := crash.Body.(*CrashAction); !ok {
		t.Fatalf("crash action body = %T, want *CrashAction", crash.Body)
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

func firstArg(t *testing.T, node Node) Node {
	t.Helper()
	args := node.Args()
	if len(args) == 0 {
		t.Fatalf("%T has no args", node)
	}
	return args[0]
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

func nodeSortString(node Node) string {
	switch n := node.(type) {
	case *Atom:
		if n.ASort == nil {
			return ""
		}
		return n.ASort.String()
	case *App:
		if n.ASort == nil {
			return ""
		}
		return n.ASort.String()
	default:
		return ""
	}
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
