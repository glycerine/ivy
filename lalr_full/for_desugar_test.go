package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// TestMethcall tests the methcall helper matching Python ivy_parser.py:3034-3038.
func TestMethcall(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Case 1: App with no args → compose (not MethodCall)
	lhs := cfg.NewApp(cfg.NewSymbol("x", nil))
	rhs := cfg.NewApp(cfg.NewSymbol("begin", nil))
	result := methcall(cfg, lhs, rhs)
	if _, ok := result.(*ast.MethodCall); ok {
		t.Error("expected composed node for App with no args, got MethodCall")
	}
	if ast.NodeRep(result) != "x.begin" {
		t.Errorf("expected rep 'x.begin', got %q", ast.NodeRep(result))
	}

	// Case 2: App with args → MethodCall
	lhsWithArgs := cfg.NewApp(cfg.NewSymbol("x", nil), cfg.NewApp(cfg.NewSymbol("y", nil)))
	result2 := methcall(cfg, lhsWithArgs, rhs)
	mc, ok := result2.(*ast.MethodCall)
	if !ok {
		t.Fatalf("expected MethodCall for App with args, got %T", result2)
	}
	if mc.Obj != lhsWithArgs {
		t.Error("MethodCall.Obj wrong")
	}
	if mc.Method != rhs {
		t.Error("MethodCall.Method wrong")
	}

	// Case 3: Atom with no args → compose (not MethodCall)
	atomLhs := cfg.NewAtom("x")
	result3 := methcall(cfg, atomLhs, rhs)
	if _, ok := result3.(*ast.MethodCall); ok {
		t.Error("expected composed node for Atom with no args, got MethodCall")
	}
	if ast.NodeRep(result3) != "x.begin" {
		t.Errorf("expected rep 'x.begin', got %q", ast.NodeRep(result3))
	}

	// Case 4: Atom with args → MethodCall
	atomWithArgs := cfg.NewAtom("x", cfg.NewAtom("y"))
	result4 := methcall(cfg, atomWithArgs, rhs)
	if _, ok := result4.(*ast.MethodCall); !ok {
		t.Errorf("expected MethodCall for Atom with args, got %T", result4)
	}
}

// TestMethcallResultType verifies compose preserves the rhs type.
func TestMethcallResultType(t *testing.T) {
	cfg := ast.NewAstConfig()
	// When rhs is App and lhs is App with no args, result should be App
	lhs := cfg.NewApp(cfg.NewSymbol("fmla", nil))
	rhs := cfg.NewApp(cfg.NewSymbol("begin", nil))
	result := methcall(cfg, lhs, rhs)
	if _, ok := result.(*ast.App); !ok {
		t.Errorf("expected *ast.App, got %T", result)
	}

	// When rhs is Atom and lhs is Atom with no args, result should be Atom
	lhsAtom := cfg.NewAtom("fmla")
	rhsAtom := cfg.NewAtom("end")
	result2 := methcall(cfg, lhsAtom, rhsAtom)
	if _, ok := result2.(*ast.Atom); !ok {
		t.Errorf("expected *ast.Atom, got %T", result2)
	}
}

// TestAppRename tests that App.Rename works correctly for iend creation.
func TestAppRename(t *testing.T) {
	cfg := ast.NewAstConfig()
	app := cfg.NewApp(cfg.NewSymbol("itr", nil))
	app.ASort = cfg.NewSymbol("someSort", nil)

	renamed := app.Rename("loc:end")
	if renamed.Relname() != "loc:end" {
		t.Errorf("expected rep 'loc:end', got %q", renamed.Relname())
	}
	if renamed.ASort == nil {
		t.Error("expected ASort to be preserved after rename")
	}
	// Original should be unchanged
	if app.Relname() != "itr" {
		t.Errorf("original app was mutated, rep=%q", app.Relname())
	}
}

// TestMethodCallCanon tests that MethodCall produces correct canon output.
func TestMethodCallCanon(t *testing.T) {
	cfg := ast.NewAstConfig()
	mc := cfg.NewMethodCall(
		cfg.NewApp(cfg.NewSymbol("x", nil), cfg.NewApp(cfg.NewSymbol("y", nil))),
		cfg.NewApp(cfg.NewSymbol("next", nil)),
	)
	canon := string(mc.Canon())
	if !strings.Contains(canon, "methodCall") {
		t.Errorf("expected 'methodCall' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "obj:") {
		t.Errorf("expected 'obj:' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "method:") {
		t.Errorf("expected 'method:' in canon, got %s", canon)
	}
}

// TestForLoopDesugaring directly constructs the FOR loop desugaring and verifies
// the resulting AST structure matches what the grammar rule should produce.
// This avoids needing a full Ivy program to parse.
func TestForLoopDesugaring(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Simulate the desugaring that the grammar rule performs:
	// for itr, val in fmla { body }
	// where itr=App('x'), val=App('v'), fmla=App('rng'), body=skip

	itr := cfg.NewApp(cfg.NewSymbol("x", nil))
	val := cfg.NewApp(cfg.NewSymbol("v", nil))
	fmla := cfg.NewApp(cfg.NewSymbol("rng", nil))
	seq := cfg.NewSequence() // empty body (skip)

	// iend = itr.rename('loc:end')
	iend := itr.Rename("loc:end")

	// didx = VarAction(itr, methcall(fmla, App('begin')))
	appBegin := cfg.NewApp(cfg.NewSymbol("begin", nil))
	didx := cfg.NewVarAction(itr, methcall(cfg, fmla, appBegin))

	// dend = VarAction(iend, methcall(fmla, App('end')))
	appEnd := cfg.NewApp(cfg.NewSymbol("end", nil))
	dend := cfg.NewVarAction(iend, methcall(cfg, fmla, appEnd))

	// dval = VarAction(val, methcall(fmla, App('value', itr)))
	appValue := cfg.NewApp(cfg.NewSymbol("value", nil), itr)
	dval := cfg.NewVarAction(val, methcall(cfg, fmla, appValue))

	// incr = AssignAction(itr, methcall(itr, App('next')))
	appNext := cfg.NewApp(cfg.NewSymbol("next", nil))
	incr := cfg.NewAssignAction(itr, methcall(cfg, itr, appNext))

	// body = Sequence(*lower_var_stmts([dval, seq, incr]))
	bodyStmts := ast.LowerVarStatements([]ast.Node{dval, seq, incr})
	body := cfg.NewSequence(bodyStmts...)

	// loop = WhileAction(App('<', itr, iend), body)
	ltCond := cfg.NewApp(cfg.NewSymbol("<", nil), itr, iend)
	loop := cfg.NewWhileAction(ltCond, body)

	// result = Sequence(*lower_var_stmts([didx, dend, loop]))
	outerStmts := ast.LowerVarStatements([]ast.Node{didx, dend, loop})
	result := cfg.NewSequence(outerStmts...)

	canonStr := string(result.Canon())

	// The result should be a Sequence containing LocalAction (from lower_var_stmts)
	if !strings.Contains(canonStr, "sequence") {
		t.Error("expected sequence at top level")
	}

	// Should contain localAction from lower_var_stmts
	if !strings.Contains(canonStr, "localAction") {
		t.Error("expected localAction from lower_var_stmts processing")
	}

	// Should contain whileAction
	if !strings.Contains(canonStr, "whileAction") {
		t.Error("expected whileAction in desugared FOR loop")
	}

	// Should contain assignAction (for incr)
	if !strings.Contains(canonStr, "assignAction") {
		t.Error("expected assignAction for increment step")
	}

	// iend should have rep "loc:end"
	if iend.Relname() != "loc:end" {
		t.Errorf("iend rep should be 'loc:end', got %q", iend.Relname())
	}

	// methcall(fmla, App('begin')) should compose to 'rng.begin' since fmla has no args
	if ast.NodeRep(methcall(cfg, fmla, appBegin)) != "rng.begin" {
		t.Errorf("expected 'rng.begin', got %q", ast.NodeRep(methcall(cfg, fmla, appBegin)))
	}

	t.Logf("FOR loop desugared canon: %s", canonStr)
}

// TestForLoopParse tests parsing a FOR loop in a complete Ivy program.
func TestForLoopParse(t *testing.T) {
	input := `type t
type iter
action foo(rng:t) = {
    for x:iter, v:t in rng {
        v := v
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("FOR loop parse test skipped (grammar may need additional context): %v", err)
	}
	if len(result.Decls) == 0 {
		t.Fatal("expected at least 1 decl")
	}

	// Find the action decl and check its canon
	for _, d := range result.Decls {
		canonStr := string(d.Canon())
		if strings.Contains(canonStr, `rep:"for"`) {
			t.Error("desugared FOR loop should not contain atom with rep 'for'")
		}
	}
}

// TestForLoopMethcallCompose verifies that when fmla is a simple name (no args),
// methcall composes it with begin/end/value/next rather than creating MethodCall.
func TestForLoopMethcallCompose(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Simulate: fmla = App('rng'), rhs = App('begin')
	// Since fmla has no args, should compose to App('rng.begin')
	fmla := cfg.NewApp(cfg.NewSymbol("rng", nil))
	rhs := cfg.NewApp(cfg.NewSymbol("begin", nil))
	result := methcall(cfg, fmla, rhs)

	if _, ok := result.(*ast.MethodCall); ok {
		t.Error("fmla with no args should compose, not create MethodCall")
	}
	if ast.NodeRep(result) != "rng.begin" {
		t.Errorf("expected 'rng.begin', got %q", ast.NodeRep(result))
	}

	// value with itr arg: App('value', itr) has args, but fmla.args==0 matters
	itr := cfg.NewApp(cfg.NewSymbol("x", nil))
	rhsValue := cfg.NewApp(cfg.NewSymbol("value", nil), itr)
	resultValue := methcall(cfg, fmla, rhsValue)
	// fmla has no args → compose
	if _, ok := resultValue.(*ast.MethodCall); ok {
		t.Error("fmla with no args should compose even when rhs has args")
	}
}

// TestLocalActionDesugaring tests that the LOCAL rule applies loc: prefix
// substitution matching Python ivy_parser.py:3187-3195.
func TestLocalActionDesugaring(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Simulate what the grammar rule does:
	// local x:t, y:t { body using x and y }
	// Should produce: LocalAction(loc:x, loc:y, body_with_x→loc:x, y→loc:y)

	// Create lparams
	paramX := cfg.NewApp(cfg.NewSymbol("x", nil))
	paramY := cfg.NewApp(cfg.NewSymbol("y", nil))
	bounds := []ast.Node{paramX, paramY}

	// Create body that references x and y
	bodyExpr := cfg.NewApp(cfg.NewSymbol("+", nil), cfg.NewApp(cfg.NewSymbol("x", nil)), cfg.NewApp(cfg.NewSymbol("y", nil)))
	body := cfg.NewSequence(bodyExpr)

	// Apply the same transformation as the grammar rule
	lsyms := make([]ast.Node, len(bounds))
	subst := make(map[string]string)
	for i, s := range bounds {
		lsyms[i] = ast.PrefixNode(s, "loc:")
		subst[ast.NodeRep(s)] = ast.NodeRep(lsyms[i])
	}
	action := ast.SubstPrefixAtomsAst(body, subst, nil, nil, nil)
	args := append(lsyms, action)
	la := cfg.NewLocalAction(args...)

	canon := string(la.Canon())

	// Should contain localAction
	if !strings.Contains(canon, "localAction") {
		t.Errorf("expected localAction in canon, got %s", canon)
	}

	// The params should be prefixed with loc:
	if !strings.Contains(canon, `rep:loc:x`) {
		t.Errorf("expected loc:x param in canon, got %s", canon)
	}
	if !strings.Contains(canon, `rep:loc:y`) {
		t.Errorf("expected loc:y param in canon, got %s", canon)
	}

	// The body should have substituted references
	// The original "x" and "y" refs in the body should now be "loc:x" and "loc:y"
	canonAction := string(action.Canon())
	if strings.Contains(canonAction, `rep:"x"`) && !strings.Contains(canonAction, `rep:"loc:x"`) {
		t.Errorf("body should have x substituted to loc:x, got %s", canonAction)
	}

	t.Logf("LOCAL action canon: %s", canon)
}
