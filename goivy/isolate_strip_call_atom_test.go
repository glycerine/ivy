package goivy

import "testing"

// Python strip_action (ivy_isolate.py) on a CallAction:
//
//	args = [strip_action(arg,...) for arg in ast.args[0].args]
//	strip_params = get_strip_params(name,ast.args[0].args,...)
//	call = ast.args[0].clone(args[len(strip_params):])
//
// The call's own atom (args[0], Go's AstCallee) loses the leading isolate
// parameters, so every later pass sees imap.set(0,0). Before the fix,
// goivy stripped only the compiled Callee and kept the old AstCallee, so
// isolate iso_dm(me:id) of sht.ivy kept prm:X in its after-init call.
func TestStripActionFullStripsCallAtomParams(t *testing.T) {
	m := mkModuleWithSig()
	id, key := isolateMkSort("id"), isolateMkSort("key.t")
	setSort, err := NewFunctionSort(id, key, id, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	const name = "proto.impl.dm.impl.imap.set"
	m.Sig.Symbols.Set(name, &SymbolEntry{Name: name, Sort: setSort})

	prmX := NewConst("prm:X", id)
	k0, v0 := NewConst("0", key), NewConst("0", id)
	acfg := NewAstConfig()
	call := NewCallActionOn(NewActionsConfig(), MustApply(NewConst(name, setSort), prmX, k0, v0))
	call.AstCallee = acfg.NewAtom(name, prmX, k0, v0)

	stripMap := StripMap{"proto.impl.dm": {"me"}}
	binding := map[NodeKey]string{Key(prmX): "me"}
	res := StripActionFull(call, stripMap, m, binding, true, nil)

	stripped, ok := res.(*LogicCallAction)
	if !ok {
		t.Fatalf("result %T", res)
	}
	atom, ok := stripped.Args()[0].(*Atom)
	if !ok {
		t.Fatalf("args[0] %T", stripped.Args()[0])
	}
	if atom.Rep != name || len(atom.Terms) != 2 || Key(atom.Terms[0].(Expr)) != Key(k0) || Key(atom.Terms[1].(Expr)) != Key(v0) {
		t.Fatalf("call atom = %v (%d args), want %s(0,0)", atom, len(atom.Terms), name)
	}
	if app, ok := stripped.Callee.(*Apply); !ok || len(app.Terms) != 2 {
		t.Fatalf("compiled callee = %v, want 2 args", stripped.Callee)
	}
}
