package goivy

import (
	"testing"
)

func newTestCompiler() *Compiler {
	sig := NewSig()
	mod := New()
	mod.Sig = sig
	return NewCompiler(sig, mod)
}

// TestCompileTrueFalse checks that "true" and "false" atoms compile correctly.
func TestCompileTrueFalse(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	trueAtom := cfg.NewAtom("true")
	result, err := c.CompileNode(trueAtom)
	if err != nil {
		t.Fatalf("compile true: %v", err)
	}
	if !IvyIsTrue(result) {
		t.Errorf("expected true (empty And), got %T: %s", result, result)
	}

	falseAtom := cfg.NewAtom("false")
	result, err = c.CompileNode(falseAtom)
	if err != nil {
		t.Fatalf("compile false: %v", err)
	}
	if !IvyIsFalse(result) {
		t.Errorf("expected false (empty Or), got %T: %s", result, result)
	}
}

// TestCompileAnd checks that And nodes compile correctly.
func TestCompileAnd(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	a := cfg.NewAtom("true")
	b := cfg.NewAtom("false")
	andNode := cfg.NewAnd(a, b)

	result, err := c.CompileNode(andNode)
	if err != nil {
		t.Fatalf("compile and: %v", err)
	}
	and, ok := result.(*And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T", result)
	}
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(and.Terms))
	}
}

// TestCompileOr checks that Or nodes compile correctly.
func TestCompileOr(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	a := cfg.NewAtom("true")
	b := cfg.NewAtom("false")
	orNode := cfg.NewOr(a, b)

	result, err := c.CompileNode(orNode)
	if err != nil {
		t.Fatalf("compile or: %v", err)
	}
	or, ok := result.(*Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
	if len(or.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(or.Terms))
	}
}

// TestCompileNot checks that Not nodes compile correctly.
func TestCompileNot(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	a := cfg.NewAtom("true")
	notNode := cfg.NewNot(a)

	result, err := c.CompileNode(notNode)
	if err != nil {
		t.Fatalf("compile not: %v", err)
	}
	not, ok := result.(*Not)
	if !ok {
		t.Fatalf("expected *lg.Not, got %T", result)
	}
	if !IvyIsTrue(not.Body) {
		t.Errorf("expected true body, got %s", not.Body)
	}
}

// TestCompileImplies checks that Implies nodes compile correctly.
func TestCompileImplies(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	a := cfg.NewAtom("true")
	b := cfg.NewAtom("false")
	implNode := cfg.NewImplies(a, b)

	result, err := c.CompileNode(implNode)
	if err != nil {
		t.Fatalf("compile implies: %v", err)
	}
	imp, ok := result.(*Implies)
	if !ok {
		t.Fatalf("expected *lg.Implies, got %T", result)
	}
	if !IvyIsTrue(imp.T1) {
		t.Errorf("expected true lhs")
	}
	if !IvyIsFalse(imp.T2) {
		t.Errorf("expected false rhs")
	}
}

// TestCompileIff checks Iff compilation.
func TestCompileIff(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	a := cfg.NewAtom("true")
	b := cfg.NewAtom("true")
	iffNode := cfg.NewIff(a, b)

	result, err := c.CompileNode(iffNode)
	if err != nil {
		t.Fatalf("compile iff: %v", err)
	}
	if _, ok := result.(*Iff); !ok {
		t.Fatalf("expected *lg.Iff, got %T", result)
	}
}

// TestCompileVariable checks variable compilation.
func TestCompileVariable(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	// Add a sort to the signature
	c.Sig.Sorts.Set("node", &UninterpretedSort{Name: "node"})

	v := cfg.NewVariable("X", "node")
	result, err := c.CompileNode(v)
	if err != nil {
		t.Fatalf("compile variable: %v", err)
	}
	lv, ok := result.(*Variable)
	if !ok {
		t.Fatalf("expected *lg.Variable, got %T", result)
	}
	if lv.Name != "X" {
		t.Errorf("expected name X, got %s", lv.Name)
	}
	if lv.VSort.String() != "node" {
		t.Errorf("expected sort node, got %s", lv.VSort)
	}
}

// TestCompileEquality checks equality compilation.
func TestCompileEquality(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	c.Sig.Sorts.Set("nat", &UninterpretedSort{Name: "nat"})

	x := cfg.NewVariable("X", "nat")
	y := cfg.NewVariable("Y", "nat")
	eq := cfg.NewAtom("=", x, y)

	result, err := c.CompileNode(eq)
	if err != nil {
		t.Fatalf("compile equality: %v", err)
	}
	eqNode, ok := result.(*Eq)
	if !ok {
		t.Fatalf("expected *lg.Eq, got %T", result)
	}
	// Use .Name (not .String()) to check identity — String() now returns
	// sort-qualified output via PrettyFmla.
	t1Var, ok1 := eqNode.T1.(*Variable)
	t2Var, ok2 := eqNode.T2.(*Variable)
	if !ok1 || !ok2 || t1Var.Name != "X" || t2Var.Name != "Y" {
		t.Errorf("expected X = Y, got %s = %s", eqNode.T1, eqNode.T2)
	}
}

// TestCompileQuantifierForall checks universal quantifier compilation.
func TestCompileQuantifierForall(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	c.Sig.Sorts.Set("node", &UninterpretedSort{Name: "node"})

	x := cfg.NewVariable("X", "node")
	body := cfg.NewAtom("true")
	forall := cfg.NewForall([]Node{x}, body)

	result, err := c.CompileNode(forall)
	if err != nil {
		t.Fatalf("compile forall: %v", err)
	}
	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("expected *lg.ForAll, got %T", result)
	}
	if len(fa.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(fa.Variables))
	}
	if fa.Variables[0].Name != "X" {
		t.Errorf("expected variable X, got %s", fa.Variables[0].Name)
	}
}

// TestCompileQuantifierExists checks existential quantifier compilation.
func TestCompileQuantifierExists(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	c.Sig.Sorts.Set("node", &UninterpretedSort{Name: "node"})

	x := cfg.NewVariable("X", "node")
	body := cfg.NewAtom("true")
	exists := cfg.NewExists([]Node{x}, body)

	result, err := c.CompileNode(exists)
	if err != nil {
		t.Fatalf("compile exists: %v", err)
	}
	ex, ok := result.(*Exists)
	if !ok {
		t.Fatalf("expected *lg.Exists, got %T", result)
	}
	if len(ex.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(ex.Variables))
	}
}

// TestCompileSymbolLookup checks that declared symbols are found.
func TestCompileSymbolLookup(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Declare a 0-arity symbol
	c.Sig.Sorts.Set("nat", &UninterpretedSort{Name: "nat"})
	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.AddSymbol("zero", natSort)

	atom := cfg.NewAtom("zero")
	result, err := c.CompileNode(atom)
	if err != nil {
		t.Fatalf("compile symbol: %v", err)
	}
	cnst, ok := result.(*Const)
	if !ok {
		t.Fatalf("expected *lg.Const, got %T", result)
	}
	if cnst.Name != "zero" {
		t.Errorf("expected name zero, got %s", cnst.Name)
	}
}

// TestCompileApply checks that function application works.
func TestCompileApply(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	funcSort, _ := NewFunctionSort(natSort, natSort)
	c.Sig.AddSymbol("succ", funcSort)

	arg := cfg.NewVariable("X", "nat")
	atom := cfg.NewAtom("succ", arg)
	result, err := c.CompileNode(atom)
	if err != nil {
		t.Fatalf("compile apply: %v", err)
	}
	app, ok := result.(*Apply)
	if !ok {
		t.Fatalf("expected *lg.Apply, got %T", result)
	}
	if len(app.Terms) != 1 {
		t.Errorf("expected 1 arg, got %d", len(app.Terms))
	}
}

// TestCompileGlobally checks temporal Globally compilation.
func TestCompileGlobally(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	body := cfg.NewAtom("true")
	g := cfg.NewGlobally(body)

	result, err := c.CompileNode(g)
	if err != nil {
		t.Fatalf("compile globally: %v", err)
	}
	glob, ok := result.(*Globally)
	if !ok {
		t.Fatalf("expected *lg.Globally, got %T", result)
	}
	if !IvyIsTrue(glob.Body) {
		t.Errorf("expected true body")
	}
}

// TestCompileEventually checks temporal Eventually compilation.
func TestCompileEventually(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	body := cfg.NewAtom("true")
	e := cfg.NewEventually(body)

	result, err := c.CompileNode(e)
	if err != nil {
		t.Fatalf("compile eventually: %v", err)
	}
	ev, ok := result.(*Eventually)
	if !ok {
		t.Fatalf("expected *lg.Eventually, got %T", result)
	}
	if !IvyIsTrue(ev.Body) {
		t.Errorf("expected true body")
	}
}

// TestResolveAlias checks alias resolution.
func TestResolveAlias(t *testing.T) {
	mod := New()
	mod.Aliases["foo"] = "bar"
	mod.Aliases["baz"] = "qux"

	if got := ResolveAlias("foo", mod); got != "bar" {
		t.Errorf("expected bar, got %s", got)
	}
	if got := ResolveAlias("unknown", mod); got != "unknown" {
		t.Errorf("expected unknown, got %s", got)
	}
	// Dotted alias resolution
	mod.Aliases["a"] = "b"
	if got := ResolveAlias("a.c", mod); got != "b.c" {
		t.Errorf("expected b.c, got %s", got)
	}
}

// TestDeclInterpType checks type declaration processing.
func TestDeclInterpType(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// Uninterpreted type
	typeDef := cfg.NewTypeDef(cfg.NewSymbol("mytype", nil), cfg.NewConstantSort())
	err := d.TypeDecl(typeDef)
	if err != nil {
		t.Fatalf("type decl: %v", err)
	}
	if _, ok := c.Sig.Sorts.Get2("mytype"); !ok {
		t.Error("mytype not in sig sorts")
	}
}

// TestDeclInterpEnumType checks enumerated type declaration processing.
func TestDeclInterpEnumType(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	elems := []Node{cfg.NewSymbol("red", nil), cfg.NewSymbol("green", nil), cfg.NewSymbol("blue", nil)}
	enumSort := cfg.NewEnumeratedSort(elems...)
	typeDef := cfg.NewTypeDef(cfg.NewSymbol("color", nil), enumSort)
	err := d.TypeDecl(typeDef)
	if err != nil {
		t.Fatalf("enum type decl: %v", err)
	}

	sort, ok := c.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("color not in sig sorts")
	}
	es, ok := sort.(*EnumeratedSort)
	if !ok {
		t.Fatalf("expected EnumeratedSort, got %T", sort)
	}
	if len(es.Extension) != 3 {
		t.Errorf("expected 3 elements, got %d", len(es.Extension))
	}

	// Check that enum values were added as symbols
	for _, name := range []string{"red", "green", "blue"} {
		if _, ok := c.Sig.Symbols.Get2(name); !ok {
			t.Errorf("%s not in sig symbols", name)
		}
	}
}

// TestDeclInterpRelation checks relation declaration processing.
func TestDeclInterpRelation(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// First add the sort
	c.Sig.Sorts.Set("node", &UninterpretedSort{Name: "node"})

	// Relation: link(X:node, Y:node)
	x := cfg.NewVariable("X", "node")
	y := cfg.NewVariable("Y", "node")
	rel := cfg.NewAtom("link", x, y)

	err := d.Relation(rel)
	if err != nil {
		t.Fatalf("relation decl: %v", err)
	}

	if _, ok := c.Sig.Symbols.Get2("link"); !ok {
		t.Error("link not in sig symbols")
	}
}

// TestDeclInterpAlias checks alias declaration processing.
func TestDeclInterpAlias(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	def := cfg.NewDefinition(cfg.NewSymbol("src", nil), cfg.NewSymbol("dst", nil))
	err := d.Alias(def)
	if err != nil {
		t.Fatalf("alias decl: %v", err)
	}

	if c.Module.Aliases["src"] != "dst" {
		t.Errorf("expected alias src->dst, got %v", c.Module.Aliases)
	}
}

// TestCompileConst checks constant declaration compilation.
func TestCompileConst(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	c.Sig.Sorts.Set("nat", &UninterpretedSort{Name: "nat"})

	atom := cfg.NewAtom("zero")
	atom.ASort = cfg.NewSymbol("nat", nil)

	sym, err := c.CompileConst(atom, c.Sig)
	if err != nil {
		t.Fatalf("compile const: %v", err)
	}
	if sym.Name != "zero" {
		t.Errorf("expected name zero, got %s", sym.Name)
	}
	if sym.CSort.String() != "nat" {
		t.Errorf("expected sort nat, got %s", sym.CSort)
	}
}

// TestCompilePolymorphicSymbol checks that polymorphic symbols are found.
func TestCompilePolymorphicSymbol(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	x := cfg.NewVariable("X", "nat")
	y := cfg.NewVariable("Y", "nat")
	lt := cfg.NewAtom("<", x, y)

	result, err := c.CompileNode(lt)
	if err != nil {
		t.Fatalf("compile <: %v", err)
	}
	app, ok := result.(*Apply)
	if !ok {
		t.Fatalf("expected *lg.Apply, got %T: %s", result, result)
	}
	if len(app.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(app.Terms))
	}
}

// TestCompileNestedFormula checks compilation of a nested formula.
func TestCompileNestedFormula(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	c.Sig.Sorts.Set("node", &UninterpretedSort{Name: "node"})

	// forall X:node. X = X -> true
	x := cfg.NewVariable("X", "node")
	eq := cfg.NewAtom("=", x, x)
	impl := cfg.NewImplies(eq, cfg.NewAtom("true"))
	forall := cfg.NewForall([]Node{x}, impl)

	result, err := c.CompileNode(forall)
	if err != nil {
		t.Fatalf("compile nested: %v", err)
	}

	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("expected *lg.ForAll, got %T", result)
	}
	imp, ok := fa.Body.(*Implies)
	if !ok {
		t.Fatalf("expected *lg.Implies body, got %T", fa.Body)
	}
	if _, ok := imp.T1.(*Eq); !ok {
		t.Errorf("expected *lg.Eq in antecedent, got %T", imp.T1)
	}
}

// TestCompileOld checks the Old operator.
func TestCompileOld(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("count", natSort)

	oldNode := cfg.NewOld(cfg.NewAtom("count"))
	result, err := c.CompileNode(oldNode)
	if err != nil {
		t.Fatalf("compile old: %v", err)
	}
	cnst, ok := result.(*Const)
	if !ok {
		t.Fatalf("expected *lg.Const, got %T", result)
	}
	if cnst.Name != "old_count" {
		t.Errorf("expected old_count, got %s", cnst.Name)
	}
}

// TestCompileIte checks if-then-else expression compilation.
func TestCompileIte(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	cond := cfg.NewAtom("true")
	then := cfg.NewAtom("true")
	els := cfg.NewAtom("false")
	ite := cfg.NewIte(cond, then, els)

	result, err := c.CompileNode(ite)
	if err != nil {
		t.Fatalf("compile ite: %v", err)
	}
	iteNode, ok := result.(*Ite)
	if !ok {
		t.Fatalf("expected *lg.Ite, got %T", result)
	}
	if !IvyIsTrue(iteNode.Cond) {
		t.Errorf("expected true condition")
	}
}

// TestNewCompiler checks that New creates a valid Compiler.
func TestNewCompiler(t *testing.T) {
	c := NewCompiler(nil, nil)
	if c.Sig == nil {
		t.Error("Sig should not be nil")
	}
	if c.Module == nil {
		t.Error("Module should not be nil")
	}
	if c.VarCtx == nil || c.VarCtx.Map == nil {
		t.Error("VarCtx should be initialized")
	}
}

// TestCompileNilNode checks that nil nodes produce an error.
func TestCompileNilNode(t *testing.T) {
	c := newTestCompiler()
	_, err := c.CompileNode(nil)
	if err == nil {
		t.Error("expected error for nil node")
	}
}

// TestCompileEmptyAnd checks that empty And = true.
func TestCompileEmptyAnd(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	andNode := cfg.NewAnd()
	result, err := c.CompileNode(andNode)
	if err != nil {
		t.Fatalf("compile empty and: %v", err)
	}
	if !IvyIsTrue(result) {
		t.Errorf("expected true, got %s", result)
	}
}

// TestCompileEmptyOr checks that empty Or = false.
func TestCompileEmptyOr(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	orNode := cfg.NewOr()
	result, err := c.CompileNode(orNode)
	if err != nil {
		t.Fatalf("compile empty or: %v", err)
	}
	if !IvyIsFalse(result) {
		t.Errorf("expected false, got %s", result)
	}
}

// TestProcessDecls checks that ProcessDecls iterates through declarations.
func TestProcessDecls(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	decls := []Node{
		cfg.NewTypeDecl(cfg.NewTypeDef(cfg.NewSymbol("t1", nil), cfg.NewConstantSort())),
		cfg.NewTypeDecl(cfg.NewTypeDef(cfg.NewSymbol("t2", nil), cfg.NewConstantSort())),
	}

	err := d.ProcessDecls(decls)
	if err != nil {
		t.Fatalf("process decls: %v", err)
	}

	if _, ok := c.Sig.Sorts.Get2("t1"); !ok {
		t.Error("t1 not in sorts")
	}
	if _, ok := c.Sig.Sorts.Get2("t2"); !ok {
		t.Error("t2 not in sorts")
	}
}

// FuzzCompileAtom fuzzes compilation of Atom nodes with various names.
func FuzzCompileAtom(f *testing.F) {
	cfg := NewAstConfig()
	f.Add("true")
	f.Add("false")
	f.Add("=")
	f.Add("foo")
	f.Add("<")
	f.Add("a.b.c")
	f.Add("")

	f.Fuzz(func(t *testing.T, name string) {
		c := newTestCompiler()
		atom := cfg.NewAtom(name)
		// We don't check the result, just that it doesn't panic.
		_, _ = c.CompileNode(atom)
	})
}

// FuzzCompilerPipeline fuzzes the full parser→compiler pipeline.
// The first byte selects the Ivy language version (1.0–1.7),
// the rest is the Ivy source. The fuzzer explores all versions
// naturally by mutating the version byte.
//
// This must never panic. Errors are expected and fine.
func FuzzCompilerPipeline(f *testing.F) {
	// Seed: version byte + Ivy source.
	// Each seed exercises features specific to that version.

	// --- Version 1.0: minimal (types, relations, axioms) ---
	f.Add(byte(0), []byte(`type t
relation r(X:t, Y:t)
axiom r(X, X)
`))

	// --- Version 1.1: adds state, local ---
	f.Add(byte(1), []byte(`type node
relation link(X:node, Y:node)
individual x : node
axiom forall X:node. link(X, X)
`))

	// --- Version 1.2: adds mixin, export, import ---
	f.Add(byte(2), []byte(`type t
relation r(X:t)
action a(x:t) = {
    r(x) := true
}
export a
`))

	// --- Version 1.3: similar to 1.2 ---
	f.Add(byte(3), []byte(`type s
type t
relation edge(X:s, Y:t)
action add(x:s, y:t) = {
    edge(x,y) := true
}
action remove(x:s, y:t) = {
    edge(x,y) := false
}
export add
export remove
`))

	// --- Version 1.4: adds struct, object, method, property, while ---
	f.Add(byte(4), []byte(`type item
relation present(X:item)
action insert(x:item) = {
    present(x) := true
}
action delete(x:item) = {
    present(x) := false
}
property forall X:item. present(X) -> present(X)
export insert
export delete
`))

	// --- Version 1.5: adds variant, globally, eventually ---
	f.Add(byte(5), []byte(`type node
relation link(X:node, Y:node)
relation pending(X:node)
individual root : node
action connect(x:node, y:node) = {
    require ~link(x, y);
    link(x, y) := true
}
conjecture forall X:node, Y:node. link(X, Y) -> link(Y, X)
export connect
`))

	// --- Version 1.6: adds specification, implementation, require, ensure ---
	f.Add(byte(6), []byte(`type money
individual balance : money
action deposit(x:money) = {
    balance := x
}
action withdraw(x:money) = {
    balance := x
}
export deposit
export withdraw
`))

	// --- Version 1.7: adds process, debug, common ---
	f.Add(byte(7), []byte(`type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init {
    semaphore(W) := true;
    link(X,Y) := false
}
action connect(x:client,y:server) = {
    require semaphore(y);
    link(x,y) := true;
    semaphore(y) := false
}
action disconnect(x:client,y:server) = {
    require link(x,y);
    link(x,y) := false;
    semaphore(y) := true
}
invariant ~(X ~= Z & link(X,Y) & link(Z,Y))
export connect
export disconnect
`))

	// --- More seeds: edge cases ---
	f.Add(byte(7), []byte(`type t`))
	f.Add(byte(7), []byte(``))
	f.Add(byte(7), []byte(`# just a comment`))
	f.Add(byte(4), []byte(`type t
object foo = {
    individual x : t
    action bar = {
        x := x
    }
}
`))
	f.Add(byte(6), []byte(`type t
axiom forall X:t. X = X
conjecture exists X:t. X = X
`))
	f.Add(byte(7), []byte(`type t
relation r(X:t)
action a(x:t) returns (y:t) = {
    r(x) := true;
    y := x
}
export a
`))
	f.Add(byte(5), []byte(`type t
relation r(X:t, Y:t)
relation s(X:t)
axiom forall X:t, Y:t. r(X,Y) -> s(X)
axiom forall X:t. s(X) -> exists Y:t. r(X,Y)
`))
	f.Add(byte(7), []byte(`type t
action a(x:t) = {
    if x = x {
        assume true
    } else {
        assert false
    }
}
export a
`))

	f.Fuzz(func(t *testing.T, vByte byte, src []byte) {
		version := Version{1, int(vByte % 8)}

		// Parse
		result, err := Parse(string(src), version)
		if err != nil || result == nil {
			return
		}

		// Compile — must not panic regardless of input
		sig := NewSig()
		mod := New()
		mod.Sig = sig
		cmplr := NewCompiler(sig, mod)
		di := NewDomainSetup(cmplr)
		for _, d := range result.Decls {
			_ = di.ProcessDecl(d)
		}
	})
}
